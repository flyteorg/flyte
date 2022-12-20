package transformers

// OutputsObjectSuffix is used when execution event data includes inline outputs but the admin deployment is configured
// to offload such data. The generated file path for the offloaded data will include the execution identifier and this suffix.
const OutputsObjectSuffix = "offloaded_outputs"
