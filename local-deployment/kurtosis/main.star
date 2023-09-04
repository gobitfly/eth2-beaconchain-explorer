parse_input = import_module("github.com/kurtosis-tech/eth2-package/src/package_io/parse_input.star")
eth_network_module = import_module("github.com/kurtosis-tech/eth-network-package/main.star")
transaction_spammer = import_module("github.com/kurtosis-tech/eth2-package/src/transaction_spammer/transaction_spammer.star")
genesis_constants = import_module("github.com/kurtosis-tech/eth-network-package/src/prelaunch_data_generator/genesis_constants/genesis_constants.star")
shared_utils = import_module("github.com/kurtosis-tech/eth2-package/src/shared_utils/shared_utils.star")

POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "db"
POSTGRES_USER = "postgres"
POSTGRES_PASSWORD = "pass"

REDIS_PORT_ID = "redis"

FRONTEND_PORT_ID = "frontend"

LITTLE_BIGTABLE_PORT_ID = "littlebigtable"

EXPLORER_CONFIG_FILENAME = "config.yml"

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

    # Spin up a local ethereum testnet
    all_participants, cl_genesis_timestamp, genesis_validators_root = eth_network_module.run(plan, args)

    all_el_client_contexts = []
    all_cl_client_contexts = []
    for participant in all_participants:
        all_el_client_contexts.append(participant.el_client_context)
        all_cl_client_contexts.append(participant.cl_client_context)

    # plan.print("Launching transaction spammer")
    # transaction_spammer.launch_transaction_spammer(plan, genesis_constants.PRE_FUNDED_ACCOUNTS, all_el_client_contexts[0])
    # plan.print("Succesfully launched transaction spammer")

    el_uri = "http://{0}:{1}".format(all_el_client_contexts[0].ip_addr, all_el_client_contexts[0].rpc_port_num)
    redis_uri = "{0}:{1}".format(redis.ip_address, 6379)

    plan.print("{0}".format(all_cl_client_contexts[0].ip_addr))

    config_template = read_file("./explorer-config-template.yml")
    template_data = new_config_template_data(all_cl_client_contexts[0], el_uri, littlebigtable.ip_address, 9000, postgres.ip_address, 5432, redis_uri)
    template_and_data = shared_utils.new_template_and_data(config_template, template_data)
    template_and_data_by_rel_dest_filepath = {}
    template_and_data_by_rel_dest_filepath[EXPLORER_CONFIG_FILENAME] = template_and_data

    config_files_artifact_name = plan.render_templates(template_and_data_by_rel_dest_filepath, "config.yml")

    # Initialize the db schema
    initdbschema = plan.add_service(
        name = "initdbschema",
        config = ServiceConfig(
            image = "gobitfly/eth2-beaconchain-explorer:kurtosis",
            files = {
                "/app/config/": config_files_artifact_name,
            },
            entrypoint = [
                "./misc"
            ],
            cmd = [
                "-config",
                "/app/config/config.yml",
                "-command",
                "applyDbSchema"
            ],
        ),
    )
    # Initialize the bigtable schema
    initbigtableschema = plan.add_service(
        name = "initbigtableschema",
        config = ServiceConfig(
            image = "gobitfly/eth2-beaconchain-explorer:kurtosis",
            files = {
                "/app/config/": config_files_artifact_name,
            },
            entrypoint = [
                "./misc"
            ],
            cmd = [
                "-config",
                "/app/config/config.yml",
                "-command",
                "initBigtableSchema"
            ],
        ),
    )
    # Start the indexer
    indexer = plan.add_service(
        name = "indexer",
        config = ServiceConfig(
            image = "gobitfly/eth2-beaconchain-explorer:kurtosis",
            files = {
                "/app/config/": config_files_artifact_name,
            },
            entrypoint = [
                "./explorer"
            ],
            cmd = [
                "-config",
                "/app/config/config.yml",
            ],
            env_vars = {
                "INDEXER_ENABLED": "TRUE",
            }
        ),
    )
    # Start the eth1indexer
    eth1indexer = plan.add_service(
        name = "eth1indexer",
        config = ServiceConfig(
            image = "gobitfly/eth2-beaconchain-explorer:kurtosis",
            files = {
                "/app/config/": config_files_artifact_name,
            },
            entrypoint = [
                "./eth1indexer"
            ],
            cmd = [
                "-config",
                "/app/config/config.yml",
                "-blocks.concurrency",
                "1",
                "-blocks.tracemode",
                "geth",
                "-data.concurrency",
                "1",
                "-balances.enabled"
            ],
        ),
    )

    rewardsexporter = plan.add_service(
        name = "rewardsexporter",
        config = ServiceConfig(
            image = "gobitfly/eth2-beaconchain-explorer:kurtosis",
            files = {
                "/app/config/": config_files_artifact_name,
            },
            entrypoint = [
                "./rewards-exporter"
            ],
            cmd = [
                "-config",
                "/app/config/config.yml",
            ],
        ),
    )

    statistics = plan.add_service(
        name = "statistics",
        config = ServiceConfig(
            image = "gobitfly/eth2-beaconchain-explorer:kurtosis",
            files = {
                "/app/config/": config_files_artifact_name,
            },
            entrypoint = [
                "./statistics"
            ],
            cmd = [
                "-config",
                "/app/config/config.yml",
                "-charts.enabled",
                "-graffiti.enabled",
                "-validators.enabled"
            ],
        ),
    )

    fdu = plan.add_service(
        name = "fdu",
        config = ServiceConfig(
            image = "gobitfly/eth2-beaconchain-explorer:kurtosis",
            files = {
                "/app/config/": config_files_artifact_name,
            },
            entrypoint = [
                "./frontend-data-updater"
            ],
            cmd = [
                "-config",
                "/app/config/config.yml",
            ],
        ),
    )

    frontend = plan.add_service(
        name = "frontend",
        config = ServiceConfig(
            image = "gobitfly/eth2-beaconchain-explorer:kurtosis",
            files = {
                "/app/config/": config_files_artifact_name,
            },
            entrypoint = [
                "./explorer"
            ],
            cmd = [
                "-config",
                "/app/config/config.yml",
            ],
            env_vars = {
                "FRONTEND_ENABLED": "TRUE",
            },
            ports = {
                FRONTEND_PORT_ID: PortSpec(8080, application_protocol = "http"),
            },
        ),
    )

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