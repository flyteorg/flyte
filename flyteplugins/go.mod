module github.com/flyteorg/flyteplugins

go 1.18

require (
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20200723154620-6f35a1152625
	github.com/Masterminds/semver v1.5.0
	github.com/adammck/venv v0.0.0-20200610172036-e77789703e7c // indirect
	github.com/aws/amazon-sagemaker-operator-for-k8s v1.2.1
	github.com/aws/aws-sdk-go v1.43.37
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.0.0
	github.com/aws/aws-sdk-go-v2/service/athena v1.0.0
	github.com/coocood/freecache v1.1.1
	github.com/flyteorg/flyteidl v1.1.8-0.20220628045429-5aaea4ba607a
	github.com/flyteorg/flytestdlib v1.0.5-0.20220628044039-2918e6121725
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/go-test/deep v1.0.7
	github.com/golang/protobuf v1.5.2
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.11
	github.com/kubeflow/common v0.4.0
	github.com/kubeflow/mpi-operator/v2 v2.0.0-20220406191845-993b010e05c4
	github.com/kubeflow/pytorch-operator v0.6.0
	github.com/kubeflow/tf-operator v0.5.3
	github.com/magiconair/properties v1.8.6
	github.com/mitchellh/mapstructure v1.4.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.1
	golang.org/x/net v0.0.0-20220412020605-290c469a71a5
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5
	google.golang.org/api v0.74.0
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.28.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.2
)

require (
	cloud.google.com/go v0.100.2 // indirect
	cloud.google.com/go/compute v1.5.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/storage v1.14.0 // indirect
	github.com/Azure/azure-sdk-for-go v62.3.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.21.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v0.8.3 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v0.3.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.25 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.18 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.0.0 // indirect
	github.com/aws/smithy-go v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/flyteorg/stow v0.3.4 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/spec v0.20.3 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/googleapis/gax-go/v2 v2.3.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/ncw/swift v1.0.53 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pelletier/go-toml/v2 v2.0.0-beta.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/cobra v1.4.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.11.0 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/xerrors v0.0.0-20220411194840-2f41105eb62f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220407144326-9054f6ed7bac // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/klog/v2 v2.5.0 // indirect
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace github.com/aws/amazon-sagemaker-operator-for-k8s => github.com/aws/amazon-sagemaker-operator-for-k8s v1.2.1
