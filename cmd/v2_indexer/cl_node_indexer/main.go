package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"eth2-exporter/exporter"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gtuk/discordwebhook"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

const MAX_CL_BLOCK_NUMBER = 1_000_000_000_000 - 1

func main() {

	url := flag.String("url", "http://localhost:4000", "")

	// discordWebhookReportUrl := flag.String("discord-url", "", "")
	// discordWebhookUser := flag.String("discord-user", "", "")
	slotNumber := flag.Int("slot-number", -1, "")
	startSlotNumber := flag.Int("start-slot-number", 0, "")
	concurrency := flag.Int("concurrency", 1, "")

	flag.Parse()

	btClient, err := gcp_bigtable.NewClient(context.Background(), "etherchain", "beaconchain-node-data-storage", option.WithGRPCConnectionPool(50))
	if err != nil {
		logrus.Fatal(err)
	}

	tableBlocksRaw := btClient.Open("blocks-raw")

	spec := &exporter.BeaconSpecResponse{}

	err = utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/config/spec", *url), nil, spec)
	if err != nil {
		logrus.Fatal(err)
	}

	chainIdUint64, err := strconv.ParseUint(spec.Data["DEPOSIT_CHAIN_ID"], 10, 64)
	if err != nil {
		logrus.Fatal(err)
	}

	slotsPerEpoch, err := strconv.Atoi(spec.Data["SLOTS_PER_EPOCH"])
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(chainIdUint64)
	logrus.Info(slotsPerEpoch)

	headHeader := &exporter.BeaconBlockHeaderResponse{}
	err = utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/headers/head", *url), nil, headHeader)
	if err != nil {
		logrus.Fatal(err)
	}

	//checkRead(tableBlocksRaw, chainIdUint64)

	httpClient := &http.Client{
		Timeout: time.Second * 60,
	}

	if *slotNumber != -1 {
		logrus.Infof("checking block %v", *slotNumber)
		_, err := getAttestationRewards(*url, httpClient, *slotNumber/32)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Info("OK")
		return
	}

	gOuter := &errgroup.Group{}
	gOuter.SetLimit(*concurrency)

	muts := []*gcp_bigtable.Mutation{}
	keys := []string{}
	mux := &sync.Mutex{}

	latestBlockNumber := headHeader.Data.Header.Message.Slot
	blocksProcessedTotal := atomic.Int64{}
	blocksProcessedIntv := atomic.Int64{}
	exportStart := time.Now()

	t := time.NewTicker(time.Second * 10)

	go func() {
		for {
			<-t.C

			remainingBlocks := int64(latestBlockNumber) - int64(*startSlotNumber) - blocksProcessedTotal.Load()
			blocksPerSecond := float64(blocksProcessedIntv.Load()) / time.Since(exportStart).Seconds()
			secondsRemaining := float64(remainingBlocks) / float64(blocksPerSecond)

			durationRemaining := time.Second * time.Duration(secondsRemaining)
			logrus.Infof("current speed: %0.1f blocks/sec, %d blocks processed, %d blocks remaining (%0.1f days to go)", blocksPerSecond, blocksProcessedIntv.Load(), remainingBlocks, durationRemaining.Hours()/24)
			blocksProcessedIntv.Store(0)
			exportStart = time.Now()
		}
	}()

	//p := message.NewPrinter(language.English)

	for i := *startSlotNumber; i <= int(latestBlockNumber); i++ {

		i := i

		gOuter.Go(func() error {
			for ; ; time.Sleep(time.Second) {

				logrus.Infof("retrieving data for slot %d", i)

				var slot, proposalRewards, syncRewards []byte

				var proposerAssignments, syncAssignments, attestationAssignments, attestationRewards, validators []byte

				var err error
				slot, err = getSlot(*url, httpClient, i)
				if err != nil {
					logrus.Error(err)
					continue
				}
				proposalRewards, err = getPropoalRewards(*url, httpClient, i)
				if err != nil {
					logrus.Error(err)
					continue
				}
				syncRewards, err = getSyncRewards(*url, httpClient, i)
				if err != nil {
					logrus.Error(err)
					continue
				}

				if i%slotsPerEpoch == 0 {
					logrus.Infof("requesting data for epoch %d", i/slotsPerEpoch)
					proposerAssignments, err = getPropoalAssignments(*url, httpClient, i/slotsPerEpoch)
					if err != nil {
						logrus.Error(err)
						continue
					}
					syncAssignments, err = getSyncCommittees(*url, httpClient, i)
					if err != nil {
						logrus.Error(err)
						continue
					}

					attestationAssignments, err = getCommittees(*url, httpClient, i)
					if err != nil {
						logrus.Error(err)
						continue
					}

					attestationRewards, err = getAttestationRewards(*url, httpClient, i/slotsPerEpoch)
					if err != nil {
						logrus.Error(err)
						continue
					}

					validators, err = getValidators(*url, httpClient, i)
					if err != nil {
						logrus.Error(err)
						continue
					}
				}

				mux.Lock()
				mut := gcp_bigtable.NewMutation()
				mut.Set("s", "s", gcp_bigtable.Timestamp(0), slot)
				mut.Set("r", "p", gcp_bigtable.Timestamp(0), proposalRewards)
				mut.Set("r", "s", gcp_bigtable.Timestamp(0), syncRewards)

				if i%slotsPerEpoch == 0 {
					mut.Set("r", "a", gcp_bigtable.Timestamp(0), attestationRewards)
					mut.Set("a", "p", gcp_bigtable.Timestamp(0), proposerAssignments)
					mut.Set("a", "a", gcp_bigtable.Timestamp(0), attestationAssignments)
					mut.Set("a", "s", gcp_bigtable.Timestamp(0), syncAssignments)
					mut.Set("v", "v", gcp_bigtable.Timestamp(0), validators)
				}

				muts = append(muts, mut)
				key := getBlockKey(i, chainIdUint64)
				keys = append(keys, key)

				if len(keys) == 32 {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
					errs, err := tableBlocksRaw.ApplyBulk(ctx, keys, muts)

					if err != nil {
						logrus.Fatalf("error writing data to bigtable: %v", err)
						cancel()
						continue
					}

					for _, err := range errs {
						logrus.Fatalf("error writing data to bigtable: %v", err)
						cancel()
						continue
					}
					cancel()

					muts = []*gcp_bigtable.Mutation{}
					keys = []string{}

				}
				mux.Unlock()

				if i%slotsPerEpoch == 0 {
					logrus.Infof("completed processing block %v (block: %v b, proposer rewards: %v b, sync rewards: %v b, attestation rewards: %v b, proposer assignments: %v b, attestation assignments: %v b, sync assignments: %v b, validators: %v b, total: %v b)",
						i,
						len(slot),
						len(proposalRewards),
						len(syncRewards),
						len(attestationRewards),
						len(proposerAssignments),
						len(attestationAssignments),
						len(syncAssignments),
						len(validators),
						len(slot)+len(proposalRewards)+len(syncRewards)+len(attestationRewards)+len(proposerAssignments)+len(attestationAssignments)+len(syncAssignments)+len(validators))
				} else {
					logrus.Infof("completed processing block %v (block: %v b, proposer rewards: %v b, sync rewards: %v b, total: %v b)",
						i,
						len(slot),
						len(proposalRewards),
						len(syncRewards),
						len(slot)+len(proposalRewards)+len(syncRewards)+len(attestationRewards)+len(proposerAssignments)+len(attestationAssignments)+len(syncAssignments)+len(validators))
				}

				if blocksProcessedTotal.Add(1)%100000 == 0 {
					//sendMessage(p.Sprintf("OP MAINNET NODE EXPORT: currently at block %v of %v (%.1f%%)", i, latestBlockNumber, float64(i)*100/float64(latestBlockNumber)), *discordWebhookReportUrl, *discordWebhookUser)
				}

				blocksProcessedIntv.Add(1)

				break

			}
			return nil
		})
	}

	gOuter.Wait()
}

func checkRead(tbl *gcp_bigtable.Table, chainId uint64) {
	ctx := context.Background()

	filter := gcp_bigtable.PrefixRange(fmt.Sprintf("%d:", chainId))

	err := tbl.ReadRows(ctx, filter, func(r gcp_bigtable.Row) bool {

		blockNumberString := strings.Replace(r.Key(), fmt.Sprintf("%d:", chainId), "", 1)
		blockNumberUint64, err := strconv.ParseUint(blockNumberString, 10, 64)
		if err != nil {
			logrus.Fatal(err)
		}
		blockNumberUint64 = MAX_CL_BLOCK_NUMBER - blockNumberUint64
		logrus.Infof("retrieved block %d", blockNumberUint64)
		blockCell := r["b"][0]

		blockDataCompressed := blockCell.Value
		blockDataDecompressed := decompress(blockDataCompressed)

		logrus.Info(string(blockDataDecompressed))

		return true
	})

	if err != nil {
		logrus.Fatal(err)
	}
}

func sendMessage(content, webhookUrl, username string) {

	message := discordwebhook.Message{
		Username: &username,
		Content:  &content,
	}

	err := discordwebhook.SendMessage(webhookUrl, message)
	if err != nil {
		log.Fatal(err)
	}
}

func getSlot(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v2/beacon/blocks/%d", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getValidators(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getCommittees(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%d/committees", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getSyncCommittees(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%d/sync_committees", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getPropoalAssignments(url string, httpClient *http.Client, epoch int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", url, epoch)
	return genericRequest("GET", requestUrl, httpClient)
}

func getPropoalRewards(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/blocks/%d", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getSyncRewards(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/sync_committee/%d", url, slot)
	return genericRequest("POST", requestUrl, httpClient)
}

func getAttestationRewards(url string, httpClient *http.Client, epoch int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/attestations/%d", url, epoch)
	return genericRequest("POST", requestUrl, httpClient)
}

func genericRequest(method string, requestUrl string, httpClient *http.Client) ([]byte, error) {
	data := []byte{}
	if method == "POST" {
		data = []byte("[]")
	}
	r, err := http.NewRequest(method, requestUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}

	if res.StatusCode != http.StatusOK {

		if res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		if res.StatusCode == http.StatusBadRequest {
			return nil, nil
		}
		if res.StatusCode == http.StatusInternalServerError {
			return nil, nil
		}
		return nil, fmt.Errorf("error unexpected status code: %v", res.StatusCode)
	}

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	// logrus.Info(string(resString))

	if strings.Contains(string(resString), `"code"`) {
		return nil, fmt.Errorf("rpc error: %s", resString)
	}

	return compress(resString), nil
}

func compress(src []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(src)
	if err != nil {
		logrus.Fatalf("error writing to gzip writer: %v", err)
	}
	if err := zw.Close(); err != nil {
		logrus.Fatalf("error closing gzip writer: %v", err)
	}
	return buf.Bytes()
}

func decompress(src []byte) []byte {
	zr, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		logrus.Fatalf("error creating gzip reader: %v", err)
	}

	data, err := io.ReadAll(zr)
	if err != nil {
		logrus.Fatalf("error reading from gzip reader: %v", err)
	}
	return data
}

func getBlockKey(blockNumber int, chainId uint64) string {
	return fmt.Sprintf("CL:%d:%12d", chainId, MAX_CL_BLOCK_NUMBER-blockNumber)
}
