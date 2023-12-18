package types

import (
	"time"

	"eth2-exporter/hexutil"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

type GetBlockTimings struct {
	Headers  time.Duration
	Receipts time.Duration
	Traces   time.Duration
}

type BulkMutations struct {
	Keys []string
	Muts []*gcp_bigtable.Mutation
}

type Eth1RpcGetBlockResponse struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  struct {
		BaseFeePerGas   hexutil.Big    `json:"baseFeePerGas"`
		Difficulty      hexutil.Big    `json:"difficulty"`
		ExtraData       hexutil.Bytes  `json:"extraData"`
		GasLimit        hexutil.Uint64 `json:"gasLimit"`
		GasUsed         hexutil.Uint64 `json:"gasUsed"`
		Hash            hexutil.Bytes  `json:"hash"`
		LogsBloom       hexutil.Bytes  `json:"logsBloom"`
		Miner           hexutil.Bytes  `json:"miner"`
		MixHash         hexutil.Bytes  `json:"mixHash"`
		Nonce           hexutil.Uint64 `json:"nonce"`
		Number          hexutil.Uint64 `json:"number"`
		ParentHash      hexutil.Bytes  `json:"parentHash"`
		ReceiptsRoot    hexutil.Bytes  `json:"receiptsRoot"`
		Sha3Uncles      hexutil.Bytes  `json:"sha3Uncles"`
		Size            hexutil.Uint64 `json:"size"`
		StateRoot       hexutil.Bytes  `json:"stateRoot"`
		Timestamp       hexutil.Uint64 `json:"timestamp"`
		TotalDifficulty hexutil.Big    `json:"totalDifficulty"`
		Transactions    []struct {
			BlockHash            hexutil.Bytes  `json:"blockHash"`
			BlockNumber          hexutil.Uint64 `json:"blockNumber"`
			From                 hexutil.Bytes  `json:"from"`
			Gas                  hexutil.Uint64 `json:"gas"`
			GasPrice             hexutil.Big    `json:"gasPrice"`
			Hash                 hexutil.Bytes  `json:"hash"`
			Input                hexutil.Bytes  `json:"input"`
			Nonce                hexutil.Uint64 `json:"nonce"`
			To                   hexutil.Bytes  `json:"to"`
			TransactionIndex     hexutil.Uint64 `json:"transactionIndex"`
			Value                hexutil.Big    `json:"value"`
			Type                 hexutil.Uint64 `json:"type"`
			V                    hexutil.Bytes  `json:"v"`
			R                    hexutil.Bytes  `json:"r"`
			S                    hexutil.Bytes  `json:"s"`
			ChainID              hexutil.Uint64 `json:"chainId"`
			MaxFeePerGas         hexutil.Big    `json:"maxFeePerGas"`
			MaxPriorityFeePerGas hexutil.Big    `json:"maxPriorityFeePerGas"`

			AccessList []struct {
				Address     hexutil.Bytes   `json:"address"`
				StorageKeys []hexutil.Bytes `json:"storageKeys"`
			} `json:"accessList"`

			// Optimism specific fields
			YParity    hexutil.Bytes `json:"yParity"`
			Mint       hexutil.Bytes `json:"mint"`       // The ETH value to mint on L2.
			SourceHash hexutil.Bytes `json:"sourceHash"` // the source-hash, uniquely identifies the origin of the deposit.

			// Arbitrum specific fields
			RequestId string `json:"requestId"` // On L1 to L2 transactions, this field is added to indicate position in the Inbox queue

		} `json:"transactions"`
		TransactionsRoot hexutil.Bytes `json:"transactionsRoot"`

		Withdrawals []struct {
			Index          hexutil.Uint64 `json:"index"`
			ValidatorIndex hexutil.Uint64 `json:"validatorIndex"`
			Address        hexutil.Bytes  `json:"address"`
			Amount         hexutil.Big    `json:"amount"`
		} `json:"withdrawals"`
		WithdrawalsRoot hexutil.Bytes `json:"withdrawalsRoot"`

		Uncles []hexutil.Bytes `json:"uncles"`

		// Optimism specific fields

		// Arbitrum specific fields
		L1BlockNumber string `json:"l1BlockNumber"` // An approximate L1 block number that occurred before this L2 block.
		SendCount     string `json:"sendCount"`     // The number of L2 to L1 messages since Nitro genesis
		SendRoot      string `json:"sendRoot"`      // The Merkle root of the outbox tree state
	} `json:"result"`

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Eth1RpcGetBlockReceiptsResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  []struct {
		BlockHash         string `json:"blockHash"`
		BlockNumber       string `json:"blockNumber"`
		ContractAddress   any    `json:"contractAddress"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"`
		EffectiveGasPrice string `json:"effectiveGasPrice"`
		From              string `json:"from"`
		GasUsed           string `json:"gasUsed"`
		Logs              []struct {
			Address          string   `json:"address"`
			Topics           []string `json:"topics"`
			Data             string   `json:"data"`
			BlockNumber      string   `json:"blockNumber"`
			TransactionHash  string   `json:"transactionHash"`
			TransactionIndex string   `json:"transactionIndex"`
			BlockHash        string   `json:"blockHash"`
			LogIndex         string   `json:"logIndex"`
			Removed          bool     `json:"removed"`
		} `json:"logs"`
		LogsBloom        string `json:"logsBloom"`
		Status           string `json:"status"`
		To               string `json:"to"`
		TransactionHash  string `json:"transactionHash"`
		TransactionIndex string `json:"transactionIndex"`
		Type             string `json:"type"`

		// Optimism specific fields
		DepositNonce string `json:"depositNonce"`
		L1Fee        string `json:"l1Fee"`       // The fee associated with a transaction on the Layer 1, it is calculated as l1GasPrice multiplied by l1GasUsed
		L1FeeScalar  string `json:"l1FeeScalar"` // A multiplier applied to the actual gas usage on Layer 1 to calculate the dynamic costs. If set to 1, it has no impact on the L1 gas usage
		L1GasPrice   string `json:"l1GasPrice"`  // The gas price for transactions on the Layer 1
		L1GasUsed    string `json:"l1GasUsed"`   // The amount of gas consumed by a transaction on the Layer 1

		// Arbitrum specific fields
		L1BlockNumber string `json:"l1BlockNumber"` // The L1 block number that would be used for block.number calls.
		GasUsedForL1  string `json:"gasUsedForL1"`  // The L1 block number that would be used for block.number calls.
	} `json:"result"`

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Eth1RpcTraceBlockResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`

	Result []struct {
		TxHash string           `json:"txHash"` // is empty on erigon generated geth traces!!!
		Result Eth1RpcTraceCall `json:"result"` // sub-calls
	} `json:"result"`

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Eth1RpcTraceCall struct {
	Calls        []Eth1RpcTraceCall `json:"calls"`        // subcalls invoked by this call
	Error        string             `json:"error"`        // error, if any
	From         string             `json:"from"`         // initiator of the call
	Gas          string             `json:"gas"`          // hex-encoded gas provided for call
	GasUsed      string             `json:"gasUsed"`      // hex-encoded gas used during call
	Input        string             `json:"input"`        // call data
	Output       string             `json:"output"`       // In case a frame reverts, the field output will contain the raw return data
	RevertReason string             `json:"revertReason"` // In case the top level frame reverts, its revertReason field will contain the parsed reason of revert as returned by the Solidity contract
	To           string             `json:"to"`           // Will contain the new contract address if type is CREATE or CREATE2, will contain the destination address for SELFDESTRUCT calls
	Type         string             `json:"type"`         // Can be either CALL, STATICCALL, DELEGATECALL, CREATE, CREATE2 or SELFDESTRUCT
	Value        string             `json:"value"`        // The ETH value involved in the trace
}
