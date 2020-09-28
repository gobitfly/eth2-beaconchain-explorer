package main

import (
	"encoding/binary"
	"fmt"

	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/tree"
)

func CreateTestValidators(count uint64, balance Gwei) []KickstartValidatorData {
	out := make([]KickstartValidatorData, 0, count)
	for i := uint64(0); i < count; i++ {
		pubkey := BLSPubkey{0xaa}
		binary.LittleEndian.PutUint64(pubkey[1:], i)
		withdrawalCred := Root{0xbb}
		binary.LittleEndian.PutUint64(withdrawalCred[1:], i)
		out = append(out, KickstartValidatorData{
			Pubkey:                pubkey,
			WithdrawalCredentials: withdrawalCred,
			Balance:               balance,
		})
	}
	return out
}

func CreateTestState(spec *Spec, validatorCount uint64, balance Gwei) (*BeaconStateView, *EpochsContext) {
	out, epc, err := spec.KickStartState(Root{123}, 1564000000, CreateTestValidators(validatorCount, balance))
	if err != nil {
		panic(err)
	}
	return out, epc
}

func main() {

	// can load other testnet configurations as well
	spec := configs.Mainnet

	state, epc := CreateTestState(spec, 1000, spec.MAX_EFFECTIVE_BALANCE)

	for i := Slot(0); i < spec.SLOTS_PER_EPOCH*2; i++ {
		count, err := epc.GetCommitteeCountAtSlot(i)
		if err != nil {
			panic(err)
		}

		fmt.Printf("slot %d, committee count: %d\n", i, count)
		for j := uint64(0); j < count; j++ {
			committee, err := epc.GetBeaconCommittee(i, CommitteeIndex(j))
			if err != nil {
				panic(err)
			}
			fmt.Printf("slot %d, committee %d: %v\n", i, j, committee)
		}
	}

	root := state.HashTreeRoot(tree.GetHashFn())
	fmt.Printf("state root: %s\n", root)

}
