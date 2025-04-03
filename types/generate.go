package types

//go:generate protoc --go_out=. --go_opt=paths=source_relative ./eth1.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./machine.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./ens.proto
