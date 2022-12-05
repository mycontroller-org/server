package root

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mycontroller-org/server/v2/cmd/client/api"
	printer "github.com/mycontroller-org/server/v2/cmd/client/printer"
	clientTY "github.com/mycontroller-org/server/v2/pkg/types/client"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ENV_PREFIX       = "MYC"
	CONFIG_FILE_NAME = ".mycontroller"
	CONFIG_FILE_EXT  = "yaml"
)

var (
	cfgFile   string
	CONFIG    *clientTY.Config   // keep MyController server details
	IOStreams clientTY.IOStreams // read and write to this stream

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

	cobra.OnInitialize(initConfig)

	Cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mycontroller.yaml)")
	Cmd.PersistentFlags().StringVarP(&OutputFormat, "output", "o", printer.OutputConsole, "output format. options: yaml, json, console, wide")
	Cmd.PersistentFlags().BoolVar(&HideHeader, "hide-header", false, "hides the header on the console output")
	Cmd.PersistentFlags().BoolVar(&Pretty, "pretty", false, "JSON pretty print")
}

func GetClient() *api.Client {
	return api.NewClient(CONFIG.URL, CONFIG.GetPassword(), CONFIG.Insecure)
}

func UpdateStreams(cmd *cobra.Command) {
	cmd.SetOut(IOStreams.Out)
	cmd.SetErr(IOStreams.ErrOut)
}

func Execute(streams clientTY.IOStreams) {
	IOStreams = streams
	if err := Cmd.Execute(); err != nil {
		fmt.Fprintln(IOStreams.ErrOut, err)
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
	// encode password field
	CONFIG.EncodePassword()

	configBytes, err := yaml.Marshal(CONFIG)
	if err != nil {
		fmt.Fprintf(IOStreams.ErrOut, "error on config file marshal. error:[%s]\n", err.Error())
	}
	err = os.WriteFile(cfgFile, configBytes, os.ModePerm)
	if err != nil {
		fmt.Fprintf(IOStreams.ErrOut, "error on writing config file to disk, filename:%s, error:[%s]\n", cfgFile, err.Error())
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.initConfig
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".jenkinsctl" (without extension).
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
			fmt.Fprint(IOStreams.ErrOut, "error on unmarshal of config\n", err)
		}
	}
}
