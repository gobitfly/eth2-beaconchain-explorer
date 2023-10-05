package exporter

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/metrics"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/coocood/freecache"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type BlobIndexer struct {
	S3Client   *s3.Client
	running    bool
	runningMu  *sync.Mutex
	clEndpoint string
	cache      *freecache.Cache
}

func NewBlobIndexer() (*BlobIndexer, error) {
	s3Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               utils.Config.BlobIndexer.S3.Endpoint,
			SigningRegion:     "us-east-2",
			HostnameImmutable: true,
		}, nil
	})
	s3Client := s3.NewFromConfig(aws.Config{
		Region: "us-east-2",
		Credentials: credentials.NewStaticCredentialsProvider(
			utils.Config.BlobIndexer.S3.AccessKeyId,
			utils.Config.BlobIndexer.S3.AccessKeySecret,
			"",
		),
		EndpointResolverWithOptions: s3Resolver,
	}, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	bi := &BlobIndexer{
		S3Client:   s3Client,
		runningMu:  &sync.Mutex{},
		clEndpoint: "http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port,
		cache:      freecache.NewCache(1024 * 1024),
	}
	return bi, nil
}

func (bi *BlobIndexer) Start() {
	bi.runningMu.Lock()
	if bi.running {
		bi.runningMu.Unlock()
		return
	}
	bi.running = true
	bi.runningMu.Unlock()

	logrus.WithFields(logrus.Fields{"version": version.Version, "clEndpoint": bi.clEndpoint, "s3Endpoint": utils.Config.BlobIndexer.S3.Endpoint}).Infof("starting blobindexer")
	for {
		err := bi.Index()
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Errorf("failed indexing blobs")
		}
		time.Sleep(time.Second * 10)
	}
}

func (bi *BlobIndexer) Index() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	headHeader := &BeaconBlockHeaderResponse{}
	finalizedHeader := &BeaconBlockHeaderResponse{}
	spec := &BeaconSpecResponse{}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(3)
	g.Go(func() error {
		err := utils.HttpReq(gCtx, http.MethodGet, fmt.Sprintf("%s/eth/v1/config/spec", bi.clEndpoint), nil, spec)
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		err := utils.HttpReq(gCtx, http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/headers/head", bi.clEndpoint), nil, headHeader)
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		err := utils.HttpReq(gCtx, http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/headers/finalized", bi.clEndpoint), nil, finalizedHeader)
		if err != nil {
			return err
		}
		return nil
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	nodeDepositNetworkId, ok := spec.Data["DEPOSIT_NETWORK_ID"]
	if !ok {
		return fmt.Errorf("missing DEPOSIT_NETWORK_ID in spec from node")
	}
	if fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositNetworkID) != nodeDepositNetworkId {
		return fmt.Errorf("config.DepositNetworkId != node.DepositNetworkId: %v != %v", utils.Config.Chain.ClConfig.DepositNetworkID, nodeDepositNetworkId)
	}

	status, err := bi.GetIndexerStatus()
	if err != nil {
		return err
	}

	denebForkSlot := utils.Config.Chain.ClConfig.DenebForkEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
	startSlot := status.LastIndexedFinalizedSlot + 1
	if status.LastIndexedFinalizedSlot <= denebForkSlot {
		startSlot = denebForkSlot
	}

	if headHeader.Data.Header.Message.Slot <= startSlot {
		return fmt.Errorf("headHeader.Data.Header.Message.Slot <= startSlot: %v < %v", headHeader.Data.Header.Message.Slot, startSlot)
	}

	start := time.Now()
	logrus.WithFields(logrus.Fields{"lastIndexedFinalizedSlot": status.LastIndexedFinalizedSlot, "headSlot": headHeader.Data.Header.Message.Slot}).Infof("indexing blobs")
	defer func() {
		logrus.WithFields(logrus.Fields{
			"startSlot": startSlot,
			"endSlot":   headHeader.Data.Header.Message.Slot,
			"duration":  time.Since(start),
		}).Infof("finished indexing blobs")
	}()

	batchSize := uint64(100)
	for batchStart := startSlot; batchStart <= headHeader.Data.Header.Message.Slot; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > headHeader.Data.Header.Message.Slot {
			batchEnd = headHeader.Data.Header.Message.Slot
		}
		g, gCtx = errgroup.WithContext(context.Background())
		g.SetLimit(4)
		for slot := batchStart; slot <= batchEnd; slot++ {
			slot := slot
			g.Go(func() error {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
				}
				err := bi.IndexBlobsAtSlot(slot)
				if err != nil {
					return err
				}
				return nil
			})
		}
		err = g.Wait()
		if err != nil {
			return err
		}
		if batchEnd <= finalizedHeader.Data.Header.Message.Slot {
			err := bi.PutIndexerStatus(BlobIndexerStatus{
				LastIndexedFinalizedSlot: batchEnd,
			})
			if err != nil {
				return fmt.Errorf("error updating indexer status at slot %v: %w", batchEnd, err)
			}
			logrus.WithFields(logrus.Fields{"lastIndexedFinalizedSlot": batchEnd}).Infof("updated indexer status")
		}
	}
	return nil
}

func (bi *BlobIndexer) IndexBlobsAtSlot(slot uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	blobSidecar := &BeaconBlobSidecarsResponse{}
	tGetBlobSidcar := time.Now()
	err := utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/blob_sidecars/%d", bi.clEndpoint, slot), nil, blobSidecar)
	if err != nil {
		var httpErr *utils.HttpReqHttpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			// no sidecar for this slot
			return nil
		}
		return err
	}
	metrics.TaskDuration.WithLabelValues("blobindexer_get_blob_sidecars").Observe(time.Since(tGetBlobSidcar).Seconds())

	if len(blobSidecar.Data) <= 0 {
		return nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(4)
	for i, d := range blobSidecar.Data {
		i := i
		d := d
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			kzgCommitment, err := hex.DecodeString(strings.Replace(d.KzgCommitment, "0x", "", -1))
			if err != nil {
				return fmt.Errorf("error decoding kzgCommitment at index %v: %s: %w", i, d.KzgCommitment, err)
			}

			blob, err := hex.DecodeString(strings.Replace(d.Blob, "0x", "", -1))
			if err != nil {
				return fmt.Errorf("error decoding blob at index %v: %w", i, err)
			}

			versionedBlobHash := fmt.Sprintf("%#x", utils.VersionedBlobHash(kzgCommitment).Bytes())
			key := fmt.Sprintf("blobs/%s", versionedBlobHash)

			tS3HeadObj := time.Now()
			_, err = bi.S3Client.HeadObject(gCtx, &s3.HeadObjectInput{
				Bucket: &utils.Config.BlobIndexer.S3.Bucket,
				Key:    &key,
			})
			metrics.TaskDuration.WithLabelValues("blobindexer_check_blob").Observe(time.Since(tS3HeadObj).Seconds())
			if err != nil {
				// Only put the object if it does not exist yet
				var httpResponseErr *awshttp.ResponseError
				if errors.As(err, &httpResponseErr) && (httpResponseErr.HTTPStatusCode() == 404 || httpResponseErr.HTTPStatusCode() == 403) {
					//logrus.WithFields(logrus.Fields{"slot": d.Slot, "index": d.Index, "key": key}).Infof("putting blob")
					tS3PutObj := time.Now()
					_, putErr := bi.S3Client.PutObject(gCtx, &s3.PutObjectInput{
						Bucket: &utils.Config.BlobIndexer.S3.Bucket,
						Key:    &key,
						Body:   bytes.NewReader(blob),
						Metadata: map[string]string{
							"slot":              fmt.Sprintf("%d", d.Slot),
							"index":             fmt.Sprintf("%d", d.Index),
							"block_root":        d.BlockRoot,
							"block_parent_root": d.BlockParentRoot,
							"proposer_index":    fmt.Sprintf("%d", d.ProposerIndex),
							"kzg_commitment":    d.KzgCommitment,
							"kzg_proof":         d.KzgProof,
						},
					})
					metrics.TaskDuration.WithLabelValues("blobindexer_put_blob").Observe(time.Since(tS3PutObj).Seconds())
					if putErr != nil {
						return fmt.Errorf("error putting object: %s (%v/%v): %w", key, d.Slot, d.Index, putErr)
					}
					return nil
				}
				return fmt.Errorf("error getting headObject: %s (%v/%v): %w", key, d.Slot, d.Index, err)
			}
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return fmt.Errorf("error indexing blobs at slot %v: %w", slot, err)
	}

	return nil
}

func (bi *BlobIndexer) GetIndexerStatus() (*BlobIndexerStatus, error) {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("blobindexer_get_indexer_status").Observe(time.Since(start).Seconds())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	key := "blob-indexer-status.json"
	obj, err := bi.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &utils.Config.BlobIndexer.S3.Bucket,
		Key:    &key,
	})
	if err != nil {
		// If the object that you request doesn’t exist, the error that Amazon S3 returns depends on whether you also have the s3:ListBucket permission. If you have the s3:ListBucket permission on the bucket, Amazon S3 returns an HTTP status code 404 (Not Found) error. If you don’t have the s3:ListBucket permission, Amazon S3 returns an HTTP status code 403 ("access denied") error.
		var httpResponseErr *awshttp.ResponseError
		if errors.As(err, &httpResponseErr) && (httpResponseErr.HTTPStatusCode() == 404 || httpResponseErr.HTTPStatusCode() == 403) {
			return &BlobIndexerStatus{}, nil
		}
		return nil, err
	}
	status := &BlobIndexerStatus{}
	err = json.NewDecoder(obj.Body).Decode(status)
	return status, err
}

func (bi *BlobIndexer) PutIndexerStatus(status BlobIndexerStatus) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("blobindexer_put_indexer_status").Observe(time.Since(start).Seconds())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	key := "blob-indexer-status.json"
	contentType := "application/json"
	body, err := json.Marshal(&status)
	if err != nil {
		return err
	}
	_, err = bi.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &utils.Config.BlobIndexer.S3.Bucket,
		Key:         &key,
		Body:        bytes.NewReader(body),
		ContentType: &contentType,
		Metadata: map[string]string{
			"last_indexed_finalized_slot": fmt.Sprintf("%d", status.LastIndexedFinalizedSlot),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

type BeaconSpecResponse struct {
	Data map[string]string `json:"data"`
}

type BlobIndexerStatus struct {
	LastIndexedFinalizedSlot uint64 `json:"last_indexed_finalized_slot"`
	// LastIndexedFinalizedRoot string `json:"last_indexed_finalized_root"`
	// IndexedUnfinalized       map[string]uint64 `json:"indexed_unfinalized"`
}

type BeaconForkScheduleResponse struct {
	Data []struct {
		PreviousVersion string `json:"previous_version"`
		CurrentVersion  string `json:"current_version"`
		Epoch           uint64 `json:"epoch,string"`
	} `json:"data"`
}

type BeaconBlobSidecarsResponse struct {
	Data []BeaconBlobSidecarsResponseBlob `json:"data"`
}
type BeaconBlobSidecarsResponseBlob struct {
	BlockRoot       string `json:"block_root"`
	Index           uint64 `json:"index,string"`
	Slot            uint64 `json:"slot,string"`
	BlockParentRoot string `json:"block_parent_root"`
	ProposerIndex   uint64 `json:"proposer_index,string"`
	Blob            string `json:"blob"`
	KzgCommitment   string `json:"kzg_commitment"`
	KzgProof        string `json:"kzg_proof"`
}

type BeaconFinalityCheckpointsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		PreviousJustified struct {
			Epoch uint64 `json:"epoch,string"`
			Root  string `json:"root"`
		} `json:"previous_justified"`
		CurrentJustified struct {
			Epoch uint64 `json:"epoch,string"`
			Root  string `json:"root"`
		} `json:"current_justified"`
		Finalized struct {
			Epoch uint64 `json:"epoch,string"`
			Root  string `json:"root"`
		} `json:"finalized"`
	} `json:"data"`
}

type BeaconBlockHeaderResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		Root      string `json:"root"`
		Canonical bool   `json:"canonical"`
		Header    struct {
			Message struct {
				Slot          uint64 `json:"slot,string"`
				ProposerIndex string `json:"proposer_index"`
				ParentRoot    string `json:"parent_root"`
				StateRoot     string `json:"state_root"`
				BodyRoot      string `json:"body_root"`
			} `json:"message"`
			Signature string `json:"signature"`
		} `json:"header"`
	} `json:"data"`
}
