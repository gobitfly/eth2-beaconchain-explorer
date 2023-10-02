package main

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"

	_ "net/http/pprof"
)

/**
* This function is for indexing smart contract function names from www.4byte.directory
* so that we can label the transction function calls instead of the "id"
**/
func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	metricsAddr := flag.String("metrics.address", "localhost:9090", "serve metrics on that addr")
	metricsEnabled := flag.Bool("metrics.enabled", false, "enable serving metrics")

	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

	db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
	})
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	if *metricsEnabled {
		go func() {
			logrus.WithFields(logrus.Fields{"addr": *metricsAddr}).Infof("Serving metrics")
			if err := metrics.Serve(*metricsAddr); err != nil {
				logrus.WithError(err).Fatal("Error serving metrics")
			}
		}()
	}

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, "1", utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Errorf("error initializing bigtable: %v", err)
		return
	}

	go ImportSignatures(bt, types.MethodSignature)
	time.Sleep(time.Second * 2) // we need a little delay, as the api does not like two requests at the same time
	go ImportSignatures(bt, types.EventSignature)

	utils.WaitForCtrlC()
}

func ImportSignatures(bt *db.Bigtable, st types.SignatureType) {

	// Per default we start with the latest signatures (first page = latest signatures)
	firstPage := "https://www.4byte.directory/api/v1/signatures/"
	if st == types.EventSignature {
		firstPage = "https://www.4byte.directory/api/v1/event-signatures/"
	}
	page := firstPage
	status, err := bt.GetSignatureImportStatus(st)
	if err != nil {
		logrus.Errorf("error getting %v signature import status from bigtable: %v", st, err)
		return
	}
	isFirst := true

	// If we never completed syncing all signatures we continue with the next page
	if !status.HasFinished && status.NextPage != nil {
		page = *status.NextPage
		isFirst = false
	}

	var latestTimestamp time.Time
	// Timestamp of the first item from the last run
	if status.LatestTimestamp != nil {
		latestTimestamp, _ = time.Parse(time.RFC3339, *status.LatestTimestamp)
	}
	sleepTime := time.Second * 4

	for ; ; time.Sleep(sleepTime) { // timout needed due to rate limit
		sleepTime = time.Second * 4
		logrus.Infof("Get signatures for: %v", page)
		start := time.Now()
		next, sigs, err := GetNextSignatures(bt, page, *status)

		if err != nil {
			metrics.Errors.WithLabelValues(fmt.Sprintf("%v_signatures_get_signatures_failed", st)).Inc()
			logrus.Errorf("error getting %v signature: %v", st, err)
			sleepTime = time.Minute
			continue
		}

		// If had a complete sync done in the past, we only need to get signatures newer then the onces from our prev. run
		if status.LatestTimestamp != nil && status.HasFinished {
			createdAt, _ := time.Parse(time.RFC3339, *status.LatestTimestamp)
			if createdAt.UnixMilli() <= latestTimestamp.UnixMilli() {
				isFirst = true
				if page != firstPage {
					logrus.Infof("Our %v signature data of page %v is up to date so we jump to the first page", st, page)
					page = firstPage
				} else {
					logrus.Infof("Our %v signature data is up to date so we wait for an hour to check again", st)
					sleepTime = time.Hour
				}
				continue
			}
		}

		err = db.BigtableClient.SaveSignatures(sigs, st)
		if err != nil {
			metrics.Errors.WithLabelValues(fmt.Sprintf("%v_signatures_save_to_bt_failed", st)).Inc()
			logrus.Errorf("error saving %v signatures into bigtable: %v", st, err)
			sleepTime = time.Minute
			continue
		}

		// Lets save the timestamp from the first (=latest) entry
		if isFirst {
			status.LatestTimestamp = &sigs[0].CreatedAt
			isFirst = false
		}

		if next == nil {
			status.NextPage = nil
			status.HasFinished = true
		} else {
			if !status.HasFinished {
				status.NextPage = next
			}
			page = *next
		}
		if status != nil && (status.HasFinished || status.NextPage != nil) {
			nextPage := "-"
			latestTimestamp := "-"
			if status.NextPage != nil {
				nextPage = *status.NextPage
			}
			if status.LatestTimestamp != nil {
				latestTimestamp = *status.LatestTimestamp
			}
			logrus.Infof("Save %v Sig ts: %v next: %v", st, latestTimestamp, nextPage)
			err = bt.SaveSignatureImportStatus(*status, st)
			if err != nil {
				metrics.Errors.WithLabelValues(fmt.Sprintf("%v_signatures_save_status_to_bt_failed", st)).Inc()
				logrus.Errorf("error saving %v signature status into bigtable: %v", st, err)
				sleepTime = time.Minute
			}
		}
		metrics.TaskDuration.WithLabelValues(fmt.Sprintf("%v_signatures_page_imported", st)).Observe(time.Since(start).Seconds())
		services.ReportStatus(fmt.Sprintf("%v_signatures", st), "Running", nil)
	}
}

func GetNextSignatures(bt *db.Bigtable, page string, status types.SignatureImportStatus) (*string, []types.Signature, error) {

	httpClient := &http.Client{Timeout: time.Second * 10}

	resp, err := httpClient.Get(page)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("error querying signatures api: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	type signatureResponse struct {
		Results []types.Signature `json:"results"`
		Next    *string           `json:"next"`
	}

	respParsed := &signatureResponse{}
	err = json.Unmarshal(body, respParsed)
	if err != nil {
		return nil, nil, err
	}

	return respParsed.Next, respParsed.Results, nil
}
