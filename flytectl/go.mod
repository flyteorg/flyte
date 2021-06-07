module github.com/flyteorg/flytectl

go 1.13

require (
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/flyteorg/flyteidl v0.18.51
	github.com/flyteorg/flytestdlib v0.3.21
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/kataras/tablewriter v0.0.0-20180708051242-e063d29b7c23
	github.com/kr/text v0.2.0 // indirect
	github.com/landoop/tableprinter v0.0.0-20180806200924-8bd8c2576d27
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/sirupsen/logrus v1.8.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0
	github.com/zalando/go-keyring v0.1.1
	golang.org/x/oauth2 v0.0.0-20210126194326-f9ce19ea3013
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/flyteorg/flyteidl => github.com/flyteorg/flyteidl v0.18.51-0.20210602050605-9ebebd25056e
