package k8s

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestGetK8sClient(t *testing.T) {
	content := `
apiVersion: v1
clusters:
- cluster:
    server: https://localhost:8080
    extensions:
    - name: client.authentication.k8s.io/exec
      extension:
        audience: foo
        other: bar
  name: foo-cluster
contexts:
- context:
    cluster: foo-cluster
    user: foo-user
    namespace: bar
  name: foo-context
current-context: foo-context
kind: Config
users:
- name: foo-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      args:
      - arg-1
      - arg-2
      command: foo-command
      provideClusterInfo: true
`
	tmpfile, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile.Name())
	if err := ioutil.WriteFile(tmpfile.Name(), []byte(content), os.ModePerm); err != nil {
		t.Error(err)
	}
	t.Run("Create client from config", func(t *testing.T) {
		client := testclient.NewSimpleClientset()
		Client = client
		c, err := GetK8sClient(tmpfile.Name(), "https://localhost:8080")
		assert.Nil(t, err)
		assert.NotNil(t, c)
	})
	t.Run("Create client from config", func(t *testing.T) {
		Client = nil
		client, err := GetK8sClient(tmpfile.Name(), "https://localhost:8080")
		assert.Nil(t, err)
		assert.NotNil(t, client)
	})

}
