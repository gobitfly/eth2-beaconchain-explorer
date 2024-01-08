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

	// make sure to have an even length hex string by prefixing odd strings with a single 0, 0x0 will become 0x00 for example
	// while hashes and addresses have always an even length, numbers usually don't
	if len(v)%2 != 0 {
		v = "0" + v
	}

	var err error
	*b, err = hex.DecodeString(v)

	if err != nil {
		return fmt.Errorf("error decoding %s: %v", string(input), err)
	}
	return err
}

func (b *Bytes) String() string {
	return fmt.Sprintf("0x%x", *b)
}
