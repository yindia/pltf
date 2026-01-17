# Image & Plugin Caching

pltf keeps both Docker image builds and Terraform provider downloads cached so repeated runs stay fast.

## Docker / BuildKit caching

- Image builds still run via Dagger’s `Directory.DockerBuild`, and every build mounts the `pltf-image-cache` volume at `/dagger/image-cache`.
- When specs reuse the same context, BuildKit layers are reused and exported images continue to benefit from the cached layers.
- During `pltf terraform plan/apply`, we build every declared image (and `apply` pushes tags) before running Terraform so the BuildKit cache keeps working across commands.

## Terraform provider cache

- Terraform runs rely on whichever cache Terraform stores inside `.terraform` within the generated workspace. Keep the workspace around to benefit from cached provider downloads, and share it across CI or machines if you like.
- Inspect or archive the workspace to reuse provider binaries elsewhere; just ensure host credentials have access to the necessary artifacts.

## Tips

- If you want to inspect the plugin cache, look inside the workspace directory and locate its cache folder or copy it elsewhere.
- Ensure your host AWS/GCP credentials have permission to read the generated `.terraform` artifacts if you share the workspace across teams.
