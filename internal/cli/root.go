// Package cli implements the `intune` command-line tool for managing Microsoft
// Intune resources. Commands follow Linux conventions (ls, get, create, apply,
// delete/rm, describe, clone) and authenticate via Azure AD client credentials.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/shumasod/IntuneProvider/internal/client"
)

// version is injected at build time via -ldflags.
var version = "dev"

// global flags (shared across all subcommands via PersistentFlags)
var (
	flagTenantID     string
	flagClientID     string
	flagClientSecret string
	flagOutput       string // "table" | "json" | "wide"
)

// globalClient is initialised once by initClient() before any command runs.
var globalClient *client.Client

var rootCmd = &cobra.Command{
	Use:   "intune",
	Short: "CLI for managing Microsoft Intune resources",
	Long: `intune — a Linux-style CLI for Microsoft Intune (via Microsoft Graph API)

AUTHENTICATION
  Set credentials via flags or environment variables:
    --tenant-id     / INTUNE_TENANT_ID
    --client-id     / INTUNE_CLIENT_ID
    --client-secret / INTUNE_CLIENT_SECRET

RESOURCE TYPES
  ` + strings.Join(AllNames(), "\n  ") + `

EXAMPLES
  intune ls config-policy
  intune ls compliance-policy -o wide
  intune get config-policy <id>
  intune get config-policy <id> -o json > backup.json
  intune describe ios-mam <id>
  intune create config-policy -f policy.json
  intune apply compliance-policy -f policy.json
  intune delete config-policy <id>
  intune rm config-policy <id>          # alias for delete
  intune clone config-policy <id> --name "Copy of policy"
  intune version`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is the public entry point called from cmd/intune/main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
}

// SetVersion allows the build to inject a version string.
func SetVersion(v string) { version = v }

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagTenantID, "tenant-id", "", "Azure AD Tenant ID (env: INTUNE_TENANT_ID)")
	pf.StringVar(&flagClientID, "client-id", "", "Service principal client ID (env: INTUNE_CLIENT_ID)")
	pf.StringVar(&flagClientSecret, "client-secret", "", "Service principal client secret (env: INTUNE_CLIENT_SECRET)")
	pf.StringVarP(&flagOutput, "output", "o", "table", "Output format: table, wide, json")

	// version subcommand
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the intune CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("intune CLI %s\n", version)
		},
	})
}

// initClient builds the Graph API client from flags / env vars.
// Call this at the start of any command that needs API access.
func initClient() error {
	if globalClient != nil {
		return nil
	}

	tenantID := resolveFlag(flagTenantID, "INTUNE_TENANT_ID")
	clientID := resolveFlag(flagClientID, "INTUNE_CLIENT_ID")
	clientSecret := resolveFlag(flagClientSecret, "INTUNE_CLIENT_SECRET")

	var missing []string
	if tenantID == "" {
		missing = append(missing, "--tenant-id / INTUNE_TENANT_ID")
	}
	if clientID == "" {
		missing = append(missing, "--client-id / INTUNE_CLIENT_ID")
	}
	if clientSecret == "" {
		missing = append(missing, "--client-secret / INTUNE_CLIENT_SECRET")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required credentials:\n  %s", strings.Join(missing, "\n  "))
	}

	c, err := client.NewClient(tenantID, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	globalClient = c
	return nil
}

// resolveFlag returns flag value if set, otherwise falls back to the environment variable.
func resolveFlag(flag, envKey string) string {
	if flag != "" {
		return flag
	}
	return os.Getenv(envKey)
}

// mustInitClient calls initClient and exits on error.
func mustInitClient() {
	if err := initClient(); err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
}

// listResponse is used to decode Graph API collection responses.
type listResponse struct {
	Value []map[string]interface{} `json:"value"`
}
