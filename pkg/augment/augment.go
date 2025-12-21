package augment

import "pltf/pkg/config"

// Context captures the data needed by augmentation plugins to generate extra
// inputs (IAM policies, trusts, etc.) for modules in a stack.
type Context struct {
	Provider    string
	EnvName     string
	ServiceName string
	Modules     []config.Module
	Vars        map[string]interface{}
}

// Augmentation describes extra inputs to apply to a module.
type Augmentation struct {
	IamPolicy        map[string]interface{}
	KubernetesTrusts []map[string]interface{}
	SourceModule     config.Module
}

// Builder generates augmentations for a stack. Builders are registered by
// module plugins.
type Builder func(Context) map[string]Augmentation

// Applicator attaches an augmentation to a module (e.g. wiring iam_policy).
type Applicator func(config.Module, Augmentation) config.Module

var (
	builders    []Builder
	applicators = map[string]Applicator{}
)

// RegisterBuilder adds an augmentation builder. Typically called from a
// module plugin init() function.
func RegisterBuilder(b Builder) {
	builders = append(builders, b)
}

// RegisterApplicator associates a module type with an applicator that knows
// how to merge augmentation data into that module's inputs.
func RegisterApplicator(moduleType string, a Applicator) {
	applicators[moduleType] = a
}

// Build runs all registered builders for the given context and merges their
// results keyed by module ID.
func Build(ctx Context) map[string]Augmentation {
	result := map[string]Augmentation{}
	for _, b := range builders {
		if b == nil {
			continue
		}
		for modID, aug := range b(ctx) {
			result[modID] = mergeAugmentation(result[modID], aug)
		}
	}
	return result
}

// Apply merges the augmentation into the given module if an applicator is
// registered for its type.
func Apply(m config.Module, aug Augmentation) config.Module {
	if app, ok := applicators[m.Type]; ok && app != nil {
		return app(m, aug)
	}
	return m
}

func mergeAugmentation(existing, next Augmentation) Augmentation {
	merged := existing

	if merged.IamPolicy == nil {
		merged.IamPolicy = next.IamPolicy
	} else if next.IamPolicy != nil {
		merged.IamPolicy = mergePolicies(merged.IamPolicy, next.IamPolicy)
	}

	if len(next.KubernetesTrusts) > 0 {
		merged.KubernetesTrusts = append(merged.KubernetesTrusts, next.KubernetesTrusts...)
	}

	if merged.SourceModule.ID == "" {
		merged.SourceModule = next.SourceModule
	}

	return merged
}

func mergePolicies(base, extra map[string]interface{}) map[string]interface{} {
	if base == nil {
		return extra
	}
	if extra == nil {
		return base
	}

	baseStmt, _ := base["Statement"].([]interface{})
	extraStmt, _ := extra["Statement"].([]interface{})

	if len(extraStmt) > 0 {
		base["Statement"] = append(baseStmt, extraStmt...)
	}
	return base
}
