# Placeholders & Wiring

Lightweight templating to keep specs DRY and wire modules together.

## What it does
- Intrinsics: `${env_name}`, `${layer_name}`.
- References: `${module.<id>.<output>}`, `${parent.<output>}` (services), `${var.<name>}`.
- Auto-wires inputs to outputs when names match within scope; missing required values fail validation.

## Examples
```yaml
public_uri: "https://${module.dns.domain}"
bucket_name: "app-${env_name}"
max_nodes: "${var.max_nodes}"
public_url: "${parent.domain}/hello"   # in a service spec
```

## Notes
- Services can reference parent env outputs via `${parent.*}`.
- Variables precedence: env vars → service envRef vars → CLI `--var`.
