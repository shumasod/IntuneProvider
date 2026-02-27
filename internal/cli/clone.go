package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var cloneName string

var cloneCmd = &cobra.Command{
	Use:   "clone <resource-type> <id>",
	Short: "Duplicate an Intune resource  (like cp in bash)",
	Long: `Copy an existing Intune resource to a new one, similar to cp(1).

The new resource gets a new name (specified with --name) and a new ID.
Read-only fields (id, createdDateTime, lastModifiedDateTime, version) are
stripped automatically before re-creating the resource.

EXAMPLES
  # Duplicate a config policy with a new name
  intune clone config-policy abc-123 --name "Copy of Windows Defender"

  # Clone a compliance policy
  intune clone compliance-policy def-456 --name "Windows 11 Compliance"

  # Clone an iOS MAM policy
  intune clone ios-mam ghi-789 --name "iOS Outlook MAM - EU"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustInitClient()

		res, err := Lookup(args[0])
		if err != nil {
			return err
		}
		srcID := args[1]

		if cloneName == "" {
			return fmt.Errorf("--name is required")
		}

		// Fetch the source resource
		var src map[string]interface{}
		path := fmt.Sprintf(res.SinglePath, srcID)
		if err := globalClient.Get(context.Background(), path, &src); err != nil {
			return fmt.Errorf("failed to fetch source %s %q: %w", res.Names[0], srcID, err)
		}

		srcDisplayName := fmt.Sprintf("%v", src["displayName"])

		// Strip fields that must not be present on create
		for _, k := range []string{"id", "createdDateTime", "lastModifiedDateTime", "version"} {
			delete(src, k)
		}

		// Set the new display name
		src["displayName"] = cloneName

		var result map[string]interface{}
		if err := globalClient.Post(context.Background(), res.ListPath, src, &result); err != nil {
			return fmt.Errorf("failed to create clone: %w", err)
		}

		newID := fmt.Sprintf("%v", result["id"])
		PrintSuccess(
			"Cloned %s\n  Source : %s (%s)\n  Clone  : %s (%s)",
			res.Names[0],
			srcDisplayName, Cyan(srcID),
			cloneName, Cyan(newID),
		)
		return nil
	},
}

// cpCmd is an alias for cloneCmd (cp = copy, like Linux cp).
var cpCmd = &cobra.Command{
	Use:    "cp <resource-type> <id> --name <new-name>",
	Short:  "Alias for clone (copy a resource)",
	Hidden: true,
	Args:   cobra.ExactArgs(2),
	RunE:   cloneCmd.RunE,
}

func init() {
	cloneCmd.Flags().StringVar(&cloneName, "name", "", "Display name for the cloned resource (required)")
	cpCmd.Flags().StringVar(&cloneName, "name", "", "Display name for the cloned resource (required)")
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(cpCmd)
}
