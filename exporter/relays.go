package exporter

import (
	"database/sql"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type BidTrace struct {
	Slot                 uint64          `json:"slot,string"`
	ParentHash           string          `json:"parent_hash"`
	BlockHash            string          `json:"block_hash"`
	BuilderPubkey        string          `json:"builder_pubkey"`
	ProposerPubkey       string          `json:"proposer_pubkey"`
	ProposerFeeRecipient string          `json:"proposer_fee_recipient"`
	GasLimit             uint64          `json:"gas_limit,string"`
	GasUsed              uint64          `json:"gas_used,string"`
	Value                types.WeiString `json:"value"`
}

func mevBoostRelaysExporter() {
	var relays []types.Relay
	for {
		// we retrieve the relays from the db each loop to prevent having to restart the exporter for changes
		relays = nil
		err := db.ReaderDb.Select(&relays, `select tag_id, endpoint, public_link, is_censoring, is_ethical, export_failure_count, last_export_try_ts, last_export_success_ts from relays`)
		wg := &sync.WaitGroup{}
		mux := &sync.Mutex{}
		if err == nil {
			for _, relay := range relays {
				if shouldTryToExportRelay(relay) {
					// create relay logger
					relay.Logger = *logrus.New().WithFields(
						logrus.Fields{"module": "exporter", "relay": relay.ID})
					wg.Add(1)
					go singleRelayExport(relay, wg, mux)
				}
			}
		} else if err != sql.ErrNoRows {
			utils.LogError(err, "failed to retrieve relays from db", 0)
		}
		wg.Wait()
		time.Sleep(time.Minute)
	}

}

func singleRelayExport(r types.Relay, wg *sync.WaitGroup, mux *sync.Mutex) {
	defer wg.Done()

	err := exportRelayBlocks(r)
	if err != nil {
		errMsg := fmt.Errorf("failed to export blocks for relay: %v", err)
		if shouldLogExportAsError(r) {
			r.Logger.Error(errMsg)
		} else {
			r.Logger.Warn(errMsg)
		}

		// Only increase the export_failure_count if we haven't already reached the maximum wait time
		_, isMaxWaitTime := waitTimeToExportRelay(r)
		mux.Lock()
		if isMaxWaitTime {
			_, err = db.WriterDb.Exec(`
			UPDATE relays SET
				last_export_try_ts = (NOW() AT TIME ZONE 'utc')
			WHERE tag_id = $1 AND endpoint = $2`, r.ID, r.Endpoint)
		} else {
			_, err = db.WriterDb.Exec(`
			UPDATE relays SET
				export_failure_count = $1,
				last_export_try_ts = (NOW() AT TIME ZONE 'utc')
			WHERE tag_id = $2 AND endpoint = $3`, r.ExportFailureCount+1, r.ID, r.Endpoint)
		}
		mux.Unlock()
		if err != nil {
			r.Logger.Errorf("Could not update failed relay export: %v", r.ID)
		}

		return
	}

	mux.Lock()
	_, err = db.WriterDb.Exec(`
			UPDATE relays SET
				export_failure_count = 0,
				last_export_try_ts = (NOW() AT TIME ZONE 'utc'),
				last_export_success_ts = (NOW() AT TIME ZONE 'utc')
			WHERE tag_id = $1 AND endpoint = $2`, r.ID, r.Endpoint)
	mux.Unlock()
	if err != nil {
		r.Logger.Errorf("Could not update successful relay eport: %v", r.ID)
	}

	r.Logger.Infof("finished syncing payloads from relay")
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
		r.Logger.Tracef("error decoding delivered payloads: %v", err)
		return nil, err
	}

	return payloads, nil
}

func exportRelayBlocks(r types.Relay) error {
	// retrieve the oldest tag usage so we know when to stop processing payloads from the head
	var lastUsage types.RelayBlock
	err := db.ReaderDb.Get(&lastUsage, `SELECT tag_id, block_slot, block_root, exec_block_hash, value, builder_pubkey, proposer_pubkey, proposer_fee_recipient FROM relays_blocks WHERE tag_id=$1 ORDER BY block_slot DESC LIMIT 1`, r.ID)
	if err != nil {
		r.Logger.Errorf("failed to retrieve last relay block from db, assuming none set: %v", err)
	}

	err = retrieveAndInsertPayloadsFromRelay(r, lastUsage.BlockSlot, 0)
	if err != nil {
		r.Logger.Tracef("failed to retrieve and insert new payloads: %v", err)
		return err
	}

	// to make sure we dont have an incomplete table, check if there are any payloads before our first tag usage
	var firstUsage types.RelayBlock
	err = db.ReaderDb.Get(&firstUsage, `SELECT tag_id, block_slot, block_root, exec_block_hash, value, builder_pubkey, proposer_pubkey, proposer_fee_recipient FROM relays_blocks WHERE tag_id=$1 ORDER BY block_slot ASC LIMIT 1`, r.ID)
	if err != nil {
		r.Logger.Errorf("failed to retrieve first relay block from db, assuming none set: %v", err)
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
	tx, err := db.WriterDb.Begin()
	if err != nil {
		r.Logger.WithFields(logrus.Fields{
			"file":       "relays.go",
			"function":   "retrieveAndInsertPayloadsFromRelay",
			"Relay ID":   r.ID,
			"low_bound":  low_bound,
			"high_bound": high_bound,
		}).WithError(err).Error("failed to start db transaction")
		return err
	}
	defer tx.Rollback()

	var min_slot uint64
	if low_bound > 10 {
		min_slot = low_bound - 10
	}

	if high_bound == 0 {
		r.Logger.Debugf("loading payloads from head till %v", min_slot)
	} else if low_bound == 0 {
		r.Logger.Debugf("loading payloads from %v till genesis", high_bound)
	}

	offset := high_bound
	for {
		r.Logger.Debugf("fetching payloads with offset %v", offset)

		resp, err := fetchDeliveredPayloads(r, offset)
		if err != nil {
			r.Logger.Tracef("failed to fetch payloads: %v", err)
			return err
		}

		if resp == nil {
			r.Logger.Error("got no payloads")
			break
		}

		for _, payload := range resp {
			// first insert the tag into the blocks_tags table
			_, err = tx.Exec(`
				insert into blocks_tags
				select blocks.slot, blocks.blockroot, $1
				from blocks
				where 
					blocks.slot = $2 and
					blocks.exec_block_hash = $3
				ON CONFLICT DO NOTHING`, r.ID, payload.Slot, utils.MustParseHex(payload.BlockHash))
			if err != nil {
				r.Logger.Error("failed to insert payload into blocks_tags table")
				return err
			}
			_, err = tx.Exec(`
				insert into relays_blocks
				(
					tag_id,
					block_slot,
					block_root,
					exec_block_hash,
					value,
					builder_pubkey,
					proposer_pubkey,
					proposer_fee_recipient
				)
				select 
					$1,	blocks.slot, blocks.blockroot, blocks.exec_block_hash, $4, $5, $6, $7
				from blocks
				where
					blocks.slot = $2 and
					blocks.exec_block_hash = $3
				ON CONFLICT (block_slot, block_root, tag_id) DO NOTHING`,
				r.ID, payload.Slot, utils.MustParseHex(payload.BlockHash),
				payload.Value, utils.MustParseHex(payload.BuilderPubkey),
				utils.MustParseHex(payload.ProposerPubkey),
				utils.MustParseHex(payload.ProposerFeeRecipient))
			if err != nil {
				r.Logger.Error("failed to insert payload into relays_blocks table")
				return err
			}
		}

		if len(resp) == 0 || resp[len(resp)-1].Slot < min_slot {
			// last payload we received is bellow than our calculated min_slot
			r.Logger.Debugf("retrieved all payloads above slot %v", min_slot)
			break
		}

		if len(resp) < 100 {
			// if the response is less than 100 payloads, we assume that we have reached the end and break
			r.Logger.Debugf("got %v, expected 100 payloads", len(resp))
			r.Logger.Debugf("no more payloads available")
			break
		}
		if resp[len(resp)-1].Slot == offset {
			return fmt.Errorf("relay doesn't follow spec, last returned slot matches offset (sort order ascending instead of descending)")
		}

		// sleep for a bit to not kill the relay
		r.Logger.Debugf("sleeping 2 seconds before next request")
		offset = resp[len(resp)-1].Slot
		time.Sleep(time.Second * 1)
	}
	return tx.Commit()
}

func shouldTryToExportRelay(r types.Relay) bool {
	if r.ExportFailureCount == 0 {
		return true
	}

	waitTime, _ := waitTimeToExportRelay(r)
	return time.Since(r.LastExportTryTs) >= waitTime
}

func waitTimeToExportRelay(r types.Relay) (waitTime time.Duration, isMaxWaitTime bool) {
	maxWaitTimeForRelayExport := utils.Day
	waitTime = time.Duration(math.Exp2(float64(r.ExportFailureCount))) * time.Minute
	if waitTime >= maxWaitTimeForRelayExport {
		waitTime = maxWaitTimeForRelayExport
		isMaxWaitTime = true
	}
	return
}

func shouldLogExportAsError(r types.Relay) bool {
	maxWaitTimeForRelayExportError := utils.Month

	return time.Since(r.LastExportSuccessTs) >= maxWaitTimeForRelayExportError
}
