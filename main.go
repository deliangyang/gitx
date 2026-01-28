package main

import (
	"log"

	"github.com/deliangyang/gitx/commands"
	"github.com/spf13/cobra"
)

var (
	version = "v1.0.2"
	rootCmd = &cobra.Command{
		Use:   "gitx",
		Short: "A CLI tool for advanced Git operations",
		Long: `A CLI tool for advanced Git operations.
It provides various commands to manage Git repositories efficiently.`,
	}
)

func main() {
	rootCmd.AddCommand(commands.ConfigCmd)
	rootCmd.AddCommand(commands.CloneCmd)
	rootCmd.AddCommand(commands.SelectCmd)
	rootCmd.AddCommand(commands.SyncCmd)
	rootCmd.AddCommand(commands.FetchCmd)
	rootCmd.AddCommand(commands.AICommitCmd)
	rootCmd.AddCommand(commands.InstallCmd)
	rootCmd.AddCommand(commands.RenameCmd)
	rootCmd.AddCommand(commands.MergeBackCmd)
	rootCmd.AddCommand(commands.UseCmd)
	rootCmd.AddCommand(commands.DocCmd)
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln("Execute rootCmd fail:", err)
	}
}
