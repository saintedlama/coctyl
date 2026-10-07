package main

import (
	"fmt"

	"github.com/saintedlama/coctyl/pkg/coctyl"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	var (
		integrityFile  string
		update         bool
		includeImports bool
		includeTests   bool
		quiet          bool
	)

	cmd := &cobra.Command{
		Use:   "check [paths...]",
		Short: "Verify or generate structural dactyl integrity checksums",
		Long: `Verify or create an integrity file (default: coctyl.sum) for Go source files.

If the integrity file does not exist, coctyl calculates the dactyl hashes of the
specified paths (or the current directory by default) and creates it.

If the integrity file exists, coctyl verifies all recorded or specified files against
their stored dactyls. Refactored variable names and comments remain equivalent,
while any functional or structural changes are flagged as mismatches.`,
		Example: `  # Check current directory against coctyl.sum (creates it if missing)
  coctyl check

  # Check specific packages or files
  coctyl check ./pkg/... main.go

  # Use a custom integrity file
  coctyl check -f .coctyl.sum

  # Force regenerate/update the integrity file
  coctyl check -u

  # Only report failures
  coctyl check -q`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := coctyl.CheckOptions{
				Options: coctyl.Options{
					IncludeImports: includeImports,
				},
				IntegrityFile: integrityFile,
				IncludeTests:  includeTests,
				Update:        update,
			}

			summary, err := coctyl.CheckIntegrity(args, opts)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()

			if summary.Created {
				fmt.Fprintf(out, "Created integrity file %s (%d file(s) recorded).\n", integrityFile, summary.Total)
				return nil
			}

			if summary.Updated {
				fmt.Fprintf(out, "Updated integrity file %s (%d file(s) recorded).\n", integrityFile, summary.Total)
				return nil
			}

			// Report verification results
			for _, r := range summary.Results {
				switch r.Status {
				case coctyl.StatusOk:
					if !quiet {
						fmt.Fprintf(out, "%s: OK\n", r.Path)
					}
				case coctyl.StatusMismatch:
					fmt.Fprintf(out, "%s: FAILED (hash mismatch)\n", r.Path)
				case coctyl.StatusMissing:
					fmt.Fprintf(out, "%s: FAILED (file not found)\n", r.Path)
				case coctyl.StatusUnrecorded:
					fmt.Fprintf(out, "%s: FAILED (unrecorded file)\n", r.Path)
				}
			}

			if summary.Failed > 0 {
				return fmt.Errorf("coctyl: %d of %d file(s) failed integrity check. Run 'coctyl check -u' to update", summary.Failed, summary.Total)
			}

			fmt.Fprintf(out, "coctyl: all %d file(s) verified successfully.\n", summary.Passed)
			return nil
		},
	}

	cmd.Flags().StringVarP(&integrityFile, "file", "f", "coctyl.sum", "path to integrity checksum file")
	cmd.Flags().BoolVarP(&update, "update", "u", false, "update/overwrite integrity file with current hashes")
	cmd.Flags().BoolVarP(&includeImports, "include-imports", "i", false, "include import declarations in dactyl calculation")
	cmd.Flags().BoolVar(&includeTests, "include-tests", false, "include *_test.go files when scanning directories")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "only print failing files")

	return cmd
}
