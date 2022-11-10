package db

import (
	"context"
	"database/sql"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

func GetStatsValidatorAll(limit, offset uint64, fromDay, toDay int) (*sql.Rows, error) {
	row, err := FrontendReaderDB.Query(
		"SELECT t.* FROM (SELECT client_build, exporter_version, created_trunc, user_id, client_name, client_version, cpu_process_seconds_total, machine, memory_process_bytes, sync_eth2_fallback_configured, sync_eth2_fallback_connected, ts as timestamp, validator_active, validator_total FROM stats_add_validator LEFT JOIN stats_process ON stats_add_validator.general_id = stats_process.id "+
			" LEFT JOIN stats_meta_p on stats_process.meta_id = stats_meta_p.id "+
			"WHERE process = 'validator' AND day >= $3 AND day <= $4 ORDER BY stats_meta_p.id desc LIMIT $1 OFFSET $2) t",
		limit, offset, fromDay, toDay,
	)
	return row, err
}

func GetStatsNodeAll(limit, offset uint64, fromDay, toDay int) (*sql.Rows, error) {
	row, err := FrontendReaderDB.Query(
		"SELECT t.* FROM (SELECT client_build, exporter_version, created_trunc, user_id, client_name, client_version, cpu_process_seconds_total, machine, memory_process_bytes, sync_eth1_fallback_configured, sync_eth1_fallback_connected, sync_eth2_fallback_configured, sync_eth2_fallback_connected, ts as timestamp, disk_beaconchain_bytes_total, network_libp2p_bytes_total_receive, network_libp2p_bytes_total_transmit, network_peers_connected, sync_eth1_connected, sync_eth2_synced, sync_beacon_head_slot FROM stats_add_beaconnode left join stats_process on stats_process.id = stats_add_beaconnode.general_id "+
			" LEFT JOIN stats_meta_p on stats_process.meta_id = stats_meta_p.id "+
			"WHERE process = 'beaconnode' AND day >= $3 AND day <= $4 ORDER BY stats_meta_p.id desc LIMIT $1 OFFSET $2) t",
		limit, offset, fromDay, toDay,
	)
	return row, err
}

func GetStatsSystemAll(limit, offset uint64, fromDay, toDay int) (*sql.Rows, error) {
	row, err := FrontendReaderDB.Query(
		"SELECT t.* FROM (SELECT exporter_version,created_trunc, user_id, cpu_cores, cpu_threads, cpu_node_system_seconds_total, cpu_node_user_seconds_total, cpu_node_iowait_seconds_total, cpu_node_idle_seconds_total, memory_node_bytes_total, memory_node_bytes_free, memory_node_bytes_cached, memory_node_bytes_buffers, disk_node_bytes_total, disk_node_bytes_free, disk_node_io_seconds, disk_node_reads_total, disk_node_writes_total, network_node_bytes_total_receive, network_node_bytes_total_transmit, misc_os, misc_node_boot_ts_seconds, ts as timestamp, machine from stats_system"+
			" LEFT JOIN stats_meta_p on stats_system.meta_id = stats_meta_p.id "+
			"WHERE process = 'system' AND day >= $3 AND day <= $4 ORDER BY stats_meta_p.id desc LIMIT $1 OFFSET $2) t",
		limit, offset, fromDay, toDay,
	)
	return row, err
}

func (bigtable *Bigtable) MigrateMachineStatsFromDBToBigtable(fromDay, toDay int, batchSize int, sleep int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*5)
	defer cancel()

	limit := uint64(batchSize)
	var err error

	logrus.Infof("Migrating Machine Data from day [%v, %v]", fromDay, toDay)

	logrus.Infof("Migrating system data")
	err = bigtable.migrateSystem(ctx, limit, fromDay, toDay, sleep)
	if err != nil {
		return err
	}
	logrus.Infof("System data migration completed")

	logrus.Infof("Migrating node data")
	err = bigtable.migrateNode(ctx, limit, fromDay, toDay, sleep)
	if err != nil {
		return err
	}
	logrus.Infof("Node data migration completed")

	logrus.Infof("Migrating validator data")
	err = bigtable.migrateValidator(ctx, limit, fromDay, toDay, sleep)
	if err != nil {
		return err
	}
	logrus.Infof("Validator data migration completed")

	logrus.Infof("\n\n=== Data migration completed ===")
	return nil
}

func (bigtable *Bigtable) migrateSystem(ctx context.Context, limit uint64, fromDay, toDay int, sleep int) error {
	for offset := uint64(0); ; offset += limit {
		rows, err := GetStatsSystemAll(limit, offset, fromDay, toDay)
		if err != nil {
			return err
		}

		var rowKeys []string
		var mutations []*gcp_bigtable.Mutation
		dataSystem, err := utils.SqlRowsToJSON(rows)
		if len(dataSystem) == 0 {
			logrus.Infof("system done, break loop")
			break // done
		}
		for _, it := range dataSystem {
			mapData := it.(map[string]interface{})
			createdTs := mapData["created_trunc"].(int64) * 1000000

			rowKeyData := fmt.Sprintf("u:%s:p:%s:m:%v", reversePaddedUserID(uint64(mapData["user_id"].(int64))), "system", mapData["machine"].(string))

			obj := types.MachineMetricSystem{
				Timestamp:       uint64(mapData["timestamp"].(int64)),
				ExporterVersion: mapData["exporter_version"].(string),
				// system
				CpuCores:                      uint64(mapData["cpu_cores"].(int64)),
				CpuThreads:                    uint64(mapData["cpu_threads"].(int64)),
				CpuNodeSystemSecondsTotal:     uint64(mapData["cpu_node_system_seconds_total"].(int64)),
				CpuNodeUserSecondsTotal:       uint64(mapData["cpu_node_user_seconds_total"].(int64)),
				CpuNodeIowaitSecondsTotal:     uint64(mapData["cpu_node_iowait_seconds_total"].(int64)),
				CpuNodeIdleSecondsTotal:       uint64(mapData["cpu_node_idle_seconds_total"].(int64)),
				MemoryNodeBytesTotal:          uint64(mapData["memory_node_bytes_total"].(int64)),
				MemoryNodeBytesFree:           uint64(mapData["memory_node_bytes_free"].(int64)),
				MemoryNodeBytesCached:         uint64(mapData["memory_node_bytes_cached"].(int64)),
				MemoryNodeBytesBuffers:        uint64(mapData["memory_node_bytes_buffers"].(int64)),
				DiskNodeBytesTotal:            uint64(mapData["disk_node_bytes_total"].(int64)),
				DiskNodeBytesFree:             uint64(mapData["disk_node_bytes_free"].(int64)),
				DiskNodeIoSeconds:             uint64(mapData["disk_node_io_seconds"].(int64)),
				DiskNodeReadsTotal:            uint64(mapData["disk_node_reads_total"].(int64)),
				DiskNodeWritesTotal:           uint64(mapData["disk_node_writes_total"].(int64)),
				NetworkNodeBytesTotalReceive:  uint64(mapData["network_node_bytes_total_receive"].(int64)),
				NetworkNodeBytesTotalTransmit: uint64(mapData["network_node_bytes_total_transmit"].(int64)),
				MiscNodeBootTsSeconds:         uint64(mapData["misc_node_boot_ts_seconds"].(int64)),
				MiscOs:                        mapData["misc_os"].(string),
			}
			data, err := proto.Marshal(&obj)
			if err != nil {
				return err
			}

			dataMut := gcp_bigtable.NewMutation()
			dataMut.Set(MACHINE_METRICS_COLUMN_FAMILY, "v1", gcp_bigtable.Timestamp(createdTs), data)

			rowKeys = append(rowKeys, rowKeyData)
			mutations = append(mutations, dataMut)
		}

		errInd, err := bigtable.tableMachineMetrics.ApplyBulk(
			ctx,
			rowKeys,
			mutations,
		)
		if err != nil {
			return err
		}
		if errInd != nil {
			logrus.Errorf("multiple inserts failed %v", errInd)
			return fmt.Errorf("multiple inserts failed %v", errInd)
		}
		logrus.Infof("Migrated system batch %v - %v", offset, offset+limit)
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	return nil
}

func (bigtable *Bigtable) migrateNode(ctx context.Context, limit uint64, fromDay, toDay int, sleep int) error {
	for offset := uint64(0); ; offset += limit {
		rows, err := GetStatsNodeAll(limit, offset, fromDay, toDay)
		if err != nil {
			return err
		}

		var rowKeys []string
		var mutations []*gcp_bigtable.Mutation
		dataNode, err := utils.SqlRowsToJSON(rows)
		if len(dataNode) == 0 {
			logrus.Infof("node done, break loop")
			break // done
		}
		for _, it := range dataNode {
			mapData := it.(map[string]interface{})
			createdTs := mapData["created_trunc"].(int64) * 1000000

			rowKeyData := fmt.Sprintf("u:%s:p:%s:m:%v", reversePaddedUserID(uint64(mapData["user_id"].(int64))), "beaconnode", mapData["machine"].(string))

			obj := types.MachineMetricNode{
				Timestamp:       uint64(mapData["timestamp"].(int64)),
				ExporterVersion: mapData["exporter_version"].(string),
				// process
				CpuProcessSecondsTotal:     uint64(mapData["cpu_process_seconds_total"].(int64)),
				MemoryProcessBytes:         uint64(mapData["memory_process_bytes"].(int64)),
				ClientName:                 mapData["client_name"].(string),
				ClientVersion:              mapData["client_version"].(string),
				ClientBuild:                uint64(mapData["client_build"].(int64)),
				SyncEth2FallbackConfigured: mapData["sync_eth2_fallback_configured"].(bool),
				SyncEth2FallbackConnected:  mapData["sync_eth2_fallback_connected"].(bool),
				// node
				DiskBeaconchainBytesTotal:       uint64(mapData["disk_beaconchain_bytes_total"].(int64)),
				NetworkLibp2PBytesTotalReceive:  uint64(mapData["network_libp2p_bytes_total_receive"].(int64)),
				NetworkLibp2PBytesTotalTransmit: uint64(mapData["network_libp2p_bytes_total_transmit"].(int64)),
				NetworkPeersConnected:           uint64(mapData["network_peers_connected"].(int64)),
				SyncEth1Connected:               mapData["sync_eth1_connected"].(bool),
				SyncEth2Synced:                  mapData["sync_eth2_synced"].(bool),
				SyncBeaconHeadSlot:              uint64(mapData["sync_beacon_head_slot"].(int64)),
				SyncEth1FallbackConfigured:      mapData["sync_eth1_fallback_configured"].(bool),
				SyncEth1FallbackConnected:       mapData["sync_eth1_fallback_connected"].(bool),
			}
			data, err := proto.Marshal(&obj)
			if err != nil {
				return err
			}

			dataMut := gcp_bigtable.NewMutation()
			dataMut.Set(MACHINE_METRICS_COLUMN_FAMILY, "v1", gcp_bigtable.Timestamp(createdTs), data)

			rowKeys = append(rowKeys, rowKeyData)
			mutations = append(mutations, dataMut)
		}

		errInd, err := bigtable.tableMachineMetrics.ApplyBulk(
			ctx,
			rowKeys,
			mutations,
		)
		if err != nil {
			return err
		}
		if errInd != nil {
			logrus.Errorf("multiple inserts failed %v", errInd)
			return fmt.Errorf("multiple inserts failed %v", errInd)
		}
		logrus.Infof("Migrated beaconnode batch %v - %v", offset, offset+limit)
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	return nil
}

func (bigtable *Bigtable) migrateValidator(ctx context.Context, limit uint64, fromDay, toDay int, sleep int) error {
	for offset := uint64(0); ; offset += limit {
		rows, err := GetStatsValidatorAll(limit, offset, fromDay, toDay)
		if err != nil {
			return err
		}

		var rowKeys []string
		var mutations []*gcp_bigtable.Mutation
		dataVali, err := utils.SqlRowsToJSON(rows)
		if len(dataVali) == 0 {
			logrus.Infof("vali done, break loop")
			break // done
		}
		for _, it := range dataVali {
			mapData := it.(map[string]interface{})
			createdTs := mapData["created_trunc"].(int64) * 1000000

			rowKeyData := fmt.Sprintf("u:%s:p:%s:m:%v", reversePaddedUserID(uint64(mapData["user_id"].(int64))), "validator", mapData["machine"].(string))

			obj := types.MachineMetricValidator{
				Timestamp:       uint64(mapData["timestamp"].(int64)),
				ExporterVersion: mapData["exporter_version"].(string),
				// process
				CpuProcessSecondsTotal:     uint64(mapData["cpu_process_seconds_total"].(int64)),
				MemoryProcessBytes:         uint64(mapData["memory_process_bytes"].(int64)),
				ClientName:                 mapData["client_name"].(string),
				ClientVersion:              mapData["client_version"].(string),
				ClientBuild:                uint64(mapData["client_build"].(int64)),
				SyncEth2FallbackConfigured: mapData["sync_eth2_fallback_configured"].(bool),
				SyncEth2FallbackConnected:  mapData["sync_eth2_fallback_connected"].(bool),
				// validator
				ValidatorTotal:  uint64(mapData["validator_total"].(int64)),
				ValidatorActive: uint64(mapData["validator_active"].(int64)),
			}
			data, err := proto.Marshal(&obj)
			if err != nil {
				return err
			}

			dataMut := gcp_bigtable.NewMutation()
			dataMut.Set(MACHINE_METRICS_COLUMN_FAMILY, "v1", gcp_bigtable.Timestamp(createdTs), data)

			rowKeys = append(rowKeys, rowKeyData)
			mutations = append(mutations, dataMut)
		}

		errInd, err := bigtable.tableMachineMetrics.ApplyBulk(
			ctx,
			rowKeys,
			mutations,
		)
		if err != nil {
			return err
		}
		if errInd != nil {
			logrus.Errorf("multiple inserts failed %v", errInd)
			return fmt.Errorf("multiple inserts failed %v", errInd)
		}
		logrus.Infof("Migrated validator batch %v - %v", offset, offset+limit)
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	return nil
}
