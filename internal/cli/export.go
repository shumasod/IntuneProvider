package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var exportDir string
var exportAll bool

var exportCmd = &cobra.Command{
	Use:   "export <resource-type> [<id>...]",
	Short: "Export resources to JSON files  (like dd / backup)",
	Long: `Export Intune resources to JSON files in the target directory.

Similar to taking a backup with dd(1), this command serialises one or all
resources of a given type to individual JSON files that can later be restored
with 'intune apply'.

File names are: <displayName>_<id>.json (spaces replaced with underscores).

EXAMPLES
  # Export all device configuration policies to ./backup/
  intune export config-policy --all -d ./backup/

  # Export two specific compliance policies
  intune export compliance-policy abc-123 def-456 -d ./backup/

  # Restore from backup
  for f in ./backup/config-policy_*.json; do
    intune apply config-policy -f "$f"
  done`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustInitClient()

		res, err := Lookup(args[0])
		if err != nil {
			return err
		}
		ids := args[1:]

		// Ensure output directory exists
		if exportDir == "" {
			exportDir = "."
		}
		if err := os.MkdirAll(exportDir, 0o755); err != nil {
			return fmt.Errorf("cannot create directory %q: %w", exportDir, err)
		}

		var items []map[string]interface{}

		if exportAll || len(ids) == 0 {
			var resp listResponse
			if err := globalClient.Get(context.Background(), res.ListPath, &resp); err != nil {
				return fmt.Errorf("failed to list %s: %w", res.Names[0], err)
			}
			items = resp.Value
		} else {
			for _, id := range ids {
				var obj map[string]interface{}
				path := fmt.Sprintf(res.SinglePath, id)
				if err := globalClient.Get(context.Background(), path, &obj); err != nil {
					PrintWarning("skipping %s: %v", id, err)
					continue
				}
				items = append(items, obj)
			}
		}

		if len(items) == 0 {
			fmt.Println("Nothing to export.")
			return nil
		}

		exported := 0
		for _, item := range items {
			id := fmt.Sprintf("%v", item["id"])
			name := sanitizeFilename(fmt.Sprintf("%v", item["displayName"]))
			filename := filepath.Join(exportDir, fmt.Sprintf("%s_%s_%s.json", res.Names[0], name, id[:8]))

			data, err := json.MarshalIndent(item, "", "  ")
			if err != nil {
				PrintWarning("could not marshal %s: %v", id, err)
				continue
			}
			if err := os.WriteFile(filename, data, 0o644); err != nil {
				PrintWarning("could not write %s: %v", filename, err)
				continue
			}
			PrintSuccess("%s → %s", id, Cyan(filename))
			exported++
		}

		fmt.Printf("\nExported %d %s resource(s) to %s\n", exported, res.Names[0], exportDir)
		return nil
	},
}

// sanitizeFilename replaces non-alphanumeric characters with underscores.
func sanitizeFilename(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}

func init() {
	exportCmd.Flags().StringVarP(&exportDir, "dir", "d", ".", "Directory to write JSON files into")
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all resources of the given type")
	rootCmd.AddCommand(exportCmd)
}
