package generate

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/zclconf/go-cty/cty"

	"pltf/pkg/config"
)

// =====================
// helpers
// =====================
func sortedKeysInterfaceMap(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mergeLabels(envLabels, svcLabels map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range envLabels {
		result[k] = v
	}
	for k, v := range svcLabels {
		result[k] = v
	}
	return result
}

func mustCtyMap(m map[string]string) cty.Value {
	out := make(map[string]cty.Value, len(m))
	for k, v := range m {
		out[k] = cty.StringVal(v)
	}
	return cty.MapVal(out)
}

func toCtyValue(v interface{}) (cty.Value, error) {
	switch val := v.(type) {
	case string:
		return cty.StringVal(val), nil
	case bool:
		return cty.BoolVal(val), nil
	case int:
		return cty.NumberIntVal(int64(val)), nil
	case int32:
		return cty.NumberIntVal(int64(val)), nil
	case int64:
		return cty.NumberIntVal(val), nil
	case float32:
		return cty.NumberFloatVal(float64(val)), nil
	case float64:
		return cty.NumberFloatVal(val), nil
	case []interface{}:
		if len(val) == 0 {
			return cty.ListValEmpty(cty.DynamicPseudoType), nil
		}
		ctyVals := make([]cty.Value, 0, len(val))
		for _, item := range val {
			cv, err := toCtyValue(item)
			if err != nil {
				return cty.NilVal, err
			}
			ctyVals = append(ctyVals, cv)
		}
		return cty.ListVal(ctyVals), nil
	case map[string]string:
		obj := make(map[string]cty.Value, len(val))
		for k, v2 := range val {
			obj[k] = cty.StringVal(v2)
		}
		return cty.ObjectVal(obj), nil
	case map[string]interface{}:
		obj := make(map[string]cty.Value, len(val))
		for k, v2 := range val {
			cv, err := toCtyValue(v2)
			if err != nil {
				return cty.NilVal, err
			}
			obj[k] = cv
		}
		return cty.ObjectVal(obj), nil
	default:
		return cty.StringVal(fmt.Sprintf("%v", val)), nil
	}
}

func inputDeclared(meta *config.ModuleMetadata, name string) bool {
	for _, in := range meta.Inputs {
		if in.Name == name {
			return true
		}
	}
	return false
}

// copyUsedModules copies only the module directories that are actually used in this stack
// from modulesRoot/<type> to outDir/modules/<type>.
func copyUsedModules(outDir string, used map[string]bool, rootByType map[string]string) error {
	types := make([]string, 0, len(used))
	for moduleType := range used {
		types = append(types, moduleType)
	}
	sort.Strings(types)

	for _, moduleType := range types {
		srcRoot, ok := rootByType[moduleType]
		if !ok {
			return fmt.Errorf("module type %s root not found", moduleType)
		}
		srcDir := filepath.Join(srcRoot, moduleType)
		dstDir := filepath.Join(outDir, "modules", moduleType)

		if err := copyDir(srcDir, dstDir); err != nil {
			return fmt.Errorf("copy module %s: %w", moduleType, err)
		}
	}
	return nil
}

// copyDir recursively copies a directory tree from src to dst.
// It skips nothing right now; if you want to skip module.yaml, add a check.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip VCS/terraform internals
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".terraform":
				return fs.SkipDir
			}
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		// Skip tfstate artifacts
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".tfstate") || strings.HasSuffix(base, ".tfstate.backup") {
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// parseVarValue tries to convert a CLI var "value" string into a sensible Go type:
// - bool
// - int
// - float
// - JSON list/map: ["a","b"] or {"k":"v"}
// - simple list: "a,b,c" -> []interface{}{"a","b","c"}
// - string (fallback)
func parseVarValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch t := v.(type) {
	case bool, int, int64, float64, float32:
		return t

	case map[string]interface{}, []interface{}:
		return t

	case string:
		trimmed := strings.TrimSpace(t)

		if b, err := strconv.ParseBool(trimmed); err == nil {
			return b
		}

		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return i
		}

		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return f
		}

		if len(trimmed) > 0 && (trimmed[0] == '[' || trimmed[0] == '{') {
			var decoded interface{}
			if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
				return decoded
			}
		}

		if strings.Contains(trimmed, ",") {
			parts := strings.Split(trimmed, ",")
			out := make([]interface{}, 0, len(parts))
			for _, p := range parts {
				out = append(out, strings.TrimSpace(p))
			}
			return out
		}

		return t

	default:
		return t
	}
}
