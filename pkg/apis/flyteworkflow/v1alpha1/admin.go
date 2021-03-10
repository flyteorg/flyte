package v1alpha1

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

// This contains an OutputLocationPrefix. When running against AWS, this should be something of the form
// s3://my-bucket, or s3://my-bucket/  A sharding string will automatically be appended to this prefix before
// handing off to plugins/tasks. Sharding behavior may change in the future.
// Background available at https://github.com/flyteorg/flyte/issues/211
type RawOutputDataConfig struct {
	*admin.RawOutputDataConfig
}

func (in *RawOutputDataConfig) DeepCopyInto(out *RawOutputDataConfig) {
	*out = *in
}
