module github.com/lyft/flyteadmin

go 1.13

require (
	github.com/NYTimes/gizmo v1.3.5
	github.com/Selvatico/go-mocket v1.0.7
	github.com/aws/aws-sdk-go v1.28.9
	github.com/benbjohnson/clock v1.0.0
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/securecookie v1.1.1
	github.com/graymeta/stow v0.2.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.2
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.3.0
	github.com/lyft/flyteidl v0.17.3
	github.com/lyft/flytepropeller v0.1.30
	github.com/lyft/flytestdlib v0.3.0
	github.com/magiconair/properties v1.8.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.15.1-0.20200117000758-b4cd77d6a56c // indirect
	google.golang.org/grpc v1.26.0
	gopkg.in/gormigrate.v1 v1.6.0
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200109141947-94aeca20bf09 // indirect
	sigs.k8s.io/controller-runtime v0.3.1-0.20191029211253-40070e2a1958
)

// Pin the version of client-go to something that's compatible with katrogan's fork of api and apimachinery
// Type the following
//   replace k8s.io/client-go => k8s.io/client-go kubernetes-1.16.2
// and it will be replaced with the 'sha' variant of the version

replace (
	k8s.io/api => github.com/lyft/api v0.0.0-20191031200350-b49a72c274e0
	k8s.io/apimachinery => github.com/lyft/apimachinery v0.0.0-20191031200210-047e3ea32d7f
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)
