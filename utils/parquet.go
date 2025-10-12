package utils

import "math/big"

func BigIntFromParquetBytes(b []byte) *big.Int {
	// reverse to big-endian for SetBytes
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	z := new(big.Int).SetBytes(b)
	return z
}
