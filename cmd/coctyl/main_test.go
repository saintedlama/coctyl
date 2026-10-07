package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIHashCommand(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "sample.go")
	content := `package sample

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test source: %v", err)
	}

	cmd := exec.Command("go", "run", "./cmd/coctyl", "hash", srcPath)
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\nOutput: %s", err, string(out))
	}

	hashStr := strings.TrimSpace(string(out))
	if len(hashStr) != 64 {
		t.Errorf("expected 64-character SHA-256 hash output, got: %q", hashStr)
	}
}

func TestCLIHelpOutput(t *testing.T) {
	// Root help
	cmd := exec.Command("go", "run", "./cmd/coctyl", "--help")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help failed: %v, %s", err, string(out))
	}

	outStr := string(out)
	if !strings.Contains(outStr, "Dactyloscopy for your Go source code") {
		t.Errorf("expected root help description, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Available Commands:") || !strings.Contains(outStr, "hash") {
		t.Errorf("expected available commands in help output, got: %s", outStr)
	}

	// Subcommand help
	cmdHashHelp := exec.Command("go", "run", "./cmd/coctyl", "hash", "--help")
	cmdHashHelp.Dir = "../.."
	outHashHelp, err := cmdHashHelp.CombinedOutput()
	if err != nil {
		t.Fatalf("hash help failed: %v, %s", err, string(outHashHelp))
	}

	hashHelpStr := string(outHashHelp)
	if !strings.Contains(hashHelpStr, "--include-imports") || !strings.Contains(hashHelpStr, "-i") {
		t.Errorf("expected --include-imports / -i in hash help, got: %s", hashHelpStr)
	}
	if !strings.Contains(hashHelpStr, "Examples:") {
		t.Errorf("expected Examples section in hash help, got: %s", hashHelpStr)
	}
}

func TestCLIHashMissingArgs(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/coctyl", "hash")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("expected non-zero exit code when calling hash without arguments")
	}
	if !strings.Contains(string(out), "accepts 1 arg(s), received 0") {
		t.Errorf("expected Cobra exact args error, got: %s", string(out))
	}
}

func TestCLIHashIgnoreImportsDefaultAndShortFlag(t *testing.T) {
	tempDir := t.TempDir()
	srcPath1 := filepath.Join(tempDir, "file1.go")
	srcPath2 := filepath.Join(tempDir, "file2.go")

	file1 := `package sample

import "fmt"

func Hello() {
	fmt.Println("hi")
}
`
	file2 := `package sample

import (
	"fmt"
	"strings"
)

func Hello() {
	fmt.Println("hi")
}
`
	if err := os.WriteFile(srcPath1, []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcPath2, []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. By default, 'coctyl hash FILE' ignores imports, so outputs MUST be identical!
	cmd1 := exec.Command("go", "run", "./cmd/coctyl", "hash", srcPath1)
	cmd1.Dir = "../.."
	out1, err := cmd1.CombinedOutput()
	if err != nil {
		t.Fatalf("cmd1 failed: %v, %s", err, out1)
	}

	cmd2 := exec.Command("go", "run", "./cmd/coctyl", "hash", srcPath2)
	cmd2.Dir = "../.."
	out2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("cmd2 failed: %v, %s", err, out2)
	}

	h1 := strings.TrimSpace(string(out1))
	h2 := strings.TrimSpace(string(out2))

	if h1 != h2 {
		t.Errorf("expected identical hashes by default (imports ignored), got %s and %s", h1, h2)
	}

	// 2. When passing short flag -i, hashes MUST differ!
	cmdWith1 := exec.Command("go", "run", "./cmd/coctyl", "hash", "-i", srcPath1)
	cmdWith1.Dir = "../.."
	outWith1, err := cmdWith1.CombinedOutput()
	if err != nil {
		t.Fatalf("cmdWith1 failed: %v, %s", err, outWith1)
	}

	cmdWith2 := exec.Command("go", "run", "./cmd/coctyl", "hash", "-i", srcPath2)
	cmdWith2.Dir = "../.."
	outWith2, err := cmdWith2.CombinedOutput()
	if err != nil {
		t.Fatalf("cmdWith2 failed: %v, %s", err, outWith2)
	}

	hWith1 := strings.TrimSpace(string(outWith1))
	hWith2 := strings.TrimSpace(string(outWith2))

	if hWith1 == hWith2 {
		t.Errorf("expected different hashes with -i, but got identical %s", hWith1)
	}
}

func TestCLIASTCommand(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "sample.go")
	content := `package mypkg

func Calc(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test source: %v", err)
	}

	// 1. Basic ast command
	cmd := exec.Command("go", "run", "./cmd/coctyl", "ast", srcPath)
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ast command failed: %v\nOutput: %s", err, string(out))
	}

	astOutput := strings.TrimSpace(string(out))
	if !strings.HasPrefix(astOutput, "(File (Pkg _)") {
		t.Errorf("expected package to be ignored by default, got: %s", astOutput)
	}
	if !strings.Contains(astOutput, "(FuncDecl (Name $0)") {
		t.Errorf("expected canonical func name in AST output, got: %s", astOutput)
	}

	// 2. With -p include-package
	cmdPkg := exec.Command("go", "run", "./cmd/coctyl", "ast", "-p", srcPath)
	cmdPkg.Dir = "../.."
	outPkg, err := cmdPkg.CombinedOutput()
	if err != nil {
		t.Fatalf("ast -p command failed: %v\nOutput: %s", err, string(outPkg))
	}

	astPkgOutput := strings.TrimSpace(string(outPkg))
	if !strings.HasPrefix(astPkgOutput, "(File (Pkg mypkg)") {
		t.Errorf("expected package name to be preserved with -p, got: %s", astPkgOutput)
	}
}
