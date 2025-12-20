# Concepts Overview

pltf raises the abstraction for Infrastructure-as-Code. You describe high-level constructs in YAML and let the CLI render the Terraform (providers, backends, locals, remote state) needed to stand them up. You keep the output and can run `terraform plan/apply` yourself or let pltf drive Terraform for you.

![Architecture](../images/hero.png)

## Core specs
- **Environment**: The shared foundation—cloud/account/region plus base modules like VPC, DNS, EKS/GKE/AKS, IAM. Often one per prod/stage/QA, or per engineer/PR sandbox.
- **Service** (Layer): A workload’s resources—databases, queues, buckets, roles, charts—wired into an Environment.

Services point at Environments via `metadata.ref` (path to env file) and `envRef` (per-environment overrides). Module links let services consume environment outputs without hand-written interpolation.

## Loop
1. Author/update YAML specs (Environment + Service).
2. `pltf validate | preview | terraform plan/apply` (regenerates Terraform each run).
3. Inspect outputs/graphs; iterate.
