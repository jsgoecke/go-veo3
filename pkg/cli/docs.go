package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate documentation",
		Long:  `Generate documentation in various formats (man pages, markdown, etc.)`,
	}

	cmd.AddCommand(newManCmd())
	cmd.AddCommand(newMarkdownCmd())

	return cmd
}

func newManCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "man",
		Short: "Generate man pages",
		Long: `Generate man pages for veo3 commands.

The man pages will be written to the specified directory (default: ./docs/man).

Example:
  veo3 docs man
  veo3 docs man --output /usr/local/share/man/man1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Parent().Parent()

			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputDir, 0750); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Generate man pages
			header := &doc.GenManHeader{
				Title:   "VEO3",
				Section: "1",
				Source:  "Veo3 CLI",
				Manual:  "Veo3 Manual",
			}

			if err := doc.GenManTree(root, header, outputDir); err != nil {
				return fmt.Errorf("failed to generate man pages: %w", err)
			}

			absPath, _ := filepath.Abs(outputDir)
			fmt.Printf("Man pages generated in: %s\n", absPath)
			fmt.Printf("\nTo install, run:\n")
			fmt.Printf("  sudo cp %s/*.1 /usr/local/share/man/man1/\n", absPath)
			fmt.Printf("  sudo mandb\n")
			fmt.Printf("\nThen use:\n")
			fmt.Printf("  man veo3\n")

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "./docs/man", "Output directory for man pages")

	return cmd
}

func newMarkdownCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "markdown",
		Short: "Generate markdown documentation",
		Long: `Generate markdown documentation for veo3 commands.

The markdown files will be written to the specified directory (default: ./docs/cli).

Example:
  veo3 docs markdown
  veo3 docs markdown --output ./docs/commands`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Parent().Parent()

			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputDir, 0750); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Generate markdown docs
			if err := doc.GenMarkdownTree(root, outputDir); err != nil {
				return fmt.Errorf("failed to generate markdown docs: %w", err)
			}

			absPath, _ := filepath.Abs(outputDir)
			fmt.Printf("Markdown documentation generated in: %s\n", absPath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "./docs/cli", "Output directory for markdown files")

	return cmd
}
