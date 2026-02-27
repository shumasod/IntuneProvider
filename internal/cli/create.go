package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var createFile string

var createCmd = &cobra.Command{
	Use:   "create <resource-type> -f <file.json>",
	Short: "Create an Intune resource from a JSON file  (like touch / install)",
	Long: `Create a new Intune resource from a JSON or stdin payload.

The JSON structure must match the Microsoft Graph API body for the resource type.
Use 'intune get <type> <id> -o json' to obtain a valid template from an existing resource.

EXAMPLES
  # Create from a JSON file
  intune create config-policy -f windows_defender.json

  # Create from stdin
  cat policy.json | intune create compliance-policy -f -

  # Quick one-liner
  echo '{"displayName":"Test","@odata.type":"#microsoft.graph.windows10CompliancePolicy"}' \
    | intune create compliance-policy -f -`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustInitClient()

		res, err := Lookup(args[0])
		if err != nil {
			return err
		}

		body, err := readInput(createFile)
		if err != nil {
			return err
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		// Strip id/timestamps that must not be sent on create
		for _, k := range []string{"id", "createdDateTime", "lastModifiedDateTime", "version"} {
			delete(payload, k)
		}

		var result map[string]interface{}
		if err := globalClient.Post(context.Background(), res.ListPath, payload, &result); err != nil {
			return fmt.Errorf("failed to create %s: %w", res.Names[0], err)
		}

		id := fmt.Sprintf("%v", result["id"])
		name := fmt.Sprintf("%v", result["displayName"])
		PrintSuccess("Created %s %q  →  ID: %s", res.Names[0], name, Cyan(id))

		if flagOutput == "json" {
			PrintJSON(result)
		}
		return nil
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply <resource-type> -f <file.json>",
	Short: "Create or update an Intune resource  (idempotent)",
	Long: `Apply a JSON file to create or update a resource.

If the JSON body contains an "id" field the resource is updated (PATCH).
If there is no "id" the resource is created (POST).

This makes apply idempotent and suitable for use in pipelines and CI/CD:

  intune get config-policy <id> -o json > policy.json
  # edit policy.json ...
  intune apply config-policy -f policy.json

EXAMPLES
  intune apply config-policy      -f windows_defender.json
  intune apply compliance-policy  -f ios_compliance.json
  intune apply ios-mam            -f mam_policy.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustInitClient()

		res, err := Lookup(args[0])
		if err != nil {
			return err
		}

		body, err := readInput(applyFile)
		if err != nil {
			return err
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		id, hasID := payload["id"].(string)

		if hasID && id != "" {
			// UPDATE (PATCH)
			for _, k := range []string{"createdDateTime", "lastModifiedDateTime", "version"} {
				delete(payload, k)
			}
			path := fmt.Sprintf(res.SinglePath, id)
			if err := globalClient.Patch(context.Background(), path, payload); err != nil {
				return fmt.Errorf("failed to update %s %q: %w", res.Names[0], id, err)
			}
			name := fmt.Sprintf("%v", payload["displayName"])
			PrintSuccess("Updated %s %q  →  ID: %s", res.Names[0], name, Cyan(id))
		} else {
			// CREATE (POST)
			for _, k := range []string{"id", "createdDateTime", "lastModifiedDateTime", "version"} {
				delete(payload, k)
			}
			var result map[string]interface{}
			if err := globalClient.Post(context.Background(), res.ListPath, payload, &result); err != nil {
				return fmt.Errorf("failed to create %s: %w", res.Names[0], err)
			}
			id = fmt.Sprintf("%v", result["id"])
			name := fmt.Sprintf("%v", result["displayName"])
			PrintSuccess("Created %s %q  →  ID: %s", res.Names[0], name, Cyan(id))

			if flagOutput == "json" {
				PrintJSON(result)
			}
		}
		return nil
	},
}

var applyFile string

// readInput reads from a file path or stdin ("-").
func readInput(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, fmt.Errorf("specify an input file with -f <file.json> or -f - to read from stdin")
	}
	if filePath == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading from stdin: %w", err)
		}
		return data, nil
	}
	if !strings.HasSuffix(filePath, ".json") {
		PrintWarning("file %q does not have a .json extension", filePath)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", filePath, err)
	}
	return data, nil
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "Path to JSON file, or - for stdin (required)")
	_ = createCmd.MarkFlagRequired("file")

	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "", "Path to JSON file, or - for stdin (required)")
	_ = applyCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(applyCmd)
}
