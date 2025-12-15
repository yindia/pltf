# Preview

Quickly inspect what will be generated without touching Terraform.

## What it does
- `pltf preview` reads your spec and shows provider, backend type, labels, and modules that will render.
- Auto-detects env vs service based on `kind`.

## Example
```bash
pltf preview -f env.yaml -e prod
```

## Notes
- No cloud credentials needed; useful in CI or pre-commit checks.
- Pair with `pltf validate` for faster feedback before generation/apply.
