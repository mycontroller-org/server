module github.com/mycontroller-org/server/v2

go 1.26.2

require (
	github.com/Masterminds/semver/v3 v3.5.0
	github.com/NYTimes/gziphandler v1.1.1
	github.com/amimof/huego v1.2.1
	github.com/btittelbach/astrotime v0.0.0-20160515101311-7ddba43aa26e
	github.com/dop251/goja v0.0.0-20260721123636-c65cf2f023c8
	github.com/dop251/goja_nodejs v0.0.0-20260212111938-1f56ff5bcf14
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/fatih/structs v1.1.0
	github.com/go-cmd/cmd v1.4.3
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/influxdata/influxdb-client-go/v2 v2.14.0
	github.com/json-iterator/go v1.1.12
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c
	github.com/mycontroller-org/esphome_api v1.4.0
	github.com/nats-io/nats.go v1.52.0
	github.com/nleeper/goment v1.4.4
	github.com/olekukonko/tablewriter v0.0.5
	github.com/robfig/cron/v3 v3.0.2-0.20210106135023-bc59245fe10e
	github.com/rs/cors v1.11.1
	github.com/shirou/gopsutil/v4 v4.26.6
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/tidwall/sjson v1.2.5
	go.mongodb.org/mongo-driver v1.17.9
	go.uber.org/zap v1.28.0
	golang.org/x/crypto v0.54.0
	golang.org/x/term v0.45.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/braydonk/yaml v0.4.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2/v2 v2.5.2 // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/flynn/noise v1.0.1-0.20220214164934-d803f5c4b0f4 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/pprof v0.0.0-20260709232956-b9395ee17fa0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/influxdata/line-protocol v0.0.0-20210922203350-b1ad95c89adf // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/nats-io/nkeys v0.4.16 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oapi-codegen/runtime v1.6.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/gjson v1.19.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/tkuchiki/go-timezone v0.2.3 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

// for now gopkg.in/yaml.v3 does not support for UTF-16
// see the issue: https://github.com/go-yaml/yaml/issues/737 and PR: https://github.com/go-yaml/yaml/pull/738
// fix included in this fork: https://github.com/braydonk/yaml
replace gopkg.in/yaml.v3 => github.com/braydonk/yaml v0.4.1-0.20230115035319-29fa296a91d4
