parse_input = import_module("github.com/kurtosis-tech/eth2-package/src/package_io/parse_input.star")
eth_network_module = import_module("github.com/kurtosis-tech/eth-network-package/main.star")

POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "db"
POSTGRES_USER = "postgres"
POSTGRES_PASSWORD = "pass"

REDIS_PORT_ID = "redis"

LITTLE_BIGTABLE_PORT_ID = "littlebigtable"

def run(plan, args):
    # Add a Postgres server
    postgres = plan.add_service(
        name = "postgres",
        config = ServiceConfig(
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
    )
    # Add a redis server
    redis = plan.add_service(
        name = "redis",
        config = ServiceConfig(
            image = "redis:7",
            ports = {
                REDIS_PORT_ID: PortSpec(6379, application_protocol = "tcp"),
            },
        ),
    )
    # Add a little bigtable server
    littlebigtable = plan.add_service(
        name = "littlebigtable",
        config = ServiceConfig(
            image = "gobitfly/little_bigtable:latest",
            ports = {
                LITTLE_BIGTABLE_PORT_ID: PortSpec(9000, application_protocol = "tcp"),
            },
        ),
    )

    eth_network_participants, cl_genesis_timestamp, cl_genesis_root_hash = eth_network_module.run(plan, args)

