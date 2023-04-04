package main

import (
	"encoding/json"
	"eth2-exporter/db"
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

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

	bt, err := db.InitBigtable(*bigtableProject, *bigtableInstance, "1")
	if err != nil {
		logrus.Errorf("error initializing bigtable: %v", err)
		return
	}

	// Per default we start with the latest signatures (first page = latest signatures)
	page := "https://www.4byte.directory/api/v1/signatures/"
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

	for ; ; time.Sleep(time.Second * 2) { // timout needed due to rate limit
		logrus.Infof("Get signatures for: %v", page)
		next, sigs, err := GetNextSignitures(bt, page, *status)

		if err != nil {
			logrus.Errorf("error getting signature: %v", err)
			break
		}

		// If had a complete sync done in the past, we only need to get signatures newer then the onces from our prev. run
		if status.LatestTimestamp != nil && status.HasFinished {
			createdAt, _ := time.Parse(time.RFC3339, *status.LatestTimestamp)
			if createdAt.UnixNano() <= latestTimestamp.UnixMilli() {
				logrus.Info("Our Signature Data is up to date")
				break
			}
		}

		err = db.BigtableClient.SaveMethodSignatures(sigs)
		if err != nil {
			logrus.Errorf("error saving signatures into bigtable: %v", err)
			break
		}

		// Lets save the timestamp from the first (=latest) entry
		if isFirst {
			status.LatestTimestamp = &sigs[0].CreatedAt
			isFirst = false
		}

		if next == nil {
			status.NextPage = nil
			status.HasFinished = true
			break
		} else {
			if !status.HasFinished {
				status.NextPage = next
			}
			page = *next
		}
	}
	if status != nil && (status.HasFinished || status.NextPage != nil) {
		logrus.Infof("Save Sig ts: %v next: %v", *status.LatestTimestamp, *status.NextPage)
		bt.SaveMethodSignatureImportStatus(*status)
	}

}

func GetNextSignitures(bt *db.Bigtable, page string, status types.MethodSignatureImportStatus) (*string, []types.MethodSignature, error) {

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
