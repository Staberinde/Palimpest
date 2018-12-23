#!/usr/bin/env bash
/Palimpest/wait-for-it/wait-for-it.sh $DATABASE_HOST:5432 -- echo "$DATABASE_HOST is up"

# Run Integration Tests
psql postgresql://$DATABASE_USER:$DATABASE_PASSWORD@$DATABASE_HOST:5432/postgres -c "CREATE DATABASE ${DATABASE_NAME}_test;"
go test ./...

# Run Application
/Go/bin/Palimpest
