package root

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	loginUsername  string
	loginPassword  string
	loginToken     string
	loginExpiresIn string
	loginInsecure  bool
)

func init() {
	Cmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Username to login")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password to login")
	loginCmd.Flags().StringVarP(&loginToken, "token", "t", "", "token to login")
	loginCmd.Flags().StringVar(&loginExpiresIn, "expires-in", "30d", "session expires in")
	loginCmd.Flags().BoolVar(&loginInsecure, "insecure", false,
		"If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")

	Cmd.AddCommand(logoutCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to a MyController server",
	Example: `  # login into the MyController server with username and password
  myc login http://localhost:8080 --username admin --password password

  # login into the MyController insecure server (with SSL certificate)
  myc login https://localhost:8443 --username admin --password password  --insecure

  # prompt username and password
  myc login http://localhost:8080

  # prompt password
  myc login http://localhost:8080 --username admin

  # token based login
  myc login http://localhost:8080 --token <token>
	`,
	PreRun: func(cmd *cobra.Command, args []string) {
		UpdateStreams(cmd)
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if loginToken == "" {
			// get username from terminal
			if loginUsername == "" {
				_username, err := promptUsername()
				if err != nil {
					fmt.Fprintln(IOStreams.ErrOut, err.Error())
					return
				}
				loginUsername = _username
			}

			// get password from terminal
			if loginPassword == "" {
				_password, err := promptPassword()
				if err != nil {
					fmt.Fprintln(IOStreams.ErrOut, err.Error())
					return
				}
				loginPassword = _password
			}
		}

		CONFIG.URL = args[0]
		CONFIG.Insecure = loginInsecure
		client := GetClient()
		res, err := client.Login(loginUsername, loginPassword, loginToken, "")
		if err != nil {
			fmt.Fprintln(IOStreams.ErrOut, "error on login", err)
			return
		}
		if res != nil {
			fmt.Fprintln(IOStreams.ErrOut, "Login successful.")
			CONFIG.URL = args[0]
			CONFIG.Username = loginUsername
			CONFIG.Password = res.Token
			CONFIG.Insecure = loginInsecure
			CONFIG.LoginTime = time.Now().Format(time.RFC3339)
			CONFIG.ExpiresIn = loginExpiresIn
			WriteConfigFile()
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from a server",
	Example: `  # logout from a server
  mc logout`,
	Run: func(cmd *cobra.Command, args []string) {
		if CONFIG.URL == "" {
			fmt.Fprintln(IOStreams.ErrOut, "There is no connection information.")
			return
		}
		CONFIG.URL = ""
		CONFIG.Username = ""
		CONFIG.Password = ""
		CONFIG.Insecure = false
		fmt.Fprintln(IOStreams.Out, "Logout successful.")
		WriteConfigFile()
	},
}

func promptUsername() (string, error) {
	_, err := fmt.Fprint(IOStreams.Out, "Username: ")
	var username string
	fmt.Fscanln(IOStreams.In, &username)
	return username, err
}

func promptPassword() (string, error) {
	fmt.Fprint(IOStreams.Out, "Password: ")
	// TODO: should use IOStreams.In in the place of os.Stdin.Fd
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(IOStreams.Out)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}
