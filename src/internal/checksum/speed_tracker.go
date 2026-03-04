package checksum

import (
	"math/big"
	"sync"
	"time"
)

const speedUpdateInterval = 1.5

type SpeedTracker struct {
	rwm        sync.RWMutex
	totalBytes big.Int
	lastBytes  big.Int
	lastTime   time.Time
	lastSpeed  float64
}

func NewSpeedTracker() *SpeedTracker {
	return &SpeedTracker{
		totalBytes: *big.NewInt(0),
		lastBytes:  *big.NewInt(0),
		lastTime:   time.Now(),
	}
}

func (t *SpeedTracker) AddBytes(n int64) {
	t.rwm.Lock()
	defer t.rwm.Unlock()

	t.totalBytes.Add(&t.totalBytes, big.NewInt(n))

	now := time.Now()
	elapsed := now.Sub(t.lastTime).Seconds()

	if elapsed >= speedUpdateInterval {
		bytesDelta := new(big.Int).Sub(&t.totalBytes, &t.lastBytes)

		t.lastBytes.Set(&t.totalBytes)
		t.lastTime = now

		t.lastSpeed = float64(bytesDelta.Int64()) / elapsed
	}
}

func (t *SpeedTracker) Speed() float64 {
	t.rwm.RLock()
	defer t.rwm.RUnlock()

	return t.lastSpeed
}

func (t *SpeedTracker) Reset() {
	t.rwm.Lock()
	defer t.rwm.Unlock()

	t.totalBytes.SetInt64(0)
	t.lastBytes.SetInt64(0)
	t.lastSpeed = 0
	t.lastTime = time.Now()
}
