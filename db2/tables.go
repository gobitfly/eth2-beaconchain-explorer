package db2

const BlocksRawTable = "blocks-raw"

const BT_COLUMNFAMILY_BLOCK = "b"
const BT_COLUMN_BLOCK = "b"
const BT_COLUMNFAMILY_RECEIPTS = "r"
const BT_COLUMN_RECEIPTS = "r"
const BT_COLUMNFAMILY_TRACES = "t"
const BT_COLUMN_TRACES = "t"
const BT_COLUMNFAMILY_UNCLES = "u"
const BT_COLUMN_UNCLES = "u"

const MAX_EL_BLOCK_NUMBER = int64(1_000_000_000_000 - 1)

var raw = map[string][]string{
	BlocksRawTable: {
		BT_COLUMNFAMILY_BLOCK,
		BT_COLUMNFAMILY_RECEIPTS,
		BT_COLUMNFAMILY_TRACES,
		BT_COLUMNFAMILY_UNCLES,
	},
}
