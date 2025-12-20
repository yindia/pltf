# FAQ

**Is pltf production-ready?**  
It’s under active development. Pin a release, review generated Terraform, and run plans in non-prod first.

**Which clouds are supported?**  
Terraform output is portable; module coverage today is focused on AWS. GCP/Azure are on the roadmap and can be added via custom modules.

**How do Environment and Service specs relate?**  
`Service.metadata.ref` points to an Environment file. `envRef` selects the environment entry (e.g., `prod`) and lets you override variables/secrets for that service.

**Can I bring my own modules?**  
Yes. Add a `module.yaml` with inputs/outputs/schema and set `source: custom` (or point `modules_root` to your catalog). pltf will wire variables/links and emit Terraform.

**Do I have to run Terraform with pltf?**  
No. You can just generate Terraform and run `terraform plan/apply` yourself. The CLI can also run Terraform for you after regeneration to keep code and state aligned.

**Which state backends can I use?**  
`s3`, `gcs`, or `azurerm` regardless of target cloud. Configure credentials via profiles or env vars just like plain Terraform.

**How are secrets handled?**  
Put secrets under `envRef.<name>.secrets`. They are rendered as Terraform variables and should be sourced from your secret manager or CI env, not hardcoded.

**How do module links work?**  
Use `links` to reference other module outputs (e.g., an IAM role ARN for a Helm chart). The generator wires those into Terraform expressions; no manual interpolation needed.

**Where do I start?**  
Clone the repo and use the samples: `example/env.yaml` and `example/service.yaml`. Run `pltf preview` then `pltf terraform plan --env prod` to see the rendered code and plan.
