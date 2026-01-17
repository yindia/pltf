package backend

import (
	"strings"
)

func defaultBackendBucket(backendType, org, envName string) string {
	var parts []string
	if strings.TrimSpace(org) != "" {
		parts = append(parts, strings.ToLower(strings.TrimSpace(org)))
	} else {
		parts = append(parts, "pltf")
	}
	if strings.TrimSpace(envName) != "" {
		parts = append(parts, strings.ToLower(strings.TrimSpace(envName)))
	}
	parts = append(parts, "tfstate")
	base := strings.Join(parts, "-")

	switch backendType {
	case "azurerm":
		return sanitizeStorageAccountName(base)
	default:
		return sanitizeBucketName(base)
	}
}

func sanitizeBucketName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "pltf-tfstate"
	}
	if len(out) < 3 {
		out = out + strings.Repeat("0", 3-len(out))
	}
	if len(out) > 63 {
		out = strings.Trim(out[:63], "-")
	}
	if len(out) < 3 {
		out = "pltf-tfstate"
	}
	return out
}

func sanitizeStorageAccountName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		out = "pltftfstate"
	}
	if len(out) < 3 {
		out = out + strings.Repeat("0", 3-len(out))
	}
	if len(out) > 24 {
		out = out[:24]
	}
	return out
}
