package exporter

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type BidTrace struct {
	Slot                 uint64 `json:"slot,string"`
	ParentHash           string `json:"parent_hash"`
	BlockHash            string `json:"block_hash"`
	BuilderPubkey        string `json:"builder_pubkey"`
	ProposerPubkey       string `json:"proposer_pubkey"`
	ProposerFeeRecipient string `json:"proposer_fee_recipient"`
	GasLimit             uint64 `json:"gas_limit,string"`
	GasUsed              uint64 `json:"gas_used,string"`
	Value                string `json:"value"`
}

func mevBoostRelaysExporter() {
	var relays []types.Relay
	for {
		// we retrieve the relays from the db each loop to prevent having to restart the exporter for changes
		relays = nil
		err := db.ReaderDb.Select(&relays, `select * from relays`)
		if err == nil {
			for _, relay := range relays {
				// create relay logger
				relay.Logger = *logrus.New().WithFields(
					logrus.Fields{"module": "exporter", "relay": relay.ID})
				go singleRelayExport(relay)
			}
		} else {
			logger.Warnf("failed to retrieve relays from db: %v", err)
		}
		time.Sleep(time.Second * 60)
	}

}

func singleRelayExport(r types.Relay) {
	t0 := time.Now()
	err := exportRelayBlocks(r)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err, "duration": time.Since(t0)}).Errorf("error exporting %v tags", r.ID)
	}
}

func fetchDeliveredPayloads(r types.Relay, offset uint64) ([]BidTrace, error) {
	var payloads []BidTrace
	url := fmt.Sprintf("%s/relay/v1/data/bidtraces/proposer_payload_delivered?limit=100", r.Endpoint)
	if offset != 0 {
		url += fmt.Sprintf("&cursor=%v", offset)
	}
	r.Logger.Debugf("calling %v", url)

	resp, err := http.Get(url)

	if err != nil {
		r.Logger.Errorf("error retrieving delivered payloads: %v", err)
		return nil, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&payloads)

	if err != nil {
		r.Logger.Errorf("error decoding delivered payloads: %v", err)
		return nil, err
	}

	return payloads, nil
}

func exportRelayBlocks(r types.Relay) error {
	// retrieve the oldest tag usage so we know when to stop processing payloads from the head
	var lastUsage types.BlockTag
	err := db.ReaderDb.Get(&lastUsage, `SELECT * FROM blocks_tags WHERE tag_id=$1 ORDER BY slot DESC LIMIT 1`, r.ID)
	if err != nil {
		r.Logger.Errorf("failed to retrieve last tag usage from db, assuming none set: %v", err)
	}

	err = retrieveAndInsertPayloadsFromRelay(r, lastUsage.BlockSlot, 0)
	if err != nil {
		r.Logger.Errorf("failed to retrieve and insert new payloads: %v", err)
		return err
	}

	// to make sure we dont have an incomplete table, check if there are any payloads before our first tag usage
	var firstUsage types.BlockTag
	err = db.ReaderDb.Get(&firstUsage, `SELECT * FROM blocks_tags WHERE tag_id=$1 ORDER BY slot ASC LIMIT 1`, r.ID)
	if err != nil {
		r.Logger.Errorf("failed to retrieve first tag usage from db, assuming none set: %v", err)
	}
	if firstUsage.BlockSlot == 0 {
		return nil
	}
	err = retrieveAndInsertPayloadsFromRelay(r, 0, firstUsage.BlockSlot)
	if err != nil {
		r.Logger.Errorf("failed to retrieve and insert possibly missing payloads")
		return err
	}

	return nil
}

func retrieveAndInsertPayloadsFromRelay(r types.Relay, low_bound uint64, high_bound uint64) error {
	var min_slot uint64
	var block_hashes []string
	var slots []string

	offset := high_bound
	if low_bound > 10 {
		min_slot = low_bound - 10
	}

	if high_bound == 0 {
		r.Logger.Infof("loading payloads from head till %v", min_slot)
	} else if low_bound == 0 {
		r.Logger.Infof("loading payloads from %v till genesis", high_bound)
	}

	for {
		r.Logger.Debugf("fetching payloads with offset %v", offset)

		resp, err := fetchDeliveredPayloads(r, offset)
		if resp == nil {
			r.Logger.Errorf("got no payloads")
			return nil
		}

		if err != nil {
			r.Logger.Errorf("failed to fetch payloads: %v", err)
			return err
		}

		for _, payload := range resp {
			slots = append(slots, strconv.FormatUint(payload.Slot, 10))
			block_hashes = append(block_hashes, "'\\"+payload.BlockHash[1:]+"'")
		}

		// add to db
		res, err := db.WriterDb.Exec(`
			insert into blocks_tags 
			select blocks.slot, blocks.blockroot, $1 
			from blocks 
			where 
				blocks.slot in (`+strings.Join(slots, ",")+`) and 
				blocks.exec_block_hash in (`+strings.Join(block_hashes, ",")+`)
			on conflict do nothing`, r.ID)

		if err != nil {
			r.Logger.Errorf("failed to insert block tags into db:", err)
			return err
		}

		changedCount, _ := res.RowsAffected()
		if changedCount > 0 {
			r.Logger.Infof("inserted %v new entries to blocks_tags table", changedCount)
		}

		if resp[len(resp)-1].Slot < min_slot {
			// last payload we received is bellow than our calculated min_slot
			r.Logger.Infof("retrieved all payloads above slot %v", min_slot)
			break
		}

		if len(resp) < 100 {
			// if the response is less than 100 payloads, we assume that we have reached the end and break
			r.Logger.Debugf("got %v, expected 100 payloads", len(resp))
			r.Logger.Infof("no more payloads avaliable")
			break
		}

		// sleep for a bit to not kill the relay
		r.Logger.Debugf("sleeping 5 seconds before next request")
		offset = resp[len(resp)-1].Slot
		time.Sleep(time.Second * 5)
	}
	return nil
}
