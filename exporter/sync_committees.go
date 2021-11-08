package exporter

import (
	"eth2-exporter/rpc"
	"time"
)

func syncCommitteesExporter(rpcClient rpc.Client) {
	for {
		exportSyncCommittee(rpcClient)
		time.Sleep(time.Second * 12)
	}
}

func exportSyncCommittee(rpcClient rpc.Client) error {
	//db.DB.Select(`select period from sync_committees`)
	return nil
}
