# tw - terminal workspace manager

Manage multiple git repositories and worktrees using tmux or [cmux](https://cmux.dev), inspired by [Superset](https://github.com/nicepkg/superset)'s workspace UI.

tmux/cmux + git worktree + fzf를 활용해 여러 git 레포지토리와 워크트리를 관리하는 CLI 도구입니다.

---

## Features / 주요 기능

- **Multi-repo management** - Register and manage multiple git repositories from a single CLI
- **Git worktree integration** - Create isolated workspaces per branch using `git worktree`
- **Multi-backend** - Supports both tmux and cmux (auto-detected)
- **Interactive switching** - fzf-powered workspace selector with diff stats
- **Diff statistics** - See `+added -removed` line counts per workspace at a glance
- **Setup automation** - Auto-copy files and run commands on workspace creation

---

## Installation / 설치

### One-liner / 한 줄 설치

```bash
curl -fsSL https://raw.githubusercontent.com/kdjun99/tw/main/install.sh | sh
```

Custom install directory / 설치 경로 지정:

```bash
TW_INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/kdjun99/tw/main/install.sh | sh
```

### Go install

```bash
go install github.com/dongjunkim/tw@latest
```

### From source / 소스에서 빌드

```bash
git clone https://github.com/kdjun99/tw.git
cd tw
make install  # builds and copies to ~/.local/bin/
```

### Dependencies / 의존성

| Tool | Required | Purpose |
|------|----------|---------|
| `git` | Yes | Worktree management |
| `tmux` or `cmux` | Yes (one of) | Session/window management |
| `fzf` | Optional | Interactive workspace switching (`tw switch`) |

---

## Terminal Backend / 터미널 백엔드

tw supports two terminal backends:

tw는 두 가지 터미널 백엔드를 지원합니다:

| Backend | Description |
|---------|-------------|
| **tmux** | Traditional terminal multiplexer with sessions and windows |
| **cmux** | [cmux](https://cmux.dev) - Ghostty-based macOS terminal with vertical sidebar |

### Auto-detection / 자동 감지

tw automatically detects the best backend:

1. If running inside cmux → uses cmux
2. If running inside tmux → uses tmux
3. If cmux is installed → uses cmux
4. Fallback → uses tmux

### Manual configuration / 수동 설정

Set the `backend` field in `~/.config/tw/config.json`:

`~/.config/tw/config.json`에서 `backend` 필드를 설정합니다:

```json
{
  "backend": "cmux",
  "projects": [...]
}
```

Values: `"tmux"`, `"cmux"`, or omit for auto-detect.

### cmux setup / cmux 설정

To use cmux as backend:

1. Install [cmux](https://cmux.dev)
2. In cmux Settings → Automation → Socket Control Mode → set to **Automation mode**
3. Install cmux CLI: `Cmd+Shift+P` → "Install CLI"

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

# 4. Attach to project / 프로젝트 세션 연결
tw attach my-app

# 5. Switch workspace interactively / 인터랙티브 전환
tw switch

# 6. Remove workspace / 워크스페이스 제거
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

# Edit project config / 프로젝트 설정 편집
tw project edit <name>   # opens in $EDITOR with comments

# Remove a project / 프로젝트 제거
tw project rm <name>
```

| Flag | Description |
|------|-------------|
| `--default-branch` | Base branch for new worktrees (auto-detected if omitted) / 새 워크트리의 기본 브랜치 (생략 시 자동 감지) |
| `--worktree-dir` | Custom directory for worktrees (default: `~/.tw/<name>/`) / 워크트리 저장 디렉토리 |
| `--copy` | Files to copy to new worktrees (e.g. `--copy '.env*'`) / 워크트리에 복사할 파일 |
| `--setup-run` | Commands to run after worktree creation (e.g. `--setup-run 'npm install'`) / 생성 후 실행 명령 |
| `--teardown-run` | Commands to run before worktree removal / 삭제 전 실행 명령 |

### `tw add` - Create workspace / 워크스페이스 생성

Creates a git worktree and a terminal window in one command.

git worktree와 터미널 윈도우를 한 번에 생성합니다.

```bash
tw add <project> <branch> [flags]
```

| Flag | Description |
|------|-------------|
| `--base <branch>` | Base branch to create from (default: project's default branch) / 분기 기준 브랜치 |
| `--existing` | Checkout an existing branch instead of creating new / 기존 브랜치 체크아웃 |
| `--no-terminal` | Create worktree only, skip terminal window / worktree만 생성 |
| `--no-setup` | Skip setup scripts / 셋업 스크립트 생략 |
| `-s, --switch` | Auto-switch to new window (default: true) / 생성 후 자동 전환 |

**Examples / 예시:**

```bash
# New feature branch from main / main에서 새 feature 브랜치
tw add my-app feature/auth --base main

# Checkout existing branch / 기존 브랜치 체크아웃
tw add my-app fix/bug-123 --existing

# Worktree only (no terminal) / worktree만 생성
tw add my-app experiment/new-idea --no-terminal
```

### `tw attach` - Attach to project / 프로젝트 연결

Attaches to a project's terminal session, creating windows for all worktrees.

프로젝트의 터미널 세션에 연결하고, 모든 워크트리에 대해 윈도우를 생성합니다.

```bash
tw attach <project>              # attach to project session
tw attach <project>/<window>     # attach to specific window
tw attach                        # interactive picker (requires fzf)
```

### `tw list` - List workspaces / 워크스페이스 목록

```bash
tw list        # alias: tw ls
```

### `tw status` - Diff statistics / 변경 통계

```bash
tw status   # alias: tw st
```

### `tw switch` - Interactive switch / 인터랙티브 전환

```bash
tw switch   # aliases: tw sw, tw s
```

### `tw rm` - Remove workspace / 워크스페이스 제거

```bash
tw rm <project> <branch> [flags]
```

| Flag | Description |
|------|-------------|
| `--keep-worktree` | Only close terminal window, keep worktree / 윈도우만 닫고 worktree 유지 |
| `--keep-terminal` | Only remove worktree, keep terminal window / worktree만 삭제, 윈도우 유지 |

---

## How it works / 동작 원리

```
Superset UI concept          →  tw
──────────────────────────────────────────
Project (repository)         →  terminal session (tmux session / cmux workspace group)
Workspace (worktree)         →  terminal window (tmux window / cmux workspace)
Left sidebar (repo tree)     →  tw list / tw status / cmux sidebar
Add Workspace button         →  tw add <project> <branch>
Click to switch              →  tw switch (fzf) / tw attach
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

### Config / 설정

Configuration is stored at `~/.config/tw/config.json`.

설정 파일은 `~/.config/tw/config.json`에 저장됩니다.

```json
{
  "backend": "auto",
  "projects": [
    {
      "name": "my-app",
      "path": "/Users/me/dev/my-app",
      "defaultBranch": "main",
      "setup": {
        "copy": [".env*"],
        "run": ["npm install"]
      },
      "teardown": {
        "run": ["docker compose down"]
      }
    }
  ]
}
```

---

## License

MIT
