parse_input = import_module("github.com/kurtosis-tech/eth2-package/src/package_io/parse_input.star")
eth_network_module = import_module("github.com/kurtosis-tech/eth-network-package/main.star")
transaction_spammer = import_module("github.com/kurtosis-tech/eth2-package/src/transaction_spammer/transaction_spammer.star")
genesis_constants = import_module("github.com/kurtosis-tech/eth-network-package/src/prelaunch_data_generator/genesis_constants/genesis_constants.star")

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

    all_participants, cl_genesis_timestamp, genesis_validators_root = eth_network_module.run(plan, args)

    all_el_client_contexts = []
    all_cl_client_contexts = []
    for participant in all_participants:
        all_el_client_contexts.append(participant.el_client_context)
        all_cl_client_contexts.append(participant.cl_client_context)

    plan.print("Launching transaction spammer")
    transaction_spammer.launch_transaction_spammer(plan, genesis_constants.PRE_FUNDED_ACCOUNTS, all_el_client_contexts[0])
    plan.print("Succesfully launched transaction spammer")

