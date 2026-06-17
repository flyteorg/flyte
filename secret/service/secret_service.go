package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flytesecret "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	commonpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	secretpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/secret"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/secret/secretconnect"
	"github.com/flyteorg/flyte/v2/secret/config"
)

const (
	managedSecretLabelKey   = "app.flyte.org/managed"
	managedSecretLabelValue = "true"
	projectLabelKey         = "app.flyte.org/project"
	domainLabelKey          = "app.flyte.org/domain"

	// defaultOrganization is a placeholder org used in the encoded secret ID
	// because Flyte OSS v2 has no organization concept. It must match the
	// organization label stamped on task pods by the executor.
	defaultOrganization = "flyte"
)

// Ensure SecretService implements SecretServiceHandler at compile time.
var _ secretconnect.SecretServiceHandler = (*SecretService)(nil)

// SecretService implements the SecretService gRPC API.
type SecretService struct {
	k8sClient client.Client
}

// NewSecretService creates a new SecretService.
func NewSecretService(k8sClient client.Client) *SecretService {
	return &SecretService{k8sClient: k8sClient}
}

// validateScope enforces the valid scope combinations supported by the secret
// fetcher: global, domain-only, or domain+project. Project is only valid when
// a domain is also set.
func validateScope(domain, project string) error {
	if domain == "" && project != "" {
		return fmt.Errorf("project-scoped secrets must also specify a domain, got project=%q domain=%q", project, domain)
	}
	return nil
}

// getK8sSecretName returns the encoded secret ID (used as the data key inside
// the Kubernetes Secret and as input to EncodeK8sSecretName) and the k8s-safe
// hashed resource name. The encoding matches what the secret fetcher looks up
// when a task pod is admitted.
func getK8sSecretName(_ context.Context, id *secretpb.SecretIdentifier) (string, string, error) {
	if err := validateScope(id.GetDomain(), id.GetProject()); err != nil {
		return "", "", err
	}
	encoded := flytesecret.EncodeSecretName(defaultOrganization, id.GetDomain(), id.GetProject(), id.GetName())
	return encoded, flytesecret.EncodeK8sSecretName(encoded), nil
}

// secretNamespace returns the single namespace used for all flyte-managed secrets.
func secretNamespace() string {
	return config.GetConfig().Kubernetes.Namespace
}

func (s *SecretService) CreateSecret(ctx context.Context, req *connect.Request[secretpb.CreateSecretRequest]) (*connect.Response[secretpb.CreateSecretResponse], error) {
	logger.Debugf(ctx, "SecretService.CreateSecret called for %v", req.Msg.GetId().GetName())

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	encodedName, k8sSecretName, err := getK8sSecretName(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logger.Debugf(ctx, "encoded secret name %v, k8s secret name %v", encodedName, k8sSecretName)

	k8sSecret, err := buildK8sSecret(encodedName, k8sSecretName, secretNamespace(), req.Msg.GetId(), req.Msg.GetSecretSpec())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := s.k8sClient.Create(ctx, k8sSecret); err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		logger.Errorf(ctx, "failed to create secret %v: %v", encodedName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&secretpb.CreateSecretResponse{}), nil
}

func (s *SecretService) UpdateSecret(ctx context.Context, req *connect.Request[secretpb.UpdateSecretRequest]) (*connect.Response[secretpb.UpdateSecretResponse], error) {
	logger.Debugf(ctx, "SecretService.UpdateSecret called for %v", req.Msg.GetId().GetName())

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	encodedName, k8sSecretName, err := getK8sSecretName(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logger.Debugf(ctx, "encoded secret name %v, k8s secret name %v", encodedName, k8sSecretName)

	// Get the existing secret first to obtain the ResourceVersion required by controller-runtime Update.
	existing := &corev1.Secret{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{Name: k8sSecretName, Namespace: secretNamespace()}, existing); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("secret %v not found", req.Msg.GetId().GetName()))
		}
		logger.Errorf(ctx, "failed to get secret %v: %v", encodedName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	switch v := req.Msg.GetSecretSpec().GetValue().(type) {
	case *secretpb.SecretSpec_StringValue:
		existing.StringData = map[string]string{encodedName: v.StringValue}
		existing.Data = nil
	case *secretpb.SecretSpec_BinaryValue:
		existing.Data = map[string][]byte{encodedName: v.BinaryValue}
		existing.StringData = nil
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("string or binary value expected, got %T", v))
	}

	if err := s.k8sClient.Update(ctx, existing); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("secret %v not found", req.Msg.GetId().GetName()))
		}
		logger.Errorf(ctx, "failed to update secret %v: %v", encodedName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&secretpb.UpdateSecretResponse{}), nil
}

func (s *SecretService) GetSecret(ctx context.Context, req *connect.Request[secretpb.GetSecretRequest]) (*connect.Response[secretpb.GetSecretResponse], error) {
	logger.Infof(ctx, "SecretService.GetSecret called for %v", req.Msg.GetId().GetName())

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	encodedName, k8sSecretName, err := getK8sSecretName(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logger.Debugf(ctx, "encoded secret name %v, k8s secret name %v", encodedName, k8sSecretName)

	k8sSecret := &corev1.Secret{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{Name: k8sSecretName, Namespace: secretNamespace()}, k8sSecret); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("secret %v not found", req.Msg.GetId().GetName()))
		}
		logger.Errorf(ctx, "failed to get secret %v: %v", encodedName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&secretpb.GetSecretResponse{
		Secret: &secretpb.Secret{
			Id: req.Msg.GetId(),
			SecretMetadata: &secretpb.SecretMetadata{
				CreatedTime:  timestamppb.New(k8sSecret.GetCreationTimestamp().Time),
				SecretStatus: fullyPresentStatus(),
			},
		},
	}), nil
}

func (s *SecretService) DeleteSecret(ctx context.Context, req *connect.Request[secretpb.DeleteSecretRequest]) (*connect.Response[secretpb.DeleteSecretResponse], error) {
	logger.Debugf(ctx, "SecretService.DeleteSecret called for %v", req.Msg.GetId().GetName())

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	encodedName, k8sSecretName, err := getK8sSecretName(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logger.Debugf(ctx, "encoded secret name %v, k8s secret name %v", encodedName, k8sSecretName)

	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sSecretName,
			Namespace: secretNamespace(),
		},
	}

	if err := s.k8sClient.Delete(ctx, k8sSecret); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("secret %v not found", req.Msg.GetId().GetName()))
		}
		logger.Errorf(ctx, "failed to delete secret %v: %v", encodedName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.Debugf(ctx, "deleted k8s secret %v", k8sSecretName)
	return connect.NewResponse(&secretpb.DeleteSecretResponse{}), nil
}

func (s *SecretService) ListSecrets(ctx context.Context, req *connect.Request[secretpb.ListSecretsRequest]) (*connect.Response[secretpb.ListSecretsResponse], error) {
	logger.Debugf(ctx, "SecretService.ListSecrets called")

	reqProject := req.Msg.GetProject()
	reqDomain := req.Msg.GetDomain()
	if err := validateScope(reqDomain, reqProject); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	matchLabels := client.MatchingLabels{managedSecretLabelKey: managedSecretLabelValue}
	if reqDomain != "" {
		matchLabels[domainLabelKey] = reqDomain
	}
	if reqProject != "" {
		matchLabels[projectLabelKey] = reqProject
	}

	k8sSecretList := &corev1.SecretList{}
	listOpts := []client.ListOption{
		client.InNamespace(secretNamespace()),
		client.Limit(int64(req.Msg.GetLimit())),
		client.Continue(req.Msg.GetToken()),
		matchLabels,
	}

	if err := s.k8sClient.List(ctx, k8sSecretList, listOpts...); err != nil {
		logger.Errorf(ctx, "failed to list secrets: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	rawSecrets := make([]*secretpb.Secret, 0, len(k8sSecretList.Items))
	for _, k8sSecret := range k8sSecretList.Items {
		if len(k8sSecret.Data) != 1 {
			logger.Errorf(ctx, "unexpected number of keys in secret %v: %d", k8sSecret.Name, len(k8sSecret.Data))
			continue
		}
		var encodedKey string
		for key := range k8sSecret.Data {
			encodedKey = key
			break
		}
		components, err := flytesecret.DecodeSecretName(encodedKey)
		if err != nil {
			logger.Errorf(ctx, "failed to decode secret key %q in %v: %v", encodedKey, k8sSecret.Name, err)
			continue
		}
		rawSecrets = append(rawSecrets, &secretpb.Secret{
			Id: &secretpb.SecretIdentifier{
				Project: components.Project,
				Domain:  components.Domain,
				Name:    components.Name,
			},
			SecretMetadata: &secretpb.SecretMetadata{
				CreatedTime:  timestamppb.New(k8sSecret.GetCreationTimestamp().Time),
				SecretStatus: fullyPresentStatus(),
			},
		})
	}

	return connect.NewResponse(&secretpb.ListSecretsResponse{
		Secrets: rawSecrets,
		Token:   k8sSecretList.Continue,
	}), nil
}

// fullyPresentStatus returns a SecretStatus indicating the secret is present in this single cluster.
func fullyPresentStatus() *secretpb.SecretStatus {
	cfg := config.GetConfig()
	return &secretpb.SecretStatus{
		OverallStatus: secretpb.OverallStatus_FULLY_PRESENT,
		ClusterStatus: []*secretpb.ClusterSecretStatus{
			{
				Cluster: &commonpb.ClusterIdentifier{
					Name: cfg.Kubernetes.ClusterName,
				},
				PresenceStatus: secretpb.SecretPresenceStatus_PRESENT,
			},
		},
	}
}

// buildK8sSecret constructs a Kubernetes Secret object. The encoded secret ID
// (full org/domain/project/name form produced by flytesecret.EncodeSecretName)
// is used as the data map key so the fetcher can read it back by the same key.
func buildK8sSecret(encodedName, k8sSecretName, namespace string, id *secretpb.SecretIdentifier, spec *secretpb.SecretSpec) (*corev1.Secret, error) {
	labels := map[string]string{
		managedSecretLabelKey: managedSecretLabelValue,
	}
	if id.GetDomain() != "" {
		labels[domainLabelKey] = id.GetDomain()
	}
	if id.GetProject() != "" {
		labels[projectLabelKey] = id.GetProject()
	}
	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sSecretName,
			Namespace: namespace,
			Labels:    labels,
		},
	}

	switch v := spec.GetValue().(type) {
	case *secretpb.SecretSpec_StringValue:
		k8sSecret.StringData = map[string]string{encodedName: v.StringValue}
	case *secretpb.SecretSpec_BinaryValue:
		k8sSecret.Data = map[string][]byte{encodedName: v.BinaryValue}
	default:
		return nil, fmt.Errorf("string or binary value expected, got %T", v)
	}

	return k8sSecret, nil
}
