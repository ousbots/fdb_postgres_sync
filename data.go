package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

// addData adds the given data to the given id in FDB and returns any generated errors.
func (state State) addData(id int64, data byte) error {
	_, err := state.fdb.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		mutable, err := state.getData(tr, id)
		if err != nil {
			return nil, err
		}

		if mutable == nil {
			mutable = &Mutable{
				ID: id,
			}
		}

		mutable.Data = append(mutable.Data, data)

		state.setData(tr, id, mutable)
		state.setDirty(tr, id)

		return nil, nil
	})

	return err
}

// getData returns the data found in FDB for the given id and any generated errors.
func (state State) getData(tr fdb.Transaction, id int64) (*Mutable, error) {
	key := state.fdb.dataDir.Pack(tuple.Tuple{id})
	bytes := tr.Get(key).MustGet()

	if len(bytes) == 0 {
		return nil, nil
	}

	var mutable Mutable
	if err := json.Unmarshal(bytes, &mutable); err != nil {
		return nil, err
	}

	return &mutable, nil
}

// setData writes the given data to the given id in FDB
func (state State) setData(tr fdb.Transaction, id int64, data *Mutable) error {
	if data == nil {
		return nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	key := state.fdb.dataDir.Pack(tuple.Tuple{id})
	tr.Set(key, bytes)

	return nil
}

// setDirtyRecord marks the given lookupID as needing to be written to Postgres.
func (state State) setDirty(tr fdb.Transaction, id int64) {
	dirtyKey := state.fdb.dirtyDir.Pack(tuple.Tuple{id})
	tr.Set(dirtyKey, nil)
}

// clearDirtyRecord marks the given lookupID as having been written to Postgres.
func (state State) clearDirty(tr fdb.Transaction, id int64) {
	dirtyKey := state.fdb.dirtyDir.Pack(tuple.Tuple{id})
	tr.Clear(dirtyKey)
}

// dataHammer repeatedly writes to the given id.
func (state State) dataHammer(id int64) {
	bytes := make([]byte, 1)

	for {
		_, err := rand.Read(bytes)
		if err != nil {
			fmt.Printf("id %d hammer just died: %s", id, err)
			return
		}

		_, err = state.fdb.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
			state.addData(id, bytes[0])

			return nil, nil
		})
		if err != nil {
			fmt.Printf("id %d hammer just died: %s", id, err)
			return
		}
	}
}
