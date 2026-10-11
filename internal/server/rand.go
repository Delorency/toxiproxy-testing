package server

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	rndMu sync.Mutex
	rnd   = rand.New(rand.NewSource(loadSeed()))
)

func loadSeed() int64 {
	if s := os.Getenv("DETERMINISTIC_SEED"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			return v
		}
	}
	return time.Now().UnixNano()
}

// randFloat возвращает число в [0, 1).
func randFloat() float64 {
	rndMu.Lock()
	defer rndMu.Unlock()
	return rnd.Float64()
}
