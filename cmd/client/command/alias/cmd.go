package alias

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/cmd/client/api"
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	clientTY "github.com/mycontroller-org/server/v2/pkg/types/client"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var aliasNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

var reservedAliasNames = map[string]struct{}{
	"alias": {}, "apply": {}, "action": {}, "completion": {}, "delete": {},
	"disable": {}, "enable": {}, "get": {}, "help": {}, "reboot": {},
	"reload": {}, "server": {}, "set": {}, "upload": {}, "myc": {},
}

var (
	aliasUsername  string
	aliasPassword  string
	aliasToken     string
	aliasExpiresIn string
	aliasInsecure  bool
)

func init() {
	rootCmd.Cmd.AddCommand(aliasCmd)
	aliasCmd.AddCommand(aliasSetCmd)
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasRemoveCmd)

	aliasSetCmd.Flags().StringVarP(&aliasUsername, "username", "u", "", "username to login")
	aliasSetCmd.Flags().StringVarP(&aliasPassword, "password", "p", "", "password to login")
	aliasSetCmd.Flags().StringVarP(&aliasToken, "token", "t", "", "service token to login")
	aliasSetCmd.Flags().StringVar(&aliasExpiresIn, "expires-in", "720h", "session expires in")
	aliasSetCmd.Flags().BoolVar(&aliasInsecure, "insecure", false, "skip TLS certificate verification")
}

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage named server connections",
	Long: `Aliases store a server URL and a logged-in user session.
Every server command takes the alias as its first argument.

  myc alias set <alias> http://localhost:8080
  myc alias set <alias> https://mc.example.com -u admin --insecure
  myc alias list
  myc get node <alias>
  myc apply <alias> -f resources.yaml
`,
	SilenceUsage: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

var aliasSetCmd = &cobra.Command{
	Use:   "set <name> <url>",
	Short: "Add or update an alias and log in",
	Example: `  myc alias set <alias> http://localhost:8080
  myc alias set <alias> http://localhost:8080 -u admin
  myc alias set <alias> https://mc.example.com -u admin -p secret --insecure
  myc alias set <alias> http://localhost:8080 --token <token>`,
	Args:          cobra.ExactArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name, url := args[0], strings.TrimRight(args[1], "/")
		if err := validateAliasName(name); err != nil {
			return err
		}
		if url == "" {
			return fmt.Errorf("url is required")
		}

		username := aliasUsername
		password := aliasPassword
		if aliasToken == "" {
			if username == "" {
				var err error
				username, err = promptUsername()
				if err != nil {
					return err
				}
			}
			if password == "" {
				var err error
				password, err = promptPassword()
				if err != nil {
					return err
				}
			}
		}

		client := api.NewClient(url, "", aliasInsecure)
		res, err := client.Login(username, password, aliasToken, aliasExpiresIn)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		rootCmd.CONFIG.EnsureAliases()
		rootCmd.CONFIG.Aliases[name] = clientTY.Alias{
			URL:       url,
			Insecure:  aliasInsecure,
			Username:  username,
			Password:  res.Token,
			LoginTime: time.Now().Format(time.RFC3339),
			ExpiresIn: aliasExpiresIn,
		}
		rootCmd.WriteConfigFile()
		_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "alias %q is ready\n", name)
		return nil
	},
}

var aliasListCmd = &cobra.Command{
	Use:           "list",
	Aliases:       []string{"ls"},
	Short:         "List configured aliases",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		rootCmd.CONFIG.EnsureAliases()
		if len(rootCmd.CONFIG.Aliases) == 0 {
			_, _ = fmt.Fprintln(rootCmd.IOStreams.Out, "no aliases. add one with: myc alias set <name> <url>")
			return nil
		}
		names := make([]string, 0, len(rootCmd.CONFIG.Aliases))
		for name := range rootCmd.CONFIG.Aliases {
			names = append(names, name)
		}
		sort.Strings(names)

		headers := []printer.Header{
			{Title: "name", ValuePath: "Name"},
			{Title: "url", ValuePath: "URL"},
			{Title: "user", ValuePath: "User"},
			{Title: "insecure", ValuePath: "Insecure"},
		}
		rows := make([]interface{}, 0, len(names))
		for _, name := range names {
			a := rootCmd.CONFIG.Aliases[name]
			rows = append(rows, aliasRow{
				Name:     name,
				URL:      a.URL,
				User:     a.Username,
				Insecure: a.Insecure,
			})
		}
		printer.Print(rootCmd.IOStreams.Out, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
		return nil
	},
}

var aliasRemoveCmd = &cobra.Command{
	Use:           "remove <name>",
	Aliases:       []string{"rm"},
	Short:         "Remove an alias",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		rootCmd.CONFIG.EnsureAliases()
		if _, ok := rootCmd.CONFIG.Aliases[name]; !ok {
			return fmt.Errorf("alias %q is not configured", name)
		}
		delete(rootCmd.CONFIG.Aliases, name)
		rootCmd.WriteConfigFile()
		_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "removed alias %q\n", name)
		return nil
	},
}

type aliasRow struct {
	Name     string
	URL      string
	User     string
	Insecure bool
}

func validateAliasName(name string) error {
	if !aliasNamePattern.MatchString(name) {
		return fmt.Errorf("invalid alias name %q (use letters, numbers, - or _ and start with a letter)", name)
	}
	if _, reserved := reservedAliasNames[strings.ToLower(name)]; reserved {
		return fmt.Errorf("alias name %q is reserved", name)
	}
	return nil
}

func promptUsername() (string, error) {
	var username string
	_, err := fmt.Fprint(rootCmd.IOStreams.Out, "Username: ")
	if err != nil {
		return "", err
	}
	_, err = fmt.Fscanln(rootCmd.IOStreams.In, &username)
	return username, err
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
