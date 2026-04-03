package app

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ostapkonst/HashVerifier/internal/checksum"
)

type PathType int

const (
	PathTypeDirectory PathType = iota
	PathTypeChecksumFile
	PathTypeInvalid
)

type PathResolver struct{}

func NewPathResolver() *PathResolver {
	return &PathResolver{}
}

func (pr *PathResolver) Resolve(path string) (PathType, string) {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." {
		return PathTypeInvalid, ""
	}

	fileInfo, err := os.Stat(cleanPath)
	if err == nil && fileInfo.IsDir() {
		return PathTypeDirectory, cleanPath
	}

	if _, err = checksum.AlgorithmFromExtension(cleanPath); err != nil {
		return PathTypeDirectory, cleanPath
	}

	return PathTypeChecksumFile, cleanPath
}

func URIToFilePath(uri string) (string, error) {
	uri = strings.TrimRight(strings.TrimSpace(uri), "\r\n")
	if uri == "" {
		return "", fmt.Errorf("empty URI")
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse URI: %w", err)
	}

	if parsedURL.Scheme != "file" {
		return "", fmt.Errorf("unsupported URI scheme: %s", parsedURL.Scheme)
	}

	path := parsedURL.Path
	if runtime.GOOS == "windows" {
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}

		path = filepath.FromSlash(path)
	}

	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		return "", fmt.Errorf("failed to unescape path: %w", err)
	}

	return decodedPath, nil
}
