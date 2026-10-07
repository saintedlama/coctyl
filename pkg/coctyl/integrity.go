package coctyl

import (
	"bufio"
	"fmt"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CheckStatus represents the verification status of an individual file.
type CheckStatus string

const (
	StatusOk         CheckStatus = "OK"
	StatusMismatch   CheckStatus = "MISMATCH"
	StatusMissing    CheckStatus = "MISSING"
	StatusUnrecorded CheckStatus = "UNRECORDED"
)

// FileCheckResult holds the verification result for a single file.
type FileCheckResult struct {
	Path     string
	Status   CheckStatus
	Expected string
	Actual   string
	Err      error
}

// CheckSummary holds the aggregate results of an integrity check or creation.
type CheckSummary struct {
	Created bool
	Updated bool
	Results []FileCheckResult
	Total   int
	Passed  int
	Failed  int
}

// CheckOptions configures the check/integrity behavior.
type CheckOptions struct {
	Options               // embeds IncludeImports
	IntegrityFile string  // path to the integrity file (default: "coctyl.sum")
	IncludeTests  bool    // include *_test.go files when walking directories
	Update        bool    // overwrite/update integrity file even if it exists
}

// IntegrityEntry represents a single hash-to-path mapping.
type IntegrityEntry struct {
	Hash string
	Path string
}

// ReadIntegrityFile parses a standard checksum file (format: `<hash>  <path>`).
// Empty lines and lines starting with '#' are ignored.
func ReadIntegrityFile(filePath string) (map[string]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries := make(map[string]string)
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: "<hash>  <path>" (two spaces, standard sha256sum format) or "<hash> <path>"
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid checksum line %d in %s: %s", lineNum, filePath, line)
		}
		hash := strings.TrimSpace(parts[0])
		targetPath := strings.TrimSpace(parts[1])

		entries[filepath.ToSlash(targetPath)] = hash
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading integrity file %s: %w", filePath, err)
	}

	return entries, nil
}

// WriteIntegrityFile writes a map of relative path -> hash to an integrity file,
// sorting entries alphabetically by path for deterministic output.
func WriteIntegrityFile(filePath string, entries map[string]string) error {
	paths := make([]string, 0, len(entries))
	for p := range entries {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var sb strings.Builder
	sb.WriteString("# coctyl integrity file\n")
	sb.WriteString("# Format: <dactyl-hash>  <relative-path>\n\n")

	for _, p := range paths {
		sb.WriteString(fmt.Sprintf("%s  %s\n", entries[p], filepath.ToSlash(p)))
	}

	// Ensure destination directory exists
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

// FindGoFiles expands a list of file/directory paths into a deduplicated, sorted list
// of relative Go file paths relative to baseDir.
func FindGoFiles(baseDir string, targetPaths []string, includeTests bool) ([]string, error) {
	if len(targetPaths) == 0 {
		targetPaths = []string{"."}
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolving base dir %s: %w", baseDir, err)
	}

	foundMap := make(map[string]struct{})

	for _, target := range targetPaths {
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return nil, fmt.Errorf("resolving target path %s: %w", target, err)
		}

		info, err := os.Stat(absTarget)
		if err != nil {
			return nil, fmt.Errorf("accessing path %s: %w", target, err)
		}

		if !info.IsDir() {
			if strings.HasSuffix(info.Name(), ".go") {
				rel, err := filepath.Rel(absBase, absTarget)
				if err != nil {
					return nil, err
				}
				foundMap[filepath.ToSlash(rel)] = struct{}{}
			}
			continue
		}

		// Walk directory
		err = filepath.WalkDir(absTarget, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if d.IsDir() {
				name := d.Name()
				if path != absTarget && strings.HasPrefix(name, ".") {
					return fs.SkipDir
				}
				if name == "vendor" || name == "node_modules" {
					return fs.SkipDir
				}
				return nil
			}

			if strings.HasSuffix(d.Name(), ".go") {
				if !includeTests && strings.HasSuffix(d.Name(), "_test.go") {
					return nil
				}
				rel, err := filepath.Rel(absBase, path)
				if err != nil {
					return err
				}
				foundMap[filepath.ToSlash(rel)] = struct{}{}
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scanning directory %s: %w", target, err)
		}
	}

	result := make([]string, 0, len(foundMap))
	for p := range foundMap {
		result = append(result, p)
	}
	sort.Strings(result)

	return result, nil
}

// CheckIntegrity performs integrity verification or file creation.
// If the integrity file does not exist (or if opts.Update is true), it hashes the target files
// and writes the integrity file.
// If the integrity file exists, it validates the files against the recorded hashes.
func CheckIntegrity(targetPaths []string, opts CheckOptions) (*CheckSummary, error) {
	integrityFile := opts.IntegrityFile
	if integrityFile == "" {
		integrityFile = "coctyl.sum"
	}

	baseDir := filepath.Dir(integrityFile)
	if baseDir == "" {
		baseDir = "."
	}

	_, statErr := os.Stat(integrityFile)
	doesNotExist := os.IsNotExist(statErr)

	// Mode 1: Create or Update integrity file
	if doesNotExist || opts.Update {
		files, err := FindGoFiles(baseDir, targetPaths, opts.IncludeTests)
		if err != nil {
			return nil, err
		}

		entries := make(map[string]string)
		results := make([]FileCheckResult, 0, len(files))

		for _, relPath := range files {
			fullPath := filepath.Join(baseDir, filepath.FromSlash(relPath))
			res, err := HashFile(fullPath, opts.Options)
			if err != nil {
				return nil, fmt.Errorf("hashing file %s: %w", fullPath, err)
			}
			entries[relPath] = res.Hash
			results = append(results, FileCheckResult{
				Path:     relPath,
				Status:   StatusOk,
				Actual:   res.Hash,
				Expected: res.Hash,
			})
		}

		if err := WriteIntegrityFile(integrityFile, entries); err != nil {
			return nil, fmt.Errorf("writing integrity file %s: %w", integrityFile, err)
		}

		return &CheckSummary{
			Created: doesNotExist,
			Updated: !doesNotExist && opts.Update,
			Results: results,
			Total:   len(results),
			Passed:  len(results),
			Failed:  0,
		}, nil
	}

	// Mode 2: Verify existing integrity file
	recordedEntries, err := ReadIntegrityFile(integrityFile)
	if err != nil {
		return nil, fmt.Errorf("reading integrity file: %w", err)
	}

	// Determine which files to check
	var filesToCheck []string
	if len(targetPaths) == 0 {
		// No specific paths provided: check all recorded files
		for p := range recordedEntries {
			filesToCheck = append(filesToCheck, p)
		}
		sort.Strings(filesToCheck)
	} else {
		// Specific paths provided: find matching files
		filesToCheck, err = FindGoFiles(baseDir, targetPaths, opts.IncludeTests)
		if err != nil {
			return nil, err
		}
	}

	summary := &CheckSummary{
		Results: make([]FileCheckResult, 0, len(filesToCheck)),
		Total:   len(filesToCheck),
	}

	for _, relPath := range filesToCheck {
		expectedHash, recorded := recordedEntries[relPath]
		fullPath := filepath.Join(baseDir, filepath.FromSlash(relPath))

		if !recorded {
			summary.Failed++
			summary.Results = append(summary.Results, FileCheckResult{
				Path:   relPath,
				Status: StatusUnrecorded,
			})
			continue
		}

		res, err := HashFile(fullPath, opts.Options)
		if err != nil {
			summary.Failed++
			if errors.Is(err, fs.ErrNotExist) {
				summary.Results = append(summary.Results, FileCheckResult{
					Path:     relPath,
					Status:   StatusMissing,
					Expected: expectedHash,
					Err:      err,
				})
			} else {
				summary.Results = append(summary.Results, FileCheckResult{
					Path:     relPath,
					Status:   StatusMismatch,
					Expected: expectedHash,
					Err:      err,
				})
			}
			continue
		}

		if res.Hash != expectedHash {
			summary.Failed++
			summary.Results = append(summary.Results, FileCheckResult{
				Path:     relPath,
				Status:   StatusMismatch,
				Expected: expectedHash,
				Actual:   res.Hash,
			})
		} else {
			summary.Passed++
			summary.Results = append(summary.Results, FileCheckResult{
				Path:     relPath,
				Status:   StatusOk,
				Expected: expectedHash,
				Actual:   res.Hash,
			})
		}
	}

	return summary, nil
}