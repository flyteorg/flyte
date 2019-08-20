package flytek8s

import "time"

const DefaultInformerResyncDuration = 30 * time.Second
const Kilobytes = 1024 * 1
const Megabytes = 1024 * Kilobytes
const Gigabytes = 1024 * Megabytes
const MaxMetadataPayloadSizeBytes = 10 * Megabytes

const finalizer = "flyte/flytek8s"
