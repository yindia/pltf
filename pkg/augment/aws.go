package augment

import (
	"fmt"
	"strings"

	"pltf/pkg/config"
)

func init() {
	RegisterBuilder(buildAWS)
}

// buildAWS generates IAM augmentations (policies, trusts) for AWS modules.
func buildAWS(ctx Context) map[string]Augmentation {
	result := map[string]Augmentation{}

	targetModules := findModulesByTypes(ctx.Modules, []string{"aws_iam_role", "aws_iam_user"})
	if len(targetModules) == 0 {
		return result
	}

	eksID := findFirstModuleByType(ctx.Modules, "aws_eks")
	roleIndex := indexModulesByID(ctx.Modules)

	for roleID := range targetModules {
		bindings := collectBindings(ctx.Modules, roleID)
		if len(bindings) == 0 {
			continue
		}

		policy := buildIamPolicy(bindings)
		var trusts []map[string]interface{}

		mod := roleIndex[roleID]
		if mod.Type == "aws_iam_role" && eksID != "" {
			ns := stringVar(ctx.Vars, "irsa_namespace", "default")
			sa := stringVar(ctx.Vars, "irsa_service_account", "default")
			trusts = buildTrusts(eksID, ns, sa, mod)
		}

		result[roleID] = Augmentation{
			IamPolicy:        policy,
			KubernetesTrusts: trusts,
			SourceModule:     mod,
		}
	}

	return result
}

// --------------------------
// Helpers
// --------------------------

type iamBinding struct {
	accessLevel string
	moduleID    string
	moduleType  string
}

func collectBindings(mods []config.Module, roleID string) []iamBinding {
	var bindings []iamBinding
	for _, m := range mods {
		for access, targets := range m.Links {
			for _, t := range targets {
				if t == roleID {
					bindings = append(bindings, iamBinding{
						accessLevel: strings.ToLower(access),
						moduleID:    m.ID,
						moduleType:  m.Type,
					})
				}
			}
		}
	}
	return bindings
}

func buildIamPolicy(bindings []iamBinding) map[string]interface{} {
	var statements []map[string]interface{}
	for _, b := range bindings {
		actions, resources := awsActionsAndResources(b)
		if len(actions) == 0 || len(resources) == 0 {
			continue
		}
		statements = append(statements, map[string]interface{}{
			"Effect":   "Allow",
			"Action":   actions,
			"Resource": resources,
		})
	}
	if len(statements) == 0 {
		return nil
	}
	return map[string]interface{}{
		"Version":   "2012-10-17",
		"Statement": statements,
	}
}

func buildTrusts(eksID, ns, sa string, role config.Module) []map[string]interface{} {
	var trusts []map[string]interface{}

	if eksID != "" {
		trusts = append(trusts, map[string]interface{}{
			"open_id_url":  fmt.Sprintf("module.%s.k8s_openid_provider_url", eksID),
			"open_id_arn":  fmt.Sprintf("module.%s.k8s_openid_provider_arn", eksID),
			"service_name": sa,
			"namespace":    ns,
		})
	}

	allowed := readStringList(role.Inputs, "allowed_k8s_services")
	if len(allowed) == 0 && eksID != "" {
		trusts = []map[string]interface{}{
			{
				"open_id_url":  fmt.Sprintf("module.%s.k8s_openid_provider_url", eksID),
				"open_id_arn":  fmt.Sprintf("module.%s.k8s_openid_provider_arn", eksID),
				"service_name": "*",
				"namespace":    "*",
			},
		}
	} else {
		for _, svc := range allowed {
			svc = strings.TrimSpace(svc)
			if svc == "" {
				continue
			}
			svcNS, svcName := parseServiceRef(ns, sa, svc)
			trusts = append(trusts, map[string]interface{}{
				"open_id_url":  fmt.Sprintf("module.%s.k8s_openid_provider_url", eksID),
				"open_id_arn":  fmt.Sprintf("module.%s.k8s_openid_provider_arn", eksID),
				"service_name": svcName,
				"namespace":    svcNS,
			})
		}
	}

	return trusts
}

func awsActionsAndResources(b iamBinding) ([]interface{}, []interface{}) {
	switch b.moduleType {
	case "aws_s3":
		return s3Actions(b.accessLevel), []interface{}{
			fmt.Sprintf("${module.%s.bucket_arn}", b.moduleID),
			fmt.Sprintf("${module.%s.bucket_arn}/*", b.moduleID),
		}
	case "aws_sqs":
		return sqsActions(b.accessLevel), []interface{}{
			fmt.Sprintf("${module.%s.queue_arn}", b.moduleID),
		}
	case "aws_sns":
		return snsActions(b.accessLevel), []interface{}{
			fmt.Sprintf("${module.%s.topic_arn}", b.moduleID),
		}
	case "aws_dynamodb":
		return dynamoActions(b.accessLevel), []interface{}{
			fmt.Sprintf("${module.%s.table_arn}", b.moduleID),
			fmt.Sprintf("${module.%s.table_arn}/index/*", b.moduleID),
		}
	case "aws_ses":
		return sesActions(b.accessLevel), []interface{}{
			fmt.Sprintf("${module.%s.identity_arn}", b.moduleID),
		}
	default:
		return nil, nil
	}
}

func s3Actions(access string) []interface{} {
	read := []interface{}{"s3:GetObject", "s3:ListBucket"}
	write := []interface{}{"s3:PutObject", "s3:DeleteObject"}
	switch access {
	case "read":
		return read
	case "write":
		return append(read, write...)
	case "readwrite", "rw", "admin":
		return append(read, write...)
	default:
		return nil
	}
}

func sqsActions(access string) []interface{} {
	read := []interface{}{"sqs:ReceiveMessage", "sqs:GetQueueAttributes", "sqs:ListQueueTags", "sqs:ChangeMessageVisibility"}
	write := []interface{}{"sqs:SendMessage", "sqs:DeleteMessage"}
	switch access {
	case "read":
		return read
	case "write":
		return append(read, write...)
	case "readwrite", "rw", "admin":
		return append(read, write...)
	default:
		return nil
	}
}

func snsActions(access string) []interface{} {
	switch access {
	case "read", "write", "readwrite", "rw", "publish", "admin":
		return []interface{}{"sns:Publish"}
	default:
		return nil
	}
}

func dynamoActions(access string) []interface{} {
	read := []interface{}{
		"dynamodb:BatchGetItem",
		"dynamodb:DescribeTable",
		"dynamodb:GetItem",
		"dynamodb:Query",
		"dynamodb:Scan",
	}
	write := []interface{}{
		"dynamodb:BatchWriteItem",
		"dynamodb:DeleteItem",
		"dynamodb:PutItem",
		"dynamodb:UpdateItem",
	}
	switch access {
	case "read":
		return read
	case "write":
		return append(read, write...)
	case "readwrite", "rw", "admin":
		return append(read, write...)
	default:
		return nil
	}
}

func sesActions(access string) []interface{} {
	switch access {
	case "write", "readwrite", "rw", "send", "admin":
		return []interface{}{"ses:SendEmail", "ses:SendRawEmail"}
	default:
		return nil
	}
}

func findModulesByTypes(mods []config.Module, moduleTypes []string) map[string]struct{} {
	out := map[string]struct{}{}
	set := map[string]struct{}{}
	for _, t := range moduleTypes {
		set[t] = struct{}{}
	}
	for _, m := range mods {
		if _, ok := set[m.Type]; ok {
			out[m.ID] = struct{}{}
		}
	}
	return out
}

func findFirstModuleByType(mods []config.Module, moduleType string) string {
	for _, m := range mods {
		if m.Type == moduleType {
			return m.ID
		}
	}
	return ""
}

func indexModulesByID(mods []config.Module) map[string]config.Module {
	out := map[string]config.Module{}
	for _, m := range mods {
		out[m.ID] = m
	}
	return out
}

func readStringList(inputs map[string]interface{}, key string) []string {
	if inputs == nil {
		return nil
	}
	raw, ok := inputs[key]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []interface{}:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}

func parseServiceRef(defaultNS, defaultSA, ref string) (string, string) {
	parts := strings.Split(ref, "/")
	if len(parts) == 2 {
		if strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" {
			return parts[0], parts[1]
		}
	}
	return defaultNS, ref
}

func stringVar(vars map[string]interface{}, name, fallback string) string {
	if v, ok := vars[name]; ok {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return fallback
}
