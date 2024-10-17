package rpc

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestName(t *testing.T) {
	client, err := NewErigonClient("http://localhost:8545")
	if err != nil {
		t.Fatal(err)
	}

	tracesG, err := client.getTraceGeth(big.NewInt(20928629), common.HexToHash("0x6d2d2f8cf1e635dcca30beaf511fb0a3d2d4df32daa671b410e7ba0014b85e6c"))
	if err != nil {
		t.Fatal(err)
	}
	tracesP, err := client.getTraceParity(big.NewInt(20928629), common.HexToHash("0x6d2d2f8cf1e635dcca30beaf511fb0a3d2d4df32daa671b410e7ba0014b85e6c"), 200)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(tracesG))
	fmt.Println(len(tracesP))

	indexG, indexP := 0, 0
	for i := 0; i < 191; i++ {
		countG, countP := 0, 0
		for ; indexG < len(tracesG) && tracesG[indexG].txPosition == i; indexG++ {
			countG++
		}
		for ; indexP < len(tracesP) && tracesP[indexP].txPosition == i; indexP++ {
			countP++
		}
		if countG != countP {
			t.Errorf("%d got %d want %d", i, countG, countP)
		}
	}

	/*for i := 0; i < len(tracesG); i++ {
		if got, want := tracesG[i].Path, tracesP[i].Path; got != want {
			fmt.Println(tracesG[i].Type)
			fmt.Println(tracesP[i].Type)
			t.Errorf("tx %d:%d got %s want %s", tracesP[i].txPosition, i, got, want)
		}
	}*/

}
