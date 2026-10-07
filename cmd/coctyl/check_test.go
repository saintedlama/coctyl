package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLICheckWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	sumFile := filepath.Join(tempDir, "coctyl.sum")
	srcFile := filepath.Join(tempDir, "math.go")

	code := `package mymath

func Multiply(a, b int) int {
	return a * b
}
`
	if err := os.WriteFile(srcFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Initial check: sumFile doesn't exist -> should create it
	cmdCreate := exec.Command("go", "run", "./cmd/coctyl", "check", "-f", sumFile, tempDir)
	cmdCreate.Dir = "../.."
	outCreate, err := cmdCreate.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to create check file: %v\nOutput: %s", err, string(outCreate))
	}

	createStr := string(outCreate)
	if !strings.Contains(createStr, "Created integrity file") {
		t.Errorf("expected creation message, got: %s", createStr)
	}

	// Verify sum file contents on disk
	sumContent, err := os.ReadFile(sumFile)
	if err != nil {
		t.Fatalf("expected sum file to exist: %v", err)
	}
	if !strings.Contains(string(sumContent), "math.go") {
		t.Errorf("expected math.go in sum file, got: %s", string(sumContent))
	}

	// 2. Second check: verify unchanged file -> should pass
	cmdCheck := exec.Command("go", "run", "./cmd/coctyl", "check", "-f", sumFile, tempDir)
	cmdCheck.Dir = "../.."
	outCheck, err := cmdCheck.CombinedOutput()
	if err != nil {
		t.Fatalf("expected check to pass, got error: %v\nOutput: %s", err, string(outCheck))
	}

	if !strings.Contains(string(outCheck), "verified successfully") {
		t.Errorf("expected verified successfully message, got: %s", string(outCheck))
	}

	// 3. Rename variables (alpha equivalence) -> check should STILL pass!
	refactoredCode := `package mymath

func Multiply(x, y int) int {
	return x * y
}
`
	if err := os.WriteFile(srcFile, []byte(refactoredCode), 0644); err != nil {
		t.Fatal(err)
	}

	cmdRefactor := exec.Command("go", "run", "./cmd/coctyl", "check", "-f", sumFile, tempDir)
	cmdRefactor.Dir = "../.."
	outRefactor, err := cmdRefactor.CombinedOutput()
	if err != nil {
		t.Fatalf("expected refactored code to pass check, got error: %v\nOutput: %s", err, string(outRefactor))
	}
	if !strings.Contains(string(outRefactor), "verified successfully") {
		t.Errorf("expected verified successfully message, got: %s", string(outRefactor))
	}

	// 4. Functional mutation -> check should FAIL with mismatch
	mutatedCode := `package mymath

func Multiply(x, y int) int {
	return x + y // Changed * to +
}
`
	if err := os.WriteFile(srcFile, []byte(mutatedCode), 0644); err != nil {
		t.Fatal(err)
	}

	cmdMutated := exec.Command("go", "run", "./cmd/coctyl", "check", "-f", sumFile, tempDir)
	cmdMutated.Dir = "../.."
	outMutated, err := cmdMutated.CombinedOutput()
	if err == nil {
		t.Fatalf("expected check to fail on mutated logic, but exited with success! Output: %s", string(outMutated))
	}

	mutatedStr := string(outMutated)
	if !strings.Contains(mutatedStr, "FAILED (hash mismatch)") {
		t.Errorf("expected 'FAILED (hash mismatch)' in output, got: %s", mutatedStr)
	}

	// 5. Update with -u -> should update integrity file and succeed
	cmdUpdate := exec.Command("go", "run", "./cmd/coctyl", "check", "-f", sumFile, "-u", tempDir)
	cmdUpdate.Dir = "../.."
	outUpdate, err := cmdUpdate.CombinedOutput()
	if err != nil {
		t.Fatalf("expected update to succeed, got error: %v\nOutput: %s", err, string(outUpdate))
	}
	if !strings.Contains(string(outUpdate), "Updated integrity file") {
		t.Errorf("expected 'Updated integrity file' in output, got: %s", string(outUpdate))
	}
}