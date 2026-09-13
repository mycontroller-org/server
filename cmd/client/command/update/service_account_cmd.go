package update

import (
	"fmt"
	"strings"

	"github.com/mycontroller-org/server/v2/cmd/client/command/common"
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	dateTimeTY "github.com/mycontroller-org/server/v2/pkg/types/cusom_datetime"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	"github.com/spf13/cobra"
)

var (
	saUser            string
	saName            string
	saDescription     string
	saNeverExpire     bool
	saExpiresOn       string
	saEffect          string
	saActions         []string
	saResources       []string
	saClearStatements bool
)

var serviceAccountUpdateCmd = &cobra.Command{
	Use:     "service-account <alias> <name-or-id>",
	Aliases: []string{"service-accounts", "sa"},
	Short:   "Updates a service account without rotating the token",
	Example: `  myc update sa <alias> ci-bot --description "CI"
  myc update sa <alias> ci-bot --user alice --name ci-bot-2
  myc update sa <alias> ci-bot --never-expire
  myc update sa <alias> ci-bot --expires-on 2027-12-31
  myc update sa <alias> ci-bot --action get --resource "node:*"
  myc update sa <alias> ci-bot --clear-statements`,
	Args:          cobra.ExactArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := updateServiceAccount(cmd, args[0], args[1]); err != nil {
			return fmt.Errorf("error:%s", err)
		}
		return nil
	},
}

func init() {
	serviceAccountUpdateCmd.Flags().StringVarP(&saUser, "user", "u", "", "username or user id when the account name is not unique")
	serviceAccountUpdateCmd.Flags().StringVar(&saName, "name", "", "new name")
	serviceAccountUpdateCmd.Flags().StringVarP(&saDescription, "description", "d", "", "description")
	serviceAccountUpdateCmd.Flags().BoolVar(&saNeverExpire, "never-expire", true, "token never expires")
	serviceAccountUpdateCmd.Flags().StringVar(&saExpiresOn, "expires-on", "", "expiry date (YYYY-MM-DD); turns off never-expire")
	serviceAccountUpdateCmd.Flags().StringVar(&saEffect, "effect", policyTY.EffectAllow, "statement effect: Allow or Deny")
	serviceAccountUpdateCmd.Flags().StringArrayVar(&saActions, "action", nil, "replace statements with this action (repeatable; requires --resource)")
	serviceAccountUpdateCmd.Flags().StringArrayVar(&saResources, "resource", nil, "replace statements with this resource (repeatable; requires --action)")
	serviceAccountUpdateCmd.Flags().BoolVar(&saClearStatements, "clear-statements", false, "remove statements (same access as the owner)")
}

func updateServiceAccount(cmd *cobra.Command, alias, selector string) error {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return fmt.Errorf("name or id is required")
	}
	client := rootCmd.MustClient(alias)
	account, err := client.FindServiceAccount(selector, selector, saUser)
	if err != nil {
		return err
	}
	if account == nil {
		return fmt.Errorf("service-account %s is not present", selector)
	}

	changed := false
	if cmd.Flags().Changed("name") {
		name := strings.TrimSpace(saName)
		if name == "" {
			return fmt.Errorf("name cannot be empty")
		}
		account.Name = name
		changed = true
	}
	if cmd.Flags().Changed("description") {
		account.Description = saDescription
		changed = true
	}
	if cmd.Flags().Changed("expires-on") {
		expires := dateTimeTY.CustomDate{}
		if err := expires.Unmarshal(saExpiresOn); err != nil {
			return fmt.Errorf("expires-on must be YYYY-MM-DD: %w", err)
		}
		account.NeverExpire = false
		account.ExpiresOn = expires
		changed = true
	}
	if cmd.Flags().Changed("never-expire") {
		account.NeverExpire = saNeverExpire
		if saNeverExpire {
			account.ExpiresOn = dateTimeTY.CustomDate{}
		} else if !cmd.Flags().Changed("expires-on") && account.ExpiresOn.IsZero() {
			return fmt.Errorf("--expires-on is required when never-expire is false")
		}
		changed = true
	}
	if saClearStatements {
		account.Statements = []policyTY.Statement{}
		changed = true
	} else {
		statements, provided, err := common.ParseOptionalStatement(saEffect, saActions, saResources)
		if err != nil {
			return err
		}
		if provided {
			account.Statements = statements
			changed = true
		}
	}
	if !changed {
		return fmt.Errorf("no fields to update")
	}
	if err := client.UpdateServiceAccount(account); err != nil {
		return err
	}
	label := account.Name
	if account.Username != "" {
		label = account.Username + "." + account.Name
	}
	_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "service-account: %s\n", label)
	return nil
}
