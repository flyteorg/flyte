package util

import (
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func ShouldFetchData(config *interfaces.RemoteDataConfig, urlBlob admin.UrlBlob) bool {
	return config.Scheme == common.Local || config.Scheme == common.None || urlBlob.Bytes < config.MaxSizeInBytes
}

func ShouldFetchOutputData(config *interfaces.RemoteDataConfig, urlBlob admin.UrlBlob, outputURI string) bool {
	return ShouldFetchData(config, urlBlob) && len(outputURI) > 0
}
