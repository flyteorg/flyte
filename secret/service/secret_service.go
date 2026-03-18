package service

import (
	"context"
	"encoding/base64"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flytesecret "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
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

// getK8sSecretName decodes the base64 URL-encoded secret name from the identifier
// and returns both the decoded name and the k8s-safe MD5-hashed name.
func getK8sSecretName(ctx context.Context, id *secretpb.SecretIdentifier) (string, string, error) {
	decoded, err := base64.URLEncoding.DecodeString(id.GetName())
	if err != nil {
		logger.Errorf(ctx, "failed to base64 decode secret name %v: %v", id.GetName(), err)
		return "", "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid secret name encoding: %w", err))
	}
	secretName := string(decoded)
	k8sSecretName := flytesecret.EncodeK8sSecretName(secretName)
	return secretName, k8sSecretName, nil
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

	k8sSecret, err := buildK8sSecret(secretName, k8sSecretName, req.Msg.GetSecretSpec())
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
	namespace := config.GetConfig().Kubernetes.Namespace
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

	namespace := config.GetConfig().Kubernetes.Namespace
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
				CreatedTime: timestamppb.New(k8sSecret.GetCreationTimestamp().Time),
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

	namespace := config.GetConfig().Kubernetes.Namespace
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

	namespace := config.GetConfig().Kubernetes.Namespace
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
			Id: &secretpb.SecretIdentifier{Name: secretName},
			SecretMetadata: &secretpb.SecretMetadata{
				CreatedTime: timestamppb.New(k8sSecret.GetCreationTimestamp().Time),
			},
		})
	}

	return connect.NewResponse(&secretpb.ListSecretsResponse{
		Secrets: rawSecrets,
		Token:   k8sSecretList.Continue,
	}), nil
}

// buildK8sSecret constructs a Kubernetes Secret object from the decoded secret name and spec.
func buildK8sSecret(secretName, k8sSecretName string, spec *secretpb.SecretSpec) (*corev1.Secret, error) {
	namespace := config.GetConfig().Kubernetes.Namespace

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
