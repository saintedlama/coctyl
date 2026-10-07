package main

import (
	"fmt"

	"github.com/saintedlama/coctyl/pkg/coctyl"
	"github.com/spf13/cobra"
)

func newHashCmd() *cobra.Command {
	var includeImports bool

	cmd := &cobra.Command{
		Use:   "hash <file.go>",
		Short: "Compute the structural dactyl hash of a Go file",
		Long: `Compute the structural dactyl hash (SHA-256) of a Go source file.

The source is parsed into an AST and alpha-normalized:
  • Local variables and parameters are mapped to canonical indexed slots ($0, $1, ...)
  • Dataflow integrity is preserved (a + b != a + a)
  • Comments, whitespace, and formatting differences are discarded
  • Import declarations are ignored by default to allow code to move dynamically`,
		Example: `  # Hash a file (imports ignored by default)
  coctyl hash main.go

  # Hash a file and explicitly include import declarations
  coctyl hash -i main.go`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			opts := coctyl.Options{
				IncludeImports: includeImports,
			}

			res, err := coctyl.HashFile(filePath, opts)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), res.Hash)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&includeImports, "include-imports", "i", false, "include import declarations in the dactyl (ignored by default)")

	return cmd
}
