package get

import (
	"fmt"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/spf13/cobra"
)

var saGetUser string

var serviceAccountGetCmd = &cobra.Command{
	Use:     "service-account <alias> [<name-or-id>]",
	Aliases: []string{"service-accounts", "sa"},
	Short:   "Print service accounts",
	Example: `  myc get service-account <alias>
  myc get service-account <alias> ci-bot
  myc get sa <alias> ci-bot --user alice`,
	Args: cobra.RangeArgs(1, 2),
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.MustClient(args[0])
		headers := []printer.Header{
			{Title: "id", IsWide: true},
			{Title: "username"},
			{Title: "name"},
			{Title: "description"},
			{Title: "disabled"},
			{Title: "never expire", ValuePath: "neverExpire"},
			{Title: "expires on", ValueFunc: formatServiceAccountExpiresOn},
			{Title: "statements", ValueFunc: formatServiceAccountStatements},
			{Title: "created on", ValuePath: "createdOn", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		if len(args) == 1 {
			executeGetCmd(headers, client.ListServiceAccount, svcAccountTY.ServiceAccount{})
			return
		}
		account, err := client.FindServiceAccount(args[1], args[1], saGetUser)
		if err != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
			return
		}
		if account == nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:service-account %s is not present\n", args[1])
			return
		}
		account.Token.Token = ""
		printOne(headers, account)
	},
}

func init() {
	serviceAccountGetCmd.Flags().StringVarP(&saGetUser, "user", "u", "", "username or user id when the account name is not unique")
}

func serviceAccountFromItem(item interface{}) *svcAccountTY.ServiceAccount {
	switch typed := item.(type) {
	case *svcAccountTY.ServiceAccount:
		return typed
	case svcAccountTY.ServiceAccount:
		return &typed
	default:
		return nil
	}
}

func formatServiceAccountExpiresOn(item interface{}) string {
	account := serviceAccountFromItem(item)
	if account == nil || account.NeverExpire || account.ExpiresOn.IsZero() {
		return "-"
	}
	return printer.FormatTimeValue(account.ExpiresOn.Time, printer.DisplayStyleRelativeTime)
}

func formatServiceAccountStatements(item interface{}) string {
	account := serviceAccountFromItem(item)
	if account == nil {
		return "-"
	}
	return formatStatements(account.Statements)
}
