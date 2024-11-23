---
stage: Create
group: Code Review
info: To determine the technical writer assigned to the Stage/Group associated with this page, see https://about.gitlab.com/handbook/product/ux/technical-writing/#assignments
---

# Shell Assistant

![Shell Assistant Demo](/docs/assets/shell-assistant-demo.gif)

The GitLab CLI includes a shell assistant that converts natural language commands into actual shell commands using GitLab Duo AI assistant. This allows you to describe commands in plain English and get the correct syntax without leaving your terminal.

## Installation

The shell assistant scripts are available in the [scripts/shell-assistant](https://gitlab.com/gitlab-org/cli/-/tree/main/scripts/shell-assistant) folder of the GitLab CLI repository.

1. Choose the appropriate script for your shell:
   - For bash users: source `assistant.bash`
   - For zsh users: source `assistant.zsh` 

2. Add to your shell's config file:
   ```shell
   # For bash (~/.bashrc)
   source /path/to/assistant.bash

   # For zsh (~/.zshrc) 
   source /path/to/assistant.zsh
   ```

## Usage

1. Type your request in natural language
2. Press the key combination for your platform:
   - Linux: Press Alt+e
   - macOS: Press Esc, then press e
3. The natural language will be replaced with the appropriate command
4. Review the command before pressing Enter to execute

## Examples

```shell
# Finding files
"show me all pdf files modified in the last week"
→ find . -name "*.pdf" -mtime -7

# System information  
"show me the processes using the most memory"
→ ps aux --sort=-%mem | head -n 10

# Git operations
"undo my last commit"
→ git reset HEAD~1
```

The assistant uses GitLab Duo to translate natural language into actual shell commands, helping you stay productive without memorizing complex command syntax.
