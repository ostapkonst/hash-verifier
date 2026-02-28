package action

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/ostapkonst/hash-verifier/internal/checksum"
)

type GeneratorStatusType int

const (
	GeneratorStatusFinished GeneratorStatusType = iota
	GeneratorStatusStated
)

type Generator struct {
	rwm    sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	root                    string
	algo                    checksum.Algorithm
	dirPrefix               string
	stats                   checksum.GeneratorStats
	currFileHashingProgress atomic.Value

	status GeneratorStatusType

	err      chan error
	resultCh chan checksum.GenerateResult
	done     chan struct{}
}

func NewGenerator(
	ctx context.Context,
	root string,
	algo checksum.Algorithm,
	dirPrefix string,
) *Generator {
	ctx, cancel := context.WithCancel(ctx)

	g := &Generator{
		ctx:       ctx,
		cancel:    cancel,
		resultCh:  make(chan checksum.GenerateResult, 1),
		done:      make(chan struct{}),
		err:       make(chan error, 1),
		root:      root,
		algo:      algo,
		dirPrefix: dirPrefix,
		status:    GeneratorStatusFinished,
	}

	g.stats = checksum.GeneratorStats{CurrentFileOrStatus: "ready to go..."}
	g.currFileHashingProgress.Store(func() float64 { return 0 })

	return g
}

func (g *Generator) Start() {
	g.rwm.Lock()
	defer g.rwm.Unlock()

	if g.status != GeneratorStatusFinished {
		return
	}

	g.status = GeneratorStatusStated

	g.stats = checksum.GeneratorStats{CurrentFileOrStatus: "ready to go..."}
	g.currFileHashingProgress.Store(func() float64 { return 0 })

	go g.run()
}

func (g *Generator) Wait() error {
	<-g.done
	return <-g.err
}

func (g *Generator) Stats() checksum.GeneratorStats {
	fileHashProgress := g.currFileHashingProgress.Load().(func() float64)

	g.rwm.RLock()
	defer g.rwm.RUnlock()

	stats := g.stats
	stats.FileHashingProgress = fileHashProgress()

	return stats
}

func (g *Generator) Results() <-chan checksum.GenerateResult {
	return g.resultCh
}

func (g *Generator) updateStats(withError bool) {
	g.rwm.Lock()
	defer g.rwm.Unlock()

	if withError {
		g.stats.WithErrors++
		return
	}

	g.stats.Processed++
}

func (g *Generator) updateCurrentFileOrStatus(file string) {
	g.rwm.Lock()
	defer g.rwm.Unlock()

	g.stats.CurrentFileOrStatus = file
}

func (g *Generator) updateStatsPending(totalFiles int) {
	g.rwm.Lock()
	defer g.rwm.Unlock()

	g.stats.TotalFiles = totalFiles
}

func (g *Generator) run() {
	defer g.updateCurrentFileOrStatus("ready to go...")
	defer func() {
		g.rwm.Lock()
		defer g.rwm.Unlock()

		g.status = GeneratorStatusFinished
	}()

	defer close(g.done)
	defer close(g.err)
	defer close(g.resultCh)
	defer g.cancel()

	g.updateCurrentFileOrStatus("forming a list of files for hashing...")

	files, err := checksum.WalkDir(g.ctx, g.root)
	if err != nil {
		g.err <- err
		return
	}

	g.updateStatsPending(len(files))

	for _, file := range files {
		g.updateCurrentFileOrStatus(file)

		hashCalc := checksum.NewHashCalculator(file, g.algo)
		g.currFileHashingProgress.Store(hashCalc.Progress)

		var fileErr error

		hastResult, err := hashCalc.Calculate(g.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				g.err <- err
				return
			}

			fileErr = err
		}

		relPath, err := filepath.Rel(g.root, file)
		if err != nil {
			g.err <- fmt.Errorf("failed to calculate relative path: %w", err)
			return
		}

		finalPath := filepath.Join(g.dirPrefix, relPath)

		g.updateStats(fileErr != nil)

		g.resultCh <- checksum.GenerateResult{
			RelPath:   finalPath,
			Hash:      hastResult.Hash,
			ReadBytes: hastResult.ReadBytes,
			Err:       fileErr,
		}
	}
}
