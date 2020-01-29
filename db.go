package main

import (
	"database/sql"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	_ "github.com/lib/pq"
)

const (
	idTableName   = "id"
	dataTableName = "data"

	insertIDSQL   = "INSERT INTO " + idTableName + " VALUES ($1) ON CONFLICT DO NOTHING"
	insertDataSQL = "UPDATE " + dataTableName + " SET data = data || $2 WHERE id = $1"
)

type fdbState struct {
	db       fdb.Database
	dataDir  directory.DirectorySubspace
	dirtyDir directory.DirectorySubspace
}

type sqlState struct {
	db         *sql.DB
	insertID   *sql.Stmt
	insertData *sql.Stmt
}

type State struct {
	fdb fdbState
	sql sqlState
}

func newState() *State {
	var state State
	var err error

	fdb.MustAPIVersion(620)
	state.fdb.db = fdb.MustOpenDefault()

	state.fdb.dataDir, err = directory.CreateOrOpen(state.fdb.db, []string{"data"}, nil)
	if err != nil {
		panic(err)
	}

	state.fdb.dirtyDir, err = directory.CreateOrOpen(state.fdb.db, []string{"dirty"}, nil)
	if err != nil {
		panic(err)
	}

	connStr := "dbname='test' user='postgres' password='' host='localhost' port='5432' sslmode=disable"
	sqldb, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	state.sql.db = sqldb

	stmt, err := sqldb.Prepare(insertIDSQL)
	if err != nil {
		panic(err)
	}
	state.sql.insertID = stmt

	stmt, err = sqldb.Prepare(insertDataSQL)
	if err != nil {
		panic(err)
	}
	state.sql.insertData = stmt

	return &state
}
