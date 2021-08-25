package db

import (
	"encoding/hex"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"testing"

	"github.com/lib/pq"
)

func TestWatchlist(t *testing.T) {
	testUsers := []int64{
		7,
		10,
	}
	pubkey := "98ea8ae1894658d30b22803804bcac681e95d0d2b88a3dd1ffaf9423ecedf2ba139e975a6633596c3e9a8a7df2081740"
	watchList := []WatchlistEntry{
		{
			UserId:              uint64(testUsers[0]),
			Validator_publickey: pubkey,
		},
	}
	err := AddToWatchlist(watchList, utils.GetNetwork())
	if err != nil {
		t.Errorf("error adding validator with pubkey: %v to user %v err: %v", pubkey, testUsers[0], err)
		return
	}
	t.Cleanup(func() {
		err := RemoveFromWatchlist(uint64(testUsers[0]), pubkey, utils.GetNetwork())
		if err != nil {
			t.Errorf("error cleaning up could not remove validator: %v from the watchlist for user: %v err: %v", pubkey, testUsers[0], err)
			return
		}
	})

	pub, err := hex.DecodeString(pubkey)
	if err != nil {
		t.Errorf("error decoding pubkey err: %v", err)
		return
	}

	filter := WatchlistFilter{
		Tag:            types.ValidatorTagsWatchlist,
		UserId:         uint64(testUsers[0]),
		Network:        utils.GetNetwork(),
		Validators:     &pq.ByteaArray{pub},
		JoinValidators: true,
	}

	taggedValidators, err := GetTaggedValidators(filter)
	if err != nil {
		t.Errorf("error getting tagged validators err: %v", err)
		return
	}

	if len(taggedValidators) != 1 {
		t.Errorf("error expected to receive 1 validator got: %v validators, %+v", len(taggedValidators), taggedValidators)
		return
	}

	validator := taggedValidators[0]
	if hex.EncodeToString(validator.PublicKey) != pubkey {
		t.Errorf("error expected validator with pubkey: \n^%v$ \nbut got: \n^%v$", pubkey, hex.EncodeToString(validator.PublicKey))
		return
	}

	if validator.UserID != uint64(testUsers[0]) {
		t.Errorf("error expected validators for user: %v but got user: %v", testUsers[0], validator.UserID)
		return
	}

	if validator.Tag != network+":"+string(types.ValidatorTagsWatchlist) {
		t.Errorf("error expected tag to be: %v but got: %v", network+":"+string(types.ValidatorTagsWatchlist), validator.Tag)
		return
	}

	if hex.EncodeToString(validator.PublicKey) != hex.EncodeToString(validator.ValidatorPublickey) {
		t.Errorf("error expected validator public key to be %v but got %v", hex.EncodeToString(validator.ValidatorPublickey), hex.EncodeToString(validator.PublicKey))
	}

	if validator.Balance == 0 {
		t.Errorf("error expected the validator balance to be greater than 0 but got: %v", validator.Balance)
	}

	if validator.Index == 0 {
		t.Errorf("error expected the validator index to be set but got: %v", validator.Index)
	}

	// t.Logf("completed test with response: %+v, validator: %+v", validator, validator.Validator)

}
