package db2

import (
	"context"
	"testing"

	"github.com/gobitfly/eth2-beaconchain-explorer/db2/store"
	"github.com/gobitfly/eth2-beaconchain-explorer/db2/storetest"
)

func TestRaw(t *testing.T) {
	client, admin := storetest.NewBigTable(t)

	s, err := store.NewBigTableWithClient(context.Background(), client, admin, raw)
	if err != nil {
		t.Fatal(err)
	}

	db := RawStore{
		store:      store.Wrap(s, BlocRawTable, ""),
		compressor: noOpCompressor{},
	}

	block := FullBlockRawData{
		ChainID:          1,
		BlockNumber:      testBlockNumber,
		BlockHash:        nil,
		BlockUnclesCount: 1,
		BlockTxs:         nil,
		Block:            []byte(testBlock),
		Receipts:         []byte(testReceipts),
		Traces:           []byte(testTraces),
		Uncles:           []byte(testUncles),
	}

	if err := db.AddBlocks([]FullBlockRawData{block}); err != nil {
		t.Fatal(err)
	}

	res, err := db.ReadBlockByNumber(block.ChainID, block.BlockNumber)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := string(res.Block), testBlock; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := string(res.Receipts), testReceipts; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := string(res.Traces), testTraces; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := string(res.Uncles), testUncles; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

var testFullBlock = FullBlockRawData{
	ChainID:          1,
	BlockNumber:      testBlockNumber,
	BlockUnclesCount: 1,
	Block:            []byte(testBlock),
	Receipts:         []byte(testReceipts),
	Traces:           []byte(testTraces),
	Uncles:           []byte(testUncles),
}

var testTwoUnclesFullBlock = FullBlockRawData{
	ChainID:          1,
	BlockNumber:      testTwoUnclesBlockNumber,
	BlockUnclesCount: 2,
	Block:            []byte(testTwoUnclesBlock),
	Receipts:         nil,
	Traces:           nil,
	Uncles:           []byte(testTwoUnclesBlockUncles),
}

const (
	testBlockNumber = 6008149
	testBlock       = `{
   "id":1,
   "jsonrpc":"2.0",
   "result":{
      "difficulty":"0xbfabcdbd93dda",
      "extraData":"0x737061726b706f6f6c2d636e2d6e6f64652d3132",
      "gasLimit":"0x79f39e",
      "gasUsed":"0x79ccd3",
      "hash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
      "logsBloom":"0x4848112002a2020aaa0812180045840210020005281600c80104264300080008000491220144461026015300100000128005018401002090a824a4150015410020140400d808440106689b29d0280b1005200007480ca950b15b010908814e01911000054202a020b05880b914642a0000300003010044044082075290283516be82504082003008c4d8d14462a8800c2990c88002a030140180036c220205201860402001014040180002006860810ec0a1100a14144148408118608200060461821802c081000042d0810104a8004510020211c088200420822a082040e10104c00d010064004c122692020c408a1aa2348020445403814002c800888208b1",
      "miner":"0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c",
      "mixHash":"0x3d1fdd16f15aeab72e7db1013b9f034ee33641d92f71c0736beab4e67d34c7a7",
      "nonce":"0x4db7a1c01d8a8072",
      "number":"0x5bad55",
      "parentHash":"0x61a8ad530a8a43e3583f8ec163f773ad370329b2375d66433eb82f005e1d6202",
      "receiptsRoot":"0x5eced534b3d84d3d732ddbc714f5fd51d98a941b28182b6efe6df3a0fe90004b",
      "sha3Uncles":"0x8a562e7634774d3e3a36698ac4915e37fc84a2cd0044cb84fa5d80263d2af4f6",
      "size":"0x41c7",
      "stateRoot":"0xf5208fffa2ba5a3f3a2f64ebd5ca3d098978bedd75f335f56b705d8715ee2305",
      "timestamp":"0x5b541449",
      "totalDifficulty":"0x12ac11391a2f3872fcd",
      "transactions":[
         {
            "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
            "blockNumber":"0x5bad55",
            "chainId":"0x1",
            "from":"0xfbb1b73c4f0bda4f67dca266ce6ef42f520fbb98",
            "gas":"0x249f0",
            "gasPrice":"0x174876e800",
            "hash":"0x8784d99762bccd03b2086eabccee0d77f14d05463281e121a62abfebcf0d2d5f",
            "input":"0x6ea056a9000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bd8d7fa6f8cc00",
            "nonce":"0x5e4724",
            "r":"0xd1556332df97e3bd911068651cfad6f975a30381f4ff3a55df7ab3512c78b9ec",
            "s":"0x66b51cbb10cd1b2a09aaff137d9f6d4255bf73cb7702b666ebd5af502ffa4410",
            "to":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
            "transactionIndex":"0x0",
            "type":"0x0",
            "v":"0x25",
            "value":"0x0"
         },
         {
            "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
            "blockNumber":"0x5bad55",
            "chainId":"0x1",
            "from":"0xc837f51a0efa33f8eca03570e3d01a4b2cf97ffd",
            "gas":"0x15f90",
            "gasPrice":"0x14b8d03a00",
            "hash":"0x311be6a9b58748717ac0f70eb801d29973661aaf1365960d159e4ec4f4aa2d7f",
            "input":"0x",
            "nonce":"0x4241",
            "r":"0xe9ef2f6fcff76e45fac6c2e8080094370082cfb47e8fde0709312f9aa3ec06ad",
            "s":"0x421ebc4ebe187c173f13b1479986dcbff5c4997c0dfeb1fd149a982ad4bcdfe7",
            "to":"0xf49bd0367d830850456d2259da366a054038dc46",
            "transactionIndex":"0x1",
            "type":"0x0",
            "v":"0x25",
            "value":"0x1bafa9ee16e78000"
         },
         {
            "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
            "blockNumber":"0x5bad55",
            "chainId":"0x1",
            "from":"0x532a2bae845abe7e5115808b832d34f9c3d41eed",
            "gas":"0x910c",
            "gasPrice":"0xe6f7cec00",
            "hash":"0xe42b0256058b7cad8a14b136a0364acda0b4c36f5b02dea7e69bfd82cef252a2",
            "input":"0xa9059cbb000000000000000000000000398a58b2e3790431fdac1ea56017e65401fa998800000000000000000000000000000000000000000007bcadb57b861109080000",
            "nonce":"0x0",
            "r":"0x4e3fdc1ad7ac52439791a8a48bc8ed70040170fa9c4b6cef6317f63d45e9a142",
            "s":"0x6e5feaefdbc8f99c5d036b31d6386fb49c1a97812f13d48742a1b77b7e690858",
            "to":"0x818fc6c2ec5986bc6e2cbf00939d90556ab12ce5",
            "transactionIndex":"0x2",
            "type":"0x0",
            "v":"0x26",
            "value":"0x0"
         },
         {
            "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
            "blockNumber":"0x5bad55",
            "from":"0x2a9847093ad514639e8cdec960b5e51686960291",
            "gas":"0x4f588",
            "gasPrice":"0xc22a75840",
            "hash":"0x4eb05376055c6456ed883fc843bc43df1dcf739c321ba431d518aecd7f98ca11",
            "input":"0x000101fa27134d5320",
            "nonce":"0xd50",
            "r":"0x980e463d70e67c49477883a55cdb42829c9e5746e95d63b738d7390c7d685551",
            "s":"0x647babbe3a96df447da960812c88833c6b5aa009f1c361c6adae818100d15007",
            "to":"0xc7ed8919c70dd8ccf1a57c0ed75b25ceb2dd22d1",
            "transactionIndex":"0x3",
            "type":"0x0",
            "v":"0x1b",
            "value":"0x0"
         },
         {
            "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
            "blockNumber":"0x5bad55",
            "chainId":"0x1",
            "from":"0xe12c32af0ca83fe12c58b1daef82ebe6333f7b10",
            "gas":"0x5208",
            "gasPrice":"0xba43b7400",
            "hash":"0x994dd9e72b212b7dc5fd0466ab75adf7d391cf4f206a65b7ad2a1fd032bb06d7",
            "input":"0x",
            "nonce":"0x1d",
            "r":"0xedf9e958bbd3f7d2fd9831678a3166cf0373a4436a63c152c4aa84f864bb7e6e",
            "s":"0xacbc0cfcc7d3264de55c0af45b6c280ea237e28d273f7ddba0eea05204c2101",
            "to":"0x5343222c6f7e2af4d9d3e844fb8f3f18f7be0e55",
            "transactionIndex":"0x4",
            "type":"0x0",
            "v":"0x26",
            "value":"0x31f64d59c01f6000"
         },
         {
            "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
            "blockNumber":"0x5bad55",
            "chainId":"0x1",
            "from":"0x80c779504c3a3a39dbd0356f5d8e851cb6dbba0a",
            "gas":"0x57e40",
            "gasPrice":"0x9c35a3cc8",
            "hash":"0xf6feecbb9ab0ac58591a4bc287059b1133089c499517e91a274e6a1f5e7dce53",
            "input":"0x010b01000d0670",
            "nonce":"0xb9bf",
            "r":"0x349d0601e24f0128ecfce3665edd2a0727a043fa62ccf587fded784aed46c3f6",
            "s":"0x77127ddf76cb2b9e12074006a504fbd6893d5bd29a18a8efb193907f4565404",
            "to":"0x3714e5671be406fc1920351984f4429237831477",
            "transactionIndex":"0x5",
            "type":"0x0",
            "v":"0x26",
            "value":"0x0"
         },
         {
            "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
            "blockNumber":"0x5bad55",
            "chainId":"0x1",
            "from":"0xc5b373618d4d01a38f822f56ca6d2ff5080cc4f2",
            "gas":"0x4f588",
            "gasPrice":"0x9c355a8e8",
            "hash":"0x7e537d687a5525259480440c6ea2e1a8469cd98906eaff8597f3d2a44422ff97",
            "input":"0x0108e9000d0670136b",
            "nonce":"0x109f3",
            "r":"0x7736e11b03c6702eb6aaea8c45ed6b8a510878bb7741028d82938b9207448e9b",
            "s":"0x70bcd4c0ec2b0c67eb9eefb53a6ff7e114b45888589a5aaf4c1a1f00fa704775",
            "to":"0xc5f60fa4613493931b605b6da1e9febbdeb61e16",
            "transactionIndex":"0x6",
            "type":"0x0",
            "v":"0x25",
            "value":"0x0"
         }
      ],
      "transactionsRoot":"0xf98631e290e88f58a46b7032f025969039aa9b5696498efc76baf436fa69b262",
      "uncles":[
         "0x824cce7c7c2ec6874b9fa9a9a898eb5f27cbaf3991dfa81084c3af60d1db618c"
      ]
   }
}`
	testTraces = `{
    "jsonrpc": "2.0",
    "id": 1,
    "result": [
        {
            "result": {
                "from": "0xa5ba45f484bc67fe293cf01f7d92d5ba3514dd42",
                "gas": "0x5208",
                "gasUsed": "0x5208",
                "input": "0x",
                "to": "0x45a318273749d6eb00f5f6ca3bc7cd3de26d642a",
                "type": "CALL",
                "value": "0x2ca186f5fda8004"
            }
        },
        {
            "result": {
                "from": "0x25f2650cc9e8ad863bf5da6a7598e24271574e29",
                "gas": "0xfe0e",
                "gasUsed": "0xafee",
                "input": "0xd0e30db0",
                "to": "0xe5d7c2a44ffddf6b295a15c148167daaaf5cf34f",
                "type": "CALL",
                "value": "0x2386f26fc10000"
            }
        }
    ]
}`
	testReceipts = `{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    {
      "blockHash": "0x19514ce955c65e4dd2cd41f435a75a46a08535b8fc16bc660f8092b32590b182",
      "blockNumber": "0x6f55",
      "contractAddress": null,
      "cumulativeGasUsed": "0x18c36",
      "from": "0x22896bfc68814bfd855b1a167255ee497006e730",
      "gasUsed": "0x18c36",
      "effectiveGasPrice": "0x9502f907",
      "logs": [
        {
          "address": "0xfd584430cafa2f451b4e2ebcf3986a21fff04350",
          "topics": [
            "0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d",
            "0x4be29e0e4eb91f98f709d98803cba271592782e293b84a625e025cbb40197ba8",
            "0x000000000000000000000000835281a2563db4ebf1b626172e085dc406bfc7d2",
            "0x00000000000000000000000022896bfc68814bfd855b1a167255ee497006e730"
          ],
          "data": "0x",
          "blockNumber": "0x6f55",
          "transactionHash": "0x4a481e4649da999d92db0585c36cba94c18a33747e95dc235330e6c737c6f975",
          "transactionIndex": "0x0",
          "blockHash": "0x19514ce955c65e4dd2cd41f435a75a46a08535b8fc16bc660f8092b32590b182",
          "logIndex": "0x0",
          "removed": false
        }
      ],
      "logsBloom": "0x00000004000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000080020000000000000200010000000000000000000001000000800000000000000000000000000000000000000000000000000000100100000000000000000000008000000000000000000000000000000002000000000000000000000",
      "status": "0x1",
      "to": "0xfd584430cafa2f451b4e2ebcf3986a21fff04350",
      "transactionHash": "0x4a481e4649da999d92db0585c36cba94c18a33747e95dc235330e6c737c6f975",
      "transactionIndex": "0x0",
      "type": "0x0"
    },
    {
      "blockHash": "0x19514ce955c65e4dd2cd41f435a75a46a08535b8fc16bc660f8092b32590b182",
      "blockNumber": "0x6f55",
      "contractAddress": null,
      "cumulativeGasUsed": "0x1de3e",
      "from": "0x712e3a792c974b3e3dbe41229ad4290791c75a82",
      "gasUsed": "0x5208",
      "effectiveGasPrice": "0x9502f907",
      "logs": [],
      "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
      "status": "0x1",
      "to": "0xd42e2b1c14d02f1df5369a9827cb8e6f3f75f338",
      "transactionHash": "0xefb83b4e3f1c317e8da0f8e2fbb2fe964f34ee184466032aeecac79f20eacaf6",
      "transactionIndex": "0x1",
      "type": "0x2"
    }
  ]
}`
	testUncles = `[{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "difficulty": "0x57f117f5c",
    "extraData": "0x476574682f76312e302e302f77696e646f77732f676f312e342e32",
    "gasLimit": "0x1388",
    "gasUsed": "0x0",
    "hash": "0x932bdf904546a2287a2c9b2ede37925f698a7657484b172d4e5184f80bdd464d",
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "miner": "0x5bf5e9cf9b456d6591073513de7fd69a9bef04bc",
    "mixHash": "0x4500aa4ee2b3044a155252e35273770edeb2ab6f8cb19ca8e732771484462169",
    "nonce": "0x24732773618192ac",
    "number": "0x299",
    "parentHash": "0xa779859b1ee558258b7008bbabff272280136c5dd3eb3ea3bfa8f6ae03bf91e5",
    "receiptsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
    "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
    "size": "0x21d",
    "stateRoot": "0x2604fbf5183f5360da249b51f1b9f1e0f315d2ff3ffa1a4143ff221ad9ca1fec",
    "timestamp": "0x55ba4827",
    "totalDifficulty": null,
    "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
    "uncles": []
  }
}]`

	testTwoUnclesBlockNumber = 141
	testTwoUnclesBlock       = `{
   "jsonrpc":"2.0",
   "id":0,
   "result":{
      "difficulty":"0x4417decf7",
      "extraData":"0x426974636f696e2069732054484520426c6f636b636861696e2e",
      "gasLimit":"0x1388",
      "gasUsed":"0x0",
      "hash":"0xeafbe76fdcadc1b69ba248589eb2a674b60b00c84374c149c9deaf5596183932",
      "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
      "miner":"0x1b7047b4338acf65be94c1a3e8c5c9338ad7d67c",
      "mixHash":"0x21eabda67c3151855389a5a968e50daa7b356b3046e2f119ef46c97d204a541e",
      "nonce":"0x85378a3fc5e608e1",
      "number":"0x8d",
      "parentHash":"0xe2c1e8200ef2e9fba09979f0b504dc52c068719623c7064904c7bd3e9365acc1",
      "receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
      "sha3Uncles":"0x393f5f01182846b91386f8b00759fd54f83998a6a1064b8ac72fc8eca1bcf81b",
      "size":"0x653",
      "stateRoot":"0x3e1eea9a01178945535230b6f5839201f594d9be20618bb4edaa383f4f0c850f",
      "timestamp":"0x55ba4444",
      "totalDifficulty":"0x24826e73469",
      "transactions":[
         
      ],
      "transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
      "uncles":[
         "0x61beeeb3e11e89d19fed2e988c8017b55c3ddb8895f531072363ce2abaf56b95",
         "0xf84d9d74415364c3a7569f315ff831b910968c7dd637fffaab51278c9e7f9306"
      ]
   }
}`
	testTwoUnclesBlockUncles = `[
   {
      "jsonrpc":"2.0",
      "id":141,
      "result":{
         "difficulty":"0x4406dc086",
         "extraData":"0x476574682f4c5649562f76312e302e302f6c696e75782f676f312e342e32",
         "gasLimit":"0x1388",
         "gasUsed":"0x0",
         "hash":"0x61beeeb3e11e89d19fed2e988c8017b55c3ddb8895f531072363ce2abaf56b95",
         "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
         "miner":"0xbb7b8287f3f0a933474a79eae42cbca977791171",
         "mixHash":"0x87547a998fe63f18b36180ca918131b6b20fc5d67390e2ac2f66be3fee8fb7d2",
         "nonce":"0x1dc5b79704350bee",
         "number":"0x8b",
         "parentHash":"0x2253b8f79c23b6ff67cb2ef6fabd9ec59e1edf2d07c16d98a19378041f96624d",
         "receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
         "sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
         "size":"0x21f",
         "stateRoot":"0x940131b162b07452ea31b5335c4dedfdddc13338142f71f261d51dea664033b4",
         "timestamp":"0x55ba4441",
         "totalDifficulty":"0x24826e73469",
         "transactions":[
            
         ],
         "transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
         "uncles":[
            
         ]
      }
   },
   {
      "jsonrpc":"2.0",
      "id":141,
      "result":{
         "difficulty":"0x4406dc086",
         "extraData":"0x476574682f6b6c6f737572652f76312e302e302d66633739643332642f6c696e",
         "gasLimit":"0x1388",
         "gasUsed":"0x0",
         "hash":"0xf84d9d74415364c3a7569f315ff831b910968c7dd637fffaab51278c9e7f9306",
         "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
         "miner":"0xd7e30ae310c1d1800f5b641baa7af95b2e1fd98c",
         "mixHash":"0x6039f236ebb70ec71091df5770aef0f0faa13ef334c4c68daaffbfdf7961a3d3",
         "nonce":"0x7d8ec05d330e6e99",
         "number":"0x8b",
         "parentHash":"0x2253b8f79c23b6ff67cb2ef6fabd9ec59e1edf2d07c16d98a19378041f96624d",
         "receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
         "sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
         "size":"0x221",
         "stateRoot":"0x302bb7708752013f46f009dec61cad586c35dc185d20cdde0071b7487f7c2008",
         "timestamp":"0x55ba4440",
         "totalDifficulty":"0x24826e73469",
         "transactions":[
            
         ],
         "transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
         "uncles":[
            
         ]
      }
   }
]`
)
