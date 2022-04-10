package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"sync"
	"time"

	"github.com/maruel/mess/internal/model"
)

type scheduler struct {
	mu     sync.Mutex
	bots   map[string]*waitingBot
	queues map[uint64][]taskQueue
}

type taskQueue struct {
	dimensions map[string]string
	bots       []string
	lastSeen   time.Time
}

type waitingBot struct {
	bot *model.Bot
	ch  chan *model.TaskRequest
}

func (s *scheduler) init(db model.DB) {
	s.bots = map[string]*waitingBot{}
	s.queues = map[uint64][]taskQueue{}
	/*
		reqs, _ := db.TaskRequestSlice("", 1000, time.Now().Sub(time.Hour), tim.Time{})
		for _, r := range reqs {
			h := hashDimensions(r.Dimensions)
			s.queues[h] = append(s.queues[h], taskQueue{dimensions: r.Dimensions, lastSeen: r.Created})
		}
	*/
}

// loop is the main omniscient scheduling loop.
func (s *scheduler) loop(ctx context.Context) {
	<-ctx.Done()
}

// enqueue registers a task and tries to assign it to a bot inline. Returns
// the resulting TaskResult.
func (s *scheduler) enqueue(ctx context.Context, r *model.TaskRequest) *model.TaskResult {
	// Try to find a bot readily available. If not, skip.
	// TODO(maruel): Precompute task queues.
	s.mu.Lock()
	// Slow naive version.
	//for id, w := range s.bots {
	//	// if dimensions match.
	//}
	s.mu.Unlock()
	// TODO(maruel): Store model.TaskResult.
	return nil
}

// poll is a bot poll, waiting for tasks.
func (s *scheduler) poll(ctx context.Context, bot *model.Bot) *model.TaskRequest {
	// Hang for 10s initially.
	// TODO(maruel): Increase to ~2 minutes?
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	w := waitingBot{bot: bot, ch: make(chan *model.TaskRequest)}
	s.mu.Lock()
	s.bots[bot.Key] = &w
	s.mu.Unlock()

	// Register a channel.
	// Wait for it.
	var t *model.TaskRequest
	select {
	case <-ctx.Done():
	case t = <-w.ch:
	}
	s.mu.Lock()
	delete(s.bots, bot.Key)
	s.mu.Unlock()
	return t
}

//

// murmurHash64A is 64bit MurmurHash2, by Austin Appleby.
func murmurHash64A(data []byte) uint64 {
	const seed uint64 = 0xDECAFBADDECAFBAD
	const m uint64 = 0xc6a4a7935bd1e995
	const r int = 47
	h := seed ^ (uint64(len(data)) * m)
	for len(data) >= 8 {
		k := binary.LittleEndian.Uint64(data)
		k *= m
		k ^= k >> r
		k *= m
		h ^= k
		h *= m
		data = data[8:]
	}
	switch len(data) & 7 {
	case 7:
		h ^= uint64(data[6]) << 48
		fallthrough
	case 6:
		h ^= uint64(data[5]) << 40
		fallthrough
	case 5:
		h ^= uint64(data[4]) << 32
		fallthrough
	case 4:
		h ^= uint64(data[3]) << 24
		fallthrough
	case 3:
		h ^= uint64(data[2]) << 16
		fallthrough
	case 2:
		h ^= uint64(data[1]) << 8
		fallthrough
	case 1:
		h ^= uint64(data[0])
		h *= m
	}
	h ^= h >> r
	h *= m
	h ^= h >> r
	return h
}

func hashDimensions(dims map[string]string) uint64 {
	b, err := json.Marshal(dims)
	if err != nil {
		panic(err)
	}
	return murmurHash64A(b)
}

func dimensionsMatch(req map[string]string, bot map[string][]string) bool {
	for k, v := range req {
		vals := bot[k]
		if len(vals) == 0 {
			return false
		}
		// Linear search since the number of items is very small.
		found := false
		for _, v2 := range vals {
			if v == v2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
