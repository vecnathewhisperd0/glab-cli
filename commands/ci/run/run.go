package run

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/ci/ciutils"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/mr/mrutils"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var (
	PipelineVarTypeEnv  = gitlab.EnvVariableType
	PipelineVarTypeFile = gitlab.FileVariableType
)

var envVariables = []string{}

func parseVarArg(s string) (*gitlab.PipelineVariableOptions, error) {
	// From https://pkg.go.dev/strings#Split:
	//
	// > If s does not contain sep and sep is not empty,
	// > Split returns a slice of length 1 whose only element is s.
	//
	// Therefore, the function will always return a slice of min length 1.
	v := strings.SplitN(s, ":", 2)
	if len(v) == 1 {
		return nil, fmt.Errorf("invalid argument structure")
	}
	return &gitlab.PipelineVariableOptions{
		Key:   &v[0],
		Value: &v[1],
	}, nil
}

func extractEnvVar(s string) (*gitlab.PipelineVariableOptions, error) {
	pvar, err := parseVarArg(s)
	if err != nil {
		return nil, err
	}
	pvar.VariableType = &PipelineVarTypeEnv
	return pvar, nil
}

func extractFileVar(s string) (*gitlab.PipelineVariableOptions, error) {
	pvar, err := parseVarArg(s)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(*pvar.Value)
	if err != nil {
		return nil, err
	}
	content := string(b)
	pvar.VariableType = &PipelineVarTypeFile
	pvar.Value = &content
	return pvar, nil
}

type runOpts struct {
	isMergeRequest bool
}

func NewCmdRun(f *cmdutils.Factory) *cobra.Command {
	opts := runOpts{}
	pipelineRunCmd := &cobra.Command{
		Use:     "run [flags]",
		Short:   `Create or run a new CI/CD pipeline.`,
		Aliases: []string{"create"},
		Example: heredoc.Doc(`
	glab ci run
	glab ci run -b main
	glab ci run -b main --variables-env key1:val1
	glab ci run -b main --variables-env key1:val1,key2:val2
	glab ci run -b main --variables-env key1:val1 --variables-env key2:val2
	glab ci run -b main --variables-file MYKEY:file1 --variables KEY2:some_value
	glab ci run --mr [mrID]
	`),
		Long: ``,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			repo, err := f.BaseRepo()
			if err != nil {
				return err
			}

			if opts.isMergeRequest {
				mrID, err := getMRIDFromArgs(cmd, args, f)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.StdOut, "Creating pipeline for MR", mrID, "...")
				mrPipeline, err := api.CreateMRPipeline(apiClient, repo.FullName(), mrID)
				if err != nil {
					return err
				}

				fmt.Fprintf(f.IO.StdOut, "Created pipeline (id: %d), status: %s, ref: %s, weburl: %s\n", mrPipeline.ID, mrPipeline.Status, mrPipeline.Ref, mrPipeline.WebURL)
				return nil
			}

			pipelineVars := []*gitlab.PipelineVariableOptions{}

			if customPipelineVars, _ := cmd.Flags().GetStringSlice("variables-env"); len(customPipelineVars) > 0 {
				for _, v := range customPipelineVars {
					pvar, err := extractEnvVar(v)
					if err != nil {
						return fmt.Errorf("parsing pipeline variable expected format KEY:VALUE: %w", err)
					}
					pipelineVars = append(pipelineVars, pvar)
				}
			}

			if customPipelineFileVars, _ := cmd.Flags().GetStringSlice("variables-file"); len(customPipelineFileVars) > 0 {
				for _, v := range customPipelineFileVars {
					pvar, err := extractFileVar(v)
					if err != nil {
						return fmt.Errorf("parsing pipeline variable. Expected format KEY:FILENAME: %w", err)
					}
					pipelineVars = append(pipelineVars, pvar)
				}
			}

			vf, err := cmd.Flags().GetString("variables-from")
			if err != nil {
				return err
			}
			if vf != "" {
				b, err := os.ReadFile(vf)
				if err != nil {
					// Return the error encountered
					return fmt.Errorf("opening variable file: %s", vf)
				}
				var result []*gitlab.PipelineVariableOptions
				err = json.Unmarshal(b, &result)
				if err != nil {
					return fmt.Errorf("loading pipeline values: %w", err)
				}
				pipelineVars = append(pipelineVars, result...)
			}

			c := &gitlab.CreatePipelineOptions{
				Variables: &pipelineVars,
			}

			branch, err := cmd.Flags().GetString("branch")
			if err != nil {
				return err
			}
			if branch != "" {
				c.Ref = gitlab.Ptr(branch)
			} else if currentBranch, err := f.Branch(); err == nil {
				c.Ref = gitlab.Ptr(currentBranch)
			} else {
				// `ci run` is running out of a git repo
				fmt.Fprintln(f.IO.StdOut, "not in a Git repository. Using repository argument.")
				c.Ref = gitlab.Ptr(ciutils.GetDefaultBranch(f))
			}

			pipe, err := api.CreatePipeline(apiClient, repo.FullName(), c)
			if err != nil {
				return err
			}

			fmt.Fprintln(f.IO.StdOut, "Created pipeline (id:", pipe.ID, "), status:", pipe.Status, ", ref:", pipe.Ref, ", weburl: ", pipe.WebURL, ")")
			return nil
		},
	}
	pipelineRunCmd.Flags().StringP("branch", "b", "", "Create pipeline on branch/ref <string>.")
	pipelineRunCmd.Flags().StringSliceVarP(&envVariables, "variables", "", []string{}, "Pass variables to pipeline in format <key>:<value>.")
	pipelineRunCmd.Flags().StringSliceVarP(&envVariables, "variables-env", "", []string{}, "Pass variables to pipeline in format <key>:<value>.")
	pipelineRunCmd.Flags().StringSliceP("variables-file", "", []string{}, "Pass file contents as a file variable to pipeline in format <key>:<filename>.")
	pipelineRunCmd.Flags().StringP("variables-from", "f", "", "JSON file containing variables for pipeline execution.")
	pipelineRunCmd.Flags().BoolVarP(&opts.isMergeRequest, "mr", "", false, "Run a pipeline for merge request with ID or current branch.")
	pipelineRunCmd.MarkFlagsMutuallyExclusive("branch", "mr")

	return pipelineRunCmd
}

func getMRIDFromArgs(cmd *cobra.Command, args []string, f *cmdutils.Factory) (int, error) {
	mrIDStr := cmd.Flags().Arg(0)
	var mrID int
	var err error
	if len(mrIDStr) > 0 {
		mrID, err = strconv.Atoi(mrIDStr)
		if err != nil {
			return 0, fmt.Errorf("MR ID id not an integer: %s (%w)", mrIDStr, err)
		}
	}
	if mrID == 0 {
		mr, _, err := mrutils.MRFromArgs(f, args, "any")
		if err != nil {
			return 0, err
		}
		mrID = mr.IID
	}
	return mrID, nil
}
