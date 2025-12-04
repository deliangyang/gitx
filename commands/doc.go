package commands

import "github.com/spf13/cobra"

var DocCmd = &cobra.Command{
	Use:   "doc",
	Short: "Show documentation",
	Run: func(cmd *cobra.Command, args []string) {
		successLog(`Documentation is available at: https://github.com/deliangyang/gitx/blob/main/README.md`)
	},
}
