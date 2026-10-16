package cmd

import (
	"fmt"
	"os"

	"github.com/ganbitlabs/walgo/internal/config"
	"github.com/ganbitlabs/walgo/internal/hugo"
	"github.com/ganbitlabs/walgo/internal/ui"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Hugo site.",
	Long: `Builds the Hugo site using the configuration found in the current directory
(or the directory specified by global --config flag if walgo.yaml is there).
This command runs the 'hugo' command to generate static files typically into the 'public' directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		fmt.Printf("%s Building site...\n", icons.Package)

		if err := hugo.BuildSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Troubleshooting:\n", icons.Lightbulb)
			if config.IsHugoSite(sitePath) {
				configFile := config.HugoConfigFile(sitePath)
				if configFile == "" {
					configFile = "your Hugo config"
				}
				fmt.Fprintf(os.Stderr, "  - Check that Hugo is installed: hugo version\n")
				fmt.Fprintf(os.Stderr, "  - Check %s for syntax errors\n", configFile)
				fmt.Fprintf(os.Stderr, "  - Run: hugo --verbose (for detailed output)\n")
			} else {
				fmt.Fprintf(os.Stderr, "  - This is a static site: walgo deploys its publish directory as-is\n")
				fmt.Fprintf(os.Stderr, "  - Set hugo.publishDir in walgo.yaml to the folder holding your built files\n")
				fmt.Fprintf(os.Stderr, "  - If it is a Hugo site, add a Hugo config or set site.kind: hugo in walgo.yaml\n")
			}
			return fmt.Errorf("build failed: %w", err)
		}

		fmt.Printf("\n%s Build complete!\n", icons.Success)
		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Printf("  - Preview: walgo serve\n")
		fmt.Printf("  - Deploy:  walgo launch\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
