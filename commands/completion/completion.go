package completion

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"github.com/spf13/cobra"
)

func NewCmdCompletion(io *iostreams.IOStreams) *cobra.Command {
	var (
		shellType string

		// description will not be added if true
		excludeDesc = false
	)

	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long: heredoc.Docf(`
		The output of this command will be computer code and is meant to be saved 
		to a file or immediately evaluated by an interactive shell. To load completions:

		### Bash

		To load completions in your current shell session:
		
		%[2]splaintext
		source <(glab completion -s bash)
		%[2]s

		To load completions for every new session, execute once:

		#### Linux

		%[2]splaintext
		glab completion -s bash > /etc/bash_completion.d/glab
		%[2]s

		#### macOS

		%[2]splaintext
		glab completion -s bash > /usr/local/etc/bash_completion.d/glab
		%[2]s
		
		### Zsh
		
		If shell completion is not already enabled in your environment you will need
		to enable it. You can execute the following once:

		%[2]splaintext
		echo "autoload -U compinit; compinit" >> ~/.zshrc
		%[2]s

		To load completions in your current shell session:
		
		%[2]splaintext
		source <(glab completion -s zsh); compdef _glab glab
		%[2]s

		If using the 1Password shell plugin <https://developer.1password.com/docs/cli/shell-plugins/gitlab/>
		to authenticate, you may need to add the following to your ~/.zshrc file so zsh does not expand
		aliases before performing completion:
		
		%[2]splaintext
		setopt completealiases
		%[2]s
		
		To load completions for every new session, execute once:

		#### Linux
		
		%[2]splaintext
		glab completion -s zsh > "${fpath[1]}/_glab"
		%[2]s

		#### macOS

		%[2]splaintext
		glab completion -s zsh > /usr/local/share/zsh/site-functions/_glab
		%[2]s

		### fish

		To load completions in your current shell session:

		%[2]splaintext
		glab completion -s fish | source
		%[2]s

		To load completions for every new session, execute once:

		%[2]splaintext
		glab completion -s fish > ~/.config/fish/completions/glab.fish
		%[2]s

		### PowerShell

		To load completions in your current shell session:

		%[2]splaintext
		glab completion -s powershell | Out-String | Invoke-Expression
		%[2]s

		To load completions for every new session, add the output of the above command
		to your powershell profile.

		When installing glab through a package manager, however, it's possible that
		no additional shell configuration is necessary to gain completion support. 
		For Homebrew, see <https://docs.brew.sh/Shell-Completion>
		`, "`", "```"),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := io.StdOut
			rootCmd := cmd.Parent()

			switch shellType {
			case "bash":
				return rootCmd.GenBashCompletionV2(out, !excludeDesc)
			case "zsh":
				if excludeDesc {
					return rootCmd.GenZshCompletionNoDesc(out)
				}
				return rootCmd.GenZshCompletion(out)
			case "powershell":
				if excludeDesc {
					return rootCmd.GenPowerShellCompletion(out)
				}
				return rootCmd.GenPowerShellCompletionWithDesc(out)
			case "fish":
				return rootCmd.GenFishCompletion(out, !excludeDesc)
			default:
				return fmt.Errorf("unsupported shell type %q", shellType)
			}
		},
	}

	completionCmd.Flags().StringVarP(&shellType, "shell", "s", "bash", "Shell type: {bash|zsh|fish|powershell}")
	completionCmd.Flags().BoolVarP(&excludeDesc, "no-desc", "", false, "Do not include shell completion description")
	return completionCmd
}
