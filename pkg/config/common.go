package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// SecretRef describes where to resolve a secret value.
type SecretRef struct {
	// Source is retained for backward compatibility but defaults to env and is otherwise ignored.
	Source string `yaml:"source,omitempty"`
	// Key is the logical name of the secret; defaults to the map key if omitted.
	Key string `yaml:"key,omitempty"`
	// Path is optional and can be used by future resolvers (e.g., SSM/Secrets Manager).
	Path string `yaml:"path,omitempty"`
}

// Backend points to the remote state bucket/prefix.
type Backend struct {
	Type          string `yaml:"type,omitempty"` // s3, gcs, azurerm (defaults to provider)
	Bucket        string `yaml:"bucket"`
	Region        string `yaml:"region,omitempty"`         // backend region override (e.g. for s3)
	Container     string `yaml:"container,omitempty"`      // for azurerm
	ResourceGroup string `yaml:"resource_group,omitempty"` // for azurerm
	Profile       string `yaml:"profile,omitempty"`        // backend profile (e.g., s3 in another account)
}

// Module declares a module instance in env/service YAML.
type Module struct {
	ID     string                 `yaml:"id"`
	Type   string                 `yaml:"type"`
	Source string                 `yaml:"source,omitempty"` // "custom" to force custom root
	Inputs map[string]interface{} `yaml:"inputs,omitempty"`
	Links  AccessLinks            `yaml:"links,omitempty"`
}

// AccessLinks maps access level → list of target module IDs.
// e.g. "readwrite" -> ["bucket", "logs"]
type AccessLinks map[string][]string

// UnmarshalYAML supports both:
// readwrite: bucket
// and
// readwrite: [bucket, logs]
func (a *AccessLinks) UnmarshalYAML(value *yaml.Node) error {
	result := make(map[string][]string)

	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("links must be a mapping (e.g. readwrite: bucket or readwrite: [bucket, logs])")
	}

	appendValues := func(key string, values []string) {
		if len(values) == 0 {
			return
		}
		result[key] = append(result[key], values...)
	}

	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valNode := value.Content[i+1]

		key := keyNode.Value

		switch valNode.Kind {
		case yaml.ScalarNode:
			// readwrite: bucket
			appendValues(key, []string{valNode.Value})

		case yaml.SequenceNode:
			// readwrite: [bucket, logs]
			var items []string
			for _, n := range valNode.Content {
				items = append(items, n.Value)
			}
			appendValues(key, items)

		default:
			return fmt.Errorf("invalid links format under %q", key)
		}
	}

	*a = result
	return nil
}
