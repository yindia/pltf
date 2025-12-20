# Installation

> pltf is in active development. Pin a release when installing for production workflows.

## Homebrew (macOS/Linux)
```bash
brew tap yindia/pltf
brew install pltf
```
If you had a different tap, `brew untap <old>` first.

## Install script (macOS/Linux/Windows via WSL or Git Bash)
```bash
curl -sSL https://raw.githubusercontent.com/yindia/pltf/main/scripts/install.sh | sh
```
Environment overrides:
- `REPO_OWNER` / `REPO_NAME`: install from a fork
- `VERSION`: pin a release tag (e.g., `v0.1.0`)
- `DEST`: install path (default `/usr/local/bin`)

## Docker
```bash
docker run --rm ghcr.io/yindia/pltf:sha-713a58e --help
```

## From source
```bash
go install ./...
# or build a local binary
go build -o bin/pltf main.go
```

## Verify the install
```bash
pltf --help
pltf validate -f example/env.yaml
```

## Upgrade
- Homebrew: `brew upgrade pltf`
- Script: rerun the install script with your desired `VERSION`
- Source: `git pull` then rebuild
