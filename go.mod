module github.com/flyteorg/flyteadmin

go 1.16

require (
	cloud.google.com/go v0.78.0
	cloud.google.com/go/storage v1.12.0
	github.com/NYTimes/gizmo v1.3.5
	github.com/Selvatico/go-mocket v1.0.7
	github.com/aws/aws-sdk-go v1.37.3
	github.com/benbjohnson/clock v1.0.0
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/flyteorg/flyteidl v0.18.15
	github.com/flyteorg/flytepropeller v0.7.0
	github.com/flyteorg/flytestdlib v0.3.14
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.3
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/securecookie v1.1.1
	github.com/graymeta/stow v0.2.7
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.3
	github.com/jinzhu/gorm v1.9.12
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.3.0
	github.com/magiconair/properties v1.8.4
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93
	google.golang.org/api v0.40.0
	google.golang.org/genproto v0.0.0-20210222152913-aa3ee6e6a81c
	google.golang.org/grpc v1.36.0
	gopkg.in/gormigrate.v1 v1.6.0
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.2
)

replace github.com/GoogleCloudPlatform/spark-on-k8s-operator => github.com/lyft/spark-on-k8s-operator v0.1.3
