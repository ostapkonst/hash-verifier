package checksum

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
)

func WalkDir(ctx context.Context, path string) ([]string, error) {
	var files []string

	err := godirwalk.Walk(path, &godirwalk.Options{
		FollowSymbolicLinks: true,
		Unsorted:            true,

		Callback: func(path string, de *godirwalk.Dirent) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if b, _ := de.IsDirOrSymlinkToDir(); b {
				return nil
			}

			files = append(files, path)

			return nil
		},

		ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
			select {
			case <-ctx.Done():
				return godirwalk.Halt
			default:
			}

			if errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrNotExist) {
				return godirwalk.SkipNode
			}

			return godirwalk.Halt
		},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if err != nil {
		return nil, err
	}

	return files, nil
}

func GetPrefixForFilesInChecksum(folder, file string) (string, error) {
	absFolder, err := filepath.Abs(folder)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for folder: %w", err)
	}

	absFile, err := filepath.Abs(file)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for file: %w", err)
	}

	if filepath.Dir(absFolder) == filepath.Dir(absFile) {
		return filepath.Base(absFolder), nil
	}

	return absFolder, nil
}

func FormatLine(relPath, hashStr string, algo Algorithm) string {
	switch algo {
	case CRC32:
		return fmt.Sprintf("%s %s", relPath, hashStr)
	default:
		return fmt.Sprintf("%s *%s", hashStr, relPath)
	}
}
