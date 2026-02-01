# Image & Caching

pltf keeps Docker builds fast via Dagger cache layers and lets Terraform reuse its own provider cache inside the generated workspace.

## Docker / BuildKit

- Image builds run through Dagger’s `Directory.DockerBuild` with a shared `pltf-image-cache` volume at `/dagger/image-cache`.
- `pltf terraform plan` builds every declared image locally; `pltf terraform apply` builds the images, pushes the tags using host registry credentials, and then runs Terraform; `destroy` skips image builds entirely.
- Reusing the same context and Dockerfile lets BuildKit layers persist across commands and sessions, so both plan and apply benefit from cached layers.

## Terraform cache

- Terraform executes on the host inside `.pltf/<environment_name>/workspace` (env) or `.pltf/<environment_name>/<service_name>/workspace` (service), and its `.terraform` directory caches providers/plugins there.
- Keep the workspace between runs (or share it across CI agents) to benefit from cached downloads; no extra `.terraform-plugin-cache` file or wrapper is required.
- The workspace also stores plan artifacts (`.pltf-plan.tfplan`, `.pltf-plan.json`) for downstream workflows or `terraform show` references.

## Tips

- Inspect or archive the workspace directory if you need to copy provider binaries or debug plugin downloads.
- Ensure your host credentials (AWS, GCP, Azure, Docker) can read the generated workspace when sharing caches across teams.
