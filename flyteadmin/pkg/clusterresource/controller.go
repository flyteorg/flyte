package clusterresource

import (
	"context"
	"crypto/md5" // #nosec
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	k8_api_err "k8s.io/apimachinery/pkg/api/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"

	impl2 "github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/impl"
	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/config"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster/impl"
	executionclusterIfaces "github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations"
	projectConfigurationPlugin "github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations/plugin"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories"
	errors2 "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	admin2 "github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const namespaceVariable = "namespace"
const projectVariable = "project"
const domainVariable = "domain"
const templateVariableFormat = "{{ %s }}"
const replaceAllInstancesOfString = -1
const noChange = "{}"
const reportNamespaceProvisionedBatchSize = 16

var (
	deleteImmediately      = int64(0)
	deleteForeground       = metav1.DeletePropagationForeground
	reservedMetadataFields = []string{"creationTimestamp", "resourceVersion", "selfLink", "uid"}
)

// The clusterresource Controller manages applying desired templatized kubernetes resource files as resources
// in the execution kubernetes cluster.
type Controller interface {
	Sync(ctx context.Context) error
	Run()
}

type controllerMetrics struct {
	Scope                                   promutils.Scope
	SyncErrors                              prometheus.Counter
	SyncStarted                             prometheus.Counter
	SyncsCompleted                          prometheus.Counter
	KubernetesResourcesCreated              prometheus.Counter
	KubernetesResourcesCreateErrors         prometheus.Counter
	KubernetesNamespaceDeleteErrors         prometheus.Counter
	ResourcesAdded                          prometheus.Counter
	NamespacesDeleted                       prometheus.Counter
	ResourceAddErrors                       prometheus.Counter
	TemplateReadErrors                      prometheus.Counter
	TemplateDecodeErrors                    prometheus.Counter
	AppliedTemplateExists                   prometheus.Counter
	TemplateUpdateErrors                    prometheus.Counter
	Panics                                  prometheus.Counter
	ReportClusterResourceStateUpdated       prometheus.Counter
	ReportClusterResourceStateUpdatedErrors prometheus.Counter
}

type FileName = string
type NamespaceName = string
type TemplateChecksums = map[FileName][16]byte
type NamespaceCache = map[NamespaceName]TemplateChecksums

type templateValuesType = map[string]string

type controller struct {
	config                 runtimeInterfaces.Configuration
	poller                 chan struct{}
	metrics                controllerMetrics
	lastAppliedTemplateDir string
	// Map of [namespace -> [templateFileName -> last modified time]]
	appliedNamespaceTemplateChecksums sync.Map
	adminDataProvider                 interfaces.FlyteAdminDataProvider
	listTargets                       executionclusterIfaces.ListTargetsInterface
	pluginRegistry                    *plugins.Registry
}

// templateAlreadyApplied checks if there is an applied template with the same checksum
func (c *controller) templateAlreadyApplied(ctx context.Context, namespace NamespaceName, templateFilename string, checksum [16]byte) bool {
	namespaceTemplateChecksumsVal, ok := c.appliedNamespaceTemplateChecksums.Load(namespace)
	if !ok {
		// There is no record of this namespace altogether.
		return false
	}

	namespacedAppliedTemplateChecksums, ok := namespaceTemplateChecksumsVal.(TemplateChecksums)
	if !ok {
		logger.Errorf(ctx, "unexpected type for namespace '%s' in appliedTemplates map: %t", namespace, namespaceTemplateChecksumsVal)
		return false
	}
	appliedChecksum, ok := namespacedAppliedTemplateChecksums[templateFilename]
	if !ok {
		// There is no record of this file having ever been applied.
		return false
	}
	// Check if the applied template matches the new one
	return appliedChecksum == checksum
}

// setTemplateChecksum records the latest checksum for the template file
func (c *controller) setTemplateChecksum(ctx context.Context, namespace NamespaceName, templateFilename string, checksum [16]byte) {
	namespaceTemplateChecksumsVal, ok := c.appliedNamespaceTemplateChecksums.Load(namespace)
	if !ok {
		c.appliedNamespaceTemplateChecksums.Store(namespace, TemplateChecksums{templateFilename: checksum})
		return
	}

	namespacedAppliedTemplateChecksums, ok := namespaceTemplateChecksumsVal.(TemplateChecksums)
	if !ok {
		logger.Errorf(ctx, "unexpected type for namespace '%s' in appliedTemplates map: %t", namespace, namespaceTemplateChecksumsVal)
		return
	}
	namespacedAppliedTemplateChecksums[templateFilename] = checksum
	c.appliedNamespaceTemplateChecksums.Store(namespace, namespacedAppliedTemplateChecksums)
}

// Given a map of templatized variable names -> data source, this function produces an output that maps the same
// variable names to their fully resolved values (from the specified data source).
func populateTemplateValues(data map[string]runtimeInterfaces.DataSource) (templateValuesType, error) {
	templateValues := make(templateValuesType, len(data))
	collectedErrs := make([]error, 0)
	for templateVar, dataSource := range data {
		if templateVar == namespaceVariable || templateVar == projectVariable || templateVar == domainVariable {
			// The namespace variable is specifically reserved for system use only.
			collectedErrs = append(collectedErrs, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"Cannot assign namespace template value in user data"))
			continue
		}
		var dataValue string
		if len(dataSource.Value) > 0 {
			dataValue = dataSource.Value
		} else if len(dataSource.ValueFrom.EnvVar) > 0 {
			dataValue = os.Getenv(dataSource.ValueFrom.EnvVar)
		} else if len(dataSource.ValueFrom.FilePath) > 0 {
			templateFile, err := ioutil.ReadFile(dataSource.ValueFrom.FilePath)
			if err != nil {
				collectedErrs = append(collectedErrs, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"failed to substitute parameterized value for %s: unable to read value from: [%+v] with err: %v",
					templateVar, dataSource.ValueFrom.FilePath, err))
				continue
			}
			dataValue = string(templateFile)
		} else {
			collectedErrs = append(collectedErrs, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"failed to substitute parameterized value for %s: unset or unrecognized ValueFrom: [%+v]", templateVar, dataSource.ValueFrom))
			continue
		}
		if len(dataValue) == 0 {
			collectedErrs = append(collectedErrs, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"failed to substitute parameterized value for %s: unset. ValueFrom: [%+v]", templateVar, dataSource.ValueFrom))
			continue
		}
		templateValues[fmt.Sprintf(templateVariableFormat, templateVar)] = dataValue
	}
	if len(collectedErrs) > 0 {
		return nil, errors.NewCollectedFlyteAdminError(codes.InvalidArgument, collectedErrs)
	}
	return templateValues, nil
}

// Produces a map of template variable names and their fully resolved values based on configured defaults for each
// system-domain in the application config file.
func populateDefaultTemplateValues(defaultData map[runtimeInterfaces.DomainName]runtimeInterfaces.TemplateData) (
	map[string]templateValuesType, error) {
	defaultTemplateValues := make(map[string]templateValuesType)
	collectedErrs := make([]error, 0)
	for domainName, templateData := range defaultData {
		domainSpecificTemplateValues, err := populateTemplateValues(templateData)
		if err != nil {
			collectedErrs = append(collectedErrs, err)
			continue
		}
		defaultTemplateValues[domainName] = domainSpecificTemplateValues
	}
	if len(collectedErrs) > 0 {
		return nil, errors.NewCollectedFlyteAdminError(codes.InvalidArgument, collectedErrs)
	}
	return defaultTemplateValues, nil
}

// Fetches user-specified overrides from the admin database for template variables and their desired value
// substitutions based on the input project and domain. These database values are overlaid on top of the configured
// variable defaults for the specific domain as defined in the admin application config file.
func (c *controller) getCustomTemplateValues(
	ctx context.Context, org, project, domain string, domainTemplateValues templateValuesType) (templateValuesType, error) {
	if len(domainTemplateValues) == 0 {
		domainTemplateValues = make(templateValuesType)
	}
	customTemplateValues := make(templateValuesType)
	for key, value := range domainTemplateValues {
		customTemplateValues[key] = value
	}
	collectedErrs := make([]error, 0)
	// All override values saved in the database take precedence over the domain-specific defaults.
	attributes, err := c.adminDataProvider.GetClusterResourceAttributes(ctx, org, project, domain)
	if err != nil {
		s, ok := status.FromError(err)
		if !ok || s.Code() != codes.NotFound {
			collectedErrs = append(collectedErrs, err)
		}
	}
	if attributes != nil && attributes.Attributes != nil {
		for templateKey, templateValue := range attributes.Attributes {
			customTemplateValues[fmt.Sprintf(templateVariableFormat, templateKey)] = templateValue
		}
	}

	if len(collectedErrs) > 0 {
		return nil, errors.NewCollectedFlyteAdminError(codes.InvalidArgument, collectedErrs)
	}
	return customTemplateValues, nil
}

// Obtains the REST interface for a GroupVersionResource
func getDynamicResourceInterface(mapping *meta.RESTMapping, dynamicClient dynamic.Interface, namespace NamespaceName) dynamic.ResourceInterface {
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		return dynamicClient.Resource(mapping.Resource).Namespace(namespace)
	}
	// for cluster-wide resources (e.g. namespaces)
	return dynamicClient.Resource(mapping.Resource)
}

// This represents the minimum closure of objects generated from a template file that
// allows for dynamically creating (or updating) the resource using server side apply.
type dynamicResource struct {
	obj     *unstructured.Unstructured
	mapping *meta.RESTMapping
}

// This function borrows heavily from the excellent example code here:
// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go#Background-Server-Side-Apply
// to dynamically discover the GroupVersionResource for the templatized k8s object from the cluster resource config files
// which a dynamic client can use to create or mutate the resource.
func prepareDynamicCreate(target executioncluster.ExecutionTarget, config string) (dynamicResource, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(&target.Config)
	if err != nil {
		return dynamicResource{}, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(config), nil, obj)
	if err != nil {
		return dynamicResource{}, err
	}

	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return dynamicResource{}, err
	}

	return dynamicResource{
		obj:     obj,
		mapping: mapping,
	}, nil
}

// This function loops through the kubernetes resource template files in the configured template directory.
// For each unapplied template file (wrt the namespace) this func attempts to
//  1. create k8s object resource from template by performing:
//     a) read template file
//     b) substitute templatized variables with their resolved values
//  2. create the resource on the kubernetes cluster and cache successful outcomes
func (c *controller) syncNamespace(ctx context.Context, project *admin.Project, domain *admin.Domain,
	namespace NamespaceName, templateValues, customTemplateValues templateValuesType) (ResourceSyncStats, error) {
	logger.Debugf(ctx, "syncing namespace: '%s' for project '%s' and domain '%s' and org '%s'", namespace, project.Id, domain.Id, project.Org)
	templateDir := c.config.ClusterResourceConfiguration().GetTemplatePath()
	if c.lastAppliedTemplateDir != templateDir {
		// Invalidate all caches
		c.lastAppliedTemplateDir = templateDir
		c.appliedNamespaceTemplateChecksums = sync.Map{}
	}
	templateFiles, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return ResourceSyncStats{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"Failed to read config template dir [%s] for namespace [%s] with err: %v",
			namespace, templateDir, err)
	}

	collectedErrs := make([]error, 0)
	stats := ResourceSyncStats{}
	for _, templateFile := range templateFiles {
		templateFile := templateFile
		templateFileName := templateFile.Name()
		if filepath.Ext(templateFileName) != ".yaml" {
			// nothing to do.
			logger.Debugf(ctx, "syncing namespace [%s]: ignoring unrecognized filetype [%s]",
				namespace, templateFile.Name())
			continue
		}

		// 1) create resource from template and check if already applied
		k8sManifest, err := c.createResourceFromTemplate(ctx, templateDir, templateFileName, project, domain, namespace, templateValues, customTemplateValues)
		if err != nil {
			collectedErrs = append(collectedErrs, err)
			continue
		}

		checksum := md5.Sum([]byte(k8sManifest)) // #nosec
		if c.templateAlreadyApplied(ctx, namespace, templateFileName, checksum) {
			// nothing to do.
			logger.Debugf(ctx, "syncing namespace [%s]: templateFile [%s] already applied, nothing to do.", namespace, templateFile.Name())
			stats.AlreadyThere++
			continue
		}

		// 2) create the resource on the kubernetes cluster and cache successful outcomes
		for _, target := range c.listTargets.GetValidTargets() {
			dynamicObj, err := prepareDynamicCreate(*target, k8sManifest)
			if err != nil {
				logger.Warningf(ctx, "Failed to transform kubernetes manifest for namespace [%s] "+
					"into a dynamic unstructured mapping with err: %v, manifest: %v", namespace, err, k8sManifest)
				collectedErrs = append(collectedErrs, err)
				c.metrics.KubernetesResourcesCreateErrors.Inc()
				stats.Errored++
				continue
			}

			logger.Debugf(ctx, "Attempting to create resource [%+v] in cluster [%v] for namespace [%s] and templateFileName: %+v with k8sManifest %v",
				dynamicObj.obj.GetKind(), target.ID, namespace, templateFileName, k8sManifest)

			dr := getDynamicResourceInterface(dynamicObj.mapping, target.DynamicClient, namespace)
			_, err = dr.Create(ctx, dynamicObj.obj, metav1.CreateOptions{})

			if err != nil {
				if k8serrors.IsAlreadyExists(err) {
					logger.Debugf(ctx, "Type [%+v] in namespace [%s] already exists - attempting update instead",
						dynamicObj.obj.GetKind(), namespace)
					c.metrics.AppliedTemplateExists.Inc()

					currentObj, err := dr.Get(ctx, dynamicObj.obj.GetName(), metav1.GetOptions{})
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Warningf(ctx, "Failed to get current resource from server [%+v] in namespace [%s] with err: %v",
							dynamicObj.obj.GetKind(), namespace, err)
						collectedErrs = append(collectedErrs, err)
						stats.Errored++
						continue
					}

					objectFieldsMap := currentObj.Object
					metadataFieldsVal, ok := objectFieldsMap["metadata"]
					if ok {
						metadataFieldsMap, isOfType := metadataFieldsVal.(map[string]interface{})
						if !isOfType {
							logger.Debugf(ctx, "failed to attempt to clear metadata for object: %+v, metadata was of type: %T", currentObj, metadataFieldsVal)
						} else {
							for _, reservedField := range reservedMetadataFields {
								delete(metadataFieldsMap, reservedField)
							}
							objectFieldsMap["metadata"] = metadataFieldsMap
							currentObj.Object = objectFieldsMap
						}
					}

					modified, err := json.Marshal(dynamicObj.obj)
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Warningf(ctx, "Failed to marshal resource [%+v] in namespace [%s] to json with err: %v",
							dynamicObj.obj.GetKind(), namespace, err)
						collectedErrs = append(collectedErrs, err)
						stats.Errored++
						continue
					}

					patch, patchType, err := c.createPatch(dynamicObj.mapping.GroupVersionKind, currentObj, modified, namespace)
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Warningf(ctx, "Failed to create patch for resource [%+v] in namespace [%s] err: %v",
							dynamicObj.obj.GetKind(), namespace, err)
						collectedErrs = append(collectedErrs, err)
						stats.Errored++
						continue
					}

					if string(patch) == noChange {
						logger.Infof(ctx, "Resource [%+v] in namespace [%s] is not modified",
							dynamicObj.obj.GetKind(), namespace)
						stats.AlreadyThere++
						c.setTemplateChecksum(ctx, namespace, templateFileName, checksum)
						continue
					}

					_, err = dr.Patch(ctx, dynamicObj.obj.GetName(),
						patchType, patch, metav1.PatchOptions{})
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Warningf(ctx, "Failed to patch resource [%+v] in namespace [%s] with err: %v",
							dynamicObj.obj.GetKind(), namespace, err)
						collectedErrs = append(collectedErrs, err)
						stats.Errored++
						continue
					}

					stats.Updated++
					logger.Debugf(ctx, "Successfully updated resource [%+v] in namespace [%s]",
						dynamicObj.obj.GetKind(), namespace)
					c.setTemplateChecksum(ctx, namespace, templateFileName, checksum)
				} else {
					// Some error other than AlreadyExists was raised when we tried to Create the k8s object.
					c.metrics.KubernetesResourcesCreateErrors.Inc()
					logger.Warningf(ctx, "Failed to create kubernetes object from config template [%s] for namespace [%s] with err: %v",
						templateFileName, namespace, err)
					err := errors.NewFlyteAdminErrorf(codes.Internal,
						"Failed to create kubernetes object from config template [%s] for namespace [%s] with err: %v",
						templateFileName, namespace, err)
					stats.Errored++
					collectedErrs = append(collectedErrs, err)
				}
			} else {
				stats.Created++
				logger.Debugf(ctx, "Created resource [%+v] for namespace [%s] in kubernetes",
					dynamicObj.obj.GetKind(), namespace)
				c.metrics.KubernetesResourcesCreated.Inc()
				c.setTemplateChecksum(ctx, namespace, templateFileName, checksum)
			}
		}
	}
	if len(collectedErrs) > 0 {
		return stats, errors.NewCollectedFlyteAdminError(codes.Internal, collectedErrs)
	}
	return stats, nil
}

func (c *controller) deleteNamespace(ctx context.Context, project *admin.Project, domain *admin.Domain,
	namespace NamespaceName) (ResourceSyncStats, error) {
	logger.Debugf(ctx, "attempting to delete namespace: '%s' for project '%s' and domain '%s' and org '%s'",
		namespace, project.Id, domain.Id, project.Org)
	collectedErrs := make([]error, 0)
	_, mapHasVal := c.appliedNamespaceTemplateChecksums.Load(namespace)
	if !mapHasVal {
		logger.Debugf(ctx, "namespace: '%s' for project '%s' and domain '%s' and org '%s' is already recorded as deleted",
			namespace, project.Id, domain.Id, project.Org)
		return ResourceSyncStats{}, nil
	}
	for _, target := range c.listTargets.GetValidTargets() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		err := target.Client.Delete(context.Background(), ns, &client.DeleteOptions{
			GracePeriodSeconds: &deleteImmediately,
			PropagationPolicy:  &deleteForeground,
		})
		if err != nil {
			if k8_api_err.IsNotFound(err) || k8_api_err.IsGone(err) {
				logger.Debugf(ctx, "namespace: '%s' for project '%s' and domain '%s' and org '%s' is already deleted",
					namespace, project.Id, domain.Id, project.Org)
				continue
			}

			c.metrics.KubernetesNamespaceDeleteErrors.Inc()
			logger.Errorf(ctx, "failed to delete namespace [%s] for project [%s] and domain [%s] and org [%s] with err: %v",
				namespace, project.Id, domain.Id, project.Org, err)
			collectedErrs = append(collectedErrs, err)
			continue
		}
		logger.Infof(ctx, "successfully deleted namespace [%s] for project [%s] and domain [%s] and org [%s]", namespace, project.Id, domain.Id, project.Org)
		c.metrics.NamespacesDeleted.Inc()
	}

	// Now clean up internal state that tracks this formerly active namespace
	c.appliedNamespaceTemplateChecksums.Delete(namespace)

	if len(collectedErrs) > 0 {
		return ResourceSyncStats{}, errors.NewCollectedFlyteAdminError(codes.Internal, collectedErrs)
	}

	return ResourceSyncStats{
		Deleted: 1,
	}, nil
}

var metadataAccessor = meta.NewAccessor()

// getLastApplied get last applied manifest from object's annotation
func getLastApplied(obj k8sruntime.Object) ([]byte, error) {
	annots, err := metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, err
	}

	if annots == nil {
		return nil, nil
	}

	lastApplied, ok := annots[corev1.LastAppliedConfigAnnotation]
	if !ok {
		return nil, nil
	}

	return []byte(lastApplied), nil
}

func addResourceVersion(patch []byte, rv string) ([]byte, error) {
	var patchMap map[string]interface{}
	err := json.Unmarshal(patch, &patchMap)
	if err != nil {
		return nil, err
	}
	u := unstructured.Unstructured{Object: patchMap}
	a, err := meta.Accessor(&u)
	if err != nil {
		return nil, err
	}
	a.SetResourceVersion(rv)

	return json.Marshal(patchMap)
}

// createResourceFromTemplate this method perform following processes:
//  1. read template file pointed by templateDir and templateFileName
//  2. substitute templatized variables with their resolved values
//
// the method will return the kubernetes raw manifest
func (c *controller) createResourceFromTemplate(ctx context.Context, templateDir string,
	templateFileName string, project *admin.Project, domain *admin.Domain, namespace NamespaceName,
	templateValues, customTemplateValues templateValuesType) (string, error) {
	// 1) read the template file
	template, err := ioutil.ReadFile(path.Join(templateDir, templateFileName))
	if err != nil {
		logger.Warningf(ctx,
			"failed to read config template from path [%s] for namespace [%s] with err: %v",
			templateFileName, namespace, err)
		err := errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to read config template from path [%s] for namespace [%s] with err: %v",
			templateFileName, namespace, err)
		c.metrics.TemplateReadErrors.Inc()
		return "", err
	}
	logger.Debugf(ctx, "successfully read template config file [%s]", templateFileName)

	// 2) substitute templatized variables with their resolved values
	// First, add the special case namespace template which is always substituted by the system
	// rather than fetched via a user-specified source.
	templateValues[fmt.Sprintf(templateVariableFormat, namespaceVariable)] = namespace
	templateValues[fmt.Sprintf(templateVariableFormat, projectVariable)] = project.Id
	templateValues[fmt.Sprintf(templateVariableFormat, domainVariable)] = domain.Id

	var k8sManifest = string(template)
	for templateKey, templateValue := range customTemplateValues {
		k8sManifest = strings.Replace(k8sManifest, templateKey, templateValue, replaceAllInstancesOfString)
	}
	// Replace remaining template variables from domain specific defaults.
	for templateKey, templateValue := range templateValues {
		k8sManifest = strings.Replace(k8sManifest, templateKey, templateValue, replaceAllInstancesOfString)
	}

	return k8sManifest, nil
}

// createPatch create 3-way merge patch of current object, original object (retrieved from last applied annotation), and the modification
// for native k8s resource, strategic merge patch is used
// for custom resource, json merge patch is used
// heavily inspired by kubectl's patcher
func (c *controller) createPatch(gvk schema.GroupVersionKind, currentObj *unstructured.Unstructured, modified []byte, namespace string) ([]byte, types.PatchType, error) {
	current, err := k8sruntime.Encode(unstructured.UnstructuredJSONScheme, currentObj)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encode [%+v] in namespace [%s] to json with err: %v",
			currentObj.GetKind(), namespace, err)
	}

	original, err := getLastApplied(currentObj)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get original resource [%+v] in namespace [%s] with err: %v",
			currentObj.GetKind(), namespace, err)
	}

	var patch []byte
	patchType := types.StrategicMergePatchType
	obj, err := scheme.Scheme.New(gvk)
	switch {
	case err == nil:
		lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(obj)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create lookup patch meta for [%+v] in namespace [%s] with err: %v",
				currentObj.GetKind(), namespace, err)
		}

		patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create 3 way merge patch for resource [%+v] in namespace [%s] with err: %v\noriginal:\n%s\nmodified:\n%s\ncurrent:\n%s",
				currentObj.GetKind(), namespace, err, original, modified, current)
		}
	case k8sruntime.IsNotRegisteredError(err):
		patchType = types.MergePatchType
		preconditions := []mergepatch.PreconditionFunc{mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"), mergepatch.RequireMetadataKeyUnchanged("name")}
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create 3 way json merge patch for resource [%+v] in namespace [%s] with err: %v",
				currentObj.GetKind(), namespace, err)
		}
	default:
		return nil, "", fmt.Errorf("failed to create get instance of versioned object [%+v] in namespace [%s] with err: %v",
			currentObj.GetKind(), namespace, err)

	}

	if string(patch) == noChange {
		// not modified
		return patch, patchType, nil
	}

	if len(currentObj.GetResourceVersion()) > 0 {
		patch, err = addResourceVersion(patch, currentObj.GetResourceVersion())
		if err != nil {
			return nil, "", fmt.Errorf("failed adding resource version for object [%+v] in namespace [%s] with err: %v",
				currentObj.GetKind(), namespace, err)
		}
	}

	return patch, patchType, nil
}

type syncNamespaceProcessor = func(ctx context.Context, project *admin.Project, domain *admin.Domain,
	namespace NamespaceName) (ResourceSyncStats, error)

// newSyncNamespacer returns a closure that can be used to sync a namespace with the given default template values and domain-specific template values
func (c *controller) newSyncNamespacer(templateValues templateValuesType, domainTemplateValues map[string]templateValuesType) syncNamespaceProcessor {
	return func(ctx context.Context, project *admin.Project, domain *admin.Domain,
		namespace NamespaceName) (ResourceSyncStats, error) {
		logger.Debugf(ctx, "sync namespacer is processing namespace: '%s' for project '%s' and domain '%s' and org '%s'",
			namespace, project.Id, domain.Id, project.Org)
		customTemplateValues, err := c.getCustomTemplateValues(
			ctx, project.Org, project.Id, domain.Id, domainTemplateValues[domain.Id])
		if err != nil {
			logger.Errorf(ctx, "Failed to get custom template values for %s with err: %v", namespace, err)
			return ResourceSyncStats{}, err
		}

		// syncNamespace actually mutates the templateValues, so we copy it for safe concurrent access.
		templateValuesCopy := maps.Clone(templateValues)
		newStats, err := c.syncNamespace(ctx, project, domain, namespace, templateValuesCopy, customTemplateValues)
		if err != nil {
			logger.Warningf(ctx, "Failed to create cluster resources for namespace [%s] with err: %v", namespace, err)
			c.metrics.ResourceAddErrors.Inc()
			return ResourceSyncStats{}, err
		}
		c.metrics.ResourcesAdded.Inc()
		logger.Debugf(ctx, "successfully created kubernetes resources for '%s' for project '%s' and domain '%s' and org '%s'",
			namespace, project.Id, domain.Id, project.Org)
		return newStats, nil
	}
}

// Report to self serve service that the namespaces have been updated
func (c *controller) reportClusterResourceStateUpdated(ctx context.Context, orgs sets.Set[string], state plugin.ClusterResourceState) {
	clusterResourcePlugin := plugins.Get[plugin.ClusterResourcePlugin](c.pluginRegistry, plugins.PluginIDClusterResource)
	if orgs.Has("") {
		// TODO: Investigate why this is happening and remove this once the root cause is fixed
		logger.Debugf(ctx, "orgs set contains empty string, removing it to avoid proto validation error")
		orgs.Delete("")
	}
	// Report cluster resource state updated for orgs in batches
	orgsList := orgs.UnsortedList()
	for i := 0; i < len(orgsList); i += reportNamespaceProvisionedBatchSize {
		batchOrgs := orgsList[i:min(i+reportNamespaceProvisionedBatchSize, len(orgsList))]
		logger.Debugf(ctx, "reporting [%v] for orgs [%v]", state, batchOrgs)
		batchUpdateProvisionedRequest := &plugin.BatchUpdateClusterResourceStateInput{
			OrgsName:             batchOrgs,
			ClusterName:          c.config.ClusterResourceConfiguration().GetClusterName(),
			ClusterResourceState: state,
		}
		_, batchUpdateProvisionedErrors, err := clusterResourcePlugin.BatchUpdateClusterResourceState(ctx, batchUpdateProvisionedRequest)
		if err != nil {
			c.metrics.ReportClusterResourceStateUpdatedErrors.Inc()
			logger.Errorf(ctx, "failed to report [%v] for orgs [%v] with err: %v", state, batchOrgs, err)
			// we don't error out here, in order to not block other namespaces from being reported
			continue
		}
		if len(batchUpdateProvisionedErrors) > 0 {
			c.metrics.ReportClusterResourceStateUpdatedErrors.Inc()
			for _, updateProvisionedError := range batchUpdateProvisionedErrors {
				logger.Errorf(ctx, "failed to report [%v] for org [%s] with err: %s %v", state, updateProvisionedError.OrgName, updateProvisionedError.ErrorCode, updateProvisionedError.ErrorMessage)
			}
			continue
		}
		c.metrics.ReportClusterResourceStateUpdated.Inc()
	}
}

func (c *controller) processProjectsInBatch(ctx context.Context, projects []*admin.Project, processor syncNamespaceProcessor) []ResourceSyncStats {
	numberOfDomains := len(projects[0].Domains)
	statsEntries := make([]ResourceSyncStats, len(projects)*numberOfDomains)
	batchSize := c.config.ClusterResourceConfiguration().GetUnionProjectSyncConfig().BatchSize
	for i := 0; i < len(projects); i += batchSize {
		logger.Debugf(ctx, "processing projects batch {%d...%d}", i, i+batchSize)
		processProjectsGroup, processProjectsCtx := errgroup.WithContext(ctx)
		for projectIdx, project := range projects[i:min(i+batchSize, len(projects))] {
			var lastFormattedNamespace NamespaceName
			for domainIdx, domain := range project.Domains {
				// redefine variables for loop closure
				i := i
				org := project.Org
				project := project
				projectIdx := projectIdx
				domain := domain
				domainIdx := domainIdx
				namespace := common.GetNamespaceName(c.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), org, project.Id, domain.Id)
				if lastFormattedNamespace == namespace {
					// This is a case where the namespace isn't customized based on project + domain + org, but rather project + org only
					continue
				}
				lastFormattedNamespace = namespace

				processProjectsGroup.Go(func() error {
					resourceStats, err := processor(processProjectsCtx, project, domain, namespace)
					if err != nil {
						logger.Warningf(ctx, "failed to create cluster resources for namespace [%s] with err: %v", namespace, err)
						c.metrics.ResourceAddErrors.Inc()
						// we don't error out here, in order to not block other namespaces from being synced
					} else {
						c.metrics.ResourcesAdded.Inc()
						statsEntries[(i+projectIdx)*numberOfDomains+domainIdx] = resourceStats
						logger.Debugf(ctx, "Completed cluster resource creation loop for namespace [%s] with stats: [%+v]", namespace, resourceStats)
					}
					return nil
				})
			}
		}
		if err := processProjectsGroup.Wait(); err != nil {
			logger.Warningf(ctx, "failed to process projects batch {%d...%d} when syncing namespaces with: %v",
				i, i+batchSize, err)
		}
		logger.Debugf(ctx, "processed projects batch {%d...%d} when syncing namespaces", i, i+batchSize)
	}
	return statsEntries
}

// processProjects accepts a generic project processor and applies it to all projects in the input list,
// optionally batching and parallelizing their processing for greater perf.
func (c *controller) processProjects(ctx context.Context, stats ResourceSyncStats, projects []*admin.Project, processor syncNamespaceProcessor) (ResourceSyncStats, error) {
	if len(projects) == 0 {
		// Nothing to do
		return stats, nil
	}
	// resourceCreatedOrgs/resourceDeletedOrgs is used to track which orgs have had resources created/deleted
	// This is only used in Serverless (self-serve) mode
	resourceCreatedOrgs := sets.Set[string]{}
	resourceDeletedOrgs := sets.Set[string]{}
	if c.config.ClusterResourceConfiguration().GetUnionProjectSyncConfig().BatchSize == 0 {
		logger.Debugf(ctx, "processing projects serially")
		for _, project := range projects {
			for _, domain := range project.Domains {
				namespace := common.GetNamespaceName(c.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), project.Org, project.Id, domain.Name)
				newStats, err := processor(ctx, project, domain, namespace)
				if err != nil {
					return stats, err
				}
				stats.Add(newStats)
				if newStats.Created > 0 {
					resourceCreatedOrgs.Insert(project.Org)
				}
				if newStats.Deleted > 0 {
					resourceDeletedOrgs.Insert(project.Org)
				}
				logger.Infof(ctx, "Completed cluster resource creation loop for namespace [%s] with stats: [%+v]", namespace, newStats)
			}
		}
		if c.config.ClusterResourceConfiguration().IsSelfServe() {
			c.reportClusterResourceStateUpdated(ctx, resourceCreatedOrgs, plugin.ClusterResourceStateProvisioned)
			c.reportClusterResourceStateUpdated(ctx, resourceDeletedOrgs, plugin.ClusterResourceStateDeprovisioned)
		}
		return stats, nil
	}

	logger.Debugf(ctx, "processing projects in parallel with batch size: %d", c.config.ClusterResourceConfiguration().GetUnionProjectSyncConfig().BatchSize)
	statsEntries := c.processProjectsInBatch(ctx, projects, processor)
	// Collect orgs that have had resources created/deleted
	numberOfDomains := len(projects[0].Domains)
	for projectIdx, project := range projects {
		for domainIdx := range project.Domains {
			idx := projectIdx*numberOfDomains + domainIdx
			if statsEntries[idx].Created > 0 {
				resourceCreatedOrgs.Insert(project.Org)
			}
			if statsEntries[idx].Deleted > 0 {
				resourceDeletedOrgs.Insert(project.Org)
			}
		}
	}
	if c.config.ClusterResourceConfiguration().IsSelfServe() {
		c.reportClusterResourceStateUpdated(ctx, resourceCreatedOrgs, plugin.ClusterResourceStateProvisioned)
		c.reportClusterResourceStateUpdated(ctx, resourceDeletedOrgs, plugin.ClusterResourceStateDeprovisioned)
	}
	statsSummary := lo.Reduce(statsEntries, func(existing ResourceSyncStats, item ResourceSyncStats, _ int) ResourceSyncStats {
		existing.Add(item)
		return existing
	}, ResourceSyncStats{})
	return statsSummary, nil
}

func (c *controller) Sync(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			c.metrics.Panics.Inc()
			logger.Warningf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()
	c.metrics.SyncStarted.Inc()
	logger.Debugf(ctx, "Running an invocation of ClusterResource Sync")

	// First handle ensuring k8s resources for active projects
	projects, err := c.adminDataProvider.GetProjects(ctx)
	if err != nil {
		return err
	}
	var errs = make([]error, 0)
	templateValues, err := populateTemplateValues(c.config.ClusterResourceConfiguration().GetTemplateData())
	if err != nil {
		logger.Warningf(ctx, "Failed to get templatized values specified in config: %v", err)
		errs = append(errs, err)
	}
	domainTemplateValues, err := populateDefaultTemplateValues(c.config.ClusterResourceConfiguration().GetCustomTemplateData())
	if err != nil {
		logger.Warningf(ctx, "Failed to get domain-specific templatized values specified in config: %v", err)
		errs = append(errs, err)
	}

	stats := ResourceSyncStats{}
	syncNamespaceProcessorFunc := c.newSyncNamespacer(templateValues, domainTemplateValues)
	stats, err = c.processProjects(ctx, stats, projects.GetProjects(), syncNamespaceProcessorFunc)
	if err != nil {
		logger.Warningf(ctx, "Failed to process projects when syncing namespaces: %v", err)
		return err
	}

	// Second, handle deleting k8s namespaces for archived projects
	if c.config.ClusterResourceConfiguration().GetUnionProjectSyncConfig().CleanupNamespace {
		archivedProjects, err := c.adminDataProvider.GetArchivedProjects(ctx)
		if err != nil {
			logger.Warningf(ctx, "failed to get archived projects: %v", err)
			return err
		}
		stats, err = c.processProjects(ctx, stats, archivedProjects.GetProjects(), c.deleteNamespace)
		if err != nil {
			logger.Warningf(ctx, "Failed to process projects when deleting namespaces: %v", err)
			return err
		}
	}

	logger.Infof(ctx, "Completed cluster resource creation loop with stats: [%+v]", stats)

	if len(errs) > 0 {
		c.metrics.SyncErrors.Add(float64(len(errs)))
		return errors.NewCollectedFlyteAdminError(codes.Internal, errs)
	}

	c.metrics.SyncsCompleted.Inc()
	return nil
}

func (c *controller) Run() {
	ctx := context.Background()
	logger.Info(ctx, "Running ClusterResourceController")
	interval := c.config.ClusterResourceConfiguration().GetRefreshInterval()
	wait.Forever(func() {
		err := c.Sync(ctx)
		if err != nil {
			logger.Errorf(ctx, "Failed cluster resource creation loop with: %v", err)
		} else {
			logger.Infof(ctx, "Successfully completed cluster resource creation loop")
		}
	}, interval)
}

func newMetrics(scope promutils.Scope) controllerMetrics {
	return controllerMetrics{
		Scope:      scope,
		SyncErrors: scope.MustNewCounter("sync_errors", "overall count of errors that occurred within a 'sync' method"),
		SyncStarted: scope.MustNewCounter("k8s_resource_syncs",
			"overall count of the number of invocations of the resource controller 'sync' method"),
		SyncsCompleted: scope.MustNewCounter("sync_success", "overall count of successful invocations of the resource controller 'sync' method"),
		KubernetesResourcesCreated: scope.MustNewCounter("k8s_resources_created",
			"overall count of successfully created resources in kubernetes"),
		KubernetesResourcesCreateErrors: scope.MustNewCounter("k8s_resource_create_errors",
			"overall count of errors encountered attempting to create resources in kubernetes"),
		KubernetesNamespaceDeleteErrors: scope.MustNewCounter("k8s_namespace_delete_errors", "overall count of errors encountered attempting to delete namespaces in kubernetes"),
		ResourcesAdded: scope.MustNewCounter("resources_added",
			"overall count of successfully added resources for namespaces"),
		NamespacesDeleted: scope.MustNewCounter("namespaces_deleted", "overall count of successfully deleted namespaces"),
		ResourceAddErrors: scope.MustNewCounter("resource_add_errors",
			"overall count of errors encountered creating resources for namespaces"),
		TemplateReadErrors: scope.MustNewCounter("template_read_errors",
			"errors encountered reading the yaml template file from the local filesystem"),
		TemplateDecodeErrors: scope.MustNewCounter("template_decode_errors",
			"errors encountered trying to decode yaml template into k8s go struct"),
		AppliedTemplateExists: scope.MustNewCounter("applied_template_exists",
			"Number of times the system to tried to apply an uncached resource the kubernetes reported as "+
				"already existing"),
		TemplateUpdateErrors: scope.MustNewCounter("template_update_errors",
			"Number of times an attempt at updating an already existing kubernetes resource with a template"+
				"file failed"),
		Panics: scope.MustNewCounter("panics",
			"overall count of panics encountered in primary ClusterResourceController loop"),
		ReportClusterResourceStateUpdated: scope.MustNewCounter("report_cluster_resource_state_updated",
			"overall count of successful reports of cluster resource state updated"),
		ReportClusterResourceStateUpdatedErrors: scope.MustNewCounter("report_cluster_resource_state_updated_errors",
			"overall count of errors encountered reporting cluster resource state updated"),
	}
}

func NewClusterResourceController(adminDataProvider interfaces.FlyteAdminDataProvider, listTargets executionclusterIfaces.ListTargetsInterface, scope promutils.Scope, pluginRegistry *plugins.Registry) Controller {
	cfg := runtime.NewConfigurationProvider()
	return &controller{
		adminDataProvider: adminDataProvider,
		config:            cfg,
		listTargets:       listTargets,
		poller:            make(chan struct{}),
		metrics:           newMetrics(scope),
		pluginRegistry:    pluginRegistry,
	}
}

func NewClusterResourceControllerFromConfig(ctx context.Context, scope promutils.Scope, configuration runtimeInterfaces.Configuration, pluginRegistry *plugins.Registry) (Controller, error) {
	initializationErrorCounter := scope.MustNewCounter(
		"flyteclient_initialization_error",
		"count of errors encountered initializing a flyte client from kube config")
	var listTargetsProvider executionclusterIfaces.ListTargetsInterface
	var err error
	if len(configuration.ClusterConfiguration().GetClusterConfigs()) == 0 {
		serverConfig := config.GetConfig()
		listTargetsProvider, err = impl.NewInCluster(initializationErrorCounter, serverConfig.KubeConfig, serverConfig.Master)
	} else {
		listTargetsProvider, err = impl.NewListTargets(initializationErrorCounter, impl.NewExecutionTargetProvider(), configuration.ClusterConfiguration())
	}
	if err != nil {
		return nil, err
	}

	var adminDataProvider interfaces.FlyteAdminDataProvider
	if configuration.ClusterResourceConfiguration().IsStandaloneDeployment() {
		clientSet, err := admin2.ClientSetBuilder().WithConfig(admin2.GetConfig(ctx)).Build(ctx)
		if err != nil {
			return nil, err
		}
		adminDataProvider = impl2.NewAdminServiceDataProvider(clientSet.AdminClient())
	} else {
		dbConfig := runtime.NewConfigurationProvider().ApplicationConfiguration().GetDbConfig()
		logConfig := logger.GetConfig()

		db, err := repositories.GetDB(ctx, dbConfig, logConfig)
		if err != nil {
			return nil, err
		}
		dbScope := scope.NewSubScope("db")

		repo := repositories.NewGormRepo(
			db, errors2.NewPostgresErrorTransformer(dbScope.NewSubScope("errors")), dbScope)

		dataStorageClient, err := storage.NewDataStore(storage.GetConfig(), scope.NewSubScope("storage"))
		if err != nil {
			logger.Errorf(ctx, "Failed to create data storage client: %v", err)
			return nil, err
		}

		defaultProjectConfigurationPlugin := projectConfigurationPlugin.NewDefaultProjectConfigurationPlugin()
		pluginRegistry.RegisterDefault(plugins.PluginIDProjectConfiguration, defaultProjectConfigurationPlugin)
		configurationManager, err := configurations.NewConfigurationManager(ctx, repo, configuration, dataStorageClient, pluginRegistry, configurations.ShouldNotBootstrapOrUpdateDefault)
		if err != nil {
			logger.Errorf(ctx, "Failed to create configuration manager: %v", err)
			return nil, err
		}
		resourceManager := resources.ConfigureResourceManager(repo, configuration.ApplicationConfiguration(), configurationManager)
		adminDataProvider = impl2.NewDatabaseAdminDataProvider(repo, configuration, resourceManager)
	}

	pluginRegistry.RegisterDefault(plugins.PluginIDClusterResource, plugin.NewNoopClusterResourcePlugin())
	return NewClusterResourceController(adminDataProvider, listTargetsProvider, scope, pluginRegistry), nil
}
