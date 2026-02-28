package cmd

import (
	"context"
	"errors"
	"fmt"
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

func runVerify(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- execVerify(ctx, args)

		gracer.GracefulShutdown()
	}()

	gracer.AddCallback(func() error {
		cancel()
		return <-done
	})

	return gracer.Wait()
}

func execVerify(ctx context.Context, args []string) error {
	checksumFile := args[0]

	cfg := action.VerifyConfig{
		ChecksumFile: checksumFile,
		OnFileVerified: func(res checksum.VerifyResult) {
			commonFields := func(event *zerolog.Event, err error) *zerolog.Event {
				logger := event.
					Str("file", res.Path).
					Str("status", res.Status.String()).
					Str("size", bytesize.New(float64(res.ReadBytes)).String()).
					Str("expected_hash", res.ExpectedHash).
					Str("actual_hash", res.ActualHash)

				if err != nil {
					logger = logger.Err(err)
				}

				return logger
			}

			if res.Status == checksum.HashMismatch {
				commonFields(log.Warn(), res.Err).Msg("Mismatch")
				return
			}

			if res.Status == checksum.Unreadable {
				commonFields(log.Error(), res.Err).Msg("Unreadable")
				return
			}

			commonFields(log.Info(), nil).Msg("Matched")
		},
	}

	log.Info().
		Str("checksum_file", checksumFile).
		Msg("Starting verification")

	result, err := action.VerifyChecksums(ctx, cfg)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Warn().Msg("Verification canceled")
			return nil
		}

		return fmt.Errorf("failed to verify checksums: %w", err)
	}

	stats := result.Stats
	log.Info().
		Int("matched", stats.Matched).
		Int("mismatch", stats.Mismatch).
		Int("unreadable", stats.Unreadable).
		Int("pending", stats.Pending()).
		Int("total_files", stats.TotalFiles).
		Msg("Verification stats")

	log.Info().Msg("Verification completed")

	return nil
}

var verifyCmd = &cobra.Command{
	Use:   "verify <checksum_file>",
	Short: "Verify files against checksum file",
	Long: strings.Trim(dedent.Dedent(`
		Verify files against checksum file.
		Algorithm is determined automatically from file extension:
		.sfv, .md4, .md5, .sha1, .sha256, .sha384, .sha512, .sha3-256, .sha3-384, .sha3-512, .blake3.`,
	), "\n"),
	Args: cobra.ExactArgs(1),
	RunE: runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
