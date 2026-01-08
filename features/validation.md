# Validation

Catch spec issues early before generation or apply.

## What it does
- `pltf validate` runs structural validation for Environment, Service, and Stack specs.
- Auto-detects `kind` and applies the right checks.

## Example
```bash
pltf validate -f env.yaml -e prod
pltf validate -f service.yaml -e dev
```

## Notes
- Combine with `pltf preview` to sanity check providers/backends/modules.
