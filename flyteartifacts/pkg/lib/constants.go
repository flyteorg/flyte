package lib

// ArtifactKey - This string is used to identify Artifacts when all you have
// is the underlying Literal. Look for this key under the literal's metadata field. This situation can arise
// when a user fetches an artifact, using something like flyte remote or flyte console, and then kicks
// off an execution using that literal.
const ArtifactKey = "_ua"
