package add

import (
	"fmt"
	"os"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	userPassword string
	userEmail    string
	userFullName string
	userPolicies []string
)

var userAddCmd = &cobra.Command{
	Use:     "user <alias> <username>",
	Aliases: []string{"users"},
	Short:   "Adds a user",
	Example: `  myc add user <alias> alice --password secret
  myc add user <alias> alice --email alice@example.com --full-name Alice --policy readonly
  myc add user <alias> alice`,
	Args:          cobra.ExactArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := addUser(args[0], args[1]); err != nil {
			return fmt.Errorf("error:%s", err)
		}
		return nil
	},
}

func init() {
	userAddCmd.Flags().StringVarP(&userPassword, "password", "p", "", "password (prompted if omitted)")
	userAddCmd.Flags().StringVar(&userEmail, "email", "", "email")
	userAddCmd.Flags().StringVar(&userFullName, "full-name", "", "full name")
	userAddCmd.Flags().StringArrayVar(&userPolicies, "policy", nil, "policy id to attach (repeatable)")
}

func addUser(alias, username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username is required")
	}
	client := rootCmd.MustClient(alias)

	existing, err := client.FindUser("", username)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("user %s is already present", username)
	}

	password := userPassword
	if strings.TrimSpace(password) == "" {
		password, err = promptPassword()
		if err != nil {
			return err
		}
	}
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password is required")
	}

	if err := client.SaveUser(&userTY.UserAdminUpdate{
		Username: username,
		Password: password,
		Email:    userEmail,
		FullName: userFullName,
		Policies: userPolicies,
	}); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "user: %s\n", username)
	return nil
}

func promptPassword() (string, error) {
	_, _ = fmt.Fprint(rootCmd.IOStreams.Out, "Password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(rootCmd.IOStreams.Out)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}
