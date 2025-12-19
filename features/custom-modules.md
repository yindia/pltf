# Custom Modules

Mix embedded modules with your own Terraform modules.

## What it does
- Uses the embedded catalog by default.
- Supports a custom modules root (`--modules` or profile `modules_root`).
- `source: custom` on a module forces lookup in your custom root; others fall back to embedded.
- `pltf module init` inspects a TF module and writes `module.yaml` metadata.
- Inventory commands: `pltf module list|get -o table|json|yaml`.

## Example
```yaml
modules:
  - name: app
    type: my_custom_service
    source: custom
    image: ghcr.io/acme/app:latest
```

## Notes
- Custom and embedded modules can coexist in the same spec.
- Module metadata (`module.yaml`) drives inputs/outputs and wiring; keep it committed.
