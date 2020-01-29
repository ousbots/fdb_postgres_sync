#! /usr/bin/env sh

dropdb test || true
createdb -O postgres test
psql -U postgres -d test -f ./postgres.sql
