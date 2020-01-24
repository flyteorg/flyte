module github.com/lyft/flytestdlib

go 1.13

require (
	cloud.google.com/go v0.52.0 // indirect
	github.com/Azure/azure-sdk-for-go v38.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.9.4 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.1 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/aws/aws-sdk-go v1.28.9
	github.com/benlaurie/objecthash v0.0.0-20180202135721-d1e3d6079fc1
	github.com/coocood/freecache v1.1.0
	github.com/dnaeon/go-vcr v1.0.1 // indirect
	github.com/ernesto-jimenez/gogen v0.0.0-20180125220232-d7d4131e6607
	github.com/fatih/color v1.9.0
	github.com/fatih/structtag v1.2.0
	github.com/fsnotify/fsnotify v1.4.8-0.20191012010759-4bf2d1fec783
	github.com/ghodss/yaml v1.0.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.2
	github.com/graymeta/stow v0.2.4
	github.com/hashicorp/golang-lru v0.5.4
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/magiconair/properties v1.8.1
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/common v0.9.1 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/sys v0.0.0-20200122134326-e047566fdf82 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20200124170513-3f4d10fc73b4
	google.golang.org/genproto v0.0.0-20200122232147-0452cf42e150 // indirect
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.0+incompatible
)

// Pin the version of client-go to something that's compatible with katrogan's fork of api and apimachinery
// Type the following
//   replace k8s.io/client-go => k8s.io/client-go kubernetes-1.16.2
// and it will be replaced with the 'sha' variant of the version
replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
