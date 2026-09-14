package update

import (
	"fmt"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/spf13/cobra"
)

var (
	userPassword      string
	userEmail         string
	userFullName      string
	userNewUsername   string
	userPolicies      []string
	userClearPolicies bool
)

var userUpdateCmd = &cobra.Command{
	Use:     "user <alias> <username-or-id>",
	Aliases: []string{"users"},
	Short:   "Updates a user",
	Example: `  myc update user <alias> alice --email alice@example.com
  myc update user <alias> alice --full-name Alice --policy readonly --policy admin
  myc update user <alias> alice --password newsecret
  myc update user <alias> alice --username alice2`,
	Args:          cobra.ExactArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := updateUser(cmd, args[0], args[1]); err != nil {
			return fmt.Errorf("error:%s", err)
		}
		return nil
	},
}

func init() {
	userUpdateCmd.Flags().StringVar(&userNewUsername, "username", "", "new username")
	userUpdateCmd.Flags().StringVarP(&userPassword, "password", "p", "", "new password")
	userUpdateCmd.Flags().StringVar(&userEmail, "email", "", "email")
	userUpdateCmd.Flags().StringVar(&userFullName, "full-name", "", "full name")
	userUpdateCmd.Flags().StringArrayVar(&userPolicies, "policy", nil, "replace attached policies (repeatable)")
	userUpdateCmd.Flags().BoolVar(&userClearPolicies, "clear-policies", false, "remove all attached policies")
}

func updateUser(cmd *cobra.Command, alias, selector string) error {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return fmt.Errorf("username or id is required")
	}
	client := rootCmd.MustClient(alias)
	user, err := client.FindUser(selector, selector)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user %s is not present", selector)
	}

	update := &userTY.UserAdminUpdate{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
		Policies: user.Policies,
		Labels:   user.Labels,
	}
	changed := false
	if cmd.Flags().Changed("username") {
		name := strings.TrimSpace(userNewUsername)
		if name == "" {
			return fmt.Errorf("username cannot be empty")
		}
		update.Username = name
		changed = true
	}
	if cmd.Flags().Changed("email") {
		update.Email = userEmail
		changed = true
	}
	if cmd.Flags().Changed("full-name") {
		update.FullName = userFullName
		changed = true
	}
	if cmd.Flags().Changed("password") {
		if strings.TrimSpace(userPassword) == "" {
			return fmt.Errorf("password cannot be empty")
		}
		update.Password = userPassword
		changed = true
	}
	if userClearPolicies {
		update.Policies = []string{}
		changed = true
	} else if cmd.Flags().Changed("policy") {
		update.Policies = userPolicies
		changed = true
	}
	if !changed {
		return fmt.Errorf("no fields to update")
	}
	if err := client.SaveUser(update); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "user: %s\n", update.Username)
	return nil
}
