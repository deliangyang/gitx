package commands

import "github.com/spf13/cobra"

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install gitx tool",
	Run: func(cmd *cobra.Command, args []string) {
		execCommand("go", "install", "github.com/deliangyang/gitx@latest")
		successLog("gitx installed successfully.")
		execCommand("gitx", "--version")
	},
}
