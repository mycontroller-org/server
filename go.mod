module github.com/mycontroller-org/server/v2

go 1.20

require (
	github.com/Masterminds/semver v1.5.0
	github.com/NYTimes/gziphandler v1.1.1
	github.com/amimof/huego v1.2.1
	github.com/btittelbach/astrotime v0.0.0-20160515101311-7ddba43aa26e
	github.com/dop251/goja v0.0.0-20230216180835-5937a312edda
	github.com/dop251/goja_nodejs v0.0.0-20230226152057-060fa99b809f
	github.com/eclipse/paho.mqtt.golang v1.4.2
	github.com/fatih/structs v1.1.0
	github.com/go-cmd/cmd v1.4.1
	github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/influxdata/influxdb-client-go/v2 v2.12.2
	github.com/jaegertracing/jaeger v1.47.0
	github.com/json-iterator/go v1.1.12
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4
	github.com/mycontroller-org/esphome_api v1.3.0
	github.com/nats-io/nats.go v1.28.0
	github.com/nleeper/goment v1.4.4
	github.com/olekukonko/tablewriter v0.0.5
	github.com/robfig/cron/v3 v3.0.2-0.20210106135023-bc59245fe10e
	github.com/rs/cors v1.9.0
	github.com/shirou/gopsutil/v3 v3.23.1
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.16.0
	github.com/stretchr/testify v1.8.4
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/tidwall/sjson v1.2.5
	go.mongodb.org/mongo-driver v1.11.6
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.14.0
	golang.org/x/term v0.13.0
	google.golang.org/protobuf v1.31.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/braydonk/yaml v0.4.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.8.2 // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/flynn/noise v1.0.1-0.20220214164934-d803f5c4b0f4 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/nats-io/nats-server/v2 v2.9.23 // indirect
	github.com/nats-io/nkeys v0.4.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tidwall/gjson v1.14.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/tkuchiki/go-timezone v0.2.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

// for now gopkg.in/yaml.v3 does not support for UTF-16
// see the issue: https://github.com/go-yaml/yaml/issues/737 and PR: https://github.com/go-yaml/yaml/pull/738
// fix included in this fork: https://github.com/braydonk/yaml
replace gopkg.in/yaml.v3 => github.com/braydonk/yaml v0.4.1-0.20230115035319-29fa296a91d4
