#!/bin/bash

# Ask for name of migration
echo "Enter name of migration file (for example add_validators_indices): "
read -r name

# This script creates a new migration file with the current timestamp
# as the filename prefix.
filename=$(date +"%Y%m%d%H%M%S")_$name.sql
touch $filename

cat <<EOF > $filename
-- +goose Up
-- +goose StatementBegin

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
EOF


