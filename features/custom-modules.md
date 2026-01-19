# Custom Modules

Mix embedded modules with your own Terraform modules.

## What it does
- Uses the embedded catalog by default.
- Supports a custom modules root (`--modules` or profile defaults) but custom git sources now make that optional.
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
- See `docs/features/custom-terraform.md` for bringing existing Terraform roots and per‑env workspace copies.
## Adding a custom module to the inventory

1. **Point at your module**  
   Run `pltf module init ./path/to/module` (or `pltf module init -m <root>` when the module sits under a shared root) to scan the Terraform code, detect inputs/outputs, and emit `module.yaml` inside the module directory. This metadata tells pltf how to wire and validate the module when it’s referenced.

2. **Review and edit `module.yaml`**  
   Confirm the generated `inputs`, `outputs`, `labels`, and `required_providers`. If the module targets a non-default provider (GCP, GitHub, etc.) add it there so pltf knows which provider block to create later.

3. **Register the module in your spec**  
   Add the module entry to your Environment or Service spec:

   ```yaml
   modules:
     - id: my-app
       type: my_custom_service
       source: custom
       inputs:
         config_bucket: module.base.bucket
   ```

   When `source: custom` is set, pltf looks under the `--modules` root (or the module directory) instead of the embedded catalog.

4. **Wire providers/variables**  
   If the module declares provider requirements, make sure the consuming spec enables and configures that provider (`providers:` block plus secrets/vars such as `github_token` or `project_id`).

5. **Commit the metadata**  
   Keep the generated `module.yaml` in version control so every run can resolve the module contract consistently.

## Using existing Terraform roots

1. Move the Terraform root into a module folder under your custom root.
2. Run `pltf module init --path <module>` to generate `module.yaml`.
3. Reference it as `type: <module_type>` in your spec and mark `source: custom`.

This lets you reuse existing Terraform code while pltf handles wiring/validation (variables, providers, outputs).

### Per-environment workspace copies

- A separate workspace exists per env/service (`.pltf/<env>/workspace` or `.pltf/<env>/<service>/workspace`).
- Module inputs that refer to file paths (relative to the spec) are copied into the workspace and rewritten, keeping per-env overrides isolated.

### Suggested layout

```
.
├─ env.yaml
├─ service.yaml
├─ modules/
│  ├─ my_network/
│  └─ my_cluster/
└─ files/
   ├─ dev/
   └─ prod/
```

Use file inputs like `./files/dev/config.json`; pltf copies them into the generated workspace for the selected env.

## Direct Git sources (future)

Each module entry can point directly at a git repo when `source` is an HTTP or SSH URL (e.g., `https://github.com/org/module.git//module-id?ref=main`). The module metadata (`module.yaml`) inside the repo defines the module `type`, so once the git fetching logic is implemented pltf will clone that repo and resolve the module without a shared `modules_root`. For now, continue to run `pltf module init` under the repo so the metadata exists inside it.

## Suggested workflow

1. Place custom modules under a shared root and run `pltf module init --path ...` for each.
2. (Optional) Point `pltf generate`/`terraform plan` at that root with `--modules` or profile defaults when you have a shared catalog; otherwise, each module can now specify its own git source URL.
3. Reference modules with `source: custom` and include any extra provider blocks required by the module (e.g., GitHub, database, or SaaS providers).
