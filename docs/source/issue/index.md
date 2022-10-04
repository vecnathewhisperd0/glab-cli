## glab issue

Work with GitLab issues

### Examples

```
$ glab issue list
$ glab issue create --label --confidential
$ glab issue view --web
$ glab issue note -m "closing because !123 was merged" <issue number>

```

### Options

```
  -R, --repo OWNER/REPO   Select another repository using the OWNER/REPO or `GROUP/NAMESPACE/REPO` format or full URL or git URL
```

### Options inherited from parent commands

```
      --help   Show help for command
```

### Subcommands

- [board](board.md)
- [close](close.md)
- [create](create.md)
- [delete](delete.md)
- [list](list.md)
- [note](note.md)
- [reopen](reopen.md)
- [subscribe](subscribe.md)
- [unsubscribe](unsubscribe.md)
- [update](update.md)
- [view](view.md)

