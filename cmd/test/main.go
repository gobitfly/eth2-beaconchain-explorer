package main

import (
	"encoding/binary"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/tree"
)

func CreateTestValidators(count uint64, balance beacon.Gwei) []beacon.KickstartValidatorData {
	out := make([]beacon.KickstartValidatorData, 0, count)
	for i := uint64(0); i < count; i++ {
		pubkey := beacon.BLSPubkey{0xaa}
		binary.LittleEndian.PutUint64(pubkey[1:], i)
		withdrawalCred := beacon.Root{0xbb}
		binary.LittleEndian.PutUint64(withdrawalCred[1:], i)
		out = append(out, beacon.KickstartValidatorData{
			Pubkey:                pubkey,
			WithdrawalCredentials: withdrawalCred,
			Balance:               balance,
		})
	}
	return out
}

func CreateTestState(spec *beacon.Spec, validatorCount uint64, balance beacon.Gwei) (*beacon.BeaconStateView, *beacon.EpochsContext) {
	out, epc, err := spec.KickStartState(beacon.Root{123}, 1564000000, CreateTestValidators(validatorCount, balance))
	if err != nil {
		panic(err)
	}
	return out, epc
}

func main() {

	// can load other testnet configurations as well
	spec := configs.Mainnet

	state, epc := CreateTestState(spec, 1000, spec.MAX_EFFECTIVE_BALANCE)

	for i := beacon.Slot(0); i < spec.SLOTS_PER_EPOCH*2; i++ {
		count, err := epc.GetCommitteeCountAtSlot(i)
		if err != nil {
			panic(err)
		}

		fmt.Printf("slot %d, committee count: %d\n", i, count)
		for j := uint64(0); j < count; j++ {
			committee, err := epc.GetBeaconCommittee(i, beacon.CommitteeIndex(j))
			if err != nil {
				panic(err)
			}
			fmt.Printf("slot %d, committee %d: %v\n", i, j, committee)
		}
	}

	root := state.HashTreeRoot(tree.GetHashFn())
	fmt.Printf("state root: %s\n", root)

}
