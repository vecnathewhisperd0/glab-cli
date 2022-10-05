## glab repo clone

Clone a Gitlab repository/project

### Synopsis

Clone a GitLab repository/project

	Clone supports these shorthands
	- repo
	- namespace/repo
	- org/group/repo
	- project ID
	

```
glab repo clone <repo> [<dir>] [-- [<gitflags>...]] [flags]
```

### Examples

```
$ glab repo clone profclems/glab

$ glab repo clone https://gitlab.com/profclems/glab

$ glab repo clone profclems/glab mydirectory  # Clones repo into mydirectory

$ glab repo clone glab   # clones repo glab for current user 

$ glab repo clone 4356677   # finds the project by the ID provided and clones it

# Clone all repos in a group
$ glab repo clone -g everyonecancontribute --paginate

# Clone from a self-hosted instance
$ GITLAB_HOST=salsa.debian.org glab repo clone myrepo  

```

### Options

```
  -g, --group string          Specify group to clone repositories from
  -p, --preserve-namespace    Clone the repo in a subdirectory based on namespace
  -a, --archived              Limit by archived status. Used with --group flag
  -G, --include-subgroups     Include projects in subgroups of this group. Default is true. Used with --group flag (default true)
  -m, --mine                  Limit by projects in the group owned by the current authenticated user. Used with --group flag
  -v, --visibility string     Limit by visibility {public, internal, or private}. Used with --group flag
  -I, --with-issues-enabled   Limit by projects with issues feature enabled. Default is false. Used with --group flag
  -M, --with-mr-enabled       Limit by projects with issues feature enabled. Default is false. Used with --group flag
  -S, --with-shared           Include projects shared to this group. Default is false. Used with --group flag
      --paginate              Make additional HTTP requests to fetch all pages of projects before cloning. Respects --per-page
      --page int              Page number (default 1)
      --per-page int          Number of items to list per page (default 30)
```

### Options inherited from parent commands

```
      --help   Show help for command
```

