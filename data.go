package main

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/rs/zerolog/log"
)

// addData adds the given data to the given id in FDB and returns any generated errors.
func (state State) addData(id int64, data byte) error {
	_, err := state.fdb.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		mutable, err := state.getData(tr, id)
		if err != nil {
			return nil, err
		}

		if mutable == nil {
			mutable = &MutableData{
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
func (state State) getData(tr fdb.Transaction, id int64) (*MutableData, error) {
	key := state.fdb.dataDir.Pack(tuple.Tuple{id})
	bytes := tr.Get(key).MustGet()

	if len(bytes) == 0 {
		return nil, nil
	}

	var mutable MutableData
	if err := json.Unmarshal(bytes, &mutable); err != nil {
		return nil, err
	}

	return &mutable, nil
}

// setData writes the given data to the given id in FDB
func (state State) setData(tr fdb.Transaction, id int64, data *MutableData) error {
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
func (state State) dataHammer(id int64, interval time.Duration) {
	bytes := make([]byte, 1)

	ticker := time.NewTicker(interval)

	for {
		<-ticker.C
		_, err := rand.Read(bytes)
		if err != nil {
			log.Error().Err(err).Str("component", "hammer").Int64("id", id).Msg("failed to generate random bytes")
			return
		}

		_, err = state.fdb.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
			state.addData(id, bytes[0])

			return nil, nil
		})
		if err != nil {
			log.Error().Err(err).Int64("id", id).Str("component", "hammer").Msg("failed to add data")
			return
		}
	}
}
