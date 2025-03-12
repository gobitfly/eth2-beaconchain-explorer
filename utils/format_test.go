// Copyright (C) 2025 Bitfly GmbH
//
// This file is part of Beaconchain Dashboard.
//
// Beaconchain Dashboard is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Beaconchain Dashboard is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Beaconchain Dashboard.  If not, see <https://www.gnu.org/licenses/>.

package utils

import "testing"

func TestBeginningOfSetWithdrawalCredentials(t *testing.T) {
	tests := []struct {
		version  int
		expected string
	}{
		{0, "000000000000000000000000"},
		{1, "010000000000000000000000"},
		{2, "020000000000000000000000"},
		{10, "0a0000000000000000000000"},
	}
	for _, tt := range tests {
		v := BeginningOfSetWithdrawalCredentials(tt.version)
		if v != tt.expected {
			t.Errorf("wrong beginning of set withdrawal credentials for version %v: %v expected %v", tt.version, v, tt.expected)
		}
	}
}
