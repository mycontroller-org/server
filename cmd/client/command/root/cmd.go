package root

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mycontroller-org/server/v2/cmd/client/api"
	clientTY "github.com/mycontroller-org/server/v2/pkg/types/client"
	printer "github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/mycontroller-org/server/v2/pkg/version"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ENV_PREFIX       = "MYC"
	ENV_CONFIG       = "MYC_CONFIG"
	CONFIG_FILE_NAME = ".mycontroller"
	CONFIG_FILE_EXT  = "yaml"
)

var (
	cfgFile      string
	CONFIG       *clientTY.Config
	IOStreams    clientTY.IOStreams
	HideHeader   bool
	Pretty       bool
	OutputFormat string

	rootCliLong = `MyController Client
  
This client helps you to control your MyController server from the command line.
`
)

var Cmd = &cobra.Command{
	Use:   "myc",
	Short: "myc",
	Long:  rootCliLong,
	PreRun: func(cmd *cobra.Command, args []string) {
		UpdateStreams(cmd)
	},
}

func init() {
	CONFIG = &clientTY.Config{}

	v := version.Get()
	Cmd.Version = v.Version
	Cmd.SetVersionTemplate(formatClientVersion(v))

	cobra.OnInitialize(initConfig)

	Cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $MYC_CONFIG or $HOME/.mycontroller.yaml)")
	Cmd.PersistentFlags().StringVarP(&OutputFormat, "output", "o", printer.OutputConsole, "output format. options: yaml, json, console, wide")
	Cmd.PersistentFlags().BoolVar(&HideHeader, "hide-header", false, "hides the header on the console output")
	Cmd.PersistentFlags().BoolVar(&Pretty, "pretty", false, "JSON pretty print")
}

func LookupClient(name string) (*api.Client, error) {
	if CONFIG == nil {
		CONFIG = &clientTY.Config{}
	}
	CONFIG.EnsureAliases()
	if name == "" {
		return nil, fmt.Errorf("alias is required")
	}
	alias, ok := CONFIG.Aliases[name]
	if !ok {
		return nil, fmt.Errorf("alias %q is not configured", name)
	}
	return api.NewClient(alias.URL, alias.GetPassword(), alias.Insecure), nil
}

func MustClient(name string) *api.Client {
	client, err := LookupClient(name)
	if err != nil {
		_, _ = fmt.Fprintln(IOStreams.ErrOut, err)
		os.Exit(1)
	}
	return client
}

func TakeAlias(args []string) (*api.Client, []string) {
	if len(args) < 1 {
		_, _ = fmt.Fprintln(IOStreams.ErrOut, "alias is required")
		os.Exit(1)
	}
	return MustClient(args[0]), args[1:]
}

func formatClientVersion(v version.Version) string {
	return fmt.Sprintf("version:      %s\nbuild date:   %s\ngit commit:   %s\ngolang:       %s\nplatform:     %s\narch:         %s\n",
		v.Version, v.BuildDate, v.GitCommit, v.GoVersion, v.Platform, v.Arch)
}

func UpdateStreams(cmd *cobra.Command) {
	cmd.SetOut(IOStreams.Out)
	cmd.SetErr(IOStreams.ErrOut)
}

func Execute(streams clientTY.IOStreams) {
	IOStreams = streams
	if err := Cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(IOStreams.ErrOut, err)
		os.Exit(1)
	}
}

func WriteConfigFile() {
	if cfgFile == "" {
		return
	}
	if CONFIG == nil {
		CONFIG = &clientTY.Config{}
	}
	CONFIG.EncodePasswords()

	configBytes, err := yaml.Marshal(CONFIG)
	if err != nil {
		_, _ = fmt.Fprintf(IOStreams.ErrOut, "error on config file marshal. error:[%s]\n", err.Error())
	}
	err = os.WriteFile(cfgFile, configBytes, os.ModePerm)
	if err != nil {
		_, _ = fmt.Fprintf(IOStreams.ErrOut, "error on writing config file to disk, filename:%s, error:[%s]\n", cfgFile, err.Error())
	}
}

func initConfig() {
	if cfgFile == "" {
		cfgFile = os.Getenv(ENV_CONFIG)
	}
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(CONFIG_FILE_NAME)
		viper.SetConfigType(CONFIG_FILE_EXT)
		cfgFile = filepath.Join(home, fmt.Sprintf("%s.%s", CONFIG_FILE_NAME, CONFIG_FILE_EXT))
	}

	viper.SetEnvPrefix(ENV_PREFIX)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		err = viper.Unmarshal(&CONFIG)
		if err != nil {
			_, _ = fmt.Fprint(IOStreams.ErrOut, "error on unmarshal of config\n", err)
		}
	}
	if CONFIG == nil {
		CONFIG = &clientTY.Config{}
	}
	CONFIG.EnsureAliases()
}
