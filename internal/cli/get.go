package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <resource-type> <id>",
	Short: "Get a single Intune resource by ID  (like cat in bash)",
	Long: `Fetch a single resource by its Graph API ID, similar to cat(1).

The default output format is JSON so the result can be piped or redirected:

  intune get config-policy <id> -o json > backup.json
  intune get config-policy <id> | python3 -m json.tool

EXAMPLES
  intune get config-policy abc-123
  intune get compliance-policy abc-123 -o json
  intune get ios-mam abc-123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustInitClient()

		res, err := Lookup(args[0])
		if err != nil {
			return err
		}
		id := args[1]

		var obj map[string]interface{}
		path := fmt.Sprintf(res.SinglePath, id)
		if err := globalClient.Get(context.Background(), path, &obj); err != nil {
			return fmt.Errorf("failed to get %s %q: %w", res.Names[0], id, err)
		}

		// For `get`, JSON is the most useful default; table/wide if explicitly requested.
		switch flagOutput {
		case "table":
			PrintTable(res, []map[string]interface{}{obj}, false)
		case "wide":
			PrintTable(res, []map[string]interface{}{obj}, true)
		default:
			PrintJSON(obj)
		}
		return nil
	},
}

var describeCmd = &cobra.Command{
	Use:   "describe <resource-type> <id>",
	Short: "Show detailed information about an Intune resource",
	Long: `Show all fields of a resource in a human-readable key:value format,
similar to kubectl describe.

EXAMPLES
  intune describe config-policy abc-123
  intune describe compliance-policy abc-123
  intune describe ios-mam abc-123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustInitClient()

		res, err := Lookup(args[0])
		if err != nil {
			return err
		}
		id := args[1]

		var obj map[string]interface{}
		path := fmt.Sprintf(res.SinglePath, id)
		if err := globalClient.Get(context.Background(), path, &obj); err != nil {
			return fmt.Errorf("failed to describe %s %q: %w", res.Names[0], id, err)
		}

		// Print a header line
		name := fmt.Sprintf("%v", obj["displayName"])
		odataType := strings.TrimPrefix(fmt.Sprintf("%v", obj["@odata.type"]), "#microsoft.graph.")
		fmt.Printf("%s  %s  (%s)\n", Bold("---"), Bold(name), odataType)
		fmt.Println(strings.Repeat("─", 60))
		PrintDescribe(obj)
		fmt.Println(strings.Repeat("─", 60))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(describeCmd)
}
