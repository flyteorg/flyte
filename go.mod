module github.com/flyteorg/flyte/v2

go 1.24.6

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250912141014-52f32327d4b0.1
	cloud.google.com/go/secretmanager v1.16.0
	connectrpc.com/connect v1.19.1
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.10.1
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets v1.4.0
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20200723154620-6f35a1152625
	github.com/aws/aws-sdk-go v1.55.8
	github.com/aws/aws-sdk-go-v2 v1.41.1
	github.com/aws/aws-sdk-go-v2/config v1.26.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.41.1
	github.com/benlaurie/objecthash v0.0.0-20180202135721-d1e3d6079fc1
	github.com/coocood/freecache v1.2.4
	github.com/dask/dask-kubernetes/v2023 v2023.0.0-20230626103304-abd02cd17b26
	github.com/eko/gocache/lib/v4 v4.2.0
	github.com/eko/gocache/store/freecache/v4 v4.2.0
	github.com/eko/gocache/store/redis/v4 v4.2.0
	github.com/fatih/color v1.13.0
	github.com/flyteorg/stow v0.3.12
	github.com/fsnotify/fsnotify v1.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-gormigrate/gormigrate/v2 v2.1.5
	github.com/go-test/deep v1.1.1
	github.com/golang/protobuf v1.5.4
	github.com/googleapis/gax-go/v2 v2.15.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.1.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.16
	github.com/jackc/pgconn v1.14.3
	github.com/jackc/pgx/v5 v5.7.6
	github.com/kubeflow/training-operator v1.9.3
	github.com/lib/pq v1.10.9
	github.com/magiconair/properties v1.8.6
	github.com/mitchellh/mapstructure v1.5.0
	github.com/onsi/ginkgo/v2 v2.22.0
	github.com/onsi/gomega v1.36.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.23.2
	github.com/prometheus/common v0.67.4
	github.com/ray-project/kuberay/ray-operator v1.1.0-rc.1
	github.com/redis/go-redis/extra/redisprometheus/v9 v9.17.2
	github.com/redis/go-redis/v9 v9.17.2
	github.com/samber/lo v1.39.0
	github.com/shamaton/msgpack/v2 v2.4.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/spf13/viper v1.11.0
	github.com/stretchr/testify v1.11.1
	github.com/vmihailenco/msgpack/v5 v5.4.1
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.37.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.37.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.37.0
	go.opentelemetry.io/otel/sdk v1.37.0
	go.opentelemetry.io/otel/trace v1.37.0
	golang.org/x/exp v0.0.0-20251209150349-8475f28825e9
	golang.org/x/net v0.48.0
	golang.org/x/oauth2 v0.32.0
	golang.org/x/time v0.12.0
	google.golang.org/api v0.247.0
	google.golang.org/genproto/googleapis/api v0.0.0-20251103181224-f26f9409b101
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/datatypes v1.2.7
	gorm.io/driver/postgres v1.6.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.34.1
	k8s.io/apimachinery v0.34.1
	k8s.io/client-go v0.34.1
	k8s.io/klog/v2 v2.130.1
	k8s.io/utils v0.0.0-20251002143259-bc988d571ff4
	sigs.k8s.io/controller-runtime v0.22.4
)

require (
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.121.6 // indirect
	cloud.google.com/go/auth v0.16.4 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.8.0 // indirect
	cloud.google.com/go/iam v1.5.3 // indirect
	cloud.google.com/go/monitoring v1.24.3 // indirect
	cloud.google.com/go/storage v1.56.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.4.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.4.2 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.29.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.53.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.53.0 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230512164433-5d1fd1a340c9 // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.12 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.5 // indirect
	github.com/aws/smithy-go v1.24.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cel-go v0.26.0 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20241029153458-d1b30febd7db // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncw/swift v1.0.53 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pelletier/go-toml/v2 v2.0.0-beta.8 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.9 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.9 // indirect
	go.etcd.io/etcd/client/v3 v3.5.9 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.36.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.37.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.1 // indirect
	go.uber.org/mock v0.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/term v0.38.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/genproto v0.0.0-20251103181224-f26f9409b101 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	k8s.io/apiextensions-apiserver v0.34.1 // indirect
	k8s.io/apiserver v0.34.1 // indirect
	k8s.io/component-base v0.34.1 // indirect
	k8s.io/kms v0.28.3 // indirect
	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.31.2 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace (
	github.com/flyteorg/flyte/flyteidl => ./flyteidl2
	github.com/flyteorg/flyte/flyteplugins => ./flyteplugins
	github.com/flyteorg/flyte/flytestdlib => ./flytestdlib
	github.com/flyteorg/flyte/v2 => ./
	github.com/flyteorg/flyte/v2/flyteplugins => ./flyteplugins
	github.com/flyteorg/flyte/v2/flytestdlib => ./flytestdlib
	github.com/google/cel-go => github.com/google/cel-go v0.16.1
	github.com/google/gnostic-models => github.com/google/gnostic-models v0.6.8
	github.com/robfig/cron/v3 => github.com/unionai/cron/v3 v3.0.2-0.20220915080349-5790c370e63a
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.54.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.54.0
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace => go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.37.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.37.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.37.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace => go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.37.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.37.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.37.0
	go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.37.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.37.0
	k8s.io/api => k8s.io/api v0.28.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.28.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.28.2
	k8s.io/apiserver => k8s.io/apiserver v0.28.2
	k8s.io/client-go => k8s.io/client-go v0.28.2
	k8s.io/component-base => k8s.io/component-base v0.28.2
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.100.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20230905202853-d090da108d2f
	k8s.io/utils => k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.16.3
)
