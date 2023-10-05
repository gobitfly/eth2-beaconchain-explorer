parse_input = import_module("github.com/kurtosis-tech/ethereum-package/src/package_io/parse_input.star")
eth_network_module = import_module("github.com/kurtosis-tech/ethereum-package/main.star")
transaction_spammer = import_module("github.com/kurtosis-tech/ethereum-package/src/transaction_spammer/transaction_spammer.star")
blob_spammer = import_module("github.com/kurtosis-tech/ethereum-package/src/blob_spammer/blob_spammer.star")
genesis_constants = import_module("github.com/kurtosis-tech/ethereum-package/src/prelaunch_data_generator/genesis_constants/genesis_constants.star")
shared_utils = import_module("github.com/kurtosis-tech/ethereum-package/src/shared_utils/shared_utils.star")

POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "db"
POSTGRES_USER = "postgres"
POSTGRES_PASSWORD = "pass"

REDIS_PORT_ID = "redis"

LITTLE_BIGTABLE_PORT_ID = "littlebigtable"

def run(plan, args):
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
	eth_network_module.run(plan, args)


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