package action

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ostapkonst/hash-verifier/internal/checksum"
)

type VerifyStreamingResult struct {
	Result           checksum.VerifyResult
	Stats            checksum.VerifierStats
	Err              error
	IsProgressUpdate bool
}

func VerifyChecksumsStreaming(ctx context.Context, checksumFile string) (<-chan VerifyStreamingResult, error) {
	if err := ValidateChecksumFile(checksumFile); err != nil {
		return nil, fmt.Errorf("invalid checksum file: %w", err)
	}

	algo, err := checksum.AlgorithmFromExtension(checksumFile)
	if err != nil {
		return nil, fmt.Errorf("unsupported algorithm: %w", err)
	}

	resultCh := make(chan VerifyStreamingResult, 1)

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		wg := sync.WaitGroup{}

		defer close(resultCh)
		defer wg.Wait()
		defer cancel()

		verifier := NewVerifier(ctx, checksumFile, algo)
		verifier.Start()

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
				case resultCh <- VerifyStreamingResult{
					Stats:            verifier.Stats(),
					IsProgressUpdate: true,
				}:
				default:
					// пропускаем, если канал полон
				}
			}
		}()

		for res := range verifier.Results() {
			resultCh <- VerifyStreamingResult{
				Result: res,
				Stats:  verifier.Stats(),
			}
		}

		if err := verifier.Wait(); err != nil {
			hasError = fmt.Errorf("verification failed: %w", err)
		}

		resultCh <- VerifyStreamingResult{
			Stats:            verifier.Stats(),
			IsProgressUpdate: true,
			Err:              hasError,
		}
	}()

	return resultCh, nil
}
