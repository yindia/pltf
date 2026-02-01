package augment

import "pltf/pkg/config"

func init() {
	RegisterProviderBuilder("gcp", buildGCP)
	RegisterProviderBuilder("google", buildGCP)
}

// buildGCP generates GCS access inputs for GCP service account modules.
func buildGCP(ctx Context) map[string]Augmentation {
	result := map[string]Augmentation{}
	roleIndex := indexModulesByID(ctx.Modules)

	for _, mod := range ctx.Modules {
		if !isGcpServiceAccountModule(mod) {
			continue
		}

		bindings := collectBindings(ctx.Modules, mod.ID)
		if len(bindings) == 0 {
			continue
		}

		var readBuckets []string
		var writeBuckets []string
		for _, b := range bindings {
			if b.moduleType != "gcp_gcs" {
				continue
			}
			bucketRef := refForModuleOutputInterpolated(ctx, b.moduleID, "bucket_name")
			switch b.accessLevel {
			case "read":
				readBuckets = append(readBuckets, bucketRef)
			case "write":
				writeBuckets = append(writeBuckets, bucketRef)
			case "readwrite", "rw", "admin":
				writeBuckets = append(writeBuckets, bucketRef)
			}
		}

		if len(readBuckets) == 0 && len(writeBuckets) == 0 {
			continue
		}

		result[mod.ID] = Augmentation{
			ReadBuckets:  readBuckets,
			WriteBuckets: writeBuckets,
			SourceModule: roleIndex[mod.ID],
		}
	}

	return result
}

func isGcpServiceAccountModule(m config.Module) bool {
	switch m.Type {
	case "gcp_service_account", "gcp_k8s_service":
		return true
	default:
		return false
	}
}
