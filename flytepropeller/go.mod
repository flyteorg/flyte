module github.com/lyft/flytepropeller

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v40.3.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.10.0 // indirect
	github.com/DiSiqueira/GoTree v1.0.1-0.20180907134536-53a8e837f295
	github.com/benlaurie/objecthash v0.0.0-20180202135721-d1e3d6079fc1
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/fatih/color v1.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0
	github.com/graymeta/stow v0.2.5 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.3 // indirect
	github.com/lyft/datacatalog v0.2.1
	github.com/lyft/flyteidl v0.18.11
	github.com/lyft/flyteplugins v0.5.28
	github.com/lyft/flytestdlib v0.3.9
	github.com/magiconair/properties v1.8.1
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/ncw/swift v1.0.50 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.35.0
	gopkg.in/ini.v1 v1.54.0 // indirect
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v11.0.1-0.20190918222721-c0e3722d5cf0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.8.2
)

// Pin the version of client-go to something that's compatible with katrogan's fork of api and apimachinery
// Type the following
//   replace k8s.io/client-go => k8s.io/client-go kubernetes-1.16.2
// and it will be replaced with the 'sha' variant of the version

replace (
	github.com/GoogleCloudPlatform/spark-on-k8s-operator => github.com/lyft/spark-on-k8s-operator v0.1.4-0.20201027003055-c76b67e3b6d0
	github.com/aws/amazon-sagemaker-operator-for-k8s => github.com/aws/amazon-sagemaker-operator-for-k8s v1.0.1-0.20210303003444-0fb33b1fd49d
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	gopkg.in/fsnotify.v1 => github.com/fsnotify/fsnotify v1.4.7
	k8s.io/api => github.com/lyft/api v0.0.0-20191031200350-b49a72c274e0
	k8s.io/apimachinery => github.com/lyft/apimachinery v0.0.0-20191031200210-047e3ea32d7f
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.1
)
