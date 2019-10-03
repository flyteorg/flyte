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

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/lyft/flyteadmin/pkg/executioncluster"

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

const namespaceFormat = "%s-%s"
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

type controller struct {
	db                     repositories.RepositoryInterface
	config                 runtimeInterfaces.Configuration
	executionCluster       executioncluster.ClusterInterface
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

func populateTemplateValues(namespace NamespaceName, data map[string]runtimeInterfaces.DataSource) (map[string]string, error) {
	templateValues := make(map[string]string, len(data)+1)
	// First, add the special case namespace template which is always substituted by the system
	// rather than fetched via a user-specified source.
	templateValues[fmt.Sprintf(templateVariableFormat, namespaceVariable)] = namespace

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

// This function loops through the kubernetes resource template files in the configured template directory.
// For each unapplied template file (wrt the namespace) this func attempts to
//   1) read the template file
//   2) substitute templatized variables with their resolved values
//   3) decode the output of the above into a kubernetes resource
//   4) create the resource on the kubernetes cluster and cache successful outcomes
func (c *controller) syncNamespace(ctx context.Context, namespace NamespaceName) error {
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
	// Template values are lazy initialized only iff a new template file must be applied for this namespace.
	var templateValues map[string]string
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
		if len(templateValues) == 0 {
			// Compute templatized values.
			var err error
			templateValues, err = populateTemplateValues(namespace, c.config.ClusterResourceConfiguration().GetTemplateData())
			if err != nil {
				return err
			}
		}
		var config = string(template)
		for templateKey, templateValue := range templateValues {
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
			err = target.Client.Create(ctx, k8sObjCopy)
			if err != nil {
				if k8serrors.IsAlreadyExists(err) {
					logger.Debugf(ctx, "Resource [%+v] in namespace [%s] already exists - attempting update instead",
						k8sObj.GetObjectKind().GroupVersionKind().Kind, namespace)
					c.metrics.AppliedTemplateExists.Inc()
					err = target.Client.Patch(ctx, k8sObjCopy, client.MergeFrom(k8sObjCopy))
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Infof(ctx, "Failed to update resource [%+v] in namespace [%s] with err :%v",
							k8sObj.GetObjectKind().GroupVersionKind().Kind, namespace, err)
						collectedErrs = append(collectedErrs, err)
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

	// Prefer to sync projects most newly created to ensure their resources get created first when other resources exist.
	projects, err := c.db.ProjectRepo().ListAll(ctx, descCreatedAtSortParam)
	if err != nil {
		return err
	}
	domains := c.config.ApplicationConfiguration().GetDomainsConfig()
	var errs = make([]error, 0)
	for _, project := range projects {
		for _, domain := range *domains {
			namespace := fmt.Sprintf(namespaceFormat, project.Identifier, domain.Name)
			err := c.syncNamespace(ctx, namespace)
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

func NewClusterResourceController(db repositories.RepositoryInterface, executionCluster executioncluster.ClusterInterface, scope promutils.Scope) Controller {
	config := runtime.NewConfigurationProvider()
	return &controller{
		db:               db,
		config:           config,
		executionCluster: executionCluster,
		poller:           make(chan struct{}),
		metrics:          newMetrics(scope),
		appliedTemplates: make(map[string]map[string]time.Time),
	}
}
