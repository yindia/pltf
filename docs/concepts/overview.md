# Concepts Overview

pltf is Infrastructure-as-Code with a higher-level abstraction. You write configuration files, then run the pltf CLI (locally or in CI/CD) to connect to your cloud account and provision resources using Terraform under the hood.

![Architecture](../images/hero.png) <!-- Replace with your hero image -->

## How It Works
1) Author YAML specs.
2) Run `pltf preview|validate|generate|terraform ...`.
3) The CLI renders Terraform (providers, backends, locals, remote state) and can execute Terraform for you.

There are two primary spec types:

- **Environment**: Specifies cloud, account, and region. Running an Environment sets up the base resources (e.g., Kubernetes cluster, networks, IAM, ingress). Typical patterns: one per staging/prod/QA, or one per engineer/PR for isolated sandboxes.
- **Service** (Layer): Specifies the workload (often a microservice) and any non-Kubernetes resources it needs (e.g., databases, queues). pltf connects these seamlessly to the Environment.

Environment and Service specs link via `metadata.ref` (path to env) and `envRef` (per-environment overrides).
