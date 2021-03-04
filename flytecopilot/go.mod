module github.com/lyft/flyteplugins/copilot

go 1.13

require (
	github.com/aws/aws-sdk-go v1.37.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.4.3
	github.com/lyft/flyteidl v0.18.12
	github.com/lyft/flyteplugins v0.4.4
	github.com/lyft/flytestdlib v0.3.9
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
)

replace (
	github.com/GoogleCloudPlatform/spark-on-k8s-operator => github.com/lyft/spark-on-k8s-operator v0.1.3
	github.com/lyft/flyteidl => ../../flyteidl
	github.com/lyft/flyteplugins => ../
	github.com/lyft/flytestdlib => ../../flytestdlib
)
