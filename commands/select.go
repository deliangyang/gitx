package commands

import (
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	SelectCmd.Flags().StringVarP(&stableBranch, "branch", "b", "stable", "Stable branch name, default is 'stable'")
}

var SelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select common projects to clone, like `gitx select` or `gitx select -b main`",
	Run: func(cmd *cobra.Command, args []string) {
		prompt := promptui.Select{
			Label: "Select Repository to Clone",
			Items: config.CommonProjects,
		}

		_, repoUrl, err := prompt.Run()
		if err != nil {
			errLog("Read repoUrl fail %v\n", err)
		}
		successLog("You selected %s", repoUrl)
		prefixPrompt := promptui.Select{
			Label: "Enter prefix for version",
			Items: prefix,
		}
		_, prefix, err := prefixPrompt.Run()
		if err != nil {
			errLog("Read prefix fail %v\n", err)
		}
		versionPrompt := promptui.Prompt{
			Label: "Enter version (e.g., 20250101 or 1.2.3, not containing prefix and -)",
		}
		version, err := versionPrompt.Run()
		if err != nil {
			errLog("Read version fail %v\n", err)
		}
		branchPrompt := promptui.Prompt{
			Label: "Enter develop branch",
		}
		branch, err := branchPrompt.Run()
		if err != nil {
			errLog("Read branch fail %v\n", err)
		} else if branch == "" {
			errLog("Branch cannot be empty")
		}
		if config.OpenInIDEAfterUse {
			cloneRepository(repoUrl, prefix+"-"+version, branch)
		}
	},
}
