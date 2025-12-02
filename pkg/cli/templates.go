package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/jasongoecke/go-veo3/pkg/templates"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Manage prompt templates",
		Long:  "Save, list, and manage reusable prompt templates with variable substitution",
	}

	cmd.AddCommand(
		newTemplatesSaveCmd(),
		newTemplatesListCmd(),
		newTemplatesGetCmd(),
		newTemplatesDeleteCmd(),
		newTemplatesExportCmd(),
		newTemplatesImportCmd(),
	)

	return cmd
}

func newTemplatesSaveCmd() *cobra.Command {
	var (
		description string
		tags        []string
	)

	cmd := &cobra.Command{
		Use:   "save <name> <prompt>",
		Short: "Save a prompt template",
		Long: `Save a prompt template with variable placeholders.

Variables are specified using {{variable_name}} syntax.
Example: "A {{style}} image of {{subject}}"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			prompt := args[1]

			manager, err := getTemplateManager()
			if err != nil {
				return err
			}

			template := &templates.Template{
				Name:        name,
				Prompt:      prompt,
				Description: description,
				Tags:        tags,
			}

			if err := manager.Save(template); err != nil {
				return fmt.Errorf("failed to save template: %w", err)
			}

			// Show template info
			vars := template.Variables()
			fmt.Printf("✓ Saved template '%s'\n", name)
			if len(vars) > 0 {
				fmt.Printf("  Variables: %s\n", strings.Join(vars, ", "))
			}
			if description != "" {
				fmt.Printf("  Description: %s\n", description)
			}
			if len(tags) > 0 {
				fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Template description")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Template tags (comma-separated)")

	return cmd
}

func newTemplatesListCmd() *cobra.Command {
	var (
		outputFormat string
		tags         []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all prompt templates",
		Long:  "List all saved prompt templates with their details",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := getTemplateManager()
			if err != nil {
				return err
			}

			allTemplates := manager.List()
			if len(allTemplates) == 0 {
				fmt.Println("No templates found")
				return nil
			}

			// Filter by tags if specified
			var filteredTemplates []*templates.Template
			if len(tags) > 0 {
				for _, tmpl := range allTemplates {
					if hasAnyTag(tmpl.Tags, tags) {
						filteredTemplates = append(filteredTemplates, tmpl)
					}
				}
			} else {
				filteredTemplates = allTemplates
			}

			if len(filteredTemplates) == 0 {
				fmt.Println("No templates found matching tags:", strings.Join(tags, ", "))
				return nil
			}

			// Sort by name
			sort.Slice(filteredTemplates, func(i, j int) bool {
				return filteredTemplates[i].Name < filteredTemplates[j].Name
			})

			switch outputFormat {
			case "json":
				return outputJSON(filteredTemplates)
			default:
				return outputTemplatesTable(filteredTemplates)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Filter by tags (comma-separated)")

	return cmd
}

func newTemplatesGetCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get a prompt template",
		Long:  "Display details of a specific prompt template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			manager, err := getTemplateManager()
			if err != nil {
				return err
			}

			template, err := manager.Get(name)
			if err != nil {
				return err
			}

			switch outputFormat {
			case "json":
				return outputJSON(template)
			default:
				return outputTemplateDetails(template)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json)")

	return cmd
}

func newTemplatesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a prompt template",
		Long:  "Remove a saved prompt template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			manager, err := getTemplateManager()
			if err != nil {
				return err
			}

			if err := manager.Delete(name); err != nil {
				return err
			}

			fmt.Printf("✓ Deleted template '%s'\n", name)
			return nil
		},
	}

	return cmd
}

func newTemplatesExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <file>",
		Short: "Export all templates to a file",
		Long:  "Export all prompt templates to a YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			manager, err := getTemplateManager()
			if err != nil {
				return err
			}

			if err := manager.Export(path); err != nil {
				return fmt.Errorf("failed to export templates: %w", err)
			}

			templates := manager.List()
			fmt.Printf("✓ Exported %d template(s) to %s\n", len(templates), path)
			return nil
		},
	}

	return cmd
}

func newTemplatesImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import templates from a file",
		Long:  "Import prompt templates from a YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			manager, err := getTemplateManager()
			if err != nil {
				return err
			}

			beforeCount := len(manager.List())

			if err := manager.Import(path); err != nil {
				return fmt.Errorf("failed to import templates: %w", err)
			}

			afterCount := len(manager.List())
			imported := afterCount - beforeCount

			fmt.Printf("✓ Imported %d template(s) from %s\n", imported, path)
			return nil
		},
	}

	return cmd
}

// Helper functions

func getTemplateManager() (*templates.Manager, error) {
	configDir := viper.GetString("config-dir")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".veo3")
	}

	return templates.NewManager(configDir)
}

func outputTemplatesTable(templates []*templates.Template) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tVARIABLES\tTAGS\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "----\t---------\t----\t-----------")

	for _, tmpl := range templates {
		vars := tmpl.Variables()
		varStr := strings.Join(vars, ", ")
		if varStr == "" {
			varStr = "-"
		}

		tagStr := strings.Join(tmpl.Tags, ", ")
		if tagStr == "" {
			tagStr = "-"
		}

		desc := tmpl.Description
		if desc == "" {
			desc = "-"
		}
		// Truncate long descriptions
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", tmpl.Name, varStr, tagStr, desc)
	}

	return w.Flush()
}

func outputTemplateDetails(template *templates.Template) error {
	fmt.Printf("Name:        %s\n", template.Name)
	fmt.Printf("Prompt:      %s\n", template.Prompt)

	vars := template.Variables()
	if len(vars) > 0 {
		fmt.Printf("Variables:   %s\n", strings.Join(vars, ", "))
	}

	if template.Description != "" {
		fmt.Printf("Description: %s\n", template.Description)
	}

	if len(template.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(template.Tags, ", "))
	}

	fmt.Printf("Created:     %s\n", template.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:     %s\n", template.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func outputJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func hasAnyTag(templateTags, filterTags []string) bool {
	for _, filterTag := range filterTags {
		for _, templateTag := range templateTags {
			if templateTag == filterTag {
				return true
			}
		}
	}
	return false
}
