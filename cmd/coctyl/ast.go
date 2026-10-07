package main

import (
	"fmt"

	"github.com/saintedlama/coctyl/pkg/coctyl"
	"github.com/spf13/cobra"
)

func newASTCmd() *cobra.Command {
	var includeImports bool
	var includePackage bool

	cmd := &cobra.Command{
		Use:   "ast <file.go>",
		Short: "Print the normalized canonical AST of a Go file",
		Long: `Print the normalized canonical AST (S-expression) of a Go source file.

The source is parsed into an AST and alpha-normalized:
  • Local variables and parameters are mapped to scope-order indices ($0, $1, ...)
  • Comments, whitespace, and formatting differences are discarded
  • Package declaration and imports are omitted by default for functional equivalence`,
		Example: `  # Print normalized AST (imports and package omitted by default)
  coctyl ast main.go

  # Print normalized AST including import declarations
  coctyl ast -i main.go

  # Print normalized AST including package declaration
  coctyl ast -p main.go`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			opts := coctyl.Options{
				IncludeImports: includeImports,
				IncludePackage: includePackage,
			}

			astStr, err := coctyl.CanonicalASTFile(filePath, opts)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), astStr)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&includeImports, "include-imports", "i", false, "include import declarations in the AST output")
	cmd.Flags().BoolVarP(&includePackage, "include-package", "p", false, "include package declaration in the AST output")

	return cmd
}
