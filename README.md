[中文](./README-cn.md)

## gitx

A command-line tool to simplify Git workflows, integrated with an AI assistant for generating commit messages.

## First-time Installation

```bash
go install github.com/deliangyang/gitx@latest
```

## Features

```
A CLI tool for advanced Git operations.
It provides various commands to manage Git repositories efficiently.

Usage:
  gitx [command]

Available Commands:
  am          Generate AI-based commit messages, then push to remote, limit to 10000 characters diff
  clone       Clone a repository
  completion  Generate the autocompletion script for the specified shell
  config      Configure gitx settings
  doc         Show documentation
  fetch       Merge main branch into current feat branch, like merge main into feat-3.4.0
  help        Help about any command
  install     Install gitx tool
  mb          Merge current branch back to other branch
  rename      Rename current project directory
  select      Select common projects to clone, like `gitx select` or `gitx select -b main`
  sync        Merge feat branch into target branch, like merge feat-3.4.0 into new-dev
  use         Switch to a specific project directory in the workspace

Flags:
  -h, --help      help for gitx
  -v, --version   version for gitx

Use "gitx [command] --help" for more information about a command.
```

## Setting Up Workspace Directory

Set the `WORKSPACE_DIR` environment variable to specify your workspace directory (default: `~/work`).

```bash
export WORKSPACE_DIR=~/my_workspace
```

## am Command

Use AI assistant to generate commit messages and push changes to remote repository.
Make sure you have set the `OPENAI_API_KEY` environment variable.
There are two types of prompts: one is [default](./commands/prompts/default.prompt), and the other is more detailed [github](./commands/prompts/github.prompt).

```bash
gitx am [default|github]  # Interactive commit confirmation

OPENAI_API_KEY="your_openai_api_key" gitx am    # Interactive commit confirmation

OPENAI_API_KEY="your_openai_api_key" gitx am -y # Auto-confirm commit
```

## clone Command
Clone a specified Git repository and switch to the specified branch.

```bash
gitx clone <repository_url> <version> <branch> [-b <base_branch>]
```
Example:

```bash
gitx clone git@github.com:deliangyang/gitx.git feat-3.4.0 new-dev              # Default branch main
gitx clone git@github.com:deliangyang/gitx.git feat-3.4.0 new-dev -b main      # Specify source branch main
``` 

## sync Command
Based on current directory pattern (deliangyang-gitx-feat-3.4.0-new-dev), merge the specified feat branch into target branch.

```bash
gitx sync
```
For example, merge feat-3.4.0 branch into new-dev branch.

## fetch Command

Based on current directory pattern (deliangyang-gitx-feat-3.4.0-new-dev), merge stable branch into current feat branch.

```bash
gitx fetch                      # Default branch main

gitx fetch -b main              # Specify source branch main
```

## mb Command

Merge current branch back to other branch. This command will:
1. Switch to the target branch
2. Pull latest changes from remote
3. Merge current branch into target branch using `--no-ff` flag
4. Push changes to remote
5. Switch back to the original branch

```bash
gitx mb <target_branch>
```

Example:

```bash
gitx mb main                    # Merge current branch back to main
gitx mb develop                 # Merge current branch back to develop
```

## select Command
Select common projects to clone.

```bash
gitx select                     # Use default branch main
gitx select -b main             # Specify source branch main
```

## rename Command

Rename current project directory and reset the new version number.

```bash
pwd
# deliangyang-gitx-feat-1.2.0-new-dev
gitx rename feat-1.3.0    # Current version feat-1.2.0 will be renamed to feat-1.3.0

pwd
# deliangyang-gitx-feat-1.3.0-new-dev
```

## use Command

Switch to a specific project directory in workspace.

```bash
gitx use
```

## install Command
Update gitx tool to latest version:

```bash
gitx install
```

## config Command
View and modify gitx configuration:

```bash
gitx config # Generate default config file and store it in ~/.gitx/config.json
gitx config view               # View current configuration
gitx config set <key> <value>  # Set configuration item
```

## doc Command

Display documentation link: [https://github.com/deliangyang/gitx/blob/main/README.md](https://github.com/deliangyang/gitx/blob/main/README.md)

## Version
View current tool version:

```bash
gitx --version
```
