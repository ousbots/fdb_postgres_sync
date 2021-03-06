package main

import (
	"fmt"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (state State) runWriter(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			ids, err := state.getDirtyIDs()
			if err != nil {
				log.Error().Err(err).Str("component", "writer").Msg("failed to get dirty IDs")
			}

			for _, id := range ids {
				if err := state.writeDirtyID(id); err != nil {
					log.Error().Err(err).Str("component", "writer").Int64("id", id).Msg("failed to write dirty data")
					continue
				}
			}
		}
	}
}

// getDirtyIDs returns a list of all ids that are marked as dirty or any errors generated.
func (state State) getDirtyIDs() ([]int64, error) {
	var dirtyIDs []int64

	_, err := state.fdb.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		dirtyIDs = nil

		iter := tr.GetRange(state.fdb.dirtyDir, fdb.RangeOptions{}).Iterator()

		for iter.Advance() {
			kvs := iter.MustGet()

			tup, err := state.fdb.dirtyDir.Unpack(kvs.Key)
			if err != nil {
				return nil, err
			}

			if len(tup) != 1 {
				return nil, errors.New("incorrect tuple size")
			}

			var id int64
			switch tup[0].(type) {
			case int64:
				id = tup[0].(int64)

			default:
				return nil, fmt.Errorf("tuple lookupID is %T not an int64", tup[0])
			}

			if !state.getWrite(tr, id) {
				state.setWrite(tr, id)
				dirtyIDs = append(dirtyIDs, id)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get lookupIDs of dirty records")
	}

	return dirtyIDs, nil
}

// writeDirtyID moves the data of the given id from FDB to Postgres and return any generated
// errors.
func (state State) writeDirtyID(id int64) error {
	var data *MutableData
	var gotoErr error

	_, err := state.fdb.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		var err error
		if data, err = state.getData(tr, id); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to clear dirty record")
	}

	// Record the parts of the data to Postgres and clear them from FDB
	wroteData := false

	if !data.IDWritten {
		if _, err := state.sql.insertID.Exec(id); err != nil {
			gotoErr = errors.Wrap(err, "failed to insert id")
			goto exit
		}

		data.IDWritten = true
	}

	if len(data.Data) > 0 {
		if _, err := state.sql.insertData.Exec(id, pq.Array(data.Data)); err != nil {
			gotoErr = errors.Wrap(err, "failed to write errors to Postgres")
			goto exit
		}

		wroteData = true
	}

exit:
	if !wroteData {
		data.Data = nil
	}

	err = state.recordDirtyDataDiff(id, data)
	if err != nil {
		return fmt.Errorf("multiple errors: %s, %s", gotoErr,
			errors.Wrap(err, "failed to record the diff of dirty data"))
	}

	return gotoErr
}

func (state State) recordDirtyDataDiff(id int64, dirty *MutableData) error {
	if dirty == nil {
		return nil
	}

	// Go through the record in FDB and the dirty record and keep any elements that are in the FDB
	// record but not in the dirty record.
	_, err := state.fdb.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		dirtyData := dirty.deepCopy()

		fdbData, err := state.getData(tr, id)
		if err != nil {
			return nil, err
		}

		var cleanData MutableData
		cleanData.IDWritten = dirtyData.IDWritten
		cleanData.ID = dirtyData.ID

		for _, element := range fdbData.Data {
			if dirtyData.containsInt(element) {
				dirtyData.deleteFirstElement(element)
				continue
			}

			cleanData.Data = append(cleanData.Data, element)
		}

		if len(cleanData.Data) == 0 {
			state.clearDirty(tr, id)
		}

		state.setData(tr, id, &cleanData)
		state.clearWrite(tr, id)

		return nil, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to clear dirty record")
	}

	return nil
}
