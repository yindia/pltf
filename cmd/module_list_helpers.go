package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
		fmt.Fprintln(tw, "TYPE\tPROVIDER\tVERSION\tDESCRIPTION")
		keys := make([]string, 0, len(metas))
		for k := range metas {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			m := metas[k]
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", m.Type, m.Provider, m.Version, m.Description)
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
		fmt.Fprintf(tw, "Module: %s (%s) provider=%s version=%s\n", meta.Name, meta.Type, meta.Provider, meta.Version)
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "INPUT\tTYPE\tREQUIRED\tDEFAULT\tDESCRIPTION")
		for _, in := range meta.Inputs {
			def := fmt.Sprintf("%v", in.Default)
			if in.Default == nil {
				def = ""
			}
			fmt.Fprintf(tw, "%s\t%s\t%t\t%s\t%s\n", in.Name, in.Type, in.Required, def, in.Description)
		}
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "OUTPUT\tTYPE\tDESCRIPTION")
		for _, out := range meta.Outputs {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", out.Name, out.Type, out.Description)
		}
		return tw.Flush()
	default:
		return fmt.Errorf("unsupported output format %q (use table|json|yaml)", format)
	}
}
