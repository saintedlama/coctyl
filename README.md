# coctyl 🪪

> Dactyloscopy for your Go source code. Hash the logic, ignore the labels.

`coctyl` is a structural fingerprinting CLI and library for Go. It parses your source code into an Abstract Syntax Tree (AST), canonicalizes local variable bindings via scope-aware alpha-normalization ($\alpha$-conversion), and generates a robust SHA-256 signature—a **dactyl**.

If two functions have entirely different variable names but perform the exact same structural logic, `coctyl` guarantees they will produce the exact same fingerprint. Crucially, `coctyl` preserves dataflow integrity: `a + b` and `a + a` produce different fingerprints.

## Why use coctyl?

Traditional checksums break the moment you rename a variable, reorder comments, or adjust formatting. `coctyl` is designed for:
* **Refactoring Verification:** Prove that renaming local variables, parameters, or receivers didn't alter the structural flow of a critical function.
* **Duplicate Detection:** Discover identical logic hiding behind renamed variables and reformatted syntax.
* **Smart Caching:** Rebuild or retest components only when *structural logic* changes.
* **Relocation & Moving Code:** By default, import declarations are excluded from the fingerprint, allowing code parts to move dynamically across files or packages without invalidating structural fingerprints.

## How it Works

`coctyl` takes a structural "cast" of your Go AST:

1. **Parse:** Reads `.go` files using `go/parser` and constructs the standard AST.
2. **Alpha-Normalize (Scope-Aware Indexing):** Walks lexical scopes (parameters, receivers, short variable declarations `:=`, block scopes, and closures). Local identifiers are mapped to deterministic canonical slots (`$0`, `$1`, ...), while preserving language built-ins (`len`, `make`), standard library packages, and struct field selectors.
3. **Skeletonize:** Structural nodes (control flow, assignments, expressions, operators) are serialized into a deterministic, formatting-agnostic canonical representation. Comments, source positions, and redundant parenthesization are discarded. Import declarations are ignored by default.
4. **Hash:** The canonical skeleton is passed through SHA-256 to produce the final **dactyl**.

## Installation

```bash
go install github.com/saintedlama/coctyl/cmd/coctyl@latest
```

## CLI Usage

### View Available Commands

```bash
coctyl --help
```

```text
coctyl 🪪
Dactyloscopy for your Go source code. Hash the logic, ignore the labels.

coctyl parses Go code into an AST, performs lexical scope-aware
alpha-normalization (canonical variable indexing), and computes a robust
SHA-256 fingerprint—a dactyl.

Two functions with completely different variable names and formatting will
produce identical dactyls as long as their structural logic is identical.

Usage:
  coctyl [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  hash        Compute the structural dactyl hash of a Go file
  help        Help about any command

Flags:
  -h, --help   help for coctyl

Use "coctyl [command] --help" for more information about a command.
```

### Compute the dactyl hash of a Go file

```bash
coctyl hash path/to/file.go
```

```text
93315cb160ced58a571006e580aacbc51ef6621d7b6cd864a6e7a49d7625fd38
```

By default, imports are ignored so code can move freely without altering the fingerprint. Output is the raw hash, making it easily pipeable to other Unix tools.

### Include imports explicitly

If you specifically want import statements included in the fingerprint, use `-i` or `--include-imports`:

```bash
coctyl hash -i path/to/file.go
```

## Library Usage

```go
package main

import (
	"fmt"
	"github.com/saintedlama/coctyl/pkg/coctyl"
)

func main() {
	res, err := coctyl.HashFile("main.go", coctyl.Options{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Dactyl: %s\n", res.Hash)
}
```
