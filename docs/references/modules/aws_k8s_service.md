# aws_k8s_service

Deploys a Kubernetes workload via Helm with service/HPA/ingress and IRSA wiring; links drive IAM policies.

## What it does

- Creates a Helm release with Deployment/Service/HPA and optional Ingress.
- Configures service account + IRSA using provided OIDC info and IAM policy.
- Supports env vars, secrets, cronjobs, persistent volumes, and pod annotations.
- Supports sticky sessions, custom probes, ingress annotations, and extra IAM policies.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
additional_iam_policies |  | [] | False
args |  |  | True
autoscaling_target_cpu_percentage | Percentage of requested cpu after which autoscaling kicks in | 80 | False
autoscaling_target_mem_percentage | Percentage of requested memory after which autoscaling kicks in | 80 | False
commands |  |  | True
consistent_hash |  |  | False
cron_jobs |  | [] | False
digest | Digest of image to be deployed |  | False
domain |  |  | False
env_vars | Environment variables to pass to the container | [] | False
healthcheck_command |  |  | True
healthcheck_path |  |  | False
http_port | The port that exposes an HTTP interface |  | False
iam_policy |  |  | True
image | External Image to be deployed |  | True
ingress_extra_annotations |  | {} | False
initial_liveness_delay |  | 30 | False
initial_readiness_delay |  | 30 | False
keep_path_prefix |  | False | False
link_secrets |  | [] | False
links |  |  | False
liveness_probe_command |  |  | True
liveness_probe_path | Url path for liveness probe |  | False
max_containers | Max value for HPA autoscaling | 3 | False
max_history |  |  | True
min_containers | Min value for HPA autoscaling | 1 | False
openid_provider_arn |  |  | True
openid_provider_url |  |  | True
persistent_storage |  | [] | False
pod_annotations | values to add to the pod annotations for the k8s-service pods | {} | False
pod_labels |  |  | True
ports | Ports to be exposed |  | True
probe_port | The port that is used for health probes |  | False
public_uri |  | [] | False
readiness_probe_command |  |  | True
readiness_probe_path | Url path for readiness probe |  | False
resource_limits |  |  | True
resource_request |  | {'cpu': 100, 'memory': 128} | False
secrets |  |  | False
service_annotations | Annotations to add to the service resource | {} | False
sticky_session |  | False | False
sticky_session_max_age |  | 86400 | False
tag | Tag of image to be deployed |  | False
timeout |  | 300 | False
tolerations |  | [] | False

## Outputs

Name | Description
--- | ---
current_digest | Current image digest deployed.
current_image | 
current_tag | 
docker_repo_url | 

