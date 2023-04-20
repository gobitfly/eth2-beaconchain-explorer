package main

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"

	_ "net/http/pprof"
)

/**
* This function is for indexing smart contract function names from www.4byte.directory
* so that we can label the transction function calls instead of the "id"
**/
func main() {
	bigtableProject := flag.String("bigtable.project", "", "Bigtable project")
	bigtableInstance := flag.String("bigtable.instance", "", "Bigtable instance")
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	metricsAddr := flag.String("metrics.address", "localhost:9090", "serve metrics on that addr")
	metricsEnabled := flag.Bool("metrics.enabled", false, "enable serving metrics")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

	db.MustInitDB(&types.DatabaseConfig{
		Username: cfg.WriterDatabase.Username,
		Password: cfg.WriterDatabase.Password,
		Name:     cfg.WriterDatabase.Name,
		Host:     cfg.WriterDatabase.Host,
		Port:     cfg.WriterDatabase.Port,
	}, &types.DatabaseConfig{
		Username: cfg.ReaderDatabase.Username,
		Password: cfg.ReaderDatabase.Password,
		Name:     cfg.ReaderDatabase.Name,
		Host:     cfg.ReaderDatabase.Host,
		Port:     cfg.ReaderDatabase.Port,
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

	bt, err := db.InitBigtable(*bigtableProject, *bigtableInstance, "1")
	if err != nil {
		logrus.Errorf("error initializing bigtable: %v", err)
		return
	}

	go ImportMethodSignatures(bt)

	utils.WaitForCtrlC()
}

func ImportMethodSignatures(bt *db.Bigtable) {

	// Per default we start with the latest signatures (first page = latest signatures)
	const firstPage = "https://www.4byte.directory/api/v1/signatures/"
	page := firstPage
	status, err := bt.GetMethodSignatureImportStatus()
	if err != nil {
		logrus.Errorf("error getting signature import status from bigtable: %v", err)
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
	sleepTime := time.Second * 2

	for ; ; time.Sleep(sleepTime) { // timout needed due to rate limit
		sleepTime = time.Second * 2
		logrus.Infof("Get signatures for: %v", page)
		start := time.Now()
		next, sigs, err := GetNextSignatures(bt, page, *status)

		if err != nil {
			metrics.Errors.WithLabelValues("method_signatures_get_signatures_failed").Inc()
			logrus.Errorf("error getting signature: %v", err)
			sleepTime = time.Minute
			continue
		}

		// If had a complete sync done in the past, we only need to get signatures newer then the onces from our prev. run
		if status.LatestTimestamp != nil && status.HasFinished {
			createdAt, _ := time.Parse(time.RFC3339, *status.LatestTimestamp)
			if createdAt.UnixNano() <= latestTimestamp.UnixMilli() {
				logrus.Info("Our Signature Data is up to date")
				sleepTime = time.Hour
				isFirst = true
				page = firstPage
				continue
			}
		}

		err = db.BigtableClient.SaveMethodSignatures(sigs)
		if err != nil {
			metrics.Errors.WithLabelValues("method_signatures_save_to_bt_failed").Inc()
			logrus.Errorf("error saving signatures into bigtable: %v", err)
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
			logrus.Infof("Save Sig ts: %v next: %v", *status.LatestTimestamp, *status.NextPage)
			err = bt.SaveMethodSignatureImportStatus(*status)
			if err != nil {
				metrics.Errors.WithLabelValues("method_signatures_save_status_to_bt_failed").Inc()
				logrus.Errorf("error saving signature status into bigtable: %v", err)
				sleepTime = time.Minute
			}
		}
		metrics.TaskDuration.WithLabelValues("method_signatures_page_imported").Observe(time.Since(start).Seconds())
		services.ReportStatus("signatures", "Running", nil)
	}
}

func GetNextSignatures(bt *db.Bigtable, page string, status types.MethodSignatureImportStatus) (*string, []types.MethodSignature, error) {

	httpClient := &http.Client{Timeout: time.Second * 10}

	resp, err := httpClient.Get(page)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("error querying signatures api: %v", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	type signatureResponse struct {
		Results []types.MethodSignature `json:"results"`
		Next    *string                 `json:"next"`
	}

	respParsed := &signatureResponse{}
	err = json.Unmarshal(body, respParsed)
	if err != nil {
		return nil, nil, err
	}

	return respParsed.Next, respParsed.Results, nil
}
