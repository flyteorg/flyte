module github.com/flyteorg/flyteplugins

go 1.16

require (
	cloud.google.com/go v0.78.0 // indirect
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20200723154620-6f35a1152625
	github.com/Masterminds/semver v1.5.0
	github.com/adammck/venv v0.0.0-20200610172036-e77789703e7c // indirect
	github.com/aws/amazon-sagemaker-operator-for-k8s v1.0.1-0.20210303003444-0fb33b1fd49d
	github.com/aws/aws-sdk-go v1.37.3
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.0.0
	github.com/aws/aws-sdk-go-v2/service/athena v1.0.0
	github.com/coocood/freecache v1.1.1
	github.com/flyteorg/flyteidl v0.19.2
	github.com/flyteorg/flytestdlib v0.3.33
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/go-test/deep v1.0.7
	github.com/golang/protobuf v1.4.3
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kubeflow/pytorch-operator v0.6.0
	github.com/kubeflow/tf-operator v0.5.3
	github.com/magiconair/properties v1.8.4
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.18.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93
	golang.org/x/sys v0.0.0-20210303074136-134d130e1a04 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/api v0.40.0
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.2
)

replace github.com/aws/amazon-sagemaker-operator-for-k8s => github.com/aws/amazon-sagemaker-operator-for-k8s v1.0.1-0.20210303003444-0fb33b1fd49d
