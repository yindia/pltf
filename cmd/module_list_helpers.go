package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"

	"pltf/pkg/config"
)

func scanModules(root string) (map[string]*config.ModuleMetadata, error) {
	metas, err := config.ScanModuleMetas(root)
	if err != nil {
		return nil, err
	}
	return metas, nil
}

func printModules(metas map[string]*config.ModuleMetadata, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(metas)
	case "yaml", "yml":
		out, err := yaml.Marshal(metas)
		if err != nil {
			return err
		}
		fmt.Print(string(out))
		return nil
	case "table", "":
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "TYPE\tPROVIDER\tVERSION\tCLUSTER\tINPUTS\tOUTPUTS\tRESOURCES\tDATA\tCAPABILITIES\tDESCRIPTION")
		keys := make([]string, 0, len(metas))
		for k := range metas {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			m := metas[k]
			fmt.Fprintf(
				tw,
				"%s\t%s\t%s\t%t\t%d\t%d\t%d\t%d\t%s\t%s\n",
				m.Type,
				m.Provider,
				m.Version,
				m.Cluster,
				len(m.Inputs),
				len(m.Outputs),
				len(m.Resources),
				len(m.DataSources),
				formatCapabilities(m.Capabilities),
				m.Description,
			)
		}
		return tw.Flush()
	default:
		return fmt.Errorf("unsupported output format %q (use table|json|yaml)", format)
	}
}

func printModuleDetail(meta *config.ModuleMetadata, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(meta)
	case "yaml", "yml":
		out, err := yaml.Marshal(meta)
		if err != nil {
			return err
		}
		fmt.Print(string(out))
		return nil
	case "table", "":
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(tw, "Module: %s (%s)\n", meta.Name, meta.Type)
		fmt.Fprintf(tw, "Provider: %s\n", meta.Provider)
		fmt.Fprintf(tw, "Version: %s\n", meta.Version)
		fmt.Fprintf(tw, "Cluster: %t\n", meta.Cluster)
		fmt.Fprintf(tw, "Capabilities: %s\n", formatCapabilities(meta.Capabilities))
		fmt.Fprintf(tw, "Resources: %d\n", len(meta.Resources))
		fmt.Fprintf(tw, "Data sources: %d\n", len(meta.DataSources))
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "INPUT\tTYPE\tREQUIRED\tDEFAULT\tCAPABILITY\tDESCRIPTION")
		inputs := append([]config.InputSpec(nil), meta.Inputs...)
		sort.Slice(inputs, func(i, j int) bool { return inputs[i].Name < inputs[j].Name })
		for _, in := range inputs {
			fmt.Fprintf(
				tw,
				"%s\t%s\t%t\t%s\t%s\t%s\n",
				in.Name,
				in.Type,
				in.Required,
				formatDefaultValue(in.Default),
				in.Capability,
				in.Description,
			)
		}
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "OUTPUT\tTYPE\tCAPABILITY\tDESCRIPTION")
		outputs := append([]config.OutputSpec(nil), meta.Outputs...)
		sort.Slice(outputs, func(i, j int) bool { return outputs[i].Name < outputs[j].Name })
		for _, out := range outputs {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", out.Name, out.Type, out.Capability, out.Description)
		}
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "RESOURCE")
		if len(meta.Resources) == 0 {
			fmt.Fprintln(tw, "(none)")
		} else {
			resources := append([]string(nil), meta.Resources...)
			sort.Strings(resources)
			for _, res := range resources {
				fmt.Fprintf(tw, "%s\n", res)
			}
		}
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "DATA")
		if len(meta.DataSources) == 0 {
			fmt.Fprintln(tw, "(none)")
		} else {
			dataSources := append([]string(nil), meta.DataSources...)
			sort.Strings(dataSources)
			for _, data := range dataSources {
				fmt.Fprintf(tw, "%s\n", data)
			}
		}
		return tw.Flush()
	default:
		return fmt.Errorf("unsupported output format %q (use table|json|yaml)", format)
	}
}

func formatCapabilities(c config.Capabilities) string {
	if len(c.Provides) == 0 && len(c.Accepts) == 0 {
		return "-"
	}
	parts := []string{}
	if len(c.Provides) > 0 {
		parts = append(parts, fmt.Sprintf("provides=%s", strings.Join(c.Provides, ",")))
	}
	if len(c.Accepts) > 0 {
		parts = append(parts, fmt.Sprintf("accepts=%s", strings.Join(c.Accepts, ",")))
	}
	return strings.Join(parts, "; ")
}

func formatDefaultValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool, int, int64, float64, uint, uint64:
		return fmt.Sprint(val)
	default:
		out, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprint(val)
		}
		return string(out)
	}
}
