package exporter

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type SSVExporterResponse struct {
	Type   string `json:"type"`
	Filter struct {
		From int `json:"from"`
		To   int `json:"to"`
	} `json:"filter"`
	Data []struct {
		Index     int    `json:"index"`
		Publickey string `json:"publicKey"`
		Operators []struct {
			Nodeid    int    `json:"nodeId"`
			Publickey string `json:"publicKey"`
		} `json:"operators"`
	} `json:"data"`
}

func ssvExporter() {
	for {
		err := exportSSV()
		if err != nil {
			logger.WithError(err).Error("error exporting ssv validators")
		}
		logger.Warning("connection to ssv-exporter closed, reconnecting")
		time.Sleep(time.Second * 10)
	}
}

func exportSSV() error {
	c, _, err := websocket.DefaultDialer.Dial(utils.Config.SSVExporter.Address, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logger.WithError(err).Error("error reading message from ssv-exporter")
				return
			}
			t0 := time.Now()
			res := SSVExporterResponse{}
			err = json.Unmarshal(message, &res)
			if err != nil {
				logger.WithError(err).Error("error unmarshaling json from ssv-exporter")
				continue
			}
			err = saveSSV(&res)
			if err != nil {
				logger.WithError(err).Error("error tagging ssv validators")
				continue
			}

			logger.WithFields(logrus.Fields{"number": len(res.Data), "duration": time.Since(t0)}).Infof("tagged ssv validators")
		}
	}()

	qryValidatorsTicker := time.NewTicker(time.Second * 600)
	defer qryValidatorsTicker.Stop()

	for {
		err := c.WriteMessage(websocket.TextMessage, []byte(`{"type":"validator","filter":{"from":0}}`))
		if err != nil {
			return err
		}
		select {
		case <-qryValidatorsTicker.C:
			continue
		case <-done:
			return nil
		}
	}

	return nil
}

func saveSSV(res *SSVExporterResponse) error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`delete from validator_tags where tag = 'ssv'`)
	if err != nil {
		return err
	}

	batchSize := 5000
	for b := 0; b < len(res.Data); b += batchSize {
		start := b
		end := b + batchSize
		if len(res.Data) < end {
			end = len(res.Data)
		}
		n := 1
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*n)
		for i, d := range res.Data[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, 'ssv')", i*n+1))
			pubkey, err := hex.DecodeString(strings.Replace(d.Publickey, "0x", "", -1))
			if err != nil {
				return err
			}
			valueArgs = append(valueArgs, pubkey)
		}
		stmt := fmt.Sprintf(`insert into validator_tags (publickey, tag) values %s on conflict (publickey, tag) do nothing`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
