# Image & Plugin Caching

pltf keeps both Docker image builds and Terraform provider downloads cached so repeated runs stay fast.

## Docker / BuildKit caching

- Image builds run via Dagger’s `Directory.DockerBuild`, and every build mounts the `pltf-image-cache` volume at `/dagger/image-cache`.
- When specs reuse the same context, BuildKit layers are reused, and exported images push/preserve tags with the cache volume intact.
- During `pltf terraform plan/apply`, we build every declared image (and `apply` pushes tags) before running Terraform; the cache volume lives for the whole command.

## Terraform provider/plugin cache

- Terraform runs share a CacheVolume named `pltf-terraform-plugin-cache`. It’s mounted at `/work/.terraform-plugin-cache` inside every command container.
- We generate `~/.terraformrc` pointing `plugin_cache_dir` to the cache so Terraform checks that directory first and writes new providers there.
- `TF_PLUGIN_CACHE_DIR` also points at the same path so the CLI and Terraform agree.
- Cache persistence spans commands and even project directories because the CacheVolume is stored by Dagger outside any single workspace.

## Tips

- If you want to inspect the cache, export the Dagger volume via the CLI (`dagger cache export ...`) or mount a host directory by modifying the runner.
- Ensure your host AWS/GCP credentials have permission to read any `.terraformrc` artifacts and plugin binaries if you share the cache across teams.
