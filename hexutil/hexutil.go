// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package hexutil

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Bytes marshals/unmarshals as a JSON string with 0x prefix.
// The empty slice marshals as "0x".
type Bytes []byte

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bytes) UnmarshalJSON(input []byte) error {
	var v string
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	v = strings.Replace(v, "0x", "", 1)
	if len(v)%2 != 0 {
		v += "0"
	}
	var err error
	*b, err = hex.DecodeString(v)
	return err
}

func (b *Bytes) String() string {
	return fmt.Sprintf("0x%x", *b)
}

// Big unmarshals as a JSON string with 0x prefix.
type Big big.Int

// UnmarshalJSON implements json.Unmarshaler.
func (b *Big) UnmarshalJSON(input []byte) error {
	var v string
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	v = strings.Replace(v, "0x", "", 1)
	if len(v)%2 != 0 {
		v += "0"
	}

	ret, ok := new(big.Int).SetString(v, 16)
	if !ok {
		return fmt.Errorf("error decoding %s to big int", v)
	}
	*b = (Big)(*ret)
	return nil
}

// ToInt converts b to a big.Int.
func (b *Big) ToInt() *big.Int {
	return (*big.Int)(b)
}

// Uint64 unmarshals as a JSON string with 0x prefix.
// The zero value marshals as "0x0".
type Uint64 uint64

// UnmarshalJSON implements json.Unmarshaler.
func (b *Uint64) UnmarshalJSON(input []byte) error {
	var v string
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	v = strings.Replace(v, "0x", "", 1)

	ret, err := strconv.ParseUint(v, 16, 64)
	if err != nil {
		return err
	}
	*b = Uint64(ret)
	return err
}
