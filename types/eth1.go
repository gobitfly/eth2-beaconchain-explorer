package types

import (
	"time"

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
		BaseFeePerGas   string `json:"baseFeePerGas"`
		Difficulty      string `json:"difficulty"`
		ExtraData       string `json:"extraData"`
		GasLimit        string `json:"gasLimit"`
		GasUsed         string `json:"gasUsed"`
		Hash            string `json:"hash"`
		LogsBloom       string `json:"logsBloom"`
		Miner           string `json:"miner"`
		MixHash         string `json:"mixHash"`
		Nonce           string `json:"nonce"`
		Number          string `json:"number"`
		ParentHash      string `json:"parentHash"`
		ReceiptsRoot    string `json:"receiptsRoot"`
		Sha3Uncles      string `json:"sha3Uncles"`
		Size            string `json:"size"`
		StateRoot       string `json:"stateRoot"`
		Timestamp       string `json:"timestamp"`
		TotalDifficulty string `json:"totalDifficulty"`
		Transactions    []struct {
			BlockHash            string `json:"blockHash"`
			BlockNumber          string `json:"blockNumber"`
			From                 string `json:"from"`
			Gas                  string `json:"gas"`
			GasPrice             string `json:"gasPrice"`
			Hash                 string `json:"hash"`
			Input                string `json:"input"`
			Nonce                string `json:"nonce"`
			To                   string `json:"to"`
			TransactionIndex     string `json:"transactionIndex"`
			Value                string `json:"value"`
			Type                 string `json:"type"`
			V                    string `json:"v"`
			R                    string `json:"r"`
			S                    string `json:"s"`
			ChainID              string `json:"chainId"`
			MaxFeePerGas         string `json:"maxFeePerGas"`
			MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`

			AccessList []struct {
				Address     string   `json:"address"`
				StorageKeys []string `json:"storageKeys"`
			} `json:"accessList"`

			// Optimism specific fields
			YParity    string `json:"yParity"`
			Mint       string `json:"mint"`       // The ETH value to mint on L2.
			SourceHash string `json:"sourceHash"` // the source-hash, uniquely identifies the origin of the deposit.

			// Arbitrum specific fields
			RequestId string `json:"requestId"` // On L1 to L2 transactions, this field is added to indicate position in the Inbox queue

		} `json:"transactions"`
		TransactionsRoot string `json:"transactionsRoot"`

		Withdrawals []struct {
			Index          string `json:"index"`
			ValidatorIndex string `json:"validatorIndex"`
			Address        string `json:"address"`
			Amount         string `json:"amount"`
		} `json:"withdrawals"`
		WithdrawalsRoot string `json:"withdrawalsRoot"`

		Uncles []string `json:"uncles"`

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
