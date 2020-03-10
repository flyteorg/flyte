package clusterresource

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/lyft/flyteadmin/pkg/manager/impl/resources"
	managerinterfaces "github.com/lyft/flyteadmin/pkg/manager/interfaces"

	"github.com/lyft/flyteadmin/pkg/executioncluster/interfaces"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc/codes"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteadmin/pkg/runtime"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const namespaceVariable = "namespace"
const templateVariableFormat = "{{ %s }}"
const replaceAllInstancesOfString = -1

// The clusterresource Controller manages applying desired templatized kubernetes resource files as resources
// in the execution kubernetes cluster.
type Controller interface {
	Sync(ctx context.Context) error
	Run()
}

type controllerMetrics struct {
	Scope                           promutils.Scope
	SyncStarted                     prometheus.Counter
	KubernetesResourcesCreated      prometheus.Counter
	KubernetesResourcesCreateErrors prometheus.Counter
	ResourcesAdded                  prometheus.Counter
	ResourceAddErrors               prometheus.Counter
	TemplateReadErrors              prometheus.Counter
	TemplateDecodeErrors            prometheus.Counter
	AppliedTemplateExists           prometheus.Counter
	TemplateUpdateErrors            prometheus.Counter
	Panics                          prometheus.Counter
}

type FileName = string
type NamespaceName = string
type LastModTimeCache = map[FileName]time.Time
type NamespaceCache = map[NamespaceName]LastModTimeCache

type templateValuesType = map[string]string

type controller struct {
	db                     repositories.RepositoryInterface
	config                 runtimeInterfaces.Configuration
	executionCluster       interfaces.ClusterInterface
	resourceManager        managerinterfaces.ResourceInterface
	poller                 chan struct{}
	metrics                controllerMetrics
	lastAppliedTemplateDir string
	// Map of [namespace -> [templateFileName -> last modified time]]
	appliedTemplates NamespaceCache
}

var descCreatedAtSortParam, _ = common.NewSortParameter(admin.Sort{
	Direction: admin.Sort_DESCENDING,
	Key:       "created_at",
})

func (c *controller) templateAlreadyApplied(namespace NamespaceName, templateFile os.FileInfo) bool {
	namespacedAppliedTemplates, ok := c.appliedTemplates[namespace]
	if !ok {
		// There is no record of this namespace altogether.
		return false
	}
	timestamp, ok := namespacedAppliedTemplates[templateFile.Name()]
	if !ok {
		// There is no record of this file having ever been applied.
		return false
	}
	// The applied template file could have been modified, in which case we will need to apply it once more.
	return timestamp.Equal(templateFile.ModTime())
}

// Given a map of templatized variable names -> data source, this function produces an output that maps the same
// variable names to their fully resolved values (from the specified data source).
func populateTemplateValues(data map[string]runtimeInterfaces.DataSource) (templateValuesType, error) {
	templateValues := make(templateValuesType, len(data))
	collectedErrs := make([]error, 0)
	for templateVar, dataSource := range data {
		if templateVar == namespaceVariable {
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
	ctx context.Context, project, domain string, domainTemplateValues templateValuesType) (templateValuesType, error) {
	if len(domainTemplateValues) == 0 {
		domainTemplateValues = make(templateValuesType)
	}
	customTemplateValues := make(templateValuesType)
	for key, value := range domainTemplateValues {
		customTemplateValues[key] = value
	}
	collectedErrs := make([]error, 0)
	// All override values saved in the database take precedence over the domain-specific defaults.
	resource, err := c.resourceManager.GetResource(ctx, managerinterfaces.ResourceRequest{
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	if err != nil {
		if _, ok := err.(errors.FlyteAdminError); !ok || err.(errors.FlyteAdminError).Code() != codes.NotFound {
			collectedErrs = append(collectedErrs, err)
		}
	}
	if resource != nil && resource.Attributes != nil && resource.Attributes.GetClusterResourceAttributes() != nil {
		for templateKey, templateValue := range resource.Attributes.GetClusterResourceAttributes().Attributes {
			customTemplateValues[fmt.Sprintf(templateVariableFormat, templateKey)] = templateValue
		}
	}
	if len(collectedErrs) > 0 {
		return nil, errors.NewCollectedFlyteAdminError(codes.InvalidArgument, collectedErrs)
	}
	return customTemplateValues, nil
}

// This function loops through the kubernetes resource template files in the configured template directory.
// For each unapplied template file (wrt the namespace) this func attempts to
//   1) read the template file
//   2) substitute templatized variables with their resolved values
//   3) decode the output of the above into a kubernetes resource
//   4) create the resource on the kubernetes cluster and cache successful outcomes
func (c *controller) syncNamespace(ctx context.Context, namespace NamespaceName,
	templateValues, customTemplateValues templateValuesType) error {
	templateDir := c.config.ClusterResourceConfiguration().GetTemplatePath()
	if c.lastAppliedTemplateDir != templateDir {
		// Invalidate all caches
		c.lastAppliedTemplateDir = templateDir
		c.appliedTemplates = make(NamespaceCache)
	}
	templateFiles, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"Failed to read config template dir [%s] for namespace [%s] with err: %v",
			namespace, templateDir, err)
	}

	collectedErrs := make([]error, 0)
	for _, templateFile := range templateFiles {
		templateFileName := templateFile.Name()
		if filepath.Ext(templateFileName) != ".yaml" {
			// nothing to do.
			logger.Infof(ctx, "syncing namespace [%s]: ignoring unrecognized filetype [%s]",
				namespace, templateFile.Name())
			continue
		}

		if c.templateAlreadyApplied(namespace, templateFile) {
			// nothing to do.
			logger.Infof(ctx, "syncing namespace [%s]: templateFile [%s] already applied, nothing to do.", namespace, templateFile.Name())
			continue
		}

		// 1) read the template file
		template, err := ioutil.ReadFile(path.Join(templateDir, templateFileName))
		if err != nil {
			logger.Warningf(ctx,
				"failed to read config template from path [%s] for namespace [%s] with err: %v",
				templateFileName, namespace, err)
			err := errors.NewFlyteAdminErrorf(
				codes.Internal, "failed to read config template from path [%s] for namespace [%s] with err: %v",
				templateFileName, namespace, err)
			collectedErrs = append(collectedErrs, err)
			c.metrics.TemplateReadErrors.Inc()
			continue
		}
		logger.Debugf(ctx, "successfully read template config file [%s]", templateFileName)

		// 2) substitute templatized variables with their resolved values
		// First, add the special case namespace template which is always substituted by the system
		// rather than fetched via a user-specified source.
		templateValues[fmt.Sprintf(templateVariableFormat, namespaceVariable)] = namespace
		var config = string(template)
		for templateKey, templateValue := range templateValues {
			config = strings.Replace(config, templateKey, templateValue, replaceAllInstancesOfString)
		}
		// Replace remaining template variables from domain specific defaults.
		for templateKey, templateValue := range customTemplateValues {
			config = strings.Replace(config, templateKey, templateValue, replaceAllInstancesOfString)
		}

		// 3) decode the kubernetes resource template file into an actual resource object
		decode := scheme.Codecs.UniversalDeserializer().Decode
		k8sObj, _, err := decode([]byte(config), nil, nil)
		if err != nil {
			logger.Warningf(ctx, "Failed to decode config template [%s] for namespace [%s] into a kubernetes object with err: %v",
				templateFileName, namespace, err)
			err := errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"Failed to decode namespace config template [%s] for namespace [%s] into a kubernetes object with err: %v",
				templateFileName, namespace, err)
			collectedErrs = append(collectedErrs, err)
			c.metrics.TemplateDecodeErrors.Inc()
			continue
		}

		// 4) create the resource on the kubernetes cluster and cache successful outcomes
		if _, ok := c.appliedTemplates[namespace]; !ok {
			c.appliedTemplates[namespace] = make(LastModTimeCache)
		}
		for _, target := range c.executionCluster.GetAllValidTargets() {
			k8sObjCopy := k8sObj.DeepCopyObject()
			logger.Debugf(ctx, "Attempting to create resource [%+v] in cluster [%v] for namespace [%s]",
				k8sObj.GetObjectKind().GroupVersionKind().Kind, target.ID, namespace)
			err = target.Client.Create(ctx, k8sObjCopy)
			if err != nil {
				if k8serrors.IsAlreadyExists(err) {
					logger.Debugf(ctx, "Type [%+v] in namespace [%s] already exists - attempting update instead",
						k8sObj.GetObjectKind().GroupVersionKind().Kind, namespace)
					c.metrics.AppliedTemplateExists.Inc()
					// Use a strategic-merge-patch to mimic `kubectl apply` behavior.
					// Kubectl defaults to using the StrategicMergePatch strategy.
					// However the controller-runtime only has an implementation for MergePatch which we were formerly
					// using but failed to actually always merge resources in the Patch call.
					err = target.Client.Patch(ctx, k8sObjCopy, StrategicMergeFrom(k8sObjCopy))
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Infof(ctx, "Failed to update resource [%+v] in namespace [%s] with err :%v",
							k8sObj.GetObjectKind().GroupVersionKind().Kind, namespace, err)
						collectedErrs = append(collectedErrs, err)
					} else {
						logger.Debugf(ctx, "Successfully updated resource [%+v] in namespace [%s]",
							k8sObj.GetObjectKind().GroupVersionKind().Kind, namespace)
					}
					c.appliedTemplates[namespace][templateFile.Name()] = templateFile.ModTime()
				} else {
					c.metrics.KubernetesResourcesCreateErrors.Inc()
					logger.Warningf(ctx, "Failed to create kubernetes object from config template [%s] for namespace [%s] with err: %v",
						templateFileName, namespace, err)
					err := errors.NewFlyteAdminErrorf(codes.Internal,
						"Failed to create kubernetes object from config template [%s] for namespace [%s] with err: %v",
						templateFileName, namespace, err)
					collectedErrs = append(collectedErrs, err)
				}
			} else {
				logger.Infof(ctx, "Created resource [%+v] for namespace [%s] in kubernetes",
					k8sObj.GetObjectKind().GroupVersionKind().Kind, namespace)
				c.metrics.KubernetesResourcesCreated.Inc()
				c.appliedTemplates[namespace][templateFile.Name()] = templateFile.ModTime()
			}
		}
	}
	if len(collectedErrs) > 0 {
		return errors.NewCollectedFlyteAdminError(codes.Internal, collectedErrs)
	}
	return nil
}

func (c *controller) Sync(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			c.metrics.Panics.Inc()
			logger.Warningf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()
	c.metrics.SyncStarted.Inc()
	logger.Info(ctx, "Running an invocation of ClusterResource Sync")

	// Prefer to sync projects most newly created to ensure their resources get created first when other resources exist.
	projects, err := c.db.ProjectRepo().ListAll(ctx, descCreatedAtSortParam)
	if err != nil {
		return err
	}
	domains := c.config.ApplicationConfiguration().GetDomainsConfig()
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

	for _, project := range projects {
		for _, domain := range *domains {
			namespace := common.GetNamespaceName(c.config.NamespaceMappingConfiguration().GetNamespaceMappingConfig(), project.Identifier, domain.Name)
			customTemplateValues, err := c.getCustomTemplateValues(
				ctx, project.Identifier, domain.ID, domainTemplateValues[domain.ID])
			if err != nil {
				logger.Warningf(ctx, "Failed to get custom template values for %s with err: %v", namespace, err)
				errs = append(errs, err)
			}
			err = c.syncNamespace(ctx, namespace, templateValues, customTemplateValues)
			if err != nil {
				logger.Warningf(ctx, "Failed to create cluster resources for namespace [%s] with err: %v", namespace, err)
				c.metrics.ResourceAddErrors.Inc()
				errs = append(errs, err)
			} else {
				logger.Infof(ctx, "Created cluster resources for namespace [%s] in kubernetes", namespace)
				c.metrics.ResourcesAdded.Inc()
				logger.Debugf(ctx, "Successfully created kubernetes resources for [%s-%s]", project.Identifier, domain.ID)
			}
		}
	}
	if len(errs) > 0 {
		return errors.NewCollectedFlyteAdminError(codes.Internal, errs)
	}
	return nil
}

func (c *controller) Run() {
	ctx := context.Background()
	logger.Infof(ctx, "Running ClusterResourceController")
	interval := c.config.ClusterResourceConfiguration().GetRefreshInterval()
	wait.Forever(func() {
		err := c.Sync(ctx)
		if err != nil {
			logger.Warningf(ctx, "Failed cluster resource creation loop with: %v", err)
		}
	}, interval)
}

func newMetrics(scope promutils.Scope) controllerMetrics {
	return controllerMetrics{
		Scope: scope,
		SyncStarted: scope.MustNewCounter("k8s_resource_syncs",
			"overall count of the number of invocations of the resource controller 'sync' method"),
		KubernetesResourcesCreated: scope.MustNewCounter("k8s_resources_created",
			"overall count of successfully created resources in kubernetes"),
		KubernetesResourcesCreateErrors: scope.MustNewCounter("k8s_resource_create_errors",
			"overall count of errors encountered attempting to create resources in kubernetes"),
		ResourcesAdded: scope.MustNewCounter("resources_added",
			"overall count of successfully added resources for namespaces"),
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
	}
}

func NewClusterResourceController(db repositories.RepositoryInterface, executionCluster interfaces.ClusterInterface, scope promutils.Scope) Controller {
	config := runtime.NewConfigurationProvider()
	return &controller{
		db:               db,
		config:           config,
		executionCluster: executionCluster,
		resourceManager:  resources.NewResourceManager(db, config.ApplicationConfiguration()),
		poller:           make(chan struct{}),
		metrics:          newMetrics(scope),
		appliedTemplates: make(map[string]map[string]time.Time),
	}
}
