package root

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	loginUsername  string
	loginPassword  string
	loginExpiresIn string
	loginInsecure  bool
)

func init() {
	Cmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Username to login")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password to login")
	loginCmd.Flags().StringVar(&loginExpiresIn, "expires-in", "30d", "session expires in")
	loginCmd.Flags().BoolVar(&loginInsecure, "insecure", false,
		"If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")

	Cmd.AddCommand(logoutCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to a MyController server",
	Example: `  # login into the MyController server with username and password
  export TOKEN=$(cat my_token.txt)
  mc login http://localhost:8080 --username admin --password ${TOKEN}

  # login into the MyController server with username and password
  mc login http://localhost:8080 --username admin --password password

  # login into the MyController insecure server (with SSL certificate)
  mc login https://localhost:8443 --username admin --password password  --insecure`,
	PreRun: func(cmd *cobra.Command, args []string) {
		UpdateStreams(cmd)
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CONFIG.URL = args[0]
		CONFIG.Insecure = loginInsecure
		client := GetClient()
		res, err := client.Login(loginUsername, loginPassword, "")
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
