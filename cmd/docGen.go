package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newGenerateDocsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doc-gen",
		Short: "Generate documentation",
		Long: heredoc.Doc(`
			Generate documentation for all commands
			to the 'docs' directory.`),
		Hidden: true,
		RunE:   generateDocs,
	}
}

func init() {
	rootCmd.AddCommand(newGenerateDocsCmd())
}

func generateDocs(cmd *cobra.Command, args []string) error {
	fmt.Println("Generating docs...")

	err := doc.GenMarkdownTree(NewRootCmd(), "doc")
	if err != nil {
		return err
	}

	fmt.Println("Done.")
	return nil
}
