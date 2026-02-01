# Documentation Map

This folder is the source for the MkDocs site. Use this file as the quick map
for where to find (or add) content.

## Start here
- Overview: `docs/index.md`
- Installation: `docs/installation.md`
- Getting started (AWS): `docs/getting-started/aws.md`
- Terraform workflows: `docs/workflows.md`
- CLI overview: `docs/platform.md`

## Specs and concepts
- Concepts: `docs/concepts/overview.md`
- Environment: `docs/concepts/environment.md`
- Stack: `docs/concepts/stack.md`
- Modules and wiring: `docs/concepts/module.md`, `docs/modules.md`
- Specs guide: `docs/specs.md`

## Features
- Feature index: `docs/features.md`
- Validation: `docs/features/validation.md`
- Backends: `docs/features/backends.md`
- Custom modules: `docs/features/custom-modules.md`
- Profiles and defaults: `docs/features/profiles.md`
- Secrets and variables: `docs/features/secrets.md`, `docs/features/variables.md`

## Workflows
- Terraform workflows: `docs/workflows.md`
- Terraform command behavior: `docs/features/terraform-commands.md`
- Terraform generator: `docs/features/terraform-generator.md`
- Image & caching: `docs/caching.md`

## CLI reference
- CLI usage: `docs/usage.md`
- Platform patterns: `docs/platform.md`
- Main entry: `docs/cli/pltf.md`
- Validate: `docs/cli/pltf_validate.md`
- Generate: `docs/cli/pltf_generate.md`
- Terraform subcommands: `docs/cli/pltf_terraform.md`
- Module subcommands: `docs/cli/pltf_module.md`

## References and modules
- Provider references: `docs/references/`
- Built-in module references: `docs/references/modules/`

## Examples and security
- Examples: `docs/example/`
- Security: `docs/security/`
- FAQ: `docs/faq.md`

## Maintaining docs
- The navigation lives in `mkdocs.yml`.
- When you add a new page, also add it to the `nav:` list so it shows up in the site.
