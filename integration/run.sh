#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
DIR_ROOT="${DIR}/.."

var_help="./run.sh <cmd> <opt>

    run|r)       .. run everything (initdb && tests)
    init_db|idb) .. initialize database
    test|t)      .. run tests

"

fn_main() {
    if test $# -eq 0; then
        echo "$var_help"
        return
    fi
    while test $# -ne 0; do
        case $1 in
            run|r) shift; fn_run "$@"; exit;;
            init_db|idb) shift; fn_init_db "$@"; exit;;
            test|t) shift; fn_test "$@"; exit;;
            *) echo "$var_help"
        esac
        shift
    done
}

fn_run() {
    sudo docker-compose down
    fn_init_db
    fn_test
}

fn_init_db() {
    echo "dropping old database"
    sudo docker-compose stop db
    sudo rm -rf $DIR/docker-volumes/db
    sudo docker-compose up -d db
    echo "waiting for database to be up and running"
    sudo docker-compose exec -T db /bin/bash -c "while ! psql -h localhost -U beaconchain -c \"\\l\"; do sleep 1; done"
    echo "initializing database (loading schema)"
    # sudo docker-compose exec -T db /bin/bash -c 'psql -h localhost -U beaconchain -d beaconchain' < $DIR_ROOT/tables.sql
    sudo docker-compose exec -T db /bin/bash -c 'psql -h localhost -U beaconchain -d beaconchain' < $DIR/beaconchain_dump.sql
}

fn_test() {
    sudo docker-compose up -d explorer
    go run main.go
    sudo docker-compose down
    echo "all done!"
}

fn_main "$@"
