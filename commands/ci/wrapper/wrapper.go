package run

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gopkg.in/yaml.v2"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const keyValuePair = ".+:.+"

var re = regexp.MustCompile(keyValuePair)

func aliasNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	switch name {
	case "variables":
		name = "variables-env"
		break
	}
	return pflag.NormalizedName(name)
}

func NewCmdWrapper(f *cmdutils.Factory) *cobra.Command {
	var pipelineRunCmd = &cobra.Command{
		Use:     "wrapper [flags]",
		Short:   `Emulate pipeline run with local command`,
		Aliases: []string{"wrap"},
		Example: heredoc.Doc(`
	glab ci wrapper <command>
	glab ci wrapper --variables-env MYKEY:some_value
	glab ci wrapper --variables-env MYKEY:some_value --variables-env KEY2:another_value
	glab ci wrapper --variables-file MYKEY:file1 --variables KEY2:some_value
	glab ci wrapper --pipeline-defaults --variables-file MYKEY:file1 --variables KEY2:some_value
	`),
		Long: ``,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// pipelineVars := []*gitlab.PipelineVariable{}
			pipelineVars := make(map[string]string)

			var shell string
			var shellCommand string

			shell, _ = cmd.Flags().GetString("shell")

			shellCommand, _ = cmd.Flags().GetString("command")
			if shellCommand == "" {
				fmt.Println("Shell command was not provided")
			}

			pipelineFile, _ := cmd.Flags().GetString("pipeline-file")
			pipelineDefaults, _ := cmd.Flags().GetBool("pipeline-defaults")

			if pipelineDefaults {
				yfile, err := ioutil.ReadFile(pipelineFile)

				if err != nil {

					log.Fatal(err)
				}

				data := make(map[interface{}]interface{})

				err2 := yaml.Unmarshal(yfile, &data)

				if err2 != nil {

					log.Fatal(err2)
				}

				yamlVariables := data["variables"].(map[interface{}]interface{})

				for k, v := range yamlVariables {
					if valueStr, ok := v.(string); ok {
						pipelineVars[k.(string)] = valueStr
					} else if valueInt, ok := v.(int); ok {
						pipelineVars[k.(string)] = strconv.Itoa(valueInt)
					} else if valueBool, ok := v.(bool); ok {
						pipelineVars[k.(string)] = strconv.FormatBool(valueBool)
					} else if valueFloat, ok := v.(float64); ok {
						pipelineVars[k.(string)] = strconv.FormatFloat(valueFloat, 'f', -1, 64)
					} else {
						mapValue := v.(map[interface{}]interface{})
						pipelineVars[k.(string)] = mapValue["value"].(string)
					}
				}

			}
			if customPipelineVars, _ := cmd.Flags().GetStringSlice("variables-env"); len(customPipelineVars) > 0 {
				for _, v := range customPipelineVars {
					if !re.MatchString(v) {
						return fmt.Errorf("Bad pipeline variable : \"%s\" should be of format KEY:VALUE", v)
					}
					s := strings.SplitN(v, ":", 2)
					pipelineVars[s[0]] = s[1]
				}
			}

			if customPipelineFileVars, _ := cmd.Flags().GetStringSlice("variables-file"); len(customPipelineFileVars) > 0 {
				for _, v := range customPipelineFileVars {
					if !re.MatchString(v) {
						return fmt.Errorf("Bad pipeline variable : \"%s\" should be of format KEY:FILENAME", v)
					}
					s := strings.SplitN(v, ":", 2)
					pipelineVars[s[0]] = s[1]
				}
			}

			if vf, _ := cmd.Flags().GetString("variables-from"); vf != "" {
				variablesFile, err := os.Open(vf)
				if err != nil {
					fmt.Println("Can't open file " + vf)
				}
				byteValue, _ := ioutil.ReadAll(variablesFile)
				var result []interface{}
				json.Unmarshal([]byte(byteValue), &result)
				for _, v := range result {
					variableType := "env_var"
					value := v.(map[string]interface{})
					if varType, ok := value["variable_type"]; ok {
						variableType = varType.(string)
					}
					varName := value["key"].(string)
					varValue := value["value"].(string)
					if variableType == "env_var" {
						pipelineVars[varName] = varValue
					} else if variableType == "file" {
						// it's a file
						// we need to dump value into a file and record pointer
						file, err := ioutil.TempFile(".", "pipeline-"+varName+"-*.var")
						pipelineVars[varName] = file.Name()
						if err != nil {
							fmt.Println("Error opening " + file.Name())
						}
						buff := bufio.NewWriter(file)
						buff.WriteString(varValue)
						buff.Flush()
						defer os.Remove(file.Name())
					}

				}
				defer variablesFile.Close()
			}

			// Run commands with env vars set

			command := exec.Command(shell, "-c", shellCommand)

			// Set up environment for the command without setting the
			// the variables outside of it's execution.
			command.Env = os.Environ()
			for envName, envValue := range pipelineVars {
				command.Env = append(command.Env, envName+"="+envValue)
			}
			out, err := command.CombinedOutput()
			if err != nil {
				return fmt.Errorf("executing cmd: %s out: %s", command.String(), out)
			}
			fmt.Println("cmd:", command.String(), "out:", string(out))

			return nil
		},
	}
	pipelineRunCmd.Flags().StringP("shell", "s", os.Getenv("SHELL"), "Use alternative shell for command execution")
	pipelineRunCmd.Flags().StringP("command", "c", "", "Command to execute")
	pipelineRunCmd.Flags().BoolP("pipeline-defaults", "p", false, "load variables from pipeline-like file with top-level 'variables:' section")
	pipelineRunCmd.Flags().String("pipeline-file", ".gitlab-ci.yml", "YAML file with root-level 'variables:' definitions")
	pipelineRunCmd.Flags().StringSlice("variables-env", []string{}, "Pass variables to pipeline in format <key>:<value> where type can be 'file' or 'env_var'")
	pipelineRunCmd.Flags().StringSlice("variables-file", []string{}, "Pass variables to pipeline in format <key>:<filename> where type can be 'file' or 'env_var'")
	pipelineRunCmd.Flags().StringP("variables-from", "f", "", "JSON file containing variables for pipeline execution")
	pipelineRunCmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)

	return pipelineRunCmd
}
