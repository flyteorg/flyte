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

// getK8sSecretName returns the secret name and its k8s-safe MD5-hashed name.
func getK8sSecretName(_ context.Context, id *secretpb.SecretIdentifier) (string, string, error) {
	secretName := id.GetName()
	// We need to encode to make sure it is a valid secret name
	k8sSecretName := flytesecret.EncodeK8sSecretName(secretName)
	return secretName, k8sSecretName, nil
}

// resolveSecretNamespace returns the Kubernetes namespace for a secret based on its project and domain.
// If both project and domain are empty, it falls back to the configured default namespace.
// If only one of them is set, it returns an error.
func resolveSecretNamespace(project, domain string) (string, error) {
	if project == "" && domain == "" {
		return config.GetConfig().Kubernetes.Namespace, nil
	}
	if project == "" || domain == "" {
		return "", fmt.Errorf("project and domain must both be set or both be empty, got project=%q domain=%q", project, domain)
	}
	return fmt.Sprintf("%s-%s", project, domain), nil
}

func (s *SecretService) CreateSecret(ctx context.Context, req *connect.Request[secretpb.CreateSecretRequest]) (*connect.Response[secretpb.CreateSecretResponse], error) {
	logger.Debugf(ctx, "SecretService.CreateSecret called for %v", req.Msg.GetId().GetName())

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	secretName, k8sSecretName, err := getK8sSecretName(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	logger.Debugf(ctx, "decoded secret name %v, k8s secret name %v", secretName, k8sSecretName)

	namespace, err := resolveSecretNamespace(req.Msg.GetId().GetProject(), req.Msg.GetId().GetDomain())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	k8sSecret, err := buildK8sSecret(secretName, k8sSecretName, namespace, req.Msg.GetSecretSpec())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := s.k8sClient.Create(ctx, k8sSecret); err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		logger.Errorf(ctx, "failed to create secret %v: %v", secretName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&secretpb.CreateSecretResponse{}), nil
}

func (s *SecretService) UpdateSecret(ctx context.Context, req *connect.Request[secretpb.UpdateSecretRequest]) (*connect.Response[secretpb.UpdateSecretResponse], error) {
	logger.Debugf(ctx, "SecretService.UpdateSecret called for %v", req.Msg.GetId().GetName())

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	secretName, k8sSecretName, err := getK8sSecretName(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	logger.Debugf(ctx, "decoded secret name %v, k8s secret name %v", secretName, k8sSecretName)

	// Get the existing secret first to obtain the ResourceVersion required by controller-runtime Update.
	namespace, err := resolveSecretNamespace(req.Msg.GetId().GetProject(), req.Msg.GetId().GetDomain())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	existing := &corev1.Secret{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{Name: k8sSecretName, Namespace: namespace}, existing); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("secret %v not found", secretName))
		}
		logger.Errorf(ctx, "failed to get secret %v: %v", secretName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	switch v := req.Msg.GetSecretSpec().GetValue().(type) {
	case *secretpb.SecretSpec_StringValue:
		existing.StringData = map[string]string{secretName: v.StringValue}
		existing.Data = nil
	case *secretpb.SecretSpec_BinaryValue:
		existing.Data = map[string][]byte{secretName: v.BinaryValue}
		existing.StringData = nil
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("string or binary value expected, got %T", v))
	}

	if err := s.k8sClient.Update(ctx, existing); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("secret %v not found", secretName))
		}
		logger.Errorf(ctx, "failed to update secret %v: %v", secretName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&secretpb.UpdateSecretResponse{}), nil
}

func (s *SecretService) GetSecret(ctx context.Context, req *connect.Request[secretpb.GetSecretRequest]) (*connect.Response[secretpb.GetSecretResponse], error) {
	logger.Infof(ctx, "SecretService.GetSecret called for %v", req.Msg.GetId().GetName())

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	secretName, k8sSecretName, err := getK8sSecretName(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	logger.Debugf(ctx, "decoded secret name %v, k8s secret name %v", secretName, k8sSecretName)

	namespace, err := resolveSecretNamespace(req.Msg.GetId().GetProject(), req.Msg.GetId().GetDomain())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	k8sSecret := &corev1.Secret{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{Name: k8sSecretName, Namespace: namespace}, k8sSecret); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("secret %v not found", secretName))
		}
		logger.Errorf(ctx, "failed to get secret %v: %v", secretName, err)
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

	secretName, k8sSecretName, err := getK8sSecretName(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	logger.Debugf(ctx, "decoded secret name %v, k8s secret name %v", secretName, k8sSecretName)

	namespace, err := resolveSecretNamespace(req.Msg.GetId().GetProject(), req.Msg.GetId().GetDomain())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sSecretName,
			Namespace: namespace,
		},
	}

	if err := s.k8sClient.Delete(ctx, k8sSecret); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("secret %v not found", secretName))
		}
		logger.Errorf(ctx, "failed to delete secret %v: %v", secretName, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.Debugf(ctx, "deleted k8s secret %v", k8sSecretName)
	return connect.NewResponse(&secretpb.DeleteSecretResponse{}), nil
}

func (s *SecretService) ListSecrets(ctx context.Context, req *connect.Request[secretpb.ListSecretsRequest]) (*connect.Response[secretpb.ListSecretsResponse], error) {
	logger.Debugf(ctx, "SecretService.ListSecrets called")

	namespace, err := resolveSecretNamespace(req.Msg.GetProject(), req.Msg.GetDomain())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	k8sSecretList := &corev1.SecretList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.Limit(int64(req.Msg.GetLimit())),
		client.Continue(req.Msg.GetToken()),
		client.MatchingLabels{managedSecretLabelKey: managedSecretLabelValue},
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
		var secretName string
		for key := range k8sSecret.Data {
			secretName = key
			break
		}
		rawSecrets = append(rawSecrets, &secretpb.Secret{
			Id: &secretpb.SecretIdentifier{
				Organization: req.Msg.GetOrganization(),
				Project:      req.Msg.GetProject(),
				Domain:       req.Msg.GetDomain(),
				Name:         secretName,
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

// buildK8sSecret constructs a Kubernetes Secret object from the decoded secret name and spec.
func buildK8sSecret(secretName, k8sSecretName, namespace string, spec *secretpb.SecretSpec) (*corev1.Secret, error) {
	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sSecretName,
			Namespace: namespace,
			Labels: map[string]string{
				managedSecretLabelKey: managedSecretLabelValue,
			},
		},
	}

	switch v := spec.GetValue().(type) {
	case *secretpb.SecretSpec_StringValue:
		k8sSecret.StringData = map[string]string{secretName: v.StringValue}
	case *secretpb.SecretSpec_BinaryValue:
		k8sSecret.Data = map[string][]byte{secretName: v.BinaryValue}
	default:
		return nil, fmt.Errorf("string or binary value expected, got %T", v)
	}

	return k8sSecret, nil
}
