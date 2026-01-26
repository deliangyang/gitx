package commands

import (
	"github.com/spf13/cobra"
)

var MergeBackCmd = &cobra.Command{
	Use:   "mb",
	Short: "Merge current branch back to other branch",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			errLog("Please specify the target branch to merge back to.")
		}
		targetBranch := args[0]
		currentBranch := execCommandWithOutput("git", "rev-parse", "--abbrev-ref", "HEAD")
		execCommand("git", "checkout", targetBranch)
		// git pull
		execCommand("git", "pull", "origin", targetBranch)
		// git merge --no-ff currentBranch
		execCommand("git", "merge", "--no-ff", currentBranch)
		successLog("Merged branch %s back to %s successfully.", currentBranch, targetBranch)
		// git push
		execCommand("git", "push", "origin", targetBranch)
		successLog("Pushed branch %s to remote successfully.", targetBranch)
		// checkout back to current branch
		execCommand("git", "checkout", currentBranch)
		successLog("Checked out back to branch %s.", currentBranch)
	},
}
