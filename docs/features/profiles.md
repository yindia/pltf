# Profiles & Defaults

Set org-wide defaults so users type fewer flags and stay consistent.

## What it does
- Reads `~/.pltf/profile.yaml` (or `PLTF_PROFILE`) for defaults like `modules_root`, `default_env`, `default_out`, and `telemetry`.
- Lets you pick a custom modules root for all commands without repeating `--modules`.
- Allows a default environment name so `--env` can be omitted when unambiguous.

## Example profile
```yaml
modules_root: /infra/modules
default_env: dev
default_out: .pltf
telemetry: false
```

## Usage
- Any CLI flags you pass override profile settings.
- Profiles are optional; when absent, embedded modules and CLI flags are used.
