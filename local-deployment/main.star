input_parser = import_module("github.com/kurtosis-tech/ethereum-package/src/package_io/input_parser.star")
eth_network_module = import_module("github.com/kurtosis-tech/eth-network-package/main.star")
transaction_spammer = import_module("github.com/kurtosis-tech/ethereum-package/src/transaction_spammer/transaction_spammer.star")
blob_spammer = import_module("github.com/kurtosis-tech/ethereum-package/src/blob_spammer/blob_spammer.star")
genesis_constants = import_module("github.com/kurtosis-tech/eth-network-package/src/prelaunch_data_generator/genesis_constants/genesis_constants.star")
shared_utils = import_module("github.com/kurtosis-tech/ethereum-package/src/shared_utils/shared_utils.star")

POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "db"
POSTGRES_USER = "postgres"
POSTGRES_PASSWORD = "pass"

REDIS_PORT_ID = "redis"

LITTLE_BIGTABLE_PORT_ID = "littlebigtable"

EXPLORER_CONFIG_FILENAME = "config.yml"

def run(plan, args):
	args_with_right_defaults = input_parser.input_parser(plan, args)
	network_params = args_with_right_defaults.network_params

	db_services = plan.add_services(
		configs={
			# Add a Postgres server
			"postgres": ServiceConfig(
				image = "postgres:15.2-alpine",
				ports = {
					POSTGRES_PORT_ID: PortSpec(5432, application_protocol = "postgresql"),
				},
				env_vars = {
					"POSTGRES_DB": POSTGRES_DB,
					"POSTGRES_USER": POSTGRES_USER,
					"POSTGRES_PASSWORD": POSTGRES_PASSWORD,
				},
			),
			# Add a Redis server
			"redis": ServiceConfig(
				image = "redis:7",
				ports = {
					REDIS_PORT_ID: PortSpec(6379, application_protocol = "tcp"),
				},
			),
			# Add a Bigtable Emulator server
			"littlebigtable": ServiceConfig(
				image = "gobitfly/little_bigtable:latest",
				ports = {
					LITTLE_BIGTABLE_PORT_ID: PortSpec(9000, application_protocol = "tcp"),
				},
			),
		}
	)

	# Spin up a local ethereum testnet
	all_participants, cl_genesis_timestamp, genesis_validators_root = eth_network_module.run(plan, args)

	all_el_client_contexts = []
	all_cl_client_contexts = []
	for participant in all_participants:
		all_el_client_contexts.append(participant.el_client_context)
		all_cl_client_contexts.append(participant.cl_client_context)

	fuzz_target = "http://{0}:{1}".format(
		all_el_client_contexts[0].ip_addr,
		all_el_client_contexts[0].rpc_port_num,
	)

	if args["start_tx_spammer"]:
		plan.print("Launching transaction spammer")
		transaction_spammer.launch_transaction_spammer(plan, genesis_constants.PRE_FUNDED_ACCOUNTS, fuzz_target, args_with_right_defaults.tx_spammer_params, network_params.electra_fork_epoch)
		plan.print("Succesfully launched transaction spammer")

	if args["start_blob_spammer"]:
		plan.print("Launching Blob spammer")
		blob_spammer.launch_blob_spammer(
			plan,
			genesis_constants.PRE_FUNDED_ACCOUNTS,
			fuzz_target,
			all_cl_client_contexts[0],
			network_params.deneb_fork_epoch,
			network_params.seconds_per_slot,
			network_params.slots_per_epoch,
			network_params.genesis_delay,
		)
		plan.print("Succesfully launched blob spammer")


def new_config_template_data(cl_node_info, el_uri, lbt_host, lbt_port, db_host, db_port, redis_uri):
	return {
		"CLNodeHost": cl_node_info.ip_addr,
		"CLNodePort": cl_node_info.http_port_num,
		"ELNodeEndpoint": el_uri,
		"LBTHost": lbt_host,
		"LBTPort": lbt_port,
		"DBHost": db_host,
		"DBPort": db_port,
		"RedisEndpoint": redis_uri,

	}
