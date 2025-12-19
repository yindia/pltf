# aws_nodegroup

Managed node group for EKS with scaling limits, instance type, disk size, and optional spot instances.

## What it does

- Adds an EKS managed node group with scaling limits and instance type/disk controls.
- Supports spot instances and custom labels/taints (via launch template inputs if set).

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
ami_type |  | AL2023_x86_64_STANDARD | False
autoscaling_tags |  | {} | False
labels |  | {} | False
max_nodes |  | 15 | False
min_nodes |  | 3 | False
node_disk_size |  | 20 | False
node_instance_type |  | t3.medium | False
spot_instances |  | False | False
taints |  | [] | False
use_gpu |  | False | False

## Outputs

Name | Description
--- | ---

