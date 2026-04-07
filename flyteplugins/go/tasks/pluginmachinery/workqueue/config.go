package workqueue

// Config for the queue
type Config struct {
	Workers            int `json:"workers" pflag:",Number of concurrent workers to start processing the queue."`
	MaxRetries         int `json:"maxRetries" pflag:",Maximum number of retries per item."`
	IndexCacheMaxItems int `json:"maxItems" pflag:",Maximum number of entries to keep in the index."`
}
