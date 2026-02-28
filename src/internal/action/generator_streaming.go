package action

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ostapkonst/hash-verifier/internal/checksum"
	"github.com/ostapkonst/hash-verifier/internal/header"
	"github.com/ostapkonst/hash-verifier/utils/eof"
)

const statsUpdateInterval = 50 * time.Millisecond

type GenerateStreamingResult struct {
	Result           checksum.GenerateResult
	Stats            checksum.GeneratorStats
	Err              error
	IsProgressUpdate bool
}

func GenerateChecksumsStreamingToFile(ctx context.Context, inputDir, outputFile string) (<-chan GenerateStreamingResult, error) {
	if err := ValidateInputDir(inputDir); err != nil {
		return nil, fmt.Errorf("invalid input dir: %w", err)
	}

	algo, err := checksum.AlgorithmFromExtension(outputFile)
	if err != nil {
		return nil, fmt.Errorf("unsupported algorithm: %w", err)
	}

	dirPrefix, err := checksum.GetPrefixForFilesInChecksum(inputDir, outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get prefix: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create checksum file: %w", err)
	}

	bw := bufio.NewWriter(f)
	if _, err := bw.WriteString(header.GetChecksumHeader()); err != nil {
		f.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to write program header: %w", err)
	}

	resultCh := make(chan GenerateStreamingResult, 1)

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		wg := sync.WaitGroup{}

		defer close(resultCh)
		defer wg.Wait()
		defer cancel()
		defer f.Close() //nolint:errcheck

		generator := NewGenerator(ctx, inputDir, algo, dirPrefix)
		generator.Start()

		var hasError error

		wg.Add(1)

		go func() {
			defer wg.Done()

			ticker := time.NewTicker(statsUpdateInterval)
			defer ticker.Stop()

			for range ticker.C {
				select {
				case <-ctx.Done():
					return
				case resultCh <- GenerateStreamingResult{
					Stats:            generator.Stats(),
					IsProgressUpdate: true,
				}:
				default:
					// пропускаем, если канал полон
				}
			}
		}()

		for res := range generator.Results() {
			line := checksum.FormatLine(res.RelPath, res.Hash, algo)

			if _, err := bw.WriteString(line + eof.PlatformEOF); err != nil {
				hasError = fmt.Errorf("failed to write line: %w", err)
				break
			}

			resultCh <- GenerateStreamingResult{
				Result: res,
				Stats:  generator.Stats(),
			}
		}

		if err := generator.Wait(); err != nil && hasError == nil {
			hasError = fmt.Errorf("failed to generate checksums: %w", err)
		}

		if err := bw.Flush(); err != nil && hasError == nil {
			hasError = fmt.Errorf("failed to flush buffer: %w", err)
		}

		resultCh <- GenerateStreamingResult{
			Stats:            generator.Stats(),
			IsProgressUpdate: true,
			Err:              hasError,
		}
	}()

	return resultCh, nil
}
