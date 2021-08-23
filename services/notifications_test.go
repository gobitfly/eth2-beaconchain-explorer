package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"testing"

	"github.com/lib/pq"
)

func TestBalanceDecrease(t *testing.T) {
	latestEpoch := LatestEpoch()
	t.Logf("Running test Balance Decrease for epoch: %v", latestEpoch)
	result, err := db.GetValidatorsBalanceDecrease(latestEpoch)
	if err != nil {
		t.Errorf("error getting validators balance decrease %v", err)
		return
	}

	testUsers := []int64{
		7,
		10,
	}

	t.Logf("found %v validators losing balance", len(result))

	if len(result) > 0 {
		valOne := result[0]
		for _, user := range testUsers {
			err := db.AddTestSubscription(uint64(user), "mainnet", types.ValidatorBalanceDecreasedEventName, valOne.Pubkey, 0, latestEpoch-1)
			if err != nil {
				t.Errorf("error creating test subscription %v", err)
				return
			}
		}

		t.Cleanup(func() {
			_, err := db.FrontendDB.Exec("DELETE FROM users_subscriptions where user_id = ANY($1)", pq.Int64Array(testUsers))
			if err != nil {
				t.Errorf("error cleaning up TestBalanceDecrease err: %v", err)
				return
			}
		})
	} else {
		t.Error("error no validators are losing a balance, this test cannot complete")
		return
	}

	notificationsByUserID := map[uint64]map[types.EventName][]types.Notification{}
	err = collectValidatorBalanceDecreasedNotifications(notificationsByUserID)
	if err != nil {
		t.Errorf("error collecting balance decrease notifications err: %v", err)
	}

	t.Logf("notifications recorded: %v", notificationsByUserID)

	subs, ok := notificationsByUserID[10]
	if !ok {
		t.Errorf("no notifications for user %v exist in %+v", 10, notificationsByUserID)
		return
	}

	t.Logf("test user has the following subs: %v", subs)

	notifications, ok := subs[types.ValidatorBalanceDecreasedEventName]
	if !ok {
		t.Errorf("no notifications for user %v exist in %+v", 10, notificationsByUserID)
		return
	}

	if len(notifications) == 0 {
		t.Errorf("error expected to receive at least one event")
		return
	}

	t.Logf("notifications for test user %v", notifications)

	expected := result[0].Pubkey
	got := notifications[0].GetEventFilter()
	if got != expected {
		t.Errorf("error unexpected event created expected: %v but got %v", expected, got)
		return
	}
}

func TestGotSlashedNotifications(t *testing.T) {
	latestEpoch := LatestEpoch()
	t.Logf("Running test for got slashed notification: %v", latestEpoch)
}

func TestAttestationViolationNotification(t *testing.T) {
	latestEpoch := LatestEpoch()
	latestSlot := LatestSlot()
	notificationsByUserID := map[uint64]map[types.EventName][]types.Notification{}
	t.Logf("Testing Attestation Violation for epoch: %v and slot: %v", latestEpoch, latestSlot)
	tx, err := db.DB.Beginx()
	if err != nil {
		t.Errorf("error creating tx err: %v", err)
	}
	defer tx.Rollback()

	// insert a test attestation violation
	rows, err := tx.Query(`
	INSERT INTO blocks_attesterslashings (
		block_slot,
		block_index,
		attestation1_indices,
		attestation2_indices,
		block_root,
		attestation1_signature,
		attestation1_slot,
		attestation1_index,
		attestation1_beaconblockroot,
		attestation1_source_epoch,
		attestation1_source_root,
		attestation1_target_epoch,
		attestation1_target_root,
		attestation2_signature,
		attestation2_slot,
		attestation2_index,
		attestation2_beaconblockroot,
		attestation2_source_epoch,
		attestation2_source_root,
		attestation2_target_epoch,
		attestation2_target_root
	) SELECT
		(SELECT slot from blocks where status = '1' order by slot desc limit 1) as block_slot,
		1 as block_index,
		$1 as attestation1_indices,
		$1 as attestation2_indices,
		b.*
		FROM (SELECT 
		  block_root,
		  attestation1_signature,
		  attestation1_slot,
		  attestation1_index,
		  attestation1_beaconblockroot,
		  attestation1_source_epoch,
		  attestation1_source_root,
		  attestation1_target_epoch,
		  attestation1_target_root,
		  attestation2_signature,
		  attestation2_slot,
		  attestation2_index,
		  attestation2_beaconblockroot,
		  attestation2_source_epoch,
		  attestation2_source_root,
		  attestation2_target_epoch,
		  attestation2_target_root
	FROM blocks_attesterslashings ORDER BY block_slot desc LIMIT 1) b 
	RETURNING block_slot`, pq.Int64Array([]int64{50, 60}))
	if err != nil {
		t.Errorf("error inserting dummy AttestationViolation err: %v", err)
		return
	}

	for rows.Next() {
		var slot uint64
		rows.Scan(&slot)
		t.Logf("Included an attestation violation in slot %v", slot)
	}

	err = collectValidatorGotSlashedNotifications(notificationsByUserID)
	if err != nil {
		t.Errorf("error collecting validator_got_slashed notifications err: %v", err)
	}

	t.Logf("ready to send: %+v notifications", notificationsByUserID)

	// we copied this query because the changes are only visible inside the transaction
	rows, err = tx.Query(`
			WITH
			slashings AS (
				SELECT DISTINCT ON (slashedvalidator) * FROM (
					SELECT
						blocks.slot, 
						blocks.epoch, 
						blocks.proposer AS slasher, 
						UNNEST(ARRAY(
							SELECT UNNEST(attestation1_indices)
								INTERSECT
							SELECT UNNEST(attestation2_indices)
						)) AS slashedvalidator, 
						'Attestation Violation' AS reason
					FROM blocks_attesterslashings 
					LEFT JOIN blocks ON blocks_attesterslashings.block_slot = blocks.slot
					WHERE blocks.status = '1' AND blocks.epoch > $1
					UNION ALL
						SELECT
							blocks.slot, 
							blocks.epoch, 
							blocks.proposer AS slasher, 
							blocks_proposerslashings.proposerindex AS slashedvalidator,
							'Proposer Violation' AS reason 
						FROM blocks_proposerslashings
						LEFT JOIN blocks ON blocks_proposerslashings.block_slot = blocks.slot
						WHERE blocks.status = '1' AND blocks.epoch > $1
				) a
				ORDER BY slashedvalidator, slot
			)
		SELECT slasher, vk.pubkey as slasher_pubkey, slashedvalidator, vv.pubkey as slashedvalidator_pubkey, epoch, reason
		FROM slashings s
		INNER JOIN validators vk ON s.slasher = vk.validatorindex
		INNER JOIN validators vv ON s.slashedvalidator = vv.validatorindex
	`, latestEpoch-10)
	if err != nil {
		t.Errorf("error getting recent slashable offences %v", err)
		return
	}

	var dbResults []struct {
		Epoch                  uint64 `db:"epoch"`
		SlasherIndex           uint64 `db:"slasher"`
		SlasherPubkey          string `db:"slasher_pubkey"`
		SlashedValidatorIndex  uint64 `db:"slashedvalidator"`
		SlashedValidatorPubkey string `db:"slashedvalidator_pubkey"`
		Reason                 string `db:"reason"`
	}

	for rows.Next() {
		var dbResult struct {
			Epoch                  uint64 `db:"epoch"`
			SlasherIndex           uint64 `db:"slasher"`
			SlasherPubkey          string `db:"slasher_pubkey"`
			SlashedValidatorIndex  uint64 `db:"slashedvalidator"`
			SlashedValidatorPubkey string `db:"slashedvalidator_pubkey"`
			Reason                 string `db:"reason"`
		}

		err = rows.Scan(&dbResult.SlasherIndex, &dbResult.SlasherPubkey, &dbResult.SlashedValidatorIndex, &dbResult.SlashedValidatorPubkey, &dbResult.Epoch, &dbResult.Reason)
		if err != nil {
			t.Errorf("error scanning for slashings err: %v", err)
			return
		}
		t.Logf("found slashing offence %+v", dbResult)
		dbResults = append(dbResults, dbResult)
	}

	if len(dbResults) == 0 {
		t.Errorf("error expected two slashing events but got %v", len(dbResults))
		return
	}

	for _, result := range dbResults {
		if result.Reason != "Attestation Violation" {
			t.Errorf("error expected slashing violation to be: %v but received: %v for slashed validator: %v", "Attestation Violation", result.Reason, result.SlashedValidatorIndex)
		}
	}

	t.Logf("found db results %+v", dbResults)
}

// TestProposerViolationNotification tests wether a notification gets created for a subscribed user
func TestProposerViolationNotification(t *testing.T) {

}

func TestBlockProposalSubmittedNotification(t *testing.T) {

}

func TestBlockProposalMissedNotification(t *testing.T) {

}

func TestEthClientNotifications(t *testing.T) {

}

func TestTaxReportNotifications(t *testing.T) {

}
