package coctyl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saintedlama/coctyl/pkg/coctyl"
)

func TestIntegrityLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	sumFile := filepath.Join(tempDir, "coctyl.sum")

	// Create a couple of Go files
	fileA := filepath.Join(tempDir, "a.go")
	fileB := filepath.Join(tempDir, "b.go")
	testFile := filepath.Join(tempDir, "b_test.go")

	codeA := "package test\n\nfunc Add(x, y int) int { return x + y }\n"
	codeB := "package test\n\nfunc Sub(x, y int) int { return x - y }\n"
	codeTest := "package test\n\nfunc TestDummy() {}\n"

	if err := os.WriteFile(fileA, []byte(codeA), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fileB, []byte(codeB), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, []byte(codeTest), 0644); err != nil {
		t.Fatal(err)
	}

	opts := coctyl.CheckOptions{
		IntegrityFile: sumFile,
		IncludeTests:  false,
	}

	// 1. Initial run: coctyl.sum does not exist -> should CREATE it
	summary, err := coctyl.CheckIntegrity([]string{tempDir}, opts)
	if err != nil {
		t.Fatalf("unexpected error creating integrity file: %v", err)
	}

	if !summary.Created {
		t.Errorf("expected summary.Created to be true")
	}
	// b_test.go should be excluded since IncludeTests is false
	if summary.Total != 2 {
		t.Errorf("expected 2 files recorded (excluding test), got: %d", summary.Total)
	}
	if summary.Failed != 0 {
		t.Errorf("expected 0 failed, got: %d", summary.Failed)
	}

	// Verify file was written on disk
	if _, err := os.Stat(sumFile); err != nil {
		t.Fatalf("expected integrity file %s to exist on disk: %v", sumFile, err)
	}

	// 2. Second run: coctyl.sum exists and files are unmodified -> should PASS
	checkSummary, err := coctyl.CheckIntegrity([]string{tempDir}, opts)
	if err != nil {
		t.Fatalf("unexpected error checking integrity: %v", err)
	}
	if checkSummary.Created {
		t.Errorf("expected Created to be false on existing file")
	}
	if checkSummary.Failed != 0 || checkSummary.Passed != 2 {
		t.Errorf("expected 2 passed and 0 failed, got passed=%d failed=%d", checkSummary.Passed, checkSummary.Failed)
	}

	// 3. Refactor variable names in a.go -> should STILL PASS (structural equality)
	refactoredA := "package test\n\nfunc Add(first, second int) int { return first + second }\n"
	if err := os.WriteFile(fileA, []byte(refactoredA), 0644); err != nil {
		t.Fatal(err)
	}

	refactorSummary, err := coctyl.CheckIntegrity([]string{tempDir}, opts)
	if err != nil {
		t.Fatalf("unexpected error checking refactored file: %v", err)
	}
	if refactorSummary.Failed != 0 {
		t.Errorf("expected refactored code to pass structural integrity check, but failed: %+v", refactorSummary.Results)
	}

	// 4. Introduce functional mutation to b.go -> should FAIL with MISMATCH
	mutatedB := "package test\n\nfunc Sub(x, y int) int { return x + y }\n"
	if err := os.WriteFile(fileB, []byte(mutatedB), 0644); err != nil {
		t.Fatal(err)
	}

	mismatchSummary, err := coctyl.CheckIntegrity([]string{tempDir}, opts)
	if err != nil {
		t.Fatalf("unexpected error checking mutated file: %v", err)
	}
	if mismatchSummary.Failed != 1 {
		t.Errorf("expected 1 failure on mutated file, got %d", mismatchSummary.Failed)
	}
	var bResult *coctyl.FileCheckResult
	for i := range mismatchSummary.Results {
		if filepath.Base(mismatchSummary.Results[i].Path) == "b.go" {
			bResult = &mismatchSummary.Results[i]
			break
		}
	}
	if bResult == nil || bResult.Status != coctyl.StatusMismatch {
		t.Errorf("expected b.go status to be MISMATCH, got: %+v", bResult)
	}

	// 5. Update integrity file with opts.Update = true -> should UPDATE and pass
	opts.Update = true
	updateSummary, err := coctyl.CheckIntegrity([]string{tempDir}, opts)
	if err != nil {
		t.Fatalf("unexpected error updating integrity file: %v", err)
	}
	if !updateSummary.Updated {
		t.Errorf("expected updateSummary.Updated to be true")
	}

	// 6. Check again with update=false -> should pass now
	opts.Update = false
	recheckSummary, err := coctyl.CheckIntegrity([]string{tempDir}, opts)
	if err != nil {
		t.Fatalf("unexpected error after update: %v", err)
	}
	if recheckSummary.Failed != 0 {
		t.Errorf("expected 0 failures after update, got: %d", recheckSummary.Failed)
	}
}

func TestIntegrityMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	sumFile := filepath.Join(tempDir, "coctyl.sum")

	fileA := filepath.Join(tempDir, "a.go")
	if err := os.WriteFile(fileA, []byte("package test\nfunc F() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := coctyl.CheckOptions{IntegrityFile: sumFile}
	_, err := coctyl.CheckIntegrity([]string{tempDir}, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Now delete a.go
	if err := os.Remove(fileA); err != nil {
		t.Fatal(err)
	}

	summary, err := coctyl.CheckIntegrity(nil, opts)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Failed != 1 {
		t.Errorf("expected 1 failure for missing file, got %d", summary.Failed)
	}
	if summary.Results[0].Status != coctyl.StatusMissing {
		t.Errorf("expected status MISSING, got: %v", summary.Results[0].Status)
	}
}
