# GLab

![GLab](docs/assets/glab-logo.png)

GLab is an open source GitLab CLI tool bringing GitLab to your terminal next to where you are already working with `git` and your code without switching between windows and browser tabs. Work with issues, merge requests, **watch running pipelines directly from your CLI** among other features.

`glab` is available for repositories hosted on GitLab.com and self-managed GitLab instances. `glab` supports multiple authenticated GitLab instances and automatically detects the authenticated hostname from the remotes available in the working Git directory.

![command example](docs/assets/glabgettingstarted.gif)

## Table of contents

- [Table of contents](#table-of-contents)
- [Usage](#usage)
- [Demo](#demo)
- [Documentation](#documentation)
- [Installation](#installation)
  - [Homebrew](#homebrew)
  - [Other installation methods](#other-installation-methods)
  - [Building from source](#building-from-source)
    - [Prerequisites for building from source](#prerequisites-for-building-from-source)
- [Authentication](#authentication)
- [Configuration](#configuration)
- [Environment variables](#environment-variables)
- [Issues](#issues)
- [Contributing](#contributing)
- [Inspiration](#inspiration)

## Usage

To get started with `glab`:

1. Follow the [installation instructions](#installation) appropriate for your operating system.
1. [Authenticate](#authentication) into your instance of GitLab.
1. Optional. Configure `glab` further to meet your needs:
   - Set any needed global, per-project, or per-host [configuration](#configuration).
   - Set any needed [environment variables](#environment-variables).

You're ready!

### Core commands

Run `glab --help` to view a list of core commands in your terminal.

- [`glab alias`](docs/source/alias)
- [`glab api`](docs/source/api)
- [`glab auth`](docs/source/auth)
- [`glab check-update`](docs/source/check-update)
- [`glab ci`](docs/source/ci)
- [`glab completion`](docs/source/completion)
- [`glab config`](docs/source/config)
- [`glab incident`](docs/source/incident)
- [`glab issue`](docs/source/issue)
- [`glab label`](docs/source/label)
- [`glab mr`](docs/source/mr)
- [`glab release`](docs/source/release)
- [`glab repo`](docs/source/repo)
- [`glab schedule`](docs/source/schedule)
- [`glab snippet`](docs/source/snippet)
- [`glab ssh-key`](docs/source/ssh-key)
- [`glab user`](docs/source/user)
- [`glab variable`](docs/source/variable)

Commands follow this pattern:

```shell
glab <command> <subcommand> [flags]
```

Many core commands also have sub-commands. Some examples:

- List merge requests assigned to you: `glab mr list --assignee=@me`
- List review requests for you: `glab mr list --reviewer=@me`
- Approve a merge request: `glab mr approve 235`
- Create an issue, and add milestone, title, and label: `glab issue create -m release-2.0.0 -t "My title here" --label important`

## Demo

[![asciicast](https://asciinema.org/a/368622.svg)](https://asciinema.org/a/368622)

## Documentation

Read the [documentation](https://gitlab.com/gitlab-org/cli/-/tree/main/docs/source) for usage instructions or check out `glab help`.

## Installation

Download a binary suitable for your OS at the [releases page](https://gitlab.com/gitlab-org/cli/-/releases).
Other installation methods depend on your operating system.

### Homebrew

Homebrew is the officially supported package manager for macOS, Linux, and Windows (through [Windows Subsystem for Linux](https://learn.microsoft.com/en-us/windows/wsl/install))

- Homebrew
  - Install with: `brew install glab`
  - Update with: `brew upgrade glab`

### Other installation methods

Other options to install the GitLab CLI that may not be officially support or are maintained by the community are [also available](docs/installation_options.md).

### Building from source

If a supported binary for your OS is not found at the [releases page](https://gitlab.com/gitlab-org/cli/-/releases), you can build from source:

#### Prerequisites for building from source

- `make`
- Go 1.18+

To build from source:

1. Run the command `go version` to verify that Go version 1.18 or later is installed.
   If `go` is not installed, follow instructions on [the Go website](https://go.dev/doc/install).
1. Run the `go install gitlab.com/gitlab-org/cli/cmd/glab@main` to install `glab` cmd in `$GOPATH/bin`.
1. The sources of `glab` will be in `$GOPATH/src/gitlab.com/gitlab-org/cli`.
1. If you do not have `$GOPATH/bin` or `$GOBIN` in your `$PATH`, run `export PATH=$PWD/bin:$PATH`
   to update your PATH with the newly compiled project.
1. Run `glab version` to confirm that it worked.

## Authentication

### OAuth (GitLab.com only)

To authenticate your installation of `glab` with OAuth:

1. Start interactive setup with `glab auth login`.
1. For the GitLab instance you want to sign in to, select **GitLab.com**.
1. For the login method, select **Web**. This selection launches your web browser
   to request authorization for the GitLab CLI to use your GitLab.com account.
1. Select **Authorize**.
1. Complete the authentication process in your terminal, selecting the appropriate options for your needs.

### Personal Access Token

To authenticate your installation of `glab` with a personal access token:

1. Get a GitLab personal access token with at least the `api`
   and `write_repository` scopes. Use the method appropriate for your instance:
   - For GitLab.com, create one at the [Personal access tokens](https://gitlab.com/-/profile/personal_access_tokens) page.
   - For self-managed instances, visit `https://gitlab.example.com/-/profile/personal_access_tokens`,
     modifying `gitlab.example.com` to match the domain name of your instance.
1. Start interactive setup: `glab auth login`
1. Authenticate with the method appropriate for your GitLab instance:
   - For GitLab SaaS, authenticate against `gitlab.com` by reading the token
     from a file: `glab auth login --stdin < myaccesstoken.txt`
   - For self-managed instances, authenticate by reading from a file:
     `glab auth login --hostname gitlab.example.com --stdin < myaccesstoken.txt`. This will allow you to perform
     authenticated `glab` commands against a self-managed instance when you are in a Git repository with a remote
     matching your self-managed instance's host. Alternatively set `GITLAB_HOST` to direct your command to your self-managed instance.
   - Authenticate with token and hostname: `glab auth login --hostname gitlab.example.org --token xxxxx`
     Not recommended for shared environments.

## Configuration

By default, `glab` follows the
[XDG Base Directory Spec](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html).
Configure it globally, locally, or per host:

- **Globally**: run `glab config set --global editor vim`.
  - The global configuration file is available at `~/.config/glab-cli/config.yml`.
  - To override this location, set the `GLAB_CONFIG_DIR` environment variable.
- **The current repository**: run `glab config set editor vim` in any folder in a Git repository.
  - The local configuration file is available at `.git/glab-cli/config.yml` in the current working Git directory.
- **Per host**: run `glab config set editor vim --host gitlab.example.org`, changing
  the `--host` parameter to meet your needs.
  - Per-host configuration info is always stored in the global configuration file, with or without the `global` flag.

### Configure `glab` to use your self-managed instance

When outside a Git repository, `glab` uses `gitlab.com` by default. For `glab` to default
to your self-managed instance when you are not in a Git repository, change the host
configuration settings. Use this command, changing `gitlab.example.com` to the domain name
of your instance:

```shell
glab config set -g host gitlab.example.com
```

Setting this configuration enables you to perform commands outside a Git repository while
using your self-managed instance. For example:

- `glab repo clone group/project`
- `glab issue list -R group/project`

If you don't set a default domain name, you can declare one for the current command with
the `GITLAB_HOST` environment variable, like this:

- `GITLAB_HOST=gitlab.example.com glab repo clone group/project`
- `GITLAB_HOST=gitlab.example.com glab issue list -R group/project`

When inside a git repository `glab` will use that repository's GitLab host by default. For example `glab issue list`
will list all issues of the current directory's git repository.

### Configure `glab` to use self-signed certificates for self-managed instances

The GitLab CLI can be configured to support self-managed instances using self-signed certificate authorities by making either of the following changes:

You can disable TLS verification with:

```shell
glab config set skip_tls_verify true --host gitlab.example.com
```

Or add the path to the self signed CA:

```shell
glab config set ca_cert /path/to/server.pem --host gitlab.example.com
```

## Environment variables

- `GITLAB_TOKEN`: an authentication token for API requests. Setting this avoids being
  prompted to authenticate and overrides any previously stored credentials.
  Can be set in the config with `glab config set token xxxxxx`
- `GITLAB_URI` or `GITLAB_HOST`: specify the URL of the GitLab server if self-managed (eg: `https://gitlab.example.com`). Default is `https://gitlab.com`.
- `GITLAB_API_HOST`: specify the host where the API endpoint is found. Useful when there are separate (sub)domains or hosts for Git and the API endpoint: defaults to the hostname found in the Git URL
- `GITLAB_REPO`: Default GitLab repository used for commands accepting the `--repo` option. Only used if no `--repo` option is given.
- `GITLAB_GROUP`: Default GitLab group used for listing merge requests, issues and variables. Only used if no `--group` option is given.
- `REMOTE_ALIAS` or `GIT_REMOTE_URL_VAR`: `git remote` variable or alias that contains the GitLab URL.
- `GLAB_CONFIG_DIR`: Directory where glab's global configuration file is located. Defaults to `~/.config/glab-cli/` if not set.
  Can be set in the config with `glab config set remote_alias origin`
- `VISUAL`, `EDITOR` (in order of precedence): the editor tool to use for authoring text.
  Can be set in the config with `glab config set editor vim`
- `BROWSER`: the web browser to use for opening links.
   Can be set in the configuration with `glab config set browser mybrowser`
- `GLAMOUR_STYLE`: environment variable to set your desired Markdown renderer style
  Available options are (`dark`|`light`|`notty`) or set a [custom style](https://github.com/charmbracelet/glamour#styles)
- `NO_COLOR`: set to any value to avoid printing ANSI escape sequences for color output.
- `FORCE_HYPERLINKS`: set to `1` to force hyperlinks to be output, even when not outputting to a TTY

## Issues

If you have an issue: report it on the [issue tracker](https://gitlab.com/gitlab-org/cli/-/issues)

## Contributing

Feel like contributing? That's awesome! We have a [contributing guide](https://gitlab.com/gitlab-org/cli/-/blob/main/CONTRIBUTING.md) and [Code of conduct](https://gitlab.com/gitlab-org/cli/-/blob/main/CODE_OF_CONDUCT.md) to help guide you.

## Inspiration

The GitLab CLI was adopted from [Clement Sam](https://gitlab.com/profclems) in 2022 to serve as the official CLI of GitLab. Over the years the project has been inspired by both the [GitHub CLI](https://github.com/cli/cli) and [Zaq? Wiedmann's](https://gitlab.com/zaquestion) [lab](https://github.com/zaquestion/lab).

Lab has served as the foundation for many of the GitLab CI/CD commands including `ci view` and `ci trace`.
