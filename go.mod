module github.com/lyft/flyteplugins

go 1.13

require (
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.1.3
	github.com/aws/aws-sdk-go v1.28.9
	github.com/coocood/freecache v1.1.0
	github.com/go-test/deep v1.0.5
	github.com/golang/protobuf v1.3.2
	github.com/hashicorp/golang-lru v0.5.4
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/lyft/flyteidl v0.17.0
	github.com/lyft/flytestdlib v0.3.0
	github.com/magiconair/properties v1.8.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa
	google.golang.org/grpc v1.26.0
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20200122174043-1e243dd1a584 // indirect
	sigs.k8s.io/controller-runtime v0.4.0
)

// Pin the version of client-go to something that's compatible with katrogan's fork of api and apimachinery
// Type the following
//   replace k8s.io/client-go => k8s.io/client-go kubernetes-1.16.2
// and it will be replaced with the 'sha' variant of the version

replace (
	github.com/GoogleCloudPlatform/spark-on-k8s-operator => github.com/lyft/spark-on-k8s-operator v0.1.3
	k8s.io/api => github.com/lyft/api v0.0.0-20191031200350-b49a72c274e0
	k8s.io/apimachinery => github.com/lyft/apimachinery v0.0.0-20191031200210-047e3ea32d7f
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)
