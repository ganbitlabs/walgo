package cmd

import (
	"fmt"
	"os"

	"github.com/ganbitlabs/walgo/internal/config"
	"github.com/ganbitlabs/walgo/internal/optimizer"
	"github.com/ganbitlabs/walgo/internal/ui"

	"github.com/spf13/cobra"
)

var optimizeCmd = &cobra.Command{
	Use:   "optimize [directory]",
	Short: "Optimize HTML, CSS, and JavaScript files for better performance.",
	Long: `Optimizes HTML, CSS, and JavaScript files in the specified directory (or current directory).
This command provides asset optimization including:

HTML Optimization:
• Minify HTML by removing unnecessary whitespace
• Remove HTML comments (preserving conditional comments)
• Compress inline CSS and JavaScript

CSS Optimization:
• Minify CSS by removing whitespace and comments
• Compress color values (e.g., #ffffff -> #fff)

JavaScript Optimization:
• Minify JavaScript by removing whitespace and comments
• Shorten local variable names (when obfuscate is enabled)

Minification uses tdewolff/minify. The settings live in walgo.yaml under the 'optimizer'
section; see docs/OPTIMIZER.md.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		var targetDir string
		var siteCfg *config.WalgoConfig
		if len(args) > 0 {
			targetDir = args[0]
			siteCfg, _ = config.LoadConfigFrom(targetDir)
		} else {
			sitePath, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
				return fmt.Errorf("error getting current directory: %w", err)
			}

			if cfg, err := config.LoadConfigFrom(sitePath); err == nil {
				siteCfg = cfg
				publishDir, err := config.ResolvePublishDir(sitePath, cfg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
					return err
				}
				if _, err := os.Stat(publishDir); err == nil {
					targetDir = publishDir
					fmt.Printf("Using Hugo publish directory: %s\n", publishDir)
				} else {
					targetDir = sitePath
				}
			} else {
				targetDir = sitePath
			}
		}

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s Error: Directory %s does not exist\n", icons.Error, targetDir)
			return fmt.Errorf("directory %s does not exist", targetDir)
		}

		fmt.Printf("Optimizing files in: %s\n", targetDir)

		var optimizerConfig optimizer.OptimizerConfig
		if siteCfg != nil {
			optimizerConfig = siteCfg.OptimizerConfig
		} else {
			optimizerConfig = optimizer.NewDefaultOptimizerConfig()
			fmt.Println("Using default optimization settings (no walgo.yaml found)")
		}

		if cmd.Flags().Changed("verbose") {
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: reading verbose flag: %v\n", icons.Error, err)
				return fmt.Errorf("error reading verbose flag: %w", err)
			}
			optimizerConfig.Verbose = verbose
		}

		if cmd.Flags().Changed("html") {
			html, err := cmd.Flags().GetBool("html")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: reading html flag: %v\n", icons.Error, err)
				return fmt.Errorf("error reading html flag: %w", err)
			}
			optimizerConfig.HTML.Enabled = html
		}

		if cmd.Flags().Changed("css") {
			css, err := cmd.Flags().GetBool("css")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: reading css flag: %v\n", icons.Error, err)
				return fmt.Errorf("error reading css flag: %w", err)
			}
			optimizerConfig.CSS.Enabled = css
		}

		if cmd.Flags().Changed("js") {
			js, err := cmd.Flags().GetBool("js")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: reading js flag: %v\n", icons.Error, err)
				return fmt.Errorf("error reading js flag: %w", err)
			}
			optimizerConfig.JS.Enabled = js
		}

		if cmd.Flags().Changed("remove-unused-css") {
			removeUnused, err := cmd.Flags().GetBool("remove-unused-css")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: reading remove-unused-css flag: %v\n", icons.Error, err)
				return fmt.Errorf("error reading remove-unused-css flag: %w", err)
			}
			optimizerConfig.CSS.RemoveUnused = removeUnused
		}

		engine := optimizer.NewEngine(optimizerConfig)
		stats, err := engine.OptimizeDirectory(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Optimization failed: %v\n", icons.Error, err)
			return fmt.Errorf("error during optimization: %w", err)
		}

		engine.PrintStats(stats)

		if stats.FilesOptimized > 0 {
			fmt.Printf("\n%s Optimization complete! %d files optimized.\n", icons.Check, stats.FilesOptimized)
		} else {
			fmt.Printf("\n%s No files needed optimization.\n", icons.Check)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(optimizeCmd)

	optimizeCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	optimizeCmd.Flags().Bool("html", true, "Enable HTML optimization")
	optimizeCmd.Flags().Bool("css", true, "Enable CSS optimization")
	optimizeCmd.Flags().Bool("js", true, "Enable JavaScript optimization")
	optimizeCmd.Flags().Bool("remove-unused-css", false, "Not implemented; has no effect")
}
