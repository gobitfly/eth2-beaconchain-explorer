package main

import (
	"eth2-exporter/eth2api"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

var clientURL string
var configPath string

func main() {
	flag.StringVar(&clientURL, "url", "", "url")
	flag.StringVar(&configPath, "config", "config-example.yml", "Path to the config file")
	flag.Parse()
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	Start(clientURL)
}

func Start(clientURL string) {
	// fmt.Println(utils.EpochToTime(1))
	client, err := eth2api.NewClient(clientURL)
	if err != nil {
		panic(err)
	}
	_ = client
	// test1(client)
	// test2(client)
	test3(clientURL)
}

func test3(clientURL string) {
	client, err := rpc.NewEth2ApiV1Client(clientURL)
	if err != nil {
		panic(err)
	}
	ch, err := client.GetChainHead()
	if err != nil {
		panic(err)
	}

	t0 := time.Now()
	epoch := ch.HeadEpoch
	ed, err := client.GetEpochData(epoch)
	fmt.Println("GetEpochData", err, ed.Epoch, time.Since(t0))

	t0 = time.Now()
	epoch = ch.HeadEpoch - 1000
	ed, err = client.GetEpochData(epoch)
	fmt.Println("GetEpochData", err, ed.Epoch, time.Since(t0))

	t0 = time.Now()
	epoch = ch.HeadEpoch - 5000
	ed, err = client.GetEpochData(epoch)
	fmt.Println("GetEpochData", err, ed.Epoch, time.Since(t0))

	t0 = time.Now()
	epoch = ch.HeadEpoch - 10000
	ed, err = client.GetEpochData(epoch)
	fmt.Println("GetEpochData", err, ed.Epoch, time.Since(t0))
}

func test1(client *eth2api.Client) {
	vs, err := client.GetValidators("0")
	if err != nil {
		panic(err)
	}
	fmt.Printf("there are %d genesis-validators\n", len(vs))
	validatorsByState(client, "0")
	validatorsByState(client, "512")
	fmt.Println("-----")
}

func test2(client *eth2api.Client) {
	// for i := 12300 * 32; i < 12300*32+1024; i += 32 {
	for i := 10000 * 32; i < 10000*32+1024; i += 32 {
		validatorsByState(client, fmt.Sprintf("%d", i))
		t0 := time.Now()
		g, err := client.GetGenesis()
		fmt.Println("g", time.Since(t0), err, g)
		t0 = time.Now()
		r, err := client.GetRoot(fmt.Sprintf("%v", i))
		fmt.Println("r", time.Since(t0), err, r)
		t0 = time.Now()
		f, err := client.GetFork(fmt.Sprintf("%v", i))
		fmt.Println("f", time.Since(t0), err, f)
		t0 = time.Now()
		fc, err := client.GetFinalityCheckpoints(fmt.Sprintf("%v", i))
		fmt.Println("fc", time.Since(t0), err, fc)
		t0 = time.Now()
		vs, err := client.GetValidators(fmt.Sprintf("%v", i))
		fmt.Println("vs", time.Since(t0), err, len(vs))
		t0 = time.Now()
		ads, err := client.GetAttesterDuties(uint64(i / 32))
		fmt.Println("ads", time.Since(t0), err, len(ads))
		// t0 = time.Now()
		// pds, err := client.GetProposerDuties(uint64(i / 32))
		// fmt.Println("pds", time.Since(t0), err, len(*pds))
		t0 = time.Now()
		b, err := client.GetBlock(fmt.Sprintf("%v", i))
		fmt.Println("b", time.Since(t0), err, b.Message.Slot)
		t0 = time.Now()
		b, err = client.GetBlock(fmt.Sprintf("%v", i+1))
		fmt.Println("b2", time.Since(t0), err, b.Message.Slot)
		t0 = time.Now()
		b, err = client.GetBlock(fmt.Sprintf("%v", i+2))
		fmt.Println("b3", time.Since(t0), err, b.Message.Slot)
		t0 = time.Now()
		cs, err := client.GetCommittees(fmt.Sprintf("%v", i), uint64(i/32))
		fmt.Println("cs", time.Since(t0), err, len(cs))
		if len(cs) > 0 {
			fmt.Println("cs[0]", *cs[0])
		}
		fmt.Println("---------")
	}
}

func validatorsByState(client *eth2api.Client, slot string) {
	t0 := time.Now()
	vs, err := client.GetValidators(slot)
	if err != nil {
		panic(err)
	}
	t1 := time.Now()
	vByStatus := map[string][]*eth2api.Validator{}
	for _, v := range vs {
		_, exists := vByStatus[string(v.Status)]
		if !exists {
			vByStatus[string(v.Status)] = []*eth2api.Validator{v}
		} else {
			vByStatus[string(v.Status)] = append(vByStatus[string(v.Status)], v)
		}
	}
	for k, v := range vByStatus {
		fmt.Printf("slot: %v, status: %v, len(vs): %v, dur: %v\n", slot, k, len(v), t1.Sub(t0))
	}
}
