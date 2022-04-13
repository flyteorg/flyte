module github.com/flyteorg/flyteplugins

go 1.16

require (
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20200723154620-6f35a1152625
	github.com/Masterminds/semver v1.5.0
	github.com/adammck/venv v0.0.0-20200610172036-e77789703e7c // indirect
	github.com/aws/amazon-sagemaker-operator-for-k8s v1.0.1-0.20210303003444-0fb33b1fd49d
	github.com/aws/aws-sdk-go v1.43.37
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.0.0
	github.com/aws/aws-sdk-go-v2/service/athena v1.0.0
	github.com/coocood/freecache v1.1.1
	github.com/flyteorg/flyteidl v0.24.19
	github.com/flyteorg/flytestdlib v0.4.24-0.20220413001941-d1364430a962
	github.com/go-test/deep v1.0.7
	github.com/golang/protobuf v1.5.2
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.12
	github.com/kubeflow/common v0.4.0
	github.com/kubeflow/mpi-operator/v2 v2.0.0-20220406191845-993b010e05c4
	github.com/kubeflow/pytorch-operator v0.6.0
	github.com/kubeflow/tf-operator v0.5.3
	github.com/magiconair/properties v1.8.5
	github.com/mitchellh/mapstructure v1.4.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/api v0.63.0
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	sigs.k8s.io/controller-runtime v0.11.2
)

replace github.com/aws/amazon-sagemaker-operator-for-k8s => github.com/aws/amazon-sagemaker-operator-for-k8s v1.0.1-0.20210303003444-0fb33b1fd49d
