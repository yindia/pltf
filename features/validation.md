# Validation & Lint

Catch spec issues early before generation or apply.

## What it does
- `pltf validate` runs structural validation for Environment and Service specs.
- Built-in lint suggests labels and flags unused variables.
- Auto-detects `kind` (env/service) and applies the right checks.

## Example
```bash
pltf validate -f env.yaml -e prod
pltf validate -f service.yaml -e dev
```

## Notes
- Lint also runs implicitly during validate.
- Combine with `pltf preview` to sanity check providers/backends/modules.
