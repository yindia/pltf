# Telemetry

Optional usage reporting.

## What it does
- Uses a global `--telemetry` flag (defaults from profile) to enable/disable future analytics.
- Currently a stub/no-op; reserved for opt-in reporting.

## Usage
- Set in profile:
  ```yaml
  telemetry: false
  ```
- Or export `PLTF_TELEMETRY=0` to disable.
