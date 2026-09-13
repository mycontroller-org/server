package get

import (
	"fmt"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/spf13/cobra"
)

var userGetCmd = &cobra.Command{
	Use:     "user <alias> [<username-or-id> [policies]]",
	Aliases: []string{"users"},
	Short:   "Print users, a user, or a user's linked policies",
	Example: `  myc get user <alias>
  myc get user <alias> alice
  myc get user <alias> alice policies`,
	Args: cobra.RangeArgs(1, 3),
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.MustClient(args[0])
		if len(args) == 1 {
			headers := []printer.Header{
				{Title: "id", IsWide: true},
				{Title: "username"},
				{Title: "email"},
				{Title: "full name", ValuePath: "fullName"},
				{Title: "disabled"},
				{Title: "policies", ValueFunc: formatUserPolicies},
			}
			executeGetCmd(headers, client.ListUser, userTY.User{})
			return
		}
		user, err := client.FindUser(args[1], args[1])
		if err != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
			return
		}
		if user == nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:user %s is not present\n", args[1])
			return
		}
		if len(args) == 3 {
			if strings.ToLower(args[2]) != "policies" {
				_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error: unknown argument %q (use policies)\n", args[2])
				return
			}
			printUserPolicies(client, user)
			return
		}
		headers := []printer.Header{
			{Title: "id"},
			{Title: "username"},
			{Title: "email"},
			{Title: "full name", ValuePath: "fullName"},
			{Title: "disabled"},
			{Title: "policies", ValueFunc: formatUserPolicies},
		}
		printOne(headers, user)
	},
}

var policyGetCmd = &cobra.Command{
	Use:     "policy <alias> [<id>]",
	Aliases: []string{"policies"},
	Short:   "Print policies",
	Example: `  myc get policy <alias>
  myc get policy <alias> admin`,
	Args: cobra.RangeArgs(1, 2),
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.MustClient(args[0])
		if len(args) == 1 {
			headers := []printer.Header{
				{Title: "id"},
				{Title: "description"},
				{Title: "system"},
				{Title: "statements", ValueFunc: formatPolicyStatements},
			}
			executeGetCmd(headers, client.ListPolicy, policyTY.Policy{})
			return
		}
		policy, err := client.FindPolicy(args[1])
		if err != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
			return
		}
		if policy == nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:policy %s is not present\n", args[1])
			return
		}
		headers := []printer.Header{
			{Title: "id"},
			{Title: "description"},
			{Title: "system"},
			{Title: "statements", ValueFunc: formatPolicyStatements},
		}
		printOne(headers, policy)
	},
}

func printOne(headers []printer.Header, item interface{}) {
	switch rootCmd.OutputFormat {
	case printer.OutputYAML, printer.OutputJSON:
		printer.Print(rootCmd.IOStreams.Out, headers, item, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
	default:
		printer.Print(rootCmd.IOStreams.Out, headers, []interface{}{item}, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
	}
}

func printUserPolicies(client interface {
	FindPolicy(id string) (*policyTY.Policy, error)
}, user *userTY.User) {
	if len(user.Policies) == 0 {
		_, _ = fmt.Fprintln(rootCmd.IOStreams.Out, "No policies attached")
		return
	}
	rows := make([]interface{}, 0, len(user.Policies))
	for _, id := range user.Policies {
		policy, err := client.FindPolicy(id)
		if err != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
			continue
		}
		if policy == nil {
			rows = append(rows, &policyTY.Policy{ID: id, Description: "(not present)"})
			continue
		}
		rows = append(rows, policy)
	}
	headers := []printer.Header{
		{Title: "id"},
		{Title: "description"},
		{Title: "system"},
		{Title: "statements", ValueFunc: formatPolicyStatements},
	}
	printer.Print(rootCmd.IOStreams.Out, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
}

func formatUserPolicies(item interface{}) string {
	switch user := item.(type) {
	case *userTY.User:
		return strings.Join(user.Policies, ",")
	case userTY.User:
		return strings.Join(user.Policies, ",")
	default:
		return ""
	}
}

func formatPolicyStatements(item interface{}) string {
	var statements []policyTY.Statement
	switch policy := item.(type) {
	case *policyTY.Policy:
		statements = policy.Statements
	case policyTY.Policy:
		statements = policy.Statements
	default:
		return ""
	}
	return fmt.Sprintf("%d", len(statements))
}
