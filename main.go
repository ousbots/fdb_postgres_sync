package main

import (
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Count struct {
	id    int64
	count map[int64]uint64
}

const RUNTIME = 4 * time.Second
const HAMMER_INTERVAL = 100 * time.Millisecond
const WRITER_INTERVAL = 1000 * time.Millisecond
const NUMBER_OF_HAMMERS = 1

func main() {
	state := newState()
	go state.runWriter(WRITER_INTERVAL)

	countChan := make(chan Count, NUMBER_OF_HAMMERS)
	stopChan := make(chan struct{})
	wg := new(sync.WaitGroup)

	for id := int64(0); id < NUMBER_OF_HAMMERS; id++ {
		wg.Add(1)
		go state.dataHammer(id, HAMMER_INTERVAL, wg, stopChan, countChan)
	}

	timer := time.NewTimer(RUNTIME)
	<-timer.C
	close(stopChan)
	wg.Wait()

	timer.Reset(2 * WRITER_INTERVAL)
	<-timer.C

	close(countChan)

countLoop:
	for count := range countChan {
		pgCount := make(map[int64]uint64)

		rows, err := state.sql.getData.Query(count.id)
		if err != nil {
			log.Error().Err(err).Int64("id", count.id).Msg("failed to get data from postgres")
			continue
		}

		for rows.Next() {
			var data pq.Int64Array
			var id int64

			rows.Scan(&id, &data)
			if id != count.id {
				log.Warn().Int64("id", count.id).Int64("pgID", id).Msg("mismatched ids")
			}

			for _, elem := range data {
				pgCount[elem] += 1
			}
		}

		if len(count.count) != len(pgCount) {
			log.Error().Int64("id", count.id).Msg("data sizes do not match")
			log.Error().Msgf("writer: %+v, pg: %+v", count.count, pgCount)
			continue countLoop
		}

		for key, _ := range count.count {
			if count.count[key] != pgCount[key] {
				log.Error().Int64("id", count.id).Msgf("data does not match: key: %d, writer: %+v, pg: %+v", key, count.count, pgCount)
				continue countLoop
			}
		}

		log.Info().Int64("id", count.id).Msg("data matches!")
	}
}
