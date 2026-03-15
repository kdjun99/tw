# tw - tmux workspace manager

Manage multiple git repositories and worktrees using tmux, inspired by [Superset](https://github.com/nicepkg/superset)'s workspace UI.

tmux + git worktree + fzf를 활용해 여러 git 레포지토리와 워크트리를 관리하는 CLI 도구입니다.

---

## Features / 주요 기능

- **Multi-repo management** - Register and manage multiple git repositories from a single CLI
- **Git worktree integration** - Create isolated workspaces per branch using `git worktree`
- **tmux session/window mapping** - Each project becomes a tmux session, each worktree becomes a window
- **Interactive switching** - fzf-powered workspace selector with diff stats
- **Diff statistics** - See `+added -removed` line counts per workspace at a glance

---

## Installation / 설치

### From source / 소스에서 빌드

Requires Go 1.21+ and [fzf](https://github.com/junegunn/fzf) for interactive switching.

Go 1.21+ 및 인터랙티브 전환을 위한 [fzf](https://github.com/junegunn/fzf)가 필요합니다.

```bash
git clone https://github.com/kdjun99/tw.git
cd tw
go build -o tw .
cp tw ~/.local/bin/  # or anywhere in your $PATH
```

### Dependencies / 의존성

| Tool | Required | Purpose |
|------|----------|---------|
| `git` | Yes | Worktree management |
| `tmux` | Yes | Session/window management |
| `fzf` | Optional | Interactive workspace switching (`tw switch`) |

---

## Quick Start / 빠른 시작

```bash
# 1. Register projects / 프로젝트 등록
tw project add my-app ~/dev/my-app
tw project add api-server ~/dev/api-server --default-branch develop

# 2. Create a workspace / 워크스페이스 생성
tw add my-app feature/login --base main

# 3. See all workspaces / 전체 워크스페이스 조회
tw list

# 4. Switch workspace interactively / 인터랙티브 전환
tw switch

# 5. Remove workspace / 워크스페이스 제거
tw rm my-app feature/login
```

---

## Commands / 명령어

### `tw project` - Manage projects / 프로젝트 관리

```bash
# Register a project / 프로젝트 등록
tw project add <name> <path> [--default-branch main] [--worktree-dir <dir>]

# List projects / 프로젝트 목록
tw project list   # alias: tw project ls

# Remove a project / 프로젝트 제거
tw project rm <name>
```

| Flag | Description |
|------|-------------|
| `--default-branch` | Base branch for new worktrees (auto-detected if omitted) / 새 워크트리의 기본 브랜치 (생략 시 자동 감지) |
| `--worktree-dir` | Custom directory for worktrees (default: `../<name>-worktrees/`) / 워크트리 저장 디렉토리 (기본: `../<이름>-worktrees/`) |

### `tw add` - Create workspace / 워크스페이스 생성

Creates a git worktree and a tmux window in one command.

git worktree와 tmux window를 한 번에 생성합니다.

```bash
tw add <project> <branch> [flags]
```

| Flag | Description |
|------|-------------|
| `--base <branch>` | Base branch to create from (default: project's default branch) / 분기 기준 브랜치 |
| `--existing` | Checkout an existing branch instead of creating new / 기존 브랜치 체크아웃 |
| `--no-tmux` | Create worktree only, skip tmux window / worktree만 생성, tmux 생략 |
| `-s, --switch` | Auto-switch to new window (default: true) / 생성 후 자동 전환 |

**Examples / 예시:**

```bash
# New feature branch from main / main에서 새 feature 브랜치
tw add my-app feature/auth --base main

# Checkout existing branch / 기존 브랜치 체크아웃
tw add my-app fix/bug-123 --existing

# Worktree only (no tmux) / worktree만 생성
tw add my-app experiment/new-idea --no-tmux
```

### `tw list` - List workspaces / 워크스페이스 목록

Displays all projects and their worktrees in a tree view with diff stats.

모든 프로젝트와 워크트리를 트리뷰로 표시합니다.

```bash
tw list        # alias: tw ls
tw list --all  # show full details
```

**Output example / 출력 예시:**

```
superset (main)
  ├── main                     ~/dev/superset
  ├── feature/test-worktree    ~/dev/superset-worktrees/feature-test-worktree

tw (main)
  ├── main                     ~/dev/tw
  ├── feature/add-readme       ~/dev/tw-worktrees/feature-add-readme
```

### `tw status` - Diff statistics / 변경 통계

Shows line-level change statistics for all workspaces.

모든 워크스페이스의 변경사항 통계를 표시합니다.

```bash
tw status   # alias: tw st
```

**Output example / 출력 예시:**

```
superset
  ├── main (local)              clean
  ├── feature/test-worktree     +1 -1

tw
  ├── main (local)              clean
  ├── feature/add-readme        clean
```

### `tw switch` - Interactive switch / 인터랙티브 전환

Opens an fzf selector to pick and switch to a workspace. Requires fzf.

fzf 선택기로 워크스페이스를 선택하고 전환합니다. fzf 필요.

```bash
tw switch   # aliases: tw sw, tw s
```

### `tw rm` - Remove workspace / 워크스페이스 제거

Removes both the git worktree and tmux window.

git worktree와 tmux window를 모두 제거합니다.

```bash
tw rm <project> <branch> [flags]
```

| Flag | Description |
|------|-------------|
| `--keep-worktree` | Only close tmux window, keep worktree / tmux만 닫고 worktree 유지 |
| `--keep-tmux` | Only remove worktree, keep tmux window / worktree만 삭제, tmux 유지 |

---

## How it works / 동작 원리

```
Superset UI concept          →  tw (tmux)
──────────────────────────────────────────
Project (repository)         →  tmux session
Workspace (worktree)         →  tmux window
Left sidebar (repo tree)     →  tw list / tw status
Add Workspace button         →  tw add <project> <branch>
Click to switch              →  tw switch (fzf)
Diff stats (+1893 -4)        →  tw status (git diff)
```

### Directory structure / 디렉토리 구조

All worktrees are stored under `~/.tw/` by default, keeping your dev directory clean.

모든 워크트리는 기본적으로 `~/.tw/` 아래에 저장되어 dev 디렉토리를 깔끔하게 유지합니다.

```
~/.tw/                             # centralized worktree storage
├── my-app/
│   ├── feature-login/             # worktree for feature/login branch
│   └── fix-bug-123/               # worktree for fix/bug-123 branch
└── api-server/
    └── feature-dashboard/         # worktree for feature/dashboard branch

~/dev/
├── my-app/                        # main repo (registered project)
└── api-server/                    # another registered project
```

You can override the worktree directory per project with `--worktree-dir`.

프로젝트별로 `--worktree-dir`로 워크트리 저장 경로를 변경할 수 있습니다.

### Config / 설정

Configuration is stored at `~/.config/tw/config.json`.

설정 파일은 `~/.config/tw/config.json`에 저장됩니다.

```json
{
  "projects": [
    {
      "name": "my-app",
      "path": "/Users/me/dev/my-app",
      "defaultBranch": "main"
    }
  ]
}
```

---

## License

MIT
