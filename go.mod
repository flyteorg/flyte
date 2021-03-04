module github.com/flyteorg/flytecopilot

go 1.13

require (
	github.com/aws/aws-sdk-go v1.37.1
	github.com/flyteorg/flyteidl v0.18.14
	github.com/flyteorg/flytestdlib v0.3.12
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/protobuf v1.4.3
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/lyft/flyteidl v0.18.14 // indirect
	github.com/mitchellh/go-ps v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.1-0.20190918222721-c0e3722d5cf0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
)

replace (
	github.com/flyteorg/flyteidl => github.com/lyft/flyteidl v0.18.14
	github.com/flyteorg/flytestdlib => github.com/lyft/flytestdlib v0.3.12
)
