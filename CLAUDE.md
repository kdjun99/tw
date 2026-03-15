# tw - tmux workspace manager

CLI tool that manages multiple git repositories and worktrees using tmux sessions/windows, with fzf-powered interactive switching.

## Architecture

```
main.go                    # entrypoint
cmd/                       # cobra command definitions
  root.go                  # root command
  add.go                   # create worktree + tmux window
  rm.go                    # remove worktree + tmux window
  switch.go                # fzf interactive workspace switcher
  list.go                  # tree-view of all projects/worktrees
  status.go                # diff stats per workspace
  project.go               # project CRUD (add/rm/list)
internal/
  config/config.go         # JSON config (~/.config/tw/config.json)
  git/git.go               # git worktree operations
  tmux/tmux.go             # tmux session/window operations
  setup/setup.go           # workspace setup/teardown automation
  ui/fzf.go                # fzf integration
```

## Tech Stack

- Go 1.24+
- [cobra](https://github.com/spf13/cobra) for CLI framework
- No external Go dependencies beyond cobra (git/tmux/fzf via exec)

## Build & Test

```bash
make build          # go build -o tw .
make install        # build + copy to ~/.local/bin/
go test ./...       # run all tests
```

## Conventions

### Code Style
- Standard `gofmt` formatting (tabs, no config needed)
- Error wrapping with `%w` for all returned errors
- Warnings printed to stdout for non-fatal errors (e.g., setup/teardown failures)
- No third-party test frameworks — use stdlib `testing` only

### Naming
- Commands use kebab-case flags: `--no-tmux`, `--keep-worktree`, `--default-branch`
- Internal packages are single-purpose: `git`, `tmux`, `config`, `setup`, `ui`
- Helper functions stay in the file that uses them (e.g., `shortBranch` in `add.go`, `shortenPath` in `list.go`)

### Git & Release
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `chore:`, `test:`, `docs:`, `refactor:`
- Version tags: `v0.x.y` (semver)
- GitHub releases include detailed changelogs categorized by: Bug Fixes, Features, Tests, Chore
- Include `Co-Authored-By` trailer when commits are AI-assisted

### Testing
- Test files use `_test.go` suffix in the same package
- Integration tests for git operations use real temp repos (`t.TempDir()` + `git init`)
- Config tests override `$HOME` to isolate from user config
- Pure function tests are table-driven

### Config
- Config path: `~/.config/tw/config.json`
- Worktrees default to: `~/.tw/<project-name>/`
- Branch slashes converted to hyphens in worktree directory names: `feature/login` → `feature-login`

### Safety
- `git worktree remove` runs without `--force` to protect uncommitted changes
- Teardown commands are best-effort (failures are warnings, not errors)
- Setup command failures are fatal and stop workspace creation
