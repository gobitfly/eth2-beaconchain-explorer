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
		BaseFeePerGas   hexutil.Bytes `json:"baseFeePerGas"`
		Difficulty      hexutil.Bytes `json:"difficulty"`
		ExtraData       hexutil.Bytes `json:"extraData"`
		GasLimit        hexutil.Bytes `json:"gasLimit"`
		GasUsed         hexutil.Bytes `json:"gasUsed"`
		Hash            hexutil.Bytes `json:"hash"`
		LogsBloom       hexutil.Bytes `json:"logsBloom"`
		Miner           hexutil.Bytes `json:"miner"`
		MixHash         hexutil.Bytes `json:"mixHash"`
		Nonce           hexutil.Bytes `json:"nonce"`
		Number          hexutil.Bytes `json:"number"`
		ParentHash      hexutil.Bytes `json:"parentHash"`
		ReceiptsRoot    hexutil.Bytes `json:"receiptsRoot"`
		Sha3Uncles      hexutil.Bytes `json:"sha3Uncles"`
		Size            hexutil.Bytes `json:"size"`
		StateRoot       hexutil.Bytes `json:"stateRoot"`
		Timestamp       hexutil.Bytes `json:"timestamp"`
		TotalDifficulty hexutil.Bytes `json:"totalDifficulty"`
		Transactions    []struct {
			BlockHash            hexutil.Bytes `json:"blockHash"`
			BlockNumber          hexutil.Bytes `json:"blockNumber"`
			From                 hexutil.Bytes `json:"from"`
			Gas                  hexutil.Bytes `json:"gas"`
			GasPrice             hexutil.Bytes `json:"gasPrice"`
			Hash                 hexutil.Bytes `json:"hash"`
			Input                hexutil.Bytes `json:"input"`
			Nonce                hexutil.Bytes `json:"nonce"`
			To                   hexutil.Bytes `json:"to"`
			TransactionIndex     hexutil.Bytes `json:"transactionIndex"`
			Value                hexutil.Bytes `json:"value"`
			Type                 hexutil.Bytes `json:"type"`
			V                    hexutil.Bytes `json:"v"`
			R                    hexutil.Bytes `json:"r"`
			S                    hexutil.Bytes `json:"s"`
			ChainID              hexutil.Bytes `json:"chainId"`
			MaxFeePerGas         hexutil.Bytes `json:"maxFeePerGas"`
			MaxPriorityFeePerGas hexutil.Bytes `json:"maxPriorityFeePerGas"`

			AccessList []struct {
				Address     hexutil.Bytes   `json:"address"`
				StorageKeys []hexutil.Bytes `json:"storageKeys"`
			} `json:"accessList"`

			// Optimism specific fields
			YParity    hexutil.Bytes `json:"yParity"`
			Mint       hexutil.Bytes `json:"mint"`       // The ETH value to mint on L2.
			SourceHash hexutil.Bytes `json:"sourceHash"` // the source-hash, uniquely identifies the origin of the deposit.

			// Arbitrum specific fields
			// Arbitrum Nitro
			RequestId           hexutil.Bytes `json:"requestId"`           // On L1 to L2 transactions, this field is added to indicate position in the Inbox queue
			RefundTo            hexutil.Bytes `json:"refundTo"`            //
			L1BaseFee           hexutil.Bytes `json:"l1BaseFee"`           //
			DepositValue        hexutil.Bytes `json:"depositValue"`        //
			RetryTo             hexutil.Bytes `json:"retryTo"`             // nil means contract creation
			RetryValue          hexutil.Bytes `json:"retryValue"`          // wei amount
			RetryData           hexutil.Bytes `json:"retryData"`           // contract invocation input data
			Beneficiary         hexutil.Bytes `json:"beneficiary"`         //
			MaxSubmissionFee    hexutil.Bytes `json:"maxSubmissionFee"`    //
			TicketId            hexutil.Bytes `json:"ticketId"`            //
			MaxRefund           hexutil.Bytes `json:"maxRefund"`           // the maximum refund sent to RefundTo (the rest goes to From)
			SubmissionFeeRefund hexutil.Bytes `json:"submissionFeeRefund"` // the submission fee to refund if successful (capped by MaxRefund)

			// Arbitrum Classic
			L1SequenceNumber hexutil.Bytes `json:"l1SequenceNumber"`
			ParentRequestId  hexutil.Bytes `json:"parentRequestId"`
			IndexInParent    hexutil.Bytes `json:"indexInParent"`
			ArbType          hexutil.Bytes `json:"arbType"`
			ArbSubType       hexutil.Bytes `json:"arbSubType"`
			L1BlockNumber    hexutil.Bytes `json:"l1BlockNumber"`
		} `json:"transactions"`
		TransactionsRoot hexutil.Bytes `json:"transactionsRoot"`

		Withdrawals []struct {
			Index          hexutil.Bytes `json:"index"`
			ValidatorIndex hexutil.Bytes `json:"validatorIndex"`
			Address        hexutil.Bytes `json:"address"`
			Amount         hexutil.Bytes `json:"amount"`
		} `json:"withdrawals"`
		WithdrawalsRoot hexutil.Bytes `json:"withdrawalsRoot"`

		Uncles []hexutil.Bytes `json:"uncles"`

		// Optimism specific fields

		// Arbitrum specific fields
		L1BlockNumber hexutil.Bytes `json:"l1BlockNumber"` // An approximate L1 block number that occurred before this L2 block.
		SendCount     hexutil.Bytes `json:"sendCount"`     // The number of L2 to L1 messages since Nitro genesis
		SendRoot      hexutil.Bytes `json:"sendRoot"`      // The Merkle root of the outbox tree state
	} `json:"result"`

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Eth1RpcGetBlockReceiptsResponse struct {
	Jsonrpc string                       `json:"jsonrpc"`
	ID      int                          `json:"id"`
	Result  []Eth1RpcGetBlockReceiptData `json:"result"`

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
type Eth1RpcGetBlockReceiptResponse struct {
	Jsonrpc string                     `json:"jsonrpc"`
	ID      int                        `json:"id"`
	Result  Eth1RpcGetBlockReceiptData `json:"result"`

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Eth1RpcGetBlockReceiptData struct {
	BlockHash         hexutil.Bytes `json:"blockHash"`
	BlockNumber       hexutil.Bytes `json:"blockNumber"`
	ContractAddress   hexutil.Bytes `json:"contractAddress"`
	CumulativeGasUsed hexutil.Bytes `json:"cumulativeGasUsed"`
	EffectiveGasPrice hexutil.Bytes `json:"effectiveGasPrice"`
	From              hexutil.Bytes `json:"from"`
	GasUsed           hexutil.Bytes `json:"gasUsed"`
	Logs              []struct {
		Address          hexutil.Bytes   `json:"address"`
		Topics           []hexutil.Bytes `json:"topics"`
		Data             hexutil.Bytes   `json:"data"`
		BlockNumber      hexutil.Bytes   `json:"blockNumber"`
		TransactionHash  hexutil.Bytes   `json:"transactionHash"`
		TransactionIndex hexutil.Bytes   `json:"transactionIndex"`
		BlockHash        hexutil.Bytes   `json:"blockHash"`
		LogIndex         hexutil.Bytes   `json:"logIndex"`
		Removed          bool            `json:"removed"`
	} `json:"logs"`
	LogsBloom        hexutil.Bytes `json:"logsBloom"`
	Status           hexutil.Bytes `json:"status"`
	To               hexutil.Bytes `json:"to"`
	TransactionHash  hexutil.Bytes `json:"transactionHash"`
	TransactionIndex hexutil.Bytes `json:"transactionIndex"`
	Type             hexutil.Bytes `json:"type"`

	// Optimism specific fields
	DepositNonce hexutil.Bytes `json:"depositNonce"`
	L1Fee        hexutil.Bytes `json:"l1Fee"`       // The fee associated with a transaction on the Layer 1, it is calculated as l1GasPrice multiplied by l1GasUsed
	L1FeeScalar  string        `json:"l1FeeScalar"` // A multiplier applied to the actual gas usage on Layer 1 to calculate the dynamic costs. If set to 1, it has no impact on the L1 gas usage
	L1GasPrice   hexutil.Bytes `json:"l1GasPrice"`  // The gas price for transactions on the Layer 1
	L1GasUsed    hexutil.Bytes `json:"l1GasUsed"`   // The amount of gas consumed by a transaction on the Layer 1

	// Arbitrum specific fields
	// Arbitrum Nitro
	L1BlockNumber hexutil.Bytes `json:"l1BlockNumber"` // The L1 block number that would be used for block.number calls.
	GasUsedForL1  hexutil.Bytes `json:"gasUsedForL1"`  // The L1 block number that would be used for block.number calls.

	// Arbitrum Classic
	ReturnCode hexutil.Bytes `json:"returnCode"`
	ReturnData hexutil.Bytes `json:"returnData"`
	FeeStats   struct {
		Prices struct {
			L1Transaction hexutil.Bytes `json:"l1Transaction"`
			L1Calldata    hexutil.Bytes `json:"l1Calldata"`
			L2Storage     hexutil.Bytes `json:"l2Storage"`
			L2Computation hexutil.Bytes `json:"l2Computation"`
		} `json:"prices"`
		UnitsUsed struct {
			L1Transaction hexutil.Bytes `json:"l1Transaction"`
			L1Calldata    hexutil.Bytes `json:"l1Calldata"`
			L2Storage     hexutil.Bytes `json:"l2Storage"`
			L2Computation hexutil.Bytes `json:"l2Computation"`
		} `json:"unitsUsed"`
		Paid struct {
			L1Transaction hexutil.Bytes `json:"l1Transaction"`
			L1Calldata    hexutil.Bytes `json:"l1Calldata"`
			L2Storage     hexutil.Bytes `json:"l2Storage"`
			L2Computation hexutil.Bytes `json:"l2Computation"`
		} `json:"paid"`
	} `json:"feeStats"`
	L1InboxBatchInfo hexutil.Bytes `json:"l1InboxBatchInfo"`
}

type Eth1RpcTraceBlockResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`

	Result []struct {
		TxHash hexutil.Bytes    `json:"txHash"` // is empty on erigon generated geth traces!!!
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
	From         hexutil.Bytes      `json:"from"`         // initiator of the call
	Gas          hexutil.Bytes      `json:"gas"`          // hex-encoded gas provided for call
	GasUsed      hexutil.Bytes      `json:"gasUsed"`      // hex-encoded gas used during call
	Input        hexutil.Bytes      `json:"input"`        // call data
	Output       hexutil.Bytes      `json:"output"`       // In case a frame reverts, the field output will contain the raw return data
	RevertReason string             `json:"revertReason"` // In case the top level frame reverts, its revertReason field will contain the parsed reason of revert as returned by the Solidity contract
	To           hexutil.Bytes      `json:"to"`           // Will contain the new contract address if type is CREATE or CREATE2, will contain the destination address for SELFDESTRUCT calls
	Type         string             `json:"type"`         // Can be either CALL, STATICCALL, DELEGATECALL, CREATE, CREATE2 or SELFDESTRUCT
	Value        hexutil.Bytes      `json:"value"`        // The ETH value involved in the trace

	// Optimism specific fields
	Time string `json:"time"`

	// Arbitrum specific fields
	BeforeEVMTransfers []struct {
		Purpose string        `json:"purpose"`
		From    hexutil.Bytes `json:"from"`
		To      hexutil.Bytes `json:"to"`
		Value   hexutil.Bytes `json:"value"`
	} `json:"beforeEVMTransfers"`

	AfterEVMTransfers []struct {
		Purpose string        `json:"purpose"`
		From    hexutil.Bytes `json:"from"`
		To      hexutil.Bytes `json:"to"`
		Value   hexutil.Bytes `json:"value"`
	} `json:"afterEVMTransfers"`
}

type Eth1RpcDebugTraceBlockResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  []struct {
		Action struct {
			CallType string        `json:"callType"`
			From     hexutil.Bytes `json:"from"`
			Gas      hexutil.Bytes `json:"gas"`
			Input    hexutil.Bytes `json:"input"`
			Init     hexutil.Bytes `json:"init"`
			To       hexutil.Bytes `json:"to"`
			Value    hexutil.Bytes `json:"value"`
		} `json:"action"`
		BlockHash   hexutil.Bytes `json:"blockHash"`
		BlockNumber uint64        `json:"blockNumber"`
		Result      struct {
			Address hexutil.Bytes `json:"address"`
			Code    hexutil.Bytes `json:"code"`
			GasUsed hexutil.Bytes `json:"gasUsed"`
			Output  hexutil.Bytes `json:"output"`
		} `json:"result"`
		Subtraces           int           `json:"subtraces"`
		TraceAddress        []uint64      `json:"traceAddress"`
		TransactionHash     hexutil.Bytes `json:"transactionHash"`
		TransactionPosition int           `json:"transactionPosition"`
		Type                string        `json:"type"`
	} `json:"result"`
}
