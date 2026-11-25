package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const (
	separator = "/"
)

// URLPathConstructor implements ReferenceConstructor that assumes paths are URL-compatible.
type URLPathConstructor struct {
}

func ensureEndingPathSeparator(path DataReference) DataReference {
	if len(path) > 0 && path[len(path)-1] == separator[0] {
		return path
	}

	return path + separator
}

func (URLPathConstructor) ConstructReference(ctx context.Context, reference DataReference, nestedKeys ...string) (DataReference, error) {
	u, err := url.Parse(string(ensureEndingPathSeparator(reference)))
	if err != nil {
		logger.Errorf(ctx, "Failed to parse prefix: %v", reference)
		return "", errors.Wrap(err, fmt.Sprintf("Reference is of an invalid format [%v]", reference))
	}

	// The nested keys are object-key segments, not a URL: keep characters such as '#', '?' and '%' in the path.
	rel := &url.URL{Path: strings.Join(MapStrings(func(s string) string {
		return strings.Trim(s, separator)
	}, nestedKeys...), separator)}

	u = u.ResolveReference(rel)

	return DataReference(u.String()), nil
}

func NewURLPathConstructor() URLPathConstructor {
	return URLPathConstructor{}
}
