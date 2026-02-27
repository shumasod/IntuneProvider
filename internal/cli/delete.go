package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var deleteYes bool

var deleteCmd = &cobra.Command{
	Use:   "delete <resource-type> <id>",
	Short: "Delete an Intune resource by ID  (like rm in bash)",
	Long: `Delete an Intune resource, similar to rm(1).

A confirmation prompt is shown unless -y / --yes is supplied.
Pass multiple IDs to delete in one command (like rm file1 file2 file3).

EXAMPLES
  intune delete config-policy abc-123
  intune delete compliance-policy abc-123 -y
  intune rm config-policy abc-123 def-456    # rm alias, multiple IDs`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustInitClient()

		res, err := Lookup(args[0])
		if err != nil {
			return err
		}
		ids := args[1:]

		if !deleteYes {
			if !confirmDelete(res.Names[0], ids) {
				fmt.Println("Aborted.")
				return nil
			}
		}

		var errs []string
		for _, id := range ids {
			path := fmt.Sprintf(res.SinglePath, id)
			if err := globalClient.Delete(context.Background(), path); err != nil {
				if client.IsNotFound(err) {
					PrintWarning("%s %q not found (already deleted?)", res.Names[0], id)
				} else {
					errs = append(errs, fmt.Sprintf("%s: %v", id, err))
				}
				continue
			}
			PrintSuccess("Deleted %s %s", res.Names[0], Cyan(id))
		}

		if len(errs) > 0 {
			return fmt.Errorf("some deletions failed:\n  %s", strings.Join(errs, "\n  "))
		}
		return nil
	},
}

// rmCmd is an alias for deleteCmd (like rm vs delete).
var rmCmd = &cobra.Command{
	Use:    "rm <resource-type> <id> [<id>...]",
	Short:  "Alias for delete",
	Hidden: true,
	Args:   cobra.MinimumNArgs(2),
	RunE:   deleteCmd.RunE,
}

// confirmDelete asks the user to type the resource type name to confirm deletion.
func confirmDelete(resourceType string, ids []string) bool {
	if len(ids) == 1 {
		fmt.Printf(Yellow("⚠  You are about to delete %s %s\n"), resourceType, ids[0])
	} else {
		fmt.Printf(Yellow("⚠  You are about to delete %d %s resource(s):\n"), len(ids), resourceType)
		for _, id := range ids {
			fmt.Printf("   • %s\n", id)
		}
	}
	fmt.Printf("Type %q to confirm: ", resourceType)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	return input == resourceType
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
	rmCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(rmCmd)
}
