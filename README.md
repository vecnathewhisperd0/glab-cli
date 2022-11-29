# GLab

![GLab](docs/assets/glab-logo.png)

[![Go Report Card](https://goreportcard.com/badge/gitlab.com/gitlab-org/cli)](https://goreportcard.com/report/gitlab.com/gitlab-org/cli)
![Coverage](https://gitlab.com/gitlab-org/cli/badges/main/coverage.svg)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#version-control)
[![Gitpod Ready-to-Code](https://img.shields.io/badge/Gitpod-Ready--to--Code-blue?style=flat&logo=gitpod&logoColor=white)](https://gitpod.io/#https://gitlab.com/gitlab-org/cli/-/tree/main/)

GLab is an open source GitLab CLI tool bringing GitLab to your terminal next to where you are already working with `git` and your code without switching between windows and browser tabs. Work with issues, merge requests, **watch running pipelines directly from your CLI** among other features.

`glab` is available for repositories hosted on GitLab.com and self-managed GitLab instances. `glab` supports multiple authenticated GitLab instances and automatically detects the authenticated hostname from the remotes available in the working Git directory.

![command example](docs/assets/command-example.png)

## Table of contents

- [Usage](#usage)
- [Demo](#demo)
- [Documentation](#documentation)
- [Installation](#installation)
  - [macOS](#macos)
  - [Windows](#windows)
  - [Linux](#linux)
    - [Linuxbrew (Homebrew)](#linuxbrew-homebrew)
    - [Snapcraft](#snapcraft)
    - [Arch Linux](#arch-linux)
    - [KISS Linux](#kiss-linux)
    - [Alpine Linux](#alpine-linux)
      - [Install a pinned version from edge](#install-a-pinned-version-from-edge)
      - [Alpine Linux Docker-way](#alpine-linux-docker-way)
    - [Nix/NixOS](#nixnixos)
    - [MPR (Debian/Ubuntu)](#mpr-debianubuntu)
      - [Prebuilt-MPR](#prebuilt-mpr)
    - [Spack](#spack)
  - [Building From Source](#building-from-source)
    - [Prerequisites](#prerequisites-for-building-from-source-are)
- [Authentication](#authentication)
- [Configuration](#configuration)
- [Environment Variables](#environment-variables)
- [What about lab](#what-about-lab)
- [Issues](#issues)
- [Contributing](#contributing)

## Usage

```shell
glab <command> <subcommand> [flags]
```

## Demo

[![asciicast](https://asciinema.org/a/368622.svg)](https://asciinema.org/a/368622)

## Documentation

Read the [documentation](https://gitlab.com/gitlab-org/cli/-/tree/main/docs/source) for usage instructions or check out `glab help`.

## Installation

Download a binary suitable for your OS at the [releases page](https://gitlab.com/gitlab-org/cli/-/releases).
Other installation methods depend on your operating system.

### macOS

- Install from Homebrew: `brew install glab`
  - Update from Homebrew: `brew upgrade glab`
- Install from [MacPorts](https://ports.macports.org/port/glab/summary): `sudo port install glab`
  - Update from MacPorts: `sudo port selfupdate && sudo port upgrade glab`
- Install into `usr/bin` with a shell script:
  `curl -s "https://gitlab.com/gitlab-org/cli/-/raw/main/scripts/install.sh" | sudo sh`

  Before running any install script, review its contents.

### Windows

- Download from [WinGet](https://github.com/microsoft/winget-cli): `winget install glab.glab`
  - Update from WinGet: `winget install glab.glab`
- Download from [scoop](https://scoop.sh): `scoop install glab`
  - Update from scoop: `scoop update glab`
- Download an EXE installer file from the [releases page](https://gitlab.com/gitlab-org/cli/-/releases)

### Linux

- Download prebuilt binaries from the [releases page](https://gitlab.com/gitlab-org/cli/-/releases)
- Install from Homebrew: `brew install glab`
- Update from Homebrew: `brew upgrade glab`

#### Snapcraft

To install `glab` from the [Snap Store](https://snapcraft.io/glab):

1. Make sure you have [snap installed](https://snapcraft.io/docs/installing-snapd) on your Linux distribution.
1. Install the package: `sudo snap install --edge glab`
1. Grant `glab` access to SSH keys: `sudo snap connect glab:ssh-keys`

[![Download from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/glab)

#### Arch Linux

For Arch Linux, `glab` is available:

- From the [`community/glab`](https://archlinux.org/packages/community/x86_64/glab/) package.
- By downloading and installing an archive from the
  [releases page](https://gitlab.com/gitlab-org/cli/-/releases).
- From the [Snap Store](https://snapcraft.io/glab), if
  [snap](https://snapcraft.io/docs/installing-snap-on-arch-linux) is installed.
- Installing with the package manager: `pacman -S glab`

#### KISS Linux

WARNING:
KISS Linux may no longer be actively maintained.

`glab` is available on the [KISS Linux Community Repository](https://github.com/kisslinux/community) as `gitlab-glab`.
If you already have the community repository configured in your `KISS_PATH` you can install `glab` through your terminal:

```shell
kiss b gitlab-glab && kiss i gitlab-glab
```

#### Alpine Linux

`glab` is available on the [Alpine Community Repository](https://git.alpinelinux.org/aports/tree/community/glab?h=master) as `glab`.

When installing, use `--no-cache` so no `apk update` is required:

```shell
apk add --no-cache glab
```

##### Install a pinned version from edge

To ensure that by default edge is used to get the latest updates. We need the edge repository in `/etc/apk/repositories`.

Afterwards you can install it with `apk add --no-cache glab@edge`

We use `--no-cache` so an `apk update` is not required.

```shell
echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
apk add --no-cache glab@edge
```

##### Alpine Linux Docker-way

Use edge directly

```shell
FROM alpine:3.13
RUN apk add --no-cache glab
```

Fetching latest glab version from edge

```shell
FROM alpine:3.13
RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
RUN apk add --no-cache glab@edge
```

#### Nix/NixOS

Nix (NixOS) users can install from [nixpkgs](https://search.nixos.org/packages?channel=unstable&show=glab&from=0&size=30&sort=relevance&query=glab):

```shell
nix-env -iA nixos.glab
```

#### MPR (Debian/Ubuntu)

`glab` is available inside the [makedeb package repository](https://mpr.makedeb.org/packages/glab). To install, run the following:

```shell
git clone 'https://mpr.makedeb.org/glab'
cd glab/
makedeb -si
```

##### Prebuilt-MPR

The above method downloads glab from source and builds it before packaging it into a `.deb` package. If you don't want to compile or just want a prebuilt package, you can also install glab from the Prebuilt-MPR.

First [set up the Prebuilt-MPR on your system](https://docs.makedeb.org/prebuilt-mpr/getting-started/#setting-up-the-repository), and then run the following:

```shell
sudo apt install glab
```

#### Spack

```shell
spack install glab
```

Updating (Spack):

```shell
spack uninstall glab && spack install glab
```

### Building from source

If a supported binary for your OS is not found at the [releases page](https://gitlab.com/gitlab-org/cli/-/releases), you can build from source:

#### Prerequisites for building from source

- `make`
- Go 1.13+

1. Verify that you have Go 1.13+ installed

   ```shell
   $ go version
   go version go1.14
   ```

   If `go` is not installed, follow instructions on [the Go website](https://golang.org/doc/install).
1. Clone this repository

   ```shell
   git clone https://gitlab.com/gitlab-org/cli.git glab
   cd glab
   ```

   If you have `$GOPATH/bin` or `$GOBIN` in your `$PATH`, you can just install with `make install` (install glab in `$GOPATH/bin`) and **skip steps 3 and 4**.
1. Build the project:

   ```shell
   make
   ```

1. Change PATH to find newly compiled `glab`

   ```shell
   export PATH=$PWD/bin:$PATH
   ```

1. Run `glab version` to confirm that it worked.

## Authentication

Get a GitLab access token at <https://gitlab.com/-/profile/personal_access_tokens> or
`https://gitlab.example.com/-/profile/personal_access_tokens` if self-managed:

1. Start interactive setup: `glab auth login`
1. Authenticate with the method appropriate for your GitLab instance:
   - For GitLab SaaS, authenticate against `gitlab.com` by reading the token
     from a file: `glab auth login --stdin < myaccesstoken.txt`
   - For self-managed instances, authenticate by reading from a file:
     `glab auth login --hostname salsa.debian.org --stdin < myaccesstoken.txt`
   - Authenticate with token and hostname: `glab auth login --hostname gitlab.example.org --token xxxxx`
     Not recommended for shared environments.

## Configuration

By default, `glab` follows the XDG Base Directory [Spec](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html):

- The global configuration file is saved at `~/.config/glab-cli`.
- The local configuration file is saved at `.git/glab-cli` in the current working Git directory.
- Advanced workflows may override the location of the global configuration by setting the `GLAB_CONFIG_DIR` environment variable.

- **To set configuration globally**

  ```shell
  glab config set --global editor vim
  ```

- **To set configuration for current directory (must be a Git repository)**

  ```shell
  glab config set editor vim
  ```

- **To set configuration for a specific host**

  Use the `--host` flag to set configuration for a specific host. This configuration is always stored in the global configuration file, with or without the `global` flag.

  ```shell
  glab config set editor vim --host gitlab.example.org
  ```

## Environment variables

- `GITLAB_TOKEN`: an authentication token for API requests. Setting this avoids being
  prompted to authenticate and overrides any previously stored credentials.
  Can be set in the config with `glab config set token xxxxxx`
- `GITLAB_URI` or `GITLAB_HOST`: specify the URL of the GitLab server if self-managed (eg: `https://gitlab.example.com`). Default is `https://gitlab.com`.
- `GITLAB_API_HOST`: specify the host where the API endpoint is found. Useful when there are separate (sub)domains or hosts for Git and the API endpoint: defaults to the hostname found in the Git URL
- `REMOTE_ALIAS` or `GIT_REMOTE_URL_VAR`: `git remote` variable or alias that contains the GitLab URL.
  Can be set in the config with `glab config set remote_alias origin`
- `VISUAL`, `EDITOR` (in order of precedence): the editor tool to use for authoring text.
  Can be set in the config with `glab config set editor vim`
- `BROWSER`: the web browser to use for opening links.
   Can be set in the configuration with `glab config set browser mybrowser`
- `GLAMOUR_STYLE`: environment variable to set your desired Markdown renderer style
  Available options are (`dark`|`light`|`notty`) or set a [custom style](https://github.com/charmbracelet/glamour#styles)
- `NO_COLOR`: set to any value to avoid printing ANSI escape sequences for color output.
- `FORCE_HYPERLINKS`: set to `1` to force hyperlinks to be output, even when not outputing to a TTY

## Issues

If you have an issue: report it on the [issue tracker](https://gitlab.com/gitlab-org/cli/-/issues)

## Contributing

Feel like contributing? That's awesome! We have a [contributing guide](https://gitlab.com/gitlab-org/cli/-/blob/main/CONTRIBUTING.md) and [Code of conduct](https://gitlab.com/gitlab-org/cli/-/blob/main/CODE_OF_CONDUCT.md) to help guide you.
