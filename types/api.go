package types

type ApiResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type StatsSystem struct {
	CPUCores                      uint64 `json:"cpu_cores"`
	CPUThreads                    uint64 `json:"cpu_threads"`
	CPUNodeSystemSecondsTotal     uint64 `json:"cpu_node_system_seconds_total"`
	CPUNodeUserSecondsTotal       uint64 `json:"cpu_node_user_seconds_total"`
	CPUNodeIowaitSecondsTotal     uint64 `json:"cpu_node_iowait_seconds_total"`
	CPUNodeIdleSecondsTotal       uint64 `json:"cpu_node_idle_seconds_total"`
	MemoryNodeBytesTotal          uint64 `json:"memory_node_bytes_total"`
	MemoryNodeBytesFree           uint64 `json:"memory_node_bytes_free"`
	MemoryNodeBytesCached         uint64 `json:"memory_node_bytes_cached"`
	MemoryNodeBytesBuffers        uint64 `json:"memory_node_bytes_buffers"`
	DiskNodeBytesTotal            uint64 `json:"disk_node_bytes_total"`
	DiskNodeBytesFree             uint64 `json:"disk_node_bytes_free"`
	DiskNodeIoSeconds             uint64 `json:"disk_node_io_seconds"`
	DiskNodeReadsTotal            uint64 `json:"disk_node_reads_total"`
	DiskNodeWritesTotal           uint64 `json:"disk_node_writes_total"`
	NetworkNodeBytesTotalReceive  uint64 `json:"network_node_bytes_total_receive"`
	NetworkNodeBytesTotalTransmit uint64 `json:"network_node_bytes_total_transmit"`
	MiscNodeBootTsSeconds         uint64 `json:"misc_node_boot_ts_seconds"`
	MiscOS                        string `json:"misc_os"`
}

type StatsProcess struct {
	CPUProcessSecondsTotal     uint64 `json:"cpu_process_seconds_total"`
	MemoryProcessBytes         uint64 `json:"memory_process_bytes"`
	ClientName                 string `json:"client_name"`
	ClientVersion              string `json:"client_version"`
	ClientBuild                uint64 `json:"client_build"`
	SyncEth1FallbackConfigured bool   `json:"sync_eth1_fallback_configured"`
	SyncEth1FallbackConnected  bool   `json:"sync_eth1_fallback_connected"`
	SyncEth2FallbackConfigured bool   `json:"sync_eth2_fallback_configured"`
	SyncEth2FallbackConnected  bool   `json:"sync_eth2_fallback_connected"`
}

type StatsAdditionalsValidator struct {
	ValidatorTotal  uint64 `json:"validator_total"`
	ValidatorActive uint64 `json:"validator_active"`
}

type StatsAdditionalsBeaconnode struct {
	DiskBeaconchainBytesTotal       uint64 `json:"disk_beaconchain_bytes_total"`
	NetworkLibp2pBytesTotalReceive  uint64 `json:"network_libp2p_bytes_total_receive"`
	NetworkLibp2pBytesTotalTransmit uint64 `json:"network_libp2p_bytes_total_transmit"`
	NetworkPeersConnected           uint64 `json:"network_peers_connected"`
	SyncEth1Connected               bool   `json:"sync_eth1_connected"`
	SyncEth2Synced                  bool   `json:"sync_eth2_synced"`
	SyncBeaconHeadSlot              uint64 `json:"sync_beacon_head_slot"`
}

type StatsMeta struct {
	Version   uint64 `json:"version"`
	Timestamp uint64 `json:"timestamp"`
	Process   string `json:"process"`
	Machine   string
}
