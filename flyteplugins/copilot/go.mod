module github.com/lyft/flyteplugins/copilot

go 1.13

require (
	github.com/aws/aws-sdk-go v1.29.23
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/lyft/flyteidl v0.18.11
	github.com/lyft/flyteplugins v0.4.4
	github.com/lyft/flytestdlib v0.3.9
	github.com/mitchellh/go-ps v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v11.0.1-0.20190918222721-c0e3722d5cf0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
)

replace (
	github.com/GoogleCloudPlatform/spark-on-k8s-operator => github.com/lyft/spark-on-k8s-operator v0.1.3
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	github.com/lyft/flyteplugins => ../
	k8s.io/api => github.com/lyft/api v0.0.0-20191031200350-b49a72c274e0
	k8s.io/apimachinery => github.com/lyft/apimachinery v0.0.0-20191031200210-047e3ea32d7f
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)
