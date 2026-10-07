package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "coctyl",
	Short: "Dactyloscopy for your Go source code",
	Long: `coctyl 🪪
Dactyloscopy for your Go source code. Hash the logic, ignore the labels.

coctyl parses Go code into an AST, performs lexical scope-aware
alpha-normalization (canonical variable indexing), and computes a robust
SHA-256 fingerprint—a dactyl.

Two functions with completely different variable names and formatting will
produce identical dactyls as long as their structural logic is identical.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(newHashCmd())
	rootCmd.AddCommand(newASTCmd())
	rootCmd.AddCommand(newCheckCmd())
}
