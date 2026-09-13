package add

import (
	"fmt"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	dateTimeTY "github.com/mycontroller-org/server/v2/pkg/types/cusom_datetime"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	"github.com/spf13/cobra"
)

var (
	saUser        string
	saDescription string
	saNeverExpire bool
	saExpiresOn   string
	saActions     []string
	saResources   []string
)

var serviceAccountAddCmd = &cobra.Command{
	Use:     "service-account <alias> <name>",
	Aliases: []string{"service-accounts", "sa"},
	Short:   "Adds a service account and prints the token once",
	Example: `  myc add service-account <alias> ci-bot
  myc add sa <alias> mobile --user alice --description "phone login"
  myc add sa <alias> ci-bot --action get --action list --resource "node:*"
  myc add sa <alias> temp --expires-on 2027-12-31`,
	Args: cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := addServiceAccount(args[0], args[1]); err != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
		}
	},
}

func init() {
	serviceAccountAddCmd.Flags().StringVarP(&saUser, "user", "u", "", "username or user id (defaults to the logged-in user)")
	serviceAccountAddCmd.Flags().StringVarP(&saDescription, "description", "d", "", "description")
	serviceAccountAddCmd.Flags().BoolVar(&saNeverExpire, "never-expire", true, "token never expires")
	serviceAccountAddCmd.Flags().StringVar(&saExpiresOn, "expires-on", "", "expiry date (YYYY-MM-DD); turns off never-expire")
	serviceAccountAddCmd.Flags().StringArrayVar(&saActions, "action", nil, "statement action (repeatable)")
	serviceAccountAddCmd.Flags().StringArrayVar(&saResources, "resource", nil, "statement resource (repeatable)")
}

func addServiceAccount(alias, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	client := rootCmd.MustClient(alias)

	existing, err := client.FindServiceAccount("", name, saUser)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("service-account %s is already present", name)
	}

	account := &svcAccountTY.ServiceAccount{
		Name:        name,
		Username:    strings.TrimSpace(saUser),
		Description: saDescription,
		NeverExpire: saNeverExpire,
	}
	if saExpiresOn != "" {
		expires := dateTimeTY.CustomDate{}
		if err := expires.Unmarshal(saExpiresOn); err != nil {
			return fmt.Errorf("expires-on must be YYYY-MM-DD: %w", err)
		}
		account.NeverExpire = false
		account.ExpiresOn = expires
	} else if !account.NeverExpire {
		return fmt.Errorf("--expires-on is required when never-expire is false")
	}

	if len(saActions) > 0 || len(saResources) > 0 {
		account.Statements = []policyTY.Statement{{
			Effect:    policyTY.EffectAllow,
			Actions:   saActions,
			Resources: saResources,
		}}
	}

	created, err := client.CreateServiceAccount(account)
	if err != nil {
		return err
	}
	if created == nil || created.Token == "" {
		return fmt.Errorf("service account created, but token was not returned")
	}

	out := rootCmd.IOStreams.Out
	fmt.Fprintf(out, "service-account: %s\n", tableName(account))
	fmt.Fprintln(out, "Save this token now. It will not be shown again.")
	fmt.Fprintln(out, created.Token)
	return nil
}

func tableName(account *svcAccountTY.ServiceAccount) string {
	if account.Username != "" {
		return account.Username + "." + account.Name
	}
	return account.Name
}
