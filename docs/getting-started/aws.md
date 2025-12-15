# Getting Started: AWS

This guide walks through provisioning a simple environment and service on AWS using pltf. You will create two YAML specs (environment and service), generate Terraform, and deploy.

![AWS](../images/hero.png) <!-- Replace with your AWS image -->

## 1) Prerequisites
- Terraform v1.5+ (installed locally or via your CI)
- Docker (to build/push your images if needed)
- AWS credentials configured in your shell (e.g., `aws configure` or environment variables)
- (Optional) Custom modules directory if you want to bring your own modules

Install pltf (Homebrew):
```bash
brew tap evalsocket1/pltf
brew install pltf
```
Or use the install script:
```bash
/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/your-org/pltf/main/install.sh)\"
```

## 2) Create an Environment spec
Create `env.yaml`:
```yaml
apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example-aws
  org: demo
  provider: aws
  labels:
    team: platform
    cost_center: shared
environments:
  prod:
    account: \"123456789012\"
    region: us-east-1
    backend:
      type: s3            # state backend (s3|gcs|azurerm)
      profile: default    # optional cross-account profile
    variables:
      base_domain: prod.demo.internal
      cluster_name: demo-eks
    modules:
      - type: aws_base
      - type: aws_eks
      - type: aws_k8s_base
```
What this does:
- Configures AWS provider/region and S3 backend.
- Creates networking, an EKS cluster, and base Kubernetes add-ons.
- Exposes outputs (e.g., cluster endpoint/CA) for services.

Generate and apply:
```bash
pltf terraform plan -f env.yaml -e prod
pltf terraform apply -f env.yaml -e prod
```
First apply can take ~15 minutes. You can inspect outputs:
```bash
pltf terraform output -f env.yaml -e prod --json
```

## 3) Create a Service spec
Create `service.yaml`:
```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  org: demo
  provider: aws
  envRef:
    name: prod
    path: ./env.yaml
spec:
  variables:
    image: ghcr.io/demo/payments:latest
  modules:
    - name: app
      type: aws_k8s_service
      port:
        http: 8080
      public_uri: \"/payments\"
      links:
        - app-bucket: [write]
    - name: app-bucket
      type: aws_s3
      bucket_name: \"payments-${env_name}\"
```
What this does:
- Deploys a Kubernetes service on the EKS cluster created above.
- Provisions an S3 bucket and links it to the app with write permissions (IRSA policy is generated).

Generate and apply:
```bash
pltf terraform plan -f service.yaml -e prod
pltf terraform apply -f service.yaml -e prod
```

## 4) Access the service
- Find the load balancer host from outputs: `pltf terraform output -f service.yaml -e prod | grep load_balancer_raw_dns`
- Curl the path: `curl http://<lb>/payments`

## 5) Cleanup
```bash
pltf terraform destroy -f service.yaml -e prod
pltf terraform destroy -f env.yaml -e prod
```

## 6) Next Steps
- Review AWS architecture and module docs in the References section.
- Add more modules (RDS, Redis, SES, SNS, SQS) and link them for IAM/IRSA wiring.
- Use profiles (`~/.pltf/profile.yaml`) to set default env/modules root and cross-account backends.
