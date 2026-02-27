// Package cli defines the resource registry that maps short resource-type names
// (used on the command line) to Microsoft Graph API paths and display metadata.
package cli

import (
	"fmt"
	"strings"
)

// Column defines one column in the table output of `intune ls`.
type Column struct {
	Header string
	Key    string // dot-separated path into the JSON object (e.g., "displayName")
	Wide   bool   // only displayed in --output=wide mode
}

// ResourceDef describes a single Intune resource type.
type ResourceDef struct {
	// Names is the primary name and all accepted aliases (first element = canonical).
	Names []string
	// ListPath is the Graph API path to list all resources of this type.
	ListPath string
	// SinglePath is the Graph API path for a single resource; %s is the ID.
	SinglePath string
	// Description is shown in help text.
	Description string
	// Columns defines the table columns used by `intune ls`.
	Columns []Column
}

// registry is the single source of truth for supported resource types.
var registry = []ResourceDef{
	{
		Names:      []string{"config-policy", "config", "cp"},
		ListPath:   "/deviceManagement/deviceConfigurations",
		SinglePath: "/deviceManagement/deviceConfigurations/%s",
		Description: "Device configuration profiles (Windows, iOS, Android)",
		Columns: []Column{
			{Header: "NAME", Key: "displayName"},
			{Header: "TYPE", Key: "@odata.type"},
			{Header: "ID", Key: "id", Wide: true},
			{Header: "CREATED", Key: "createdDateTime", Wide: true},
			{Header: "MODIFIED", Key: "lastModifiedDateTime"},
		},
	},
	{
		Names:      []string{"compliance-policy", "compliance", "cop"},
		ListPath:   "/deviceManagement/deviceCompliancePolicies",
		SinglePath: "/deviceManagement/deviceCompliancePolicies/%s",
		Description: "Device compliance policies with non-compliance actions",
		Columns: []Column{
			{Header: "NAME", Key: "displayName"},
			{Header: "TYPE", Key: "@odata.type"},
			{Header: "ID", Key: "id", Wide: true},
			{Header: "CREATED", Key: "createdDateTime", Wide: true},
			{Header: "MODIFIED", Key: "lastModifiedDateTime"},
		},
	},
	{
		Names:      []string{"ios-mam", "ios-app-protection", "ios"},
		ListPath:   "/deviceAppManagement/iosManagedAppProtections",
		SinglePath: "/deviceAppManagement/iosManagedAppProtections/%s",
		Description: "iOS App Protection Policies (MAM without enrollment)",
		Columns: []Column{
			{Header: "NAME", Key: "displayName"},
			{Header: "PIN", Key: "pinRequired"},
			{Header: "DATA_BACKUP_BLOCKED", Key: "dataBackupBlocked"},
			{Header: "ID", Key: "id", Wide: true},
			{Header: "MODIFIED", Key: "lastModifiedDateTime"},
		},
	},
	{
		Names:      []string{"android-mam", "android-app-protection", "android"},
		ListPath:   "/deviceAppManagement/androidManagedAppProtections",
		SinglePath: "/deviceAppManagement/androidManagedAppProtections/%s",
		Description: "Android App Protection Policies (MAM without enrollment)",
		Columns: []Column{
			{Header: "NAME", Key: "displayName"},
			{Header: "PIN", Key: "pinRequired"},
			{Header: "ENCRYPTED", Key: "encryptAppData"},
			{Header: "ID", Key: "id", Wide: true},
			{Header: "MODIFIED", Key: "lastModifiedDateTime"},
		},
	},
}

// Lookup finds a ResourceDef by name or alias. Returns an error if not found.
func Lookup(name string) (*ResourceDef, error) {
	name = strings.ToLower(name)
	for i := range registry {
		for _, n := range registry[i].Names {
			if n == name {
				return &registry[i], nil
			}
		}
	}
	return nil, fmt.Errorf("unknown resource type %q\n\nAvailable types:\n%s", name, resourceList())
}

// AllNames returns a sorted slice of all canonical resource names for help text.
func AllNames() []string {
	var out []string
	for _, r := range registry {
		out = append(out, r.Names[0])
	}
	return out
}

func resourceList() string {
	var sb strings.Builder
	for _, r := range registry {
		aliases := ""
		if len(r.Names) > 1 {
			aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(r.Names[1:], ", "))
		}
		fmt.Fprintf(&sb, "  %-30s %s%s\n", r.Names[0], r.Description, aliases)
	}
	return sb.String()
}

// fieldValue extracts a value from a decoded JSON object using a dot-separated key path.
func fieldValue(obj map[string]interface{}, key string) string {
	parts := strings.SplitN(key, ".", 2)
	v, ok := obj[parts[0]]
	if !ok {
		return ""
	}
	if len(parts) == 1 {
		// Trim the OData type prefix (#microsoft.graph.) for readability
		s := fmt.Sprintf("%v", v)
		if strings.HasPrefix(s, "#microsoft.graph.") {
			s = s[len("#microsoft.graph."):]
		}
		return s
	}
	nested, ok := v.(map[string]interface{})
	if !ok {
		return ""
	}
	return fieldValue(nested, parts[1])
}
