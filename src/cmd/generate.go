package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/inhies/go-bytesize"
	"github.com/lithammer/dedent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/ostapkonst/hash-verifier/internal/action"
	"github.com/ostapkonst/hash-verifier/internal/checksum"
	"github.com/ostapkonst/hash-verifier/utils/gracer"
)

var (
	generateNoFollowSymlinks bool
	generateNoSortPaths      bool
)

func runGenerate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- execGenerate(ctx, args)

		gracer.GracefulShutdown()
	}()

	gracer.AddCallback(func() error {
		cancel()
		return <-done
	})

	return gracer.Wait()
}

func execGenerate(ctx context.Context, args []string) error {
	inputDir := filepath.Clean(args[0])
	outputFile := filepath.Clean(args[1])

	cfg := action.GenerateConfig{
		InputDir:            inputDir,
		OutputFile:          outputFile,
		FollowSymbolicLinks: !generateNoFollowSymlinks,
		SortPaths:           !generateNoSortPaths,
		OnFileHashed: func(res checksum.GenerateResult) {
			commonFields := func(event *zerolog.Event, err error) *zerolog.Event {
				logger := event.
					Str("file", res.RelPath).
					Str("hash", res.Hash).
					Str("size", bytesize.New(float64(res.ReadBytes)).String())

				if err != nil {
					logger = logger.Err(err)
				}

				return logger
			}

			if res.Err != nil {
				commonFields(log.Error(), res.Err).Msg("Failed to hash file")
				return
			}

			commonFields(log.Info(), nil).Msg("Hashed")
		},
	}

	log.Info().
		Str("input_dir", inputDir).
		Str("output_file", outputFile).
		Msg("Starting checksum generation")

	result, err := action.GenerateChecksums(ctx, cfg)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Info().Msg("Checksum generation canceled")
			return nil
		}

		return fmt.Errorf("failed to generate checksums: %w", err)
	}

	stats := result.Stats
	log.Info().
		Int("processed", stats.Processed).
		Int("pending", stats.Pending()).
		Int("total_files", stats.TotalFiles).
		Msg("Checksum generation stats")

	log.Info().
		Str("file", outputFile).
		Int("files", stats.Processed).
		Msg("Checksum generation completed")

	return nil
}

var generateCmd = &cobra.Command{
	Use:   "generate <input_dir> <checksum_file>",
	Short: "Generate checksum file recursively from directory",
	Long: strings.Trim(dedent.Dedent(`
		Generate checksum file recursively from directory.
		Algorithm is determined automatically from file extension:
		.sfv, .md4, .md5, .sha1, .sha256, .sha384, .sha512, .sha3-256, .sha3-384, .sha3-512, .blake3.`,
	), "\n"),
	Args: cobra.ExactArgs(2),
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().BoolVar(&generateNoFollowSymlinks, "no-follow-symlinks", false,
		"do not follow symbolic links when scanning directories")
	generateCmd.Flags().BoolVar(&generateNoSortPaths, "no-sort", false,
		"do not sort files before generating checksums")

	rootCmd.AddCommand(generateCmd)
}
