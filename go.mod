module github.com/flyteorg/flyte/v2

go 1.24.6

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250912141014-52f32327d4b0.1
	connectrpc.com/connect v1.19.1
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20200723154620-6f35a1152625
	github.com/aws/aws-sdk-go v1.47.11
	github.com/aws/aws-sdk-go-v2 v1.24.1
	github.com/aws/aws-sdk-go-v2/config v1.26.1
	github.com/benlaurie/objecthash v0.0.0-20180202135721-d1e3d6079fc1
	github.com/coocood/freecache v1.2.3
	github.com/dask/dask-kubernetes/v2023 v2023.0.0-20230626103304-abd02cd17b26
	github.com/eko/gocache/lib/v4 v4.2.0
	github.com/eko/gocache/store/freecache/v4 v4.2.2
	github.com/eko/gocache/store/redis/v4 v4.2.2
	github.com/ernesto-jimenez/gogen v0.0.0-20180125220232-d7d4131e6607
	github.com/fatih/color v1.13.0
	github.com/fatih/structtag v1.2.0
	github.com/flyteorg/stow v0.3.11
	github.com/fsnotify/fsnotify v1.7.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-gormigrate/gormigrate/v2 v2.1.5
	github.com/go-logr/logr v1.4.3
	github.com/go-test/deep v1.1.1
	github.com/golang/protobuf v1.5.4
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.1.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.13
	github.com/jackc/pgconn v1.14.3
	github.com/jackc/pgx/v5 v5.7.6
	github.com/magiconair/properties v1.8.6
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.22.0
	github.com/prometheus/common v0.62.0
	github.com/ray-project/kuberay/ray-operator v1.1.0-rc.1
	github.com/redis/go-redis/extra/redisprometheus/v9 v9.12.1
	github.com/redis/go-redis/v9 v9.12.1
	github.com/samber/lo v1.39.0
	github.com/shamaton/msgpack/v2 v2.4.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.11.0
	github.com/stretchr/testify v1.11.1
	github.com/vmihailenco/msgpack/v5 v5.4.1
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.24.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.22.0
	go.opentelemetry.io/otel/sdk v1.37.0
	go.opentelemetry.io/otel/trace v1.37.0
	golang.org/x/oauth2 v0.30.0
	golang.org/x/time v0.5.0
	golang.org/x/tools v0.37.0
	google.golang.org/api v0.160.0
	google.golang.org/genproto/googleapis/api v0.0.0-20251103181224-f26f9409b101
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.29.3
	k8s.io/apimachinery v0.29.3
	k8s.io/client-go v0.29.3
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	sigs.k8s.io/controller-runtime v0.17.2
)

require (
	cloud.google.com/go v0.112.0 // indirect
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	cloud.google.com/go/iam v1.1.5 // indirect
	cloud.google.com/go/storage v1.36.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.13.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.7.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.4.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.12 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.5 // indirect
	github.com/aws/smithy-go v1.19.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/emicklei/go-restful/v3 v3.12.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.8.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
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
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncw/swift v1.0.53 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pelletier/go-toml/v2 v2.0.0-beta.8 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.47.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/genproto v0.0.0-20240123012728-ef4313101c80 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.29.0 // indirect
	k8s.io/component-base v0.29.0 // indirect
	k8s.io/klog/v2 v2.110.1 // indirect
	k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace (
	github.com/flyteorg/flyte/flyteidl => ./flyteidl2
	github.com/flyteorg/flyte/flyteplugins => ./flyteplugins
	github.com/flyteorg/flyte/flytestdlib => ./flytestdlib
	github.com/flyteorg/flyte/v2 => ./
	github.com/flyteorg/flyte/v2/flyteplugins => ./flyteplugins
	github.com/flyteorg/flyte/v2/flytestdlib => ./flytestdlib
	github.com/google/gnostic-models => github.com/google/gnostic-models v0.6.8
	github.com/robfig/cron/v3 => github.com/unionai/cron/v3 v3.0.2-0.20220915080349-5790c370e63a
	k8s.io/api => k8s.io/api v0.28.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.28.2
	k8s.io/client-go => k8s.io/client-go v0.28.2
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.60.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20230905202853-d090da108d2f
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.16.3
)
