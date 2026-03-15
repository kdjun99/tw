# tw - terminal workspace manager

CLI tool that manages multiple git repositories and worktrees using tmux or cmux, with fzf-powered interactive switching.

## Architecture

```
main.go                    # entrypoint
cmd/                       # cobra command definitions
  root.go                  # root command + backend initialization
  add.go                   # create worktree + terminal window
  rm.go                    # remove worktree + terminal window
  attach.go                # attach to project session
  switch.go                # fzf interactive workspace switcher
  list.go                  # tree-view of all projects/worktrees
  status.go                # diff stats per workspace
  project.go               # project CRUD (add/rm/list)
  project_edit.go          # edit project config in $EDITOR
internal/
  terminal/                # multi-backend terminal abstraction
    backend.go             # Backend interface definition
    tmux.go                # tmux backend (wraps internal/tmux)
    cmux.go                # cmux backend (cmux CLI calls)
    detect.go              # auto-detection + factory
  config/config.go         # JSON config (~/.config/tw/config.json)
  git/git.go               # git worktree operations
  tmux/tmux.go             # low-level tmux operations
  setup/setup.go           # workspace setup/teardown automation
  ui/fzf.go                # fzf integration
```

## Tech Stack

- Go 1.24+
- [cobra](https://github.com/spf13/cobra) for CLI framework
- Multi-backend: tmux (default) or cmux (auto-detected)
- No external Go dependencies beyond cobra (git/tmux/cmux/fzf via exec)

## Build & Test

```bash
make build          # go build -o tw .
make install        # build + copy to ~/.local/bin/
go test ./...       # run all tests
```

## Terminal Backend

tw supports two terminal backends via the `Backend` interface:

- **tmux** (default): traditional tmux sessions/windows
- **cmux**: native macOS terminal with vertical sidebar (uses cmux CLI)

Backend selection (in `~/.config/tw/config.json`):
```json
{
  "backend": "auto",
  "projects": [...]
}
```

Values: `"tmux"`, `"cmux"`, `"auto"` (or omit for auto-detect).
Auto-detect priority: cmux env → tmux env → cmux installed → tmux fallback.

## Conventions

### Code Style
- Standard `gofmt` formatting (tabs, no config needed)
- Error wrapping with `%w` for all returned errors
- Warnings printed to stdout for non-fatal errors (e.g., setup/teardown failures)
- No third-party test frameworks — use stdlib `testing` only

### Naming
- Commands use kebab-case flags: `--no-terminal`, `--keep-worktree`, `--default-branch`
- Internal packages are single-purpose: `git`, `tmux`, `terminal`, `config`, `setup`, `ui`
- Helper functions stay in the file that uses them (e.g., `shortBranch` in `add.go`, `shortenPath` in `list.go`)
- Backend-agnostic naming: "terminal window" instead of "tmux window"

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
- Backend tests verify interface compliance and factory logic

### Config
- Config path: `~/.config/tw/config.json`
- Worktrees default to: `~/.tw/<project-name>/`
- Branch slashes converted to hyphens in worktree directory names: `feature/login` → `feature-login`
- Backend field: `"backend"` in config.json (optional, defaults to auto-detect)

### Safety
- `git worktree remove` runs without `--force` to protect uncommitted changes
- Teardown commands are best-effort (failures are warnings, not errors)
- Setup command failures are fatal and stop workspace creation
