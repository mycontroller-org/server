package commands

import (
	"fmt"
	"os"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"gopkg.in/yaml.v3"
)

// PrintVersion prints the version and exits
func PrintDefaultConfig(caller string) {
	if caller != CallerServer {
		fmt.Println("generate default config implemented only for the server. For other componenet please refer documentation")
		return
	}
	verBytes, err := yaml.Marshal(getDefaultServerConfig())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(verBytes))
}

func getDefaultServerConfig() *config.Config {
	cfg := &config.Config{
		Secret:    utils.RandIDWithLength(32),
		Analytics: config.AnalyticsConfig{Enabled: true},
		Web: config.WebConfig{
			WebDirectory:     "",
			DocumentationURL: "",
			EnableProfiling:  false,
			Http: config.HttpConfig{
				Enabled:     true,
				BindAddress: "0.0.0.0",
				Port:        8080,
			},
			HttpsSSL: config.HttpsSSLConfig{
				Enabled:     true,
				BindAddress: "0.0.0.0",
				Port:        8443,
				CertDir:     "/mc_home/certs/https_ssl",
			},
			HttpsACME: config.HttpsACMEConfig{
				Enabled:     false,
				BindAddress: "0.0.0.0",
				Port:        9443,
				Email:       "mycontroller@example.com",
				Domains:     []string{"mycontroller.example.com"},
				CacheDir:    "/mc_home/certs/https_acme",
			},
		},
		Logger: config.LoggerConfig{
			Mode:     "record_all",
			Encoding: "console",
			Level: config.LogLevelConfig{
				Core:       "info",
				WebHandler: "info",
				Storage:    "info",
				Metric:     "warn",
			},
		},
		Directories: config.Directories{
			Data:          "/mc_home/data",
			Logs:          "/mc_home/logs",
			Tmp:           "/mc_home/tmp",
			SecureShare:   "/mc_home/secure_share",
			InsecureShare: "/mc_home/insecure_share",
		},

		Bus: cmap.CustomMap{
			"type":               "embedded", // other option: natsio
			"topic_prefix":       "mc_bus",
			"server_url":         "nats://192.168.1.21:4222",
			"insecure":           false,
			"connection_timeout": "10s",
		},
		Gateway: sfTY.ServiceFilter{
			Disabled: false,
			IDs:      []string{},
			Labels:   cmap.CustomStringMap{"location": "server"},
		},
		Handler: sfTY.ServiceFilter{
			Disabled: false,
			IDs:      []string{},
			Labels:   cmap.CustomStringMap{"location": "server"},
		},
		Database: config.Database{
			Storage: cmap.CustomMap{
				"type":          "memory",
				"dump_enabled":  true,
				"dump_interval": "30m",
				"dump_dir":      "memory_db",
				"dump_format":   []string{"yaml", "json"},
				"load_format":   "yaml",
			},
			Metric: cmap.CustomMap{
				"disabled":          true,
				"type":              "influxdb",
				"uri":               "http://192.168.1.21:8086",
				"token":             "",
				"username":          "",
				"password":          "",
				"organization_name": "",
				"bucket_name":       "mycontroller",
				"batch_size":        "",
				"flush_interval":    "5s",
			},
		},
	}
	return cfg
}
