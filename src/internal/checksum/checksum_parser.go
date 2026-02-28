package checksum

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	hashFirstRe = regexp.MustCompile(`^([a-fA-F0-9]+)\s+\*?(.+)$`)
	sfvRe       = regexp.MustCompile(`^(.+?)\s+([a-fA-F0-9]{8})$`)
)

type CheckSumLine struct {
	RelPath string
	Hash    string
}

func ParseCheckSum(ctx context.Context, filename string, algo Algorithm) ([]CheckSumLine, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close() //nolint:errcheck

	var lines []CheckSumLine

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		relPath, hash, err := parseLine(line, algo)
		if err != nil {
			return nil, err
		}

		lines = append(lines, CheckSumLine{
			RelPath: relPath,
			Hash:    hash,
		})
	}

	return lines, scanner.Err()
}

func parseLine(line string, algo Algorithm) (relPath, expectedHash string, err error) {
	format := formatFromAlgorithm(algo)

	switch format {
	case FormatHashFirst:
		matches := hashFirstRe.FindStringSubmatch(line)
		if len(matches) != 3 {
			return "", "", fmt.Errorf("invalid hash-first line: %q", line)
		}

		expectedHash = matches[1]
		relPath = matches[2]
	case FormatPathFirst:
		matches := sfvRe.FindStringSubmatch(line)
		if len(matches) != 3 {
			return "", "", fmt.Errorf("invalid SFV line: %q", line)
		}

		relPath = matches[1]
		expectedHash = matches[2]
	default:
		return "", "", errors.New("unknown format")
	}

	if !isValidHashLength(expectedHash, algo) {
		return "", "", fmt.Errorf("invalid hash length %d for %s", len(expectedHash), algo.String())
	}

	return fixPathSeparator(relPath), expectedHash, nil
}

func fixPathSeparator(path string) string {
	// стараемся сделать путь кросс-платформенным...
	return strings.ReplaceAll(path, "\\", string(os.PathSeparator))
}
