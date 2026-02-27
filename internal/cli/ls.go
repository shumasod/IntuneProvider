package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var lsFilter string

var lsCmd = &cobra.Command{
	Use:   "ls <resource-type>",
	Short: "List Intune resources  (like ls in bash)",
	Long: `List all resources of the given type, similar to ls(1) in bash.

Output is a table by default. Use -o wide for more columns or -o json for
machine-readable output.

RESOURCE TYPES
  ` + strings.Join(AllNames(), "\n  ") + `

EXAMPLES
  intune ls config-policy
  intune ls compliance-policy -o wide
  intune ls ios-mam --filter "Outlook"
  intune ls android-mam -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustInitClient()

		res, err := Lookup(args[0])
		if err != nil {
			return err
		}

		var resp listResponse
		if err := globalClient.Get(context.Background(), res.ListPath, &resp); err != nil {
			return fmt.Errorf("failed to list %s: %w", res.Names[0], err)
		}

		items := resp.Value

		// Apply --filter: case-insensitive substring match on displayName
		if lsFilter != "" {
			filter := strings.ToLower(lsFilter)
			var filtered []map[string]interface{}
			for _, item := range items {
				name := strings.ToLower(fmt.Sprintf("%v", item["displayName"]))
				if strings.Contains(name, filter) {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}

		if len(items) == 0 {
			if lsFilter != "" {
				fmt.Printf("No %s resources matching %q\n", res.Names[0], lsFilter)
			} else {
				fmt.Printf("No %s resources found\n", res.Names[0])
			}
			return nil
		}

		switch flagOutput {
		case "json":
			PrintJSON(items)
		case "wide":
			PrintTable(res, items, true)
		default: // "table"
			PrintTable(res, items, false)
		}

		fmt.Printf("\n%s%d %s resource(s)%s\n", ansiCyan, len(items), res.Names[0], ansiReset)
		return nil
	},
}

func init() {
	lsCmd.Flags().StringVarP(&lsFilter, "filter", "f", "", "Filter by display name (case-insensitive substring)")
	rootCmd.AddCommand(lsCmd)

	// `intune list` is a long-form alias for `intune ls`
	listCmd := *lsCmd
	listCmd.Use = "list <resource-type>"
	listCmd.Short = "List Intune resources (alias for ls)"
	listCmd.Hidden = true
	rootCmd.AddCommand(&listCmd)
}
