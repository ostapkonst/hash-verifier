package action

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/ostapkonst/hash-verifier/internal/checksum"
	"github.com/ostapkonst/hash-verifier/internal/header"
	"github.com/ostapkonst/hash-verifier/utils/eof"
)

type GenerateConfig struct {
	InputDir            string
	OutputFile          string
	FollowSymbolicLinks bool
	SortPaths           bool
	OnFileHashed        func(result checksum.GenerateResult)
}

type GenerateResultStats struct {
	Stats checksum.GeneratorStats
}

func formatStatsFooter(stats checksum.GeneratorStats) string {
	status := "Success"
	if stats.WithErrors > 0 {
		status = "Completed with errors"
	}

	optionalNewLine := eof.PlatformEOF
	if stats.TotalFiles == 0 {
		optionalNewLine = ""
	}

	return fmt.Sprintf(
		"%s"+
			"; Statistics:%s"+
			";   Status: %s%s"+
			";   Total files: %d%s"+
			";   Hashed: %d%s"+
			";   With errors: %d%s",
		optionalNewLine,
		eof.PlatformEOF,
		status,
		eof.PlatformEOF,
		stats.TotalFiles,
		eof.PlatformEOF,
		stats.Processed,
		eof.PlatformEOF,
		stats.WithErrors,
		eof.PlatformEOF,
	)
}

func ValidateInputDir(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat input directory: %w", err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("input path is not a directory")
	}

	return nil
}

func GenerateChecksums(ctx context.Context, cfg GenerateConfig) (GenerateResultStats, error) {
	if err := ValidateInputDir(cfg.InputDir); err != nil {
		return GenerateResultStats{}, fmt.Errorf("invalid input dir: %w", err)
	}

	algo, err := checksum.AlgorithmFromExtension(cfg.OutputFile)
	if err != nil {
		return GenerateResultStats{}, fmt.Errorf("unsupported algorithm: %w", err)
	}

	dirPrefix, err := checksum.GetPrefixForFilesInChecksum(cfg.InputDir, cfg.OutputFile)
	if err != nil {
		return GenerateResultStats{}, fmt.Errorf("failed to get prefix: %w", err)
	}

	f, err := os.Create(cfg.OutputFile)
	if err != nil {
		return GenerateResultStats{}, fmt.Errorf("failed to create checksum file: %w", err)
	}
	defer f.Close() //nolint:errcheck

	bw := bufio.NewWriter(f)

	if _, err := bw.WriteString(header.GetChecksumHeader()); err != nil {
		return GenerateResultStats{}, fmt.Errorf("failed to write program header: %w", err)
	}

	generator := NewGenerator(ctx, cfg.InputDir, algo, dirPrefix, cfg.FollowSymbolicLinks, cfg.SortPaths)
	generator.Start()

	var hasError error

	for res := range generator.Results() {
		line := checksum.FormatLine(res.RelPath, res.Hash, algo)

		if _, err = bw.WriteString(line + eof.PlatformEOF); err != nil {
			hasError = fmt.Errorf("failed to write line: %w", err)
			break
		}

		if cfg.OnFileHashed != nil {
			cfg.OnFileHashed(res)
		}
	}

	if err := generator.Wait(); err != nil && hasError == nil {
		hasError = fmt.Errorf("failed to generate checksums: %w", err)
	}

	if _, err := bw.WriteString(formatStatsFooter(generator.Stats())); err != nil && hasError == nil {
		hasError = fmt.Errorf("failed to write stats footer: %w", err)
	}

	if err := bw.Flush(); err != nil && hasError == nil {
		hasError = fmt.Errorf("failed to flush buffer: %w", err)
	}

	return GenerateResultStats{Stats: generator.Stats()}, hasError
}
