package resolver

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"net"
	"strings"
	"sync"
)

const (
	KubernetesSchema = "kubernetes"
)

type targetInfo struct {
	serviceName      string
	serviceNamespace string
	port             string
}

// NewBuilder creates a kubeBuilder which is used by grpc resolver.
func NewBuilder(client kubernetes.Interface, schema string) resolver.Builder {
	return &kubeBuilder{
		k8sClient: client,
		schema:    schema,
	}
}

type kubeBuilder struct {
	k8sClient kubernetes.Interface
	schema    string
}

func splitServicePortNamespace(hpn string) (service, port, namespace string) {
	service = hpn

	colon := strings.LastIndexByte(service, ':')
	if colon != -1 {
		service, port = service[:colon], service[colon+1:]
	}

	// we want to split into the service name, namespace, and whatever else is left
	// this will support fully qualified service names, e.g. {service-name}.<namespace>.svc.<cluster-domain-name>.
	// Note that since we look up the endpoints by service name and namespace, we don't care about the
	// cluster-domain-name, only that we can parse out the service name and namespace properly.
	parts := strings.SplitN(service, ".", 3)
	if len(parts) >= 2 {
		service, namespace = parts[0], parts[1]
	}

	return
}

func parseResolverTarget(target resolver.Target) (targetInfo, error) {
	var service, port, namespace string
	if target.URL.Host == "" {
		// kubernetes:///service.namespace:port
		service, port, namespace = splitServicePortNamespace(target.Endpoint())
	} else if target.URL.Port() == "" && target.Endpoint() != "" {
		// kubernetes://namespace/service:port
		service, port, _ = splitServicePortNamespace(target.Endpoint())
		namespace = target.URL.Hostname()
	} else {
		// kubernetes://service.namespace:port
		service, port, namespace = splitServicePortNamespace(target.URL.Host)
	}

	if service == "" {
		return targetInfo{}, fmt.Errorf("target %s must specify a service", &target.URL)
	}

	return targetInfo{
		serviceName:      service,
		serviceNamespace: namespace,
		port:             port,
	}, nil
}

// Build creates a new resolver for the given target, e.g. kubernetes:///flyteagent:flyte:8000.
func (b *kubeBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	ti, err := parseResolverTarget(target)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &kResolver{
		target:    ti,
		ctx:       ctx,
		cancel:    cancel,
		cc:        cc,
		k8sClient: b.k8sClient,
	}
	go wait.Until(r.run, 0, ctx.Done())
	return r, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *kubeBuilder) Scheme() string {
	return b.schema
}

type kResolver struct {
	target    targetInfo
	ctx       context.Context
	cancel    context.CancelFunc
	cc        resolver.ClientConn
	k8sClient kubernetes.Interface
	// wg is used to enforce Close() to return after the watcher() goroutine has finished.
	wg sync.WaitGroup
}

// ResolveNow is a no-op at this point.
func (k *kResolver) ResolveNow(resolver.ResolveNowOptions) {}

// Close closes the resolver.
func (k *kResolver) Close() {
	k.cancel()
	k.wg.Wait()
	logger.Infof(k.ctx, "k8s resolver: closed")
}

func (k *kResolver) resolve(e *v1.Endpoints) {
	var newAddrs []resolver.Address
	for _, subset := range e.Subsets {
		port := k.target.port

		for _, address := range subset.Addresses {
			newAddrs = append(newAddrs, resolver.Address{
				Addr:       net.JoinHostPort(address.IP, port),
				ServerName: fmt.Sprintf("%s.%s", k.target.serviceName, k.target.serviceNamespace),
				Metadata:   nil,
			})
		}
	}
	err := k.cc.UpdateState(resolver.State{Addresses: newAddrs})
	if err != nil {
		grpclog.Errorf("k8s resolver: failed  : %v", err)
	}
}

func (k *kResolver) run() {
	k.wg.Add(1)
	defer k.wg.Done()
	logger.Infof(k.ctx, "Starting k8s resolver for %s", k.target.serviceName)
	watcher, err := k.k8sClient.CoreV1().Endpoints(k.target.serviceNamespace).Watch(k.ctx, metav1.ListOptions{FieldSelector: "metadata.name=" + k.target.serviceName})
	if err != nil {
		grpclog.Errorf("k8s resolver: failed to create watcher: %v", err)
		return
	}

	for {
		logger.Infof(k.ctx, "for loop")
		select {
		case <-k.ctx.Done():
			logger.Infof(k.ctx, "done")
			return
		case event, ok := <-watcher.ResultChan():
			logger.Infof(k.ctx, "Event: %v\n", event)
			if !ok {
				logger.Debugf(k.ctx, "k8s resolver: watcher closed")
				return
			}
			if event.Object == nil {
				continue
			}
			k.resolve(event.Object.(*v1.Endpoints))
		}
	}
}
