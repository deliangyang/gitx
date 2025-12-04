
## gitx

一个用于简化 Git 工作流程的命令行工具，集成了 AI 助手以生成提交信息。

## 首次安装

```bash
go install github.com/deliangyang/gitx@latest
```

## 功能

```
A CLI tool for advanced Git operations.
It provides various commands to manage Git repositories efficiently.

Usage:
  gitx [command]

Available Commands:
  am          Generate AI-based commit messages, then push to remote, limit to 4000 characters diff
  clone       Clone a repository
  completion  Generate the autocompletion script for the specified shell
  doc         Show documentation
  fetch       Merge stable branch into current feat branch, like merge stable into feat-3.4.0
  help        Help about any command
  install     Install gitx tool
  rename      Rename current project directory
  select      Select common projects to clone, like `gitx select` or `gitx select -b main`
  sync        Merge feat branch into target branch, like merge feat-3.4.0 into new-dev
  use         Switch to a specific project directory in the workspace

Flags:
  -h, --help      help for gitx
  -v, --version   version for gitx

Use "gitx [command] --help" for more information about a command.
```

## 设置工作目录

设置环境变量 `WORKSPACE_DIR`，指定工作目录，默认为 `~/work`。

```bash
export WORKSPACE_DIR=~/my_workspace
```

## am 命令

使用 AI 助手生成提交信息，并将更改推送到远程仓库。
确保已设置 `OPENAI_API_KEY` 环境变量。

```bash
gitx am   # 交互式确认提交

OPENAI_API_KEY="your_openai_api_key" gitx am    # 交互式确认提交

OPENAI_API_KEY="your_openai_api_key" gitx am -y # 自动确认提交
```

## clone 命令
克隆指定的 Git 仓库，并切换到指定的分支。

```bash
gitx clone <repository_url> <version> <branch> [-b <base_branch>]
```
例如：

```bash
gitx clone git@github.com:deliangyang/gitx.git feat-3.4.0 new-dev              # 默认分支 stable
gitx clone git@github.com:deliangyang/gitx.git feat-3.4.0 new-dev -b main      # 指定源分支 main
``` 

## sync 命令
基于当前目录的特征 (api-site-feat-3.4.0-new-dev)，将指定的 feat 分支合并到目标分支。

```bash
gitx sync
```
例如，将 feat-3.4.0 分支合并到 new-dev 分支。

## fetch 命令

基于当前目录的特征 (api-site-feat-3.4.0-new-dev)，将稳定分支合并到当前的 feat 分支。

```bash
gitx fetch                      # 默认分支 stable

gitx fetch -b main              # 指定源分支 main
```

## select 命令
选择常用项目进行克隆。

```bash
gitx select                     # 使用默认分支 stable
gitx select -b main             # 指定源分支 main
```

## rename 命令

重命名当前项目目录，重置新的版本号

```bash
pwd
# api-site-feat-1.2.0-new-dev
gitx rename feat-1.3.0    # 当前版本为 feat-1.2.0，则重命名为 feat-1.3.0

pwd
# api-site-feat-1.3.0-new-dev
```

## use 命令

切换到工作区中的特定项目目录。

```bash
gitx use
```

## install 命令
更新 gitx 工具到最新版本：

```bash
gitx install
```

## doc 命令

显示文档链接：[https://github.com/deliangyang/gitx/-/tree/main?ref_type=heads#%E5%8A%9F%E8%83%BD](https://github.com/deliangyang/gitx/-/tree/main?ref_type=heads#%E5%8A%9F%E8%83%BD)

## 版本
查看当前工具的版本：

```bash
gitx --version
```