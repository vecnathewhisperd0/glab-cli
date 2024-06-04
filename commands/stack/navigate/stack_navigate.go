package navigate

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/xanzy/go-gitlab"

	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/internal/glrepo"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
)

type MoveOptions struct {
	IO          *iostreams.IOStreams
	LabClient   *gitlab.Client
	CurrentUser *gitlab.User
	BaseRepo    func() (glrepo.Interface, error)
	Remotes     func() (glrepo.Remotes, error)
	Config      func() (config.Config, error)
}

func baseCommand(f *cmdutils.Factory) (git.Stack, error) {
	title, err := git.GetCurrentStackTitle()
	if err != nil {
		return git.Stack{}, err
	}

	stack, err := git.GatherStackRefs(title)
	if err != nil {
		return git.Stack{}, err
	}

	return stack, nil
}

func NewCmdStackFirst(f *cmdutils.Factory) *cobra.Command {
	stackFirstCmd := &cobra.Command{
		Use:   "first",
		Short: `moves to the first diff in the stack`,
		Long:  `Moves to the first diff in the stack and checks out that branch`,
		Example: heredoc.Doc(`
			glab stack first
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			stack, err := baseCommand(f)
			if err != nil {
				return err
			}

			ref, err := stack.First()
			if err != nil {
				return err
			}

			err = git.CheckoutBranch(ref.Branch)
			if err != nil {
				return err
			}

			switchMessage(f, ref)

			return nil
		},
	}

	return stackFirstCmd
}

func NewCmdStackNext(f *cmdutils.Factory) *cobra.Command {
	stackFirstCmd := &cobra.Command{
		Use:   "next",
		Short: `moves to the next diff`,
		Long:  `Moves to the next diff in the stack and checks out that branch`,
		Example: heredoc.Doc(`
			glab stack next
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			stack, err := baseCommand(f)
			if err != nil {
				return err
			}

			ref, err := git.CurrentStackFromBranch(stack.Title)
			if err != nil {
				return err
			}

			if ref.Next != "" {
				err = git.CheckoutBranch(ref.Branch)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("no next diff")
			}

			switchMessage(f, ref)

			return nil
		},
	}

	return stackFirstCmd
}

func NewCmdStackPrev(f *cmdutils.Factory) *cobra.Command {
	stackFirstCmd := &cobra.Command{
		Use:   "prev",
		Short: `moves to the previous diff`,
		Long:  `Moves to the previous diff in the stack and checks out that branch`,
		Example: heredoc.Doc(`
			glab stack prev
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			stack, err := baseCommand(f)
			if err != nil {
				return err
			}

			ref, err := git.CurrentStackFromBranch(stack.Title)
			if err != nil {
				return err
			}

			if ref.Prev != "" {
				err = git.CheckoutBranch(ref.Branch)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("no previous diff")
			}

			switchMessage(f, ref)

			return nil
		},
	}

	return stackFirstCmd
}

func NewCmdStackLast(f *cmdutils.Factory) *cobra.Command {
	stackLastCmd := &cobra.Command{
		Use:   "last",
		Short: `moves to the last diff`,
		Long:  `Moves to the last diff in the stack and checks out that branch`,
		Example: heredoc.Doc(`
			glab stack last
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			stack, err := baseCommand(f)
			if err != nil {
				return err
			}

			ref, err := stack.Last()
			if err != nil {
				return err
			}

			err = git.CheckoutBranch(ref.Branch)
			if err != nil {
				return err
			}

			switchMessage(f, ref)

			return nil
		},
	}

	return stackLastCmd
}

type BranchChoice struct {
	branch      string
	description string
}

func NewCmdStackMove(f *cmdutils.Factory) *cobra.Command {
	stackLastCmd := &cobra.Command{
		Use:   "move",
		Short: `moves to any selected entry in the stack`,
		Long:  `Brings up a menu to select a stack with a fuzzy finder`,
		Example: heredoc.Doc(`
			glab stack move
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			stack, err := baseCommand(f)
			if err != nil {
				return err
			}

			var branches []string
			var descriptions []string

			firstRef, err := stack.First()
			if err != nil {
				return err
			}

			i := 1
			ref := firstRef
			for {
				branches = append(branches, ref.Branch)
				message := fmt.Sprintf("%v: %v", i, ref.Description)
				descriptions = append(descriptions, message)

				i++

				if ref.Next == "" {
					break
				} else {
					ref = stack.Refs[ref.Next]
				}
			}

			var branch string
			prompt := &survey.Select{
				Message: "Choose a diff:",
				Options: branches,
				Description: func(value string, index int) string {
					return descriptions[index]
				},
			}
			survey.AskOne(prompt, &branch)

			err = git.CheckoutBranch(branch)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return stackLastCmd
}

func switchMessage(f *cmdutils.Factory, ref git.StackRef) {
	color := f.IO.Color()
	fmt.Printf(
		"%v Switched to branch: %v - %v\n",
		color.ProgressIcon(),
		color.Blue(ref.Branch),
		color.Bold(ref.Description),
	)
}
