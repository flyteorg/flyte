package backoff

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func SyncResourceListFromResourceList(list v1.ResourceList) *SyncResourceList {
	s := NewSyncResourceList()
	s.AddResourceList(list)
	return s
}

func TestSyncResourceList_AddResourceList(t *testing.T) {
	s := NewSyncResourceList()
	s.Store(v1.ResourceCPU, resource.MustParse("1"))
	expectedQuantity := resource.MustParse("2")
	s.AddResourceList(v1.ResourceList{
		v1.ResourceCPU: expectedQuantity,
	})

	actualQuantity, found := s.Load(v1.ResourceCPU)
	assert.True(t, found)
	assert.Equal(t, expectedQuantity, actualQuantity)
}

func TestSyncResourceList_AsResourceList(t *testing.T) {
	t.Run("With items", func(t *testing.T) {
		s := NewSyncResourceList()
		expectedList := v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("100Mi"),
		}

		s.Store(v1.ResourceCPU, expectedList[v1.ResourceCPU])
		s.Store(v1.ResourceMemory, expectedList[v1.ResourceMemory])

		assert.Equal(t, expectedList, s.AsResourceList())
	})

	t.Run("Empty", func(t *testing.T) {
		s := NewSyncResourceList()
		assert.Equal(t, v1.ResourceList{}, s.AsResourceList())
	})
}

func BenchmarkSyncResourceList_Load(b *testing.B) {
	s := NewSyncResourceList()
	for i := 0; i < b.N; i++ {
		s.Store(v1.ResourceName(fmt.Sprintf("Resource-%v", i)), resource.MustParse(fmt.Sprintf("%v", i)))
	}

	for i := 0; i < b.N; i++ {
		actualQuantity, found := s.Load(v1.ResourceName(fmt.Sprintf("Resource-%v", i)))
		assert.True(b, found)
		assert.Equal(b, resource.MustParse(fmt.Sprintf("%v", i)), actualQuantity)
	}
}

func TestSyncResourceList_Range(t *testing.T) {
	s := NewSyncResourceList()
	for i := 0; i < 100; i++ {
		resourceValue := fmt.Sprintf("%v", i)
		s.Store(v1.ResourceName(resourceValue), resource.MustParse(resourceValue))
	}

	s.Range(func(key v1.ResourceName, value resource.Quantity) bool {
		return assert.Equal(t, resource.MustParse(string(key)), value)
	})
}

func TestSyncResourceList_String(t *testing.T) {
	s := NewSyncResourceList()
	for i := 0; i < 100; i++ {
		resourceValue := fmt.Sprintf("%v", i)
		s.Store(v1.ResourceName(resourceValue), resource.MustParse(resourceValue))
	}

	assert.Equal(t, "0:0, 1:1, 10:10, 11:11, 12:12, 13:13, 14:14, 15:15, 16:16, 17:17, 18:18, 19:19, 2:2, 20:20, 21:21, 22:22, 23:23, 24:24, 25:25, 26:26, 27:27, 28:28, 29:29, 3:3, 30:30, 31:31, 32:32, 33:33, 34:34, 35:35, 36:36, 37:37, 38:38, 39:39, 4:4, 40:40, 41:41, 42:42, 43:43, 44:44, 45:45, 46:46, 47:47, 48:48, 49:49, 5:5, 50:50, 51:51, 52:52, 53:53, 54:54, 55:55, 56:56, 57:57, 58:58, 59:59, 6:6, 60:60, 61:61, 62:62, 63:63, 64:64, 65:65, 66:66, 67:67, 68:68, 69:69, 7:7, 70:70, 71:71, 72:72, 73:73, 74:74, 75:75, 76:76, 77:77, 78:78, 79:79, 8:8, 80:80, 81:81, 82:82, 83:83, 84:84, 85:85, 86:86, 87:87, 88:88, 89:89, 9:9, 90:90, 91:91, 92:92, 93:93, 94:94, 95:95, 96:96, 97:97, 98:98, 99:99, ", s.String())
}
