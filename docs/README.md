# Pltf documentation

Pltf is a new kind of Infrastructure-as-Code framework built for fast-moving startups. It lets teams work with high-level concepts like microservices, environments, and databases, instead of low-level configuration such as VPC, IAM, ELB, or Kubernetes.

We've always been frustrated by the amount of manual effort required to manage infrastructure. We strongly believe in developer productivity, and empowering engineers has been our mission for the past few years.

With Pltf, we're reimagining how infrastructure should be managed in modern cloud environments. Pltf enables anyone to build automated, scalable, and secure infrastructure across AWS, GCP, and Azure. Our early users save countless hours every week and are able to scale their companies with minimal investment in DevOps.

Pltf gives you:

- SOC2 compliance from day one
- AWS, GCP, and Azure support
- Continuous deployment
- Hardened network and security configurations
- Support for multiple environments
- Built-in auto-scaling and high availability (HA)
- Support for spot instances
- Zero lock-in
- Out-of-the-box wiring between modules
- Out-of-the-box provider management
- Bring-your-own modules
- Out-of-the-box support for tfsec, tflint, infracost, and rover (https://github.com/yindia/rover)

## How it works

The idea is simple:

1. Platform teams define the core infrastructure using either their own modules or existing CLI modules.
2. Application teams deploy services on top of these base environments using higher-level abstractions.
3. Services become layered components within the Pltf ecosystem.

Our CLI reads these environments, services, and stacks to generate Terraform automatically. Once generated, teams can either commit the Terraform code or use our CLI to run Terraform commands directly.

In addition, Pltf integrates with infracost, tfsec, and tflint, and provides an AI-powered summary of the plan and risk assessment directly in pull requests.

## Documentation map

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
