package resolver

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"testing"
	"time"

	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func parseTarget(target string) resolver.Target {
	u, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	return resolver.Target{
		URL: *u,
	}
}

type fakeConn struct {
	cmp   chan struct{}
	found []string
}

func (fc *fakeConn) UpdateState(state resolver.State) error {
	for _, a := range state.Addresses {
		fc.found = append(fc.found, a.Addr)
	}
	fc.cmp <- struct{}{}
	return nil
}

func (fc *fakeConn) ReportError(e error) {
	log.Println(e)
}

func (fc *fakeConn) ParseServiceConfig(_ string) *serviceconfig.ParseResult {
	return nil
}

func (fc *fakeConn) NewAddress(_ []resolver.Address) {}

func (*fakeConn) NewServiceConfig(serviceConfig string) {
	fmt.Printf("serviceConfig: %s\n", serviceConfig)
}

func TestBuilder(t *testing.T) {
	k8sClient := testclient.NewSimpleClientset()
	builder := NewBuilder(k8sClient, KubernetesSchema)
	fc := &fakeConn{
		cmp: make(chan struct{}),
	}
	k8sResolver, err := builder.Build(parseTarget("kubernetes://flyteagent.flyte.svc.cluster.local:8000"), fc, resolver.BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Make sure watcher is started before we create the endpoint
	time.Sleep(5 * time.Second)

	_, err = k8sClient.CoreV1().Endpoints("flyte").Create(context.Background(), &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flyteagent",
			Namespace: "flyte",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "10.0.0.1",
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name: "grpc",
						Port: 8000,
					},
				},
			},
		},
	}, metav1.CreateOptions{})
	assert.NilError(t, err)
	assert.Equal(t, len(fc.found), 1)
	assert.Equal(t, fc.found[0], "10.0.0.1:8000")

	<-fc.cmp
	k8sResolver.Close()
	close(fc.cmp)
}

func TestWatcher(t *testing.T) {
	k8sClient := testclient.NewSimpleClientset()
	go func() {
		for i := 0; i < 10; i++ {
			_, err := k8sClient.CoreV1().Endpoints("flyte").Create(context.Background(), &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "flyteagent",
					Namespace: "flyte",
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{
								IP: "10.0.0.1",
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name: "grpc",
								Port: 8000,
							},
						},
					},
				},
			}, metav1.CreateOptions{})
			assert.NilError(t, err)
		}
	}()
}

func TestParseResolverTargets(t *testing.T) {
	for i, test := range []struct {
		target string
		want   targetInfo
		err    bool
	}{
		{"", targetInfo{}, true},
		{"kubernetes:///", targetInfo{}, true},
		{"kubernetes://a:30", targetInfo{"a", "", "30"}, false},
		{"kubernetes://a/", targetInfo{"a", "", ""}, false},
		{"kubernetes:///a", targetInfo{"a", "", ""}, false},
		{"kubernetes://a/b", targetInfo{"b", "a", ""}, false},
		{"kubernetes://a.b/", targetInfo{"a", "b", ""}, false},
		{"kubernetes:///a.b:80", targetInfo{"a", "b", "80"}, false},
		{"kubernetes:///a.b:port", targetInfo{"a", "b", "port"}, false},
		{"kubernetes:///a:port", targetInfo{"a", "", "port"}, false},
		{"kubernetes://x/a:port", targetInfo{"a", "x", "port"}, false},
		{"kubernetes://a.x:30/", targetInfo{"a", "x", "30"}, false},
		{"kubernetes://a.b.svc.cluster.local", targetInfo{"a", "b", ""}, false},
		{"kubernetes://a.b.svc.cluster.local:80", targetInfo{"a", "b", "80"}, false},
		{"kubernetes:///a.b.svc.cluster.local", targetInfo{"a", "b", ""}, false},
		{"kubernetes:///a.b.svc.cluster.local:80", targetInfo{"a", "b", "80"}, false},
		{"kubernetes:///a.b.svc.cluster.local:port", targetInfo{"a", "b", "port"}, false},
	} {
		got, err := parseResolverTarget(parseTarget(test.target))
		if err == nil && test.err {
			t.Errorf("case %d: want error but got nil", i)
			continue
		}
		if err != nil && !test.err {
			t.Errorf("case %d:got '%v' error but don't want an error", i, err)
			continue
		}
		if got != test.want {
			t.Errorf("case %d: parseTarget(%q) = %+v, want %+v", i, test.target, got, test.want)
		}
	}
}

func TestParseTargets(t *testing.T) {
	for i, test := range []struct {
		target string
		want   targetInfo
		err    bool
	}{
		{"", targetInfo{}, true},
		{"kubernetes:///", targetInfo{}, true},
		{"kubernetes://a:30", targetInfo{"a", "", "30"}, false},
		{"kubernetes://a/", targetInfo{"a", "", ""}, false},
		{"kubernetes:///a", targetInfo{"a", "", ""}, false},
		{"kubernetes://a/b", targetInfo{"b", "a", ""}, false},
		{"kubernetes://a.b/", targetInfo{"a", "b", ""}, false},
		{"kubernetes:///a.b:80", targetInfo{"a", "b", "80"}, false},
		{"kubernetes:///a.b:port", targetInfo{"a", "b", "port"}, false},
		{"kubernetes:///a:port", targetInfo{"a", "", "port"}, false},
		{"kubernetes://x/a:port", targetInfo{"a", "x", "port"}, false},
		{"kubernetes://a.x:30/", targetInfo{"a", "x", "30"}, false},
		{"kubernetes://a.b.svc.cluster.local", targetInfo{"a", "b", ""}, false},
		{"kubernetes://a.b.svc.cluster.local:80", targetInfo{"a", "b", "80"}, false},
		{"kubernetes:///a.b.svc.cluster.local", targetInfo{"a", "b", ""}, false},
		{"kubernetes:///a.b.svc.cluster.local:80", targetInfo{"a", "b", "80"}, false},
		{"kubernetes:///a.b.svc.cluster.local:port", targetInfo{"a", "b", "port"}, false},
	} {
		got, err := parseResolverTarget(parseTarget(test.target))
		if err == nil && test.err {
			t.Errorf("case %d: want error but got nil", i)
			continue
		}
		if err != nil && !test.err {
			t.Errorf("case %d:got '%v' error but don't want an error", i, err)
			continue
		}
		if got != test.want {
			t.Errorf("case %d: parseTarget(%q) = %+v, want %+v", i, test.target, got, test.want)
		}
	}
}
