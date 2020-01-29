#! /usr/bin/env sh

fdbcli --exec 'writemode on; clearrange "" "\xFF"' || true
dropdb test || true
createdb -O postgres test
psql -U postgres -d test -f ./postgres.sql
