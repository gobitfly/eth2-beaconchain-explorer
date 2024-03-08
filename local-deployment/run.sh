#!/bin/bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd $DIR
touch .env
. .env

var_help="./run.sh <cmd> <options>

run.sh start  # start local chain and explorer (will run stop first to make sure everything is clean)
run.sh stop   # stop everything and clean up
run.sh sql    # connect to database (only works if running)
"

fn_main() {
    if test $# -eq 0; then
        echo "$var_help"
        return
    fi
    while test $# -ne 0; do
        case $1 in
            start) shift; fn_start "$@"; exit;;
            stop) shift; fn_stop "$@"; exit;;
            sql) shift; fn_sql "$@"; exit;;
            redis) shift; fn_redis "$@"; exit;;
            misc) shift; fn_misc "$@"; exit;;
            *) echo "$var_help"
        esac
        shift
    done
}

fn_misc() {
    docker compose exec misc go run ./cmd/misc -config /app/local-deployment/config.yml $@
}

fn_sql() {
    if [ -z "${1}" ]; then
        PGPASSWORD=pass psql -h localhost -p$POSTGRES_PORT -U postgres -d db
    else
        PGPASSWORD=pass psql -h localhost -p$POSTGRES_PORT -U postgres -d db -c "$@" --csv --pset=pager=off
    fi
}

fn_redis() {
    if [ -z "${1}" ]; then
        docker compose exec redis-sessions redis-cli
    else
        docker compose exec redis-sessions redis-cli "$@"
    fi
    #redis-cli -h localhost -p $REDIS_PORT
}

fn_start() {
    fn_stop
    # build once before starting all services to prevent multiple parallel builds
    docker compose --profile=build-once run -T build-once &
    kurtosis run --enclave my-testnet . "$(cat network-params.json)" &
    wait
    bash provision-explorer-config.sh
    docker compose up -d
    echo "Waiting for explorer to start, then browse http://localhost:8080"
}

fn_stop() {
    docker compose down -v
    kurtosis clean -a
}

fn_main "$@"
