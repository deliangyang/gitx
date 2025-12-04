package commands

import (
	"os"
	"path"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var UseCmd = &cobra.Command{
	Use:   "use",
	Short: "Switch to a specific project directory in the workspace",
	Run: func(cmd *cobra.Command, args []string) {
		prompt := promptui.Select{
			Label: "Select Project Directory",
			Items: func() []string {
				dirs, err := os.ReadDir(workspaceDir)
				if err != nil {
					errLog("failed to read workspace directory: %v", err)
				}
				var projectDirs []string
				for _, dir := range dirs {
					if dir.IsDir() {
						projectDirs = append(projectDirs, dir.Name())
					}
				}
				return projectDirs
			}(),
		}

		_, selectedDir, err := prompt.Run()
		if err != nil {
			errLog("Read selectedDir fail %v\n", err)
		}
		successLog("You selected %s", selectedDir)
		projectPath := path.Join(workspaceDir, selectedDir)
		if err := os.Chdir(projectPath); err != nil {
			errLog("Change directory to %s failed: %v", projectPath, err)
		}
		successLog("Changed directory to: %s", projectPath)

		openByIDEA(projectPath)
	},
}
