module github.com/lyft/flyteplugins

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v39.0.0+incompatible // indirect
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.1.3
	github.com/aws/aws-sdk-go v1.28.11
	github.com/coocood/freecache v1.1.0
	github.com/go-test/deep v1.0.5
	github.com/golang/protobuf v1.3.3
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/lyft/flyteidl v0.17.6
	github.com/lyft/flytestdlib v0.3.2
	github.com/magiconair/properties v1.8.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/crypto v0.0.0-20200204104054-c9f3fb736b72 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	google.golang.org/api v0.16.0 // indirect
	google.golang.org/genproto v0.0.0-20200205142000-a86caf926a67 // indirect
	google.golang.org/grpc v1.27.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.2.0 // indirect
)

// Pin the version of client-go to something that's compatible with katrogan's fork of api and apimachinery
// Type the following
//   replace k8s.io/client-go => k8s.io/client-go kubernetes-1.16.2
// and it will be replaced with the 'sha' variant of the version

replace (
	github.com/GoogleCloudPlatform/spark-on-k8s-operator => github.com/lyft/spark-on-k8s-operator v0.1.3
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	k8s.io/api => github.com/lyft/api v0.0.0-20191031200350-b49a72c274e0
	k8s.io/apimachinery => github.com/lyft/apimachinery v0.0.0-20191031200210-047e3ea32d7f
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)
