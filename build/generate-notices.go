//go:build ignore

// Generates THIRD_PARTY_NOTICES file with license information
// Creates a comprehensive report of all third-party dependencies and their licenses
// Copyright (c) 2026 Ostap Konstantinov. Licensed under MIT License.

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	srcDir              = "../src"
	outputFile          = "../THIRD_PARTY_NOTICES"
	goLicenses          = "../.bin/go-licenses"
	extraDepsFile       = "extra-dependencies.txt"
	httpTimeout         = 30 * time.Second
	licenseNotAvailable = "License text not available"
)

var httpClient = &http.Client{
	Timeout: httpTimeout,
}

type LicenseInfo struct {
	Module      string
	Version     string
	LicenseType string
	LicenseText string
	LicenseURL  string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ THIRD_PARTY_NOTICES generated successfully")
}

func run() error {
	fmt.Println("Getting license report...")
	licenses, err := getLicenses()
	if err != nil {
		return err
	}

	fmt.Println("Loading extra dependencies...")
	extraLicenses, err := loadExtraDependencies()
	if err != nil {
		return err
	}

	licenses = append(licenses, extraLicenses...)

	// Remove duplicates
	seen := make(map[string]bool)
	unique := make([]LicenseInfo, 0, len(licenses))
	for _, lic := range licenses {
		if !seen[lic.Module] {
			seen[lic.Module] = true
			unique = append(unique, lic)
		}
	}
	licenses = unique

	slices.SortFunc(licenses, func(a, b LicenseInfo) int {
		return strings.Compare(a.Module, b.Module)
	})

	if err := writeReport(licenses); err != nil {
		return err
	}

	return nil
}

func getLicenses() ([]LicenseInfo, error) {
	cmd := exec.Command(goLicenses, "report", "./...")
	cmd.Dir = srcDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go-licenses report: %w", err)
	}

	var licenses []LicenseInfo
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" || strings.Contains(line, "Unknown") {
			continue
		}

		parts := strings.SplitN(line, ",", 3)
		if len(parts) != 3 {
			continue
		}

		module, licenseURL, licenseType := parts[0], parts[1], parts[2]
		version, rawURL := extractRepoInfo(licenseURL)
		licenseText := downloadFile(rawURL)

		licenses = append(licenses, LicenseInfo{
			Module:      module,
			Version:     version,
			LicenseType: licenseType,
			LicenseText: licenseText,
			LicenseURL:  licenseURL,
		})
	}

	return licenses, nil
}

func writeReport(licenses []LicenseInfo) error {
	var sb strings.Builder

	sb.WriteString("THIRD-PARTY SOFTWARE NOTICES AND INFORMATION\n")
	sb.WriteString("============================================\n\n")
	sb.WriteString("This project incorporates components from the projects listed below.\n\n")
	sb.WriteString("Table of Contents\n")
	sb.WriteString("-----------------\n")

	for i, lic := range licenses {
		sb.WriteString(fmt.Sprintf("%d. %s (%s) - %s\n", i+1, lic.Module, lic.Version, lic.LicenseType))
	}
	sb.WriteString("\n")

	for i, lic := range licenses {
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("%d. %s (%s) - %s\n", i+1, lic.Module, lic.Version, lic.LicenseType))
		sb.WriteString(strings.Repeat("=", 80) + "\n\n")
		repoLine := "Repository: " + lic.LicenseURL
		if len(repoLine) <= 80 {
			sb.WriteString(repoLine + "\n\n")
		}
		sb.WriteString("LICENSE TEXT:\n")
		sb.WriteString(strings.Repeat("-", 80) + "\n")
		sb.WriteString(wrapText(lic.LicenseText, 80) + "\n")
		sb.WriteString(strings.Repeat("-", 80) + "\n\n")
	}

	return os.WriteFile(outputFile, []byte(sb.String()), 0644)
}

func wrapText(text string, maxWidth int) string {
	if text == "" || text == licenseNotAvailable {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(wrapLine(line, maxWidth))
	}

	return result.String()
}

func wrapLine(line string, maxWidth int) string {
	if len(line) <= maxWidth {
		return line
	}

	var result strings.Builder
	words := strings.Fields(line)
	currentLen := 0

	for _, word := range words {
		wordLen := len(word)
		if currentLen == 0 {
			result.WriteString(word)
			currentLen = wordLen
		} else if currentLen+1+wordLen <= maxWidth {
			result.WriteString(" " + word)
			currentLen += 1 + wordLen
		} else {
			result.WriteString("\n" + word)
			currentLen = wordLen
		}
	}

	return result.String()
}

func loadExtraDependencies() ([]LicenseInfo, error) {
	extraDepsPath := extraDepsFile
	if _, err := os.Stat(extraDepsPath); os.IsNotExist(err) {
		scriptDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			scriptDir = "."
		}
		extraDepsPath = filepath.Join(scriptDir, extraDepsFile)
	}

	if _, err := os.Stat(extraDepsPath); os.IsNotExist(err) {
		fmt.Printf("Extra dependencies file not found: %s (skipping)\n", extraDepsPath)
		return nil, nil
	}

	file, err := os.Open(extraDepsPath)
	if err != nil {
		return nil, fmt.Errorf("open extra dependencies file: %w", err)
	}
	defer file.Close()

	var licenses []LicenseInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ",", 3)
		if len(parts) != 3 {
			fmt.Printf("Warning: skipping invalid line %d in extra dependencies: %s\n", lineNum, line)
			continue
		}

		module, licenseURL, licenseType := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])
		version, rawURL := extractRepoInfo(licenseURL)

		if rawURL == licenseNotAvailable {
			fmt.Printf("Warning: cannot extract license URL for %s: %s\n", module, licenseURL)
		}

		licenseText := downloadFile(rawURL)

		licenses = append(licenses, LicenseInfo{
			Module:      module,
			Version:     version,
			LicenseType: licenseType,
			LicenseText: licenseText,
			LicenseURL:  licenseURL,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read extra dependencies file: %w", err)
	}

	return licenses, nil
}

func extractRepoInfo(licenseURL string) (version, rawURL string) {
	if strings.Contains(licenseURL, "raw.githubusercontent.com") {
		// https://raw.githubusercontent.com/owner/repo/VERSION/LICENSE
		// or https://raw.githubusercontent.com/owner/repo/refs/tags/VERSION/LICENSE
		parts := strings.Split(strings.TrimPrefix(licenseURL, "https://raw.githubusercontent.com/"), "/")
		if len(parts) >= 3 {
			repoPath := parts[0] + "/" + parts[1]
			// Handle /refs/tags/VERSION or /refs/heads/VERSION format
			if len(parts) > 3 && parts[2] == "refs" && len(parts) > 4 {
				version = parts[4]
				path := strings.Join(parts[5:], "/")
				rawURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repoPath, version, path)
			} else {
				version = parts[2]
				path := strings.Join(parts[3:], "/")
				rawURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repoPath, version, path)
			}
		}
	} else if strings.Contains(licenseURL, "github.com") {
		// https://github.com/owner/repo/blob/VERSION/LICENSE
		//   -> https://raw.githubusercontent.com/owner/repo/VERSION/LICENSE
		if idx := strings.Index(licenseURL, "/blob/"); idx >= 0 {
			rest := licenseURL[idx+6:]
			if slashIdx := strings.Index(rest, "/"); slashIdx > 0 {
				version = rest[:slashIdx]
				path := rest[slashIdx+1:]
				repo := licenseURL[:idx]
				repoPath := strings.TrimPrefix(repo, "https://github.com/")
				rawURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repoPath, version, path)
			}
		}
	} else if strings.Contains(licenseURL, "cs.opensource.google") {
		// https://cs.opensource.google/go/x/crypto/+/v0.48.0:LICENSE
		//   -> https://raw.githubusercontent.com/golang/crypto/v0.48.0/LICENSE
		if idx := strings.Index(licenseURL, "/+/"); idx >= 0 {
			rest := licenseURL[idx+3:]
			if colonIdx := strings.Index(rest, ":"); colonIdx > 0 {
				version = rest[:colonIdx]
				path := rest[colonIdx+1:]
				parts := strings.Split(licenseURL, "/")
				for i, part := range parts {
					if part == "go" && i+2 < len(parts) && parts[i+1] == "x" {
						rawURL = fmt.Sprintf("https://raw.githubusercontent.com/golang/%s/%s/%s", parts[i+2], version, path)
						break
					}
				}
			}
		}
	} else if strings.Contains(licenseURL, "gist.githubusercontent.com") {
		// https://gist.githubusercontent.com/user/id/raw/commit/file
		// Already a raw URL, use 'master' as version
		version = "master"
		rawURL = licenseURL
	}

	if rawURL == "" {
		rawURL = licenseNotAvailable
	}
	return
}

func downloadFile(fileURL string) string {
	if fileURL == "" || fileURL == licenseNotAvailable {
		return licenseNotAvailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return licenseNotAvailable
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return licenseNotAvailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return licenseNotAvailable
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return licenseNotAvailable
	}

	return strings.TrimSpace(string(body))
}
