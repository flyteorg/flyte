module github.com/flyteorg/flytepropeller

go 1.16

require (
	github.com/DiSiqueira/GoTree v1.0.1-0.20180907134536-53a8e837f295
	github.com/benlaurie/objecthash v0.0.0-20180202135721-d1e3d6079fc1
	github.com/fatih/color v1.10.0
	github.com/flyteorg/flyteidl v0.19.14
	github.com/flyteorg/flyteplugins v0.5.59
	github.com/flyteorg/flytestdlib v0.3.27
	github.com/ghodss/yaml v1.0.0
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/go-test/deep v1.0.7
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/magiconair/properties v1.8.4
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.36.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.8.2
)

replace github.com/aws/amazon-sagemaker-operator-for-k8s => github.com/aws/amazon-sagemaker-operator-for-k8s v1.0.1-0.20210303003444-0fb33b1fd49d
