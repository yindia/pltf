# Installation

## Homebrew (macOS/Linux)
```bash
brew tap yindia/pltf
brew install pltf
```

If you previously tapped another repo, run `brew untap <old>` before tapping `yindia/pltf`.

## Install script (macOS/Linux/Windows via WSL or Git Bash)
```bash
curl -sSL https://raw.githubusercontent.com/yindia/pltf/main/scripts/install.sh | bash
```
Environment overrides:
- `REPO_OWNER` / `REPO_NAME` to point at a fork
- `VERSION` to pin a release tag (e.g. `v1.2.3`)
- `DEST` to change install path (defaults to `/usr/local/bin`)

## Docker
```bash
docker build -t pltf .
docker run --rm pltf --help
```

## From source
```bash
go install ./...
# or to build a local binary
go build -o bin/pltf main.go
```

## Verify
```bash
pltf --help
pltf validate -f env.yaml
```

## Upgrade
- Homebrew: `brew upgrade pltf`
- Script: rerun the install script with the desired `VERSION`
- From source: `git pull` then rebuild
