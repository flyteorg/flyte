package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestQuantityScaled(t *testing.T) {
	t.Run("1Mi", func(t *testing.T) {
		quantity := resource.NewScaledQuantity(1, resource.Mega)
		require.Equal(t, int64(1000000), quantity.Value())
	})

	t.Run("10Gi", func(t *testing.T) {
		quantity := resource.NewScaledQuantity(10, resource.Giga)
		require.Equal(t, int64(10000000000), quantity.Value())
	})
}
