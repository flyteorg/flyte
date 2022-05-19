package register

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/md5" //#nosec
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	errors2 "github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	g "github.com/flyteorg/flytectl/pkg/github"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/utils"

	"github.com/flyteorg/flytectl/cmd/config"
	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/printer"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/google/go-github/v42/github"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
)

// Variable define in serialized proto that needs to be replace in registration time
const registrationProjectPattern = "{{ registration.project }}"
const registrationDomainPattern = "{{ registration.domain }}"
const registrationVersionPattern = "{{ registration.version }}"

// Additional variable define in fast serialized proto that needs to be replace in registration time
const registrationRemotePackagePattern = "{{ .remote_package_path }}"
const registrationDestDirPattern = "{{ .dest_dir }}"

// All supported extensions for compress
var supportedExtensions = []string{".tar", ".tgz", ".tar.gz"}

// All supported extensions for gzip compress
var validGzipExtensions = []string{".tgz", ".tar.gz"}

type SignedURLPatternMatcher = *regexp.Regexp

var (
	SignedURLPattern SignedURLPatternMatcher = regexp.MustCompile(`https://((storage\.googleapis\.com/(?P<bucket_gcs>[^/]+))|((?P<bucket_s3>[^\.]+)\.s3\.amazonaws\.com)|(.*\.blob\.core\.windows\.net/(?P<bucket_az>[^/]+)))/(?P<path>[^?]*)`)
)

type Result struct {
	Name   string
	Status string
	Info   string
}

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var Client *storage.DataStore

var httpClient HTTPClient

func init() {
	httpClient = &http.Client{}
}

var projectColumns = []printer.Column{
	{Header: "Name", JSONPath: "$.Name"},
	{Header: "Status", JSONPath: "$.Status"},
	{Header: "Additional Info", JSONPath: "$.Info"},
}

func unMarshalContents(ctx context.Context, fileContents []byte, fname string) (proto.Message, error) {
	workflowSpec := &admin.WorkflowSpec{}
	errCollection := errors2.ErrorCollection{}
	err := proto.Unmarshal(fileContents, workflowSpec)
	if err == nil {
		return workflowSpec, nil
	}

	errCollection.Append(fmt.Errorf("as a Workflow: %w", err))

	logger.Debugf(ctx, "Failed to unmarshal file %v for workflow type", fname)
	taskSpec := &admin.TaskSpec{}
	err = proto.Unmarshal(fileContents, taskSpec)
	if err == nil {
		return taskSpec, nil
	}

	errCollection.Append(fmt.Errorf("as a Task: %w", err))

	logger.Debugf(ctx, "Failed to unmarshal file %v for task type", fname)
	launchPlan := &admin.LaunchPlan{}
	err = proto.Unmarshal(fileContents, launchPlan)
	if err == nil {
		return launchPlan, nil
	}

	errCollection.Append(fmt.Errorf("as a Launchplan: %w", err))

	logger.Debugf(ctx, "Failed to unmarshal file %v for launch plan type", fname)
	return nil, fmt.Errorf("failed unmarshalling file %v. Errors: %w", fname, errCollection.ErrorOrDefault())

}

func register(ctx context.Context, message proto.Message, cmdCtx cmdCore.CommandContext, dryRun, enableSchedule bool) error {
	switch v := message.(type) {
	case *admin.LaunchPlan:
		launchPlan := message.(*admin.LaunchPlan)
		if dryRun {
			logger.Debugf(ctx, "skipping CreateLaunchPlan request (DryRun)")
			return nil
		}
		_, err := cmdCtx.AdminClient().CreateLaunchPlan(ctx,
			&admin.LaunchPlanCreateRequest{
				Id: &core.Identifier{
					ResourceType: core.ResourceType_LAUNCH_PLAN,
					Project:      config.GetConfig().Project,
					Domain:       config.GetConfig().Domain,
					Name:         launchPlan.Id.Name,
					Version:      launchPlan.Id.Version,
				},
				Spec: launchPlan.Spec,
			})
		if err != nil {
			return err
		}
		// Activate the launchplan
		if enableSchedule {
			_, err = cmdCtx.AdminClient().UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
				Id: &core.Identifier{
					Project: config.GetConfig().Project,
					Domain:  config.GetConfig().Domain,
					Name:    launchPlan.Id.Name,
					Version: launchPlan.Id.Version,
				},
				State: admin.LaunchPlanState_ACTIVE,
			})
			return err
		}
		return nil
	case *admin.WorkflowSpec:
		workflowSpec := message.(*admin.WorkflowSpec)
		if dryRun {
			logger.Debugf(ctx, "skipping CreateWorkflow request (DryRun)")
			return nil
		}
		_, err := cmdCtx.AdminClient().CreateWorkflow(ctx,
			&admin.WorkflowCreateRequest{
				Id: &core.Identifier{
					ResourceType: core.ResourceType_WORKFLOW,
					Project:      config.GetConfig().Project,
					Domain:       config.GetConfig().Domain,
					Name:         workflowSpec.Template.Id.Name,
					Version:      workflowSpec.Template.Id.Version,
				},
				Spec: workflowSpec,
			})
		return err
	case *admin.TaskSpec:
		taskSpec := message.(*admin.TaskSpec)
		if dryRun {
			logger.Debugf(ctx, "skipping CreateTask request (DryRun)")
			return nil
		}
		_, err := cmdCtx.AdminClient().CreateTask(ctx,
			&admin.TaskCreateRequest{
				Id: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      config.GetConfig().Project,
					Domain:       config.GetConfig().Domain,
					Name:         taskSpec.Template.Id.Name,
					Version:      taskSpec.Template.Id.Version,
				},
				Spec: taskSpec,
			})
		return err
	default:
		return fmt.Errorf("Failed registering unknown entity  %v", v)
	}
}

func hydrateNode(node *core.Node, version string, force bool) error {
	targetNode := node.Target
	switch v := targetNode.(type) {
	case *core.Node_TaskNode:
		taskNodeWrapper := targetNode.(*core.Node_TaskNode)
		taskNodeReference := taskNodeWrapper.TaskNode.Reference.(*core.TaskNode_ReferenceId)
		hydrateIdentifier(taskNodeReference.ReferenceId, version, force)
	case *core.Node_WorkflowNode:
		workflowNodeWrapper := targetNode.(*core.Node_WorkflowNode)
		switch workflowNodeWrapper.WorkflowNode.Reference.(type) {
		case *core.WorkflowNode_SubWorkflowRef:
			subWorkflowNodeReference := workflowNodeWrapper.WorkflowNode.Reference.(*core.WorkflowNode_SubWorkflowRef)
			hydrateIdentifier(subWorkflowNodeReference.SubWorkflowRef, version, force)
		case *core.WorkflowNode_LaunchplanRef:
			launchPlanNodeReference := workflowNodeWrapper.WorkflowNode.Reference.(*core.WorkflowNode_LaunchplanRef)
			hydrateIdentifier(launchPlanNodeReference.LaunchplanRef, version, force)
		default:
			return fmt.Errorf("unknown type %T", workflowNodeWrapper.WorkflowNode.Reference)
		}
	case *core.Node_BranchNode:
		branchNodeWrapper := targetNode.(*core.Node_BranchNode)
		if err := hydrateNode(branchNodeWrapper.BranchNode.IfElse.Case.ThenNode, version, force); err != nil {
			return fmt.Errorf("failed to hydrateNode")
		}
		if len(branchNodeWrapper.BranchNode.IfElse.Other) > 0 {
			for _, ifBlock := range branchNodeWrapper.BranchNode.IfElse.Other {
				if err := hydrateNode(ifBlock.ThenNode, version, force); err != nil {
					return fmt.Errorf("failed to hydrateNode")
				}
			}
		}
		switch branchNodeWrapper.BranchNode.IfElse.Default.(type) {
		case *core.IfElseBlock_ElseNode:
			elseNodeReference := branchNodeWrapper.BranchNode.IfElse.Default.(*core.IfElseBlock_ElseNode)
			if err := hydrateNode(elseNodeReference.ElseNode, version, force); err != nil {
				return fmt.Errorf("failed to hydrateNode")
			}

		case *core.IfElseBlock_Error:
			// Do nothing.
		default:
			return fmt.Errorf("unknown type %T", branchNodeWrapper.BranchNode.IfElse.Default)
		}
	default:
		return fmt.Errorf("unknown type %T", v)
	}
	return nil
}

func hydrateIdentifier(identifier *core.Identifier, version string, force bool) {
	if identifier.Project == "" || identifier.Project == registrationProjectPattern {
		identifier.Project = config.GetConfig().Project
	}
	if identifier.Domain == "" || identifier.Domain == registrationDomainPattern {
		identifier.Domain = config.GetConfig().Domain
	}
	if force || identifier.Version == "" || identifier.Version == registrationVersionPattern {
		identifier.Version = version
	}
}

func hydrateTaskSpec(task *admin.TaskSpec, sourceUploadedLocation storage.DataReference, destinationDir string) error {
	if task.Template.GetContainer() != nil {
		for k := range task.Template.GetContainer().Args {
			if task.Template.GetContainer().Args[k] == registrationRemotePackagePattern {
				task.Template.GetContainer().Args[k] = sourceUploadedLocation.String()
			}
			if task.Template.GetContainer().Args[k] == registrationDestDirPattern {
				task.Template.GetContainer().Args[k] = "."
				if len(destinationDir) > 0 {
					task.Template.GetContainer().Args[k] = destinationDir
				}
			}
		}
	} else if task.Template.GetK8SPod() != nil && task.Template.GetK8SPod().PodSpec != nil {
		var podSpec = v1.PodSpec{}
		err := utils.UnmarshalStructToObj(task.Template.GetK8SPod().PodSpec, &podSpec)
		if err != nil {
			return err
		}
		for containerIdx, container := range podSpec.Containers {
			for argIdx, arg := range container.Args {
				if arg == registrationRemotePackagePattern {
					podSpec.Containers[containerIdx].Args[argIdx] = sourceUploadedLocation.String()
				}
				if arg == registrationDestDirPattern {
					podSpec.Containers[containerIdx].Args[argIdx] = "."
					if len(destinationDir) > 0 {
						podSpec.Containers[containerIdx].Args[argIdx] = destinationDir
					}
				}
			}
		}
		podSpecStruct, err := utils.MarshalObjToStruct(podSpec)
		if err != nil {
			return err
		}
		task.Template.Target = &core.TaskTemplate_K8SPod{
			K8SPod: &core.K8SPod{
				Metadata: task.Template.GetK8SPod().Metadata,
				PodSpec:  podSpecStruct,
			},
		}
	}
	return nil
}

func validateLPWithSchedule(lpSpec *admin.LaunchPlanSpec, wf *admin.Workflow) error {
	schedule := lpSpec.EntityMetadata.Schedule
	var scheduleRequiredParams []string
	if wf != nil && wf.Closure != nil && wf.Closure.CompiledWorkflow != nil &&
		wf.Closure.CompiledWorkflow.Primary != nil && wf.Closure.CompiledWorkflow.Primary.Template != nil &&
		wf.Closure.CompiledWorkflow.Primary.Template.Interface != nil &&
		wf.Closure.CompiledWorkflow.Primary.Template.Interface.Inputs != nil {
		variables := wf.Closure.CompiledWorkflow.Primary.Template.Interface.Inputs.Variables
		for varName := range variables {
			if varName != schedule.KickoffTimeInputArg {
				scheduleRequiredParams = append(scheduleRequiredParams, varName)
			}
		}

	}
	// Either the scheduled param should have default or fixed values
	var scheduleParamsWithValues []string
	// Check for default values
	if lpSpec.DefaultInputs != nil {
		for paramName := range lpSpec.DefaultInputs.Parameters {
			if paramName != schedule.KickoffTimeInputArg {
				scheduleParamsWithValues = append(scheduleParamsWithValues, paramName)
			}
		}
	}
	// Check for fixed values
	if lpSpec.FixedInputs != nil && lpSpec.FixedInputs.Literals != nil {
		for fixedLiteralName := range lpSpec.FixedInputs.Literals {
			scheduleParamsWithValues = append(scheduleParamsWithValues, fixedLiteralName)
		}
	}

	diffSet := leftDiff(scheduleRequiredParams, scheduleParamsWithValues)
	if len(diffSet) > 0 {
		return fmt.Errorf("param values are missing on scheduled workflow "+
			"for the following params %v. Either specify them having a default or fixed value", diffSet)
	}
	return nil
}

func validateLaunchSpec(ctx context.Context, lpSpec *admin.LaunchPlanSpec, cmdCtx cmdCore.CommandContext) error {
	if lpSpec == nil || lpSpec.WorkflowId == nil || lpSpec.EntityMetadata == nil ||
		lpSpec.EntityMetadata.Schedule == nil {
		return nil
	}
	// Fetch the workflow spec using the identifier
	workflowID := lpSpec.WorkflowId
	wf, err := cmdCtx.AdminFetcherExt().FetchWorkflowVersion(ctx, workflowID.Name, workflowID.Version,
		workflowID.Project, workflowID.Domain)
	if err != nil {
		return err
	}

	return validateLPWithSchedule(lpSpec, wf)
}

// Finds the left diff between to two string slices
// If a and b are two sets then the o/p c is defined as :
// c = a - a ^ b
// where ^ is intersection slice of a and b
// and - removes all the common elements and returns a new slice
// a= {1,2,3}
// b = {3,4,5}
// o/p c = {1,2}
func leftDiff(a, b []string) []string {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		delete(m, item)
	}
	// nil semantics on return
	if len(m) == 0 {
		return nil
	}
	c := make([]string, len(m))
	index := 0
	for item := range m {
		c[index] = item
		index++
	}
	return c
}

func hydrateLaunchPlanSpec(configAssumableIamRole string, configK8sServiceAccount string, configOutputLocationPrefix string, lpSpec *admin.LaunchPlanSpec) error {
	assumableIamRole := len(configAssumableIamRole) > 0
	k8sServiceAcct := len(configK8sServiceAccount) > 0
	outputLocationPrefix := len(configOutputLocationPrefix) > 0
	if assumableIamRole || k8sServiceAcct {
		lpSpec.AuthRole = &admin.AuthRole{
			AssumableIamRole:         configAssumableIamRole,
			KubernetesServiceAccount: configK8sServiceAccount,
		}
	}
	if outputLocationPrefix {
		lpSpec.RawOutputDataConfig = &admin.RawOutputDataConfig{
			OutputLocationPrefix: configOutputLocationPrefix,
		}
	}
	return nil
}

// Validate the spec before sending it to admin.
func validateSpec(ctx context.Context, message proto.Message, cmdCtx cmdCore.CommandContext) error {
	switch v := message.(type) {
	case *admin.LaunchPlan:
		launchPlan := v
		if err := validateLaunchSpec(ctx, launchPlan.Spec, cmdCtx); err != nil {
			return err
		}
	}
	return nil
}

func hydrateSpec(message proto.Message, uploadLocation storage.DataReference, config rconfig.FilesConfig) error {
	switch v := message.(type) {
	case *admin.LaunchPlan:
		launchPlan := message.(*admin.LaunchPlan)
		hydrateIdentifier(launchPlan.Id, config.Version, config.Force)
		hydrateIdentifier(launchPlan.Spec.WorkflowId, config.Version, config.Force)
		if err := hydrateLaunchPlanSpec(config.AssumableIamRole, config.K8sServiceAccount, config.OutputLocationPrefix, launchPlan.Spec); err != nil {
			return err
		}
	case *admin.WorkflowSpec:
		workflowSpec := message.(*admin.WorkflowSpec)
		for _, Noderef := range workflowSpec.Template.Nodes {
			if err := hydrateNode(Noderef, config.Version, config.Force); err != nil {
				return err
			}
		}
		hydrateIdentifier(workflowSpec.Template.Id, config.Version, config.Force)
		for _, subWorkflow := range workflowSpec.SubWorkflows {
			for _, Noderef := range subWorkflow.Nodes {
				if err := hydrateNode(Noderef, config.Version, config.Force); err != nil {
					return err
				}
			}
			hydrateIdentifier(subWorkflow.Id, config.Version, config.Force)
		}
	case *admin.TaskSpec:
		taskSpec := message.(*admin.TaskSpec)
		hydrateIdentifier(taskSpec.Template.Id, config.Version, config.Force)
		// In case of fast serialize input proto also have on additional variable to substitute i.e destination bucket for source code
		if err := hydrateTaskSpec(taskSpec, uploadLocation, config.DestinationDirectory); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown type %T", v)
	}
	return nil
}

func DownloadFileFromHTTP(ctx context.Context, ref storage.DataReference) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

/*
Get serialize output file list from the args list.
If the archive flag is on then download the archives to temp directory and extract it. In case of fast register it will also return the compressed source code
The o/p of this function would be sorted list of the file locations.
*/
func getSerializeOutputFiles(ctx context.Context, args []string, archive bool) ([]string, string, error) {
	if !archive {
		/*
		 * Sorting is required for non-archived case since its possible for the user to pass in a list of unordered
		 * serialized protobuf files , but flyte expects them to be registered in topologically sorted order that it had
		 * generated otherwise the registration can fail if the dependent files are not registered earlier.
		 */

		finalList := make([]string, 0, len(args))
		for _, arg := range args {
			matches, err := filepath.Glob(arg)
			if err != nil {
				return nil, "", fmt.Errorf("failed to glob [%v]. Error: %w", arg, err)
			}

			finalList = append(finalList, matches...)
		}

		sort.Strings(finalList)
		return finalList, "", nil
	}

	tempDir, err := ioutil.TempDir("/tmp", "register")

	if err != nil {
		return nil, tempDir, err
	}
	var unarchivedFiles []string
	for _, v := range args {
		dataRefReaderCloser, err := getArchiveReaderCloser(ctx, v)
		if err != nil {
			return unarchivedFiles, tempDir, err
		}
		archiveReader := tar.NewReader(dataRefReaderCloser)
		if unarchivedFiles, err = readAndCopyArchive(archiveReader, tempDir, unarchivedFiles); err != nil {
			return unarchivedFiles, tempDir, err
		}
		if err = dataRefReaderCloser.Close(); err != nil {
			return unarchivedFiles, tempDir, err
		}
	}

	/*
	 * Similarly in case of archived files, it possible to have an archive created in totally different order than the
	 * listing order of the serialized files which is required by flyte. Hence we explicitly sort here after unarchiving it.
	 */
	sort.Strings(unarchivedFiles)
	return unarchivedFiles, tempDir, nil
}

func readAndCopyArchive(src io.Reader, tempDir string, unarchivedFiles []string) ([]string, error) {
	for {
		tarReader := src.(*tar.Reader)
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return unarchivedFiles, nil
		case err != nil:
			return unarchivedFiles, err
		}
		// Location to untar. FilePath couldnt be used here due to,
		// G305: File traversal when extracting zip archive
		target := tempDir + "/" + header.Name
		if header.Typeflag == tar.TypeDir {
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return unarchivedFiles, err
				}
			}
		} else if header.Typeflag == tar.TypeReg {
			dest, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return unarchivedFiles, err
			}
			if _, err := io.Copy(dest, src); err != nil {
				return unarchivedFiles, err
			}
			unarchivedFiles = append(unarchivedFiles, dest.Name())
			if err := dest.Close(); err != nil {
				return unarchivedFiles, err
			}
		}
	}
}

func registerFile(ctx context.Context, fileName string, registerResults []Result,
	cmdCtx cmdCore.CommandContext, uploadLocation storage.DataReference, config rconfig.FilesConfig) ([]Result, error) {

	var registerResult Result
	var fileContents []byte
	var err error

	if fileContents, err = ioutil.ReadFile(fileName); err != nil {
		registerResults = append(registerResults, Result{Name: fileName, Status: "Failed", Info: fmt.Sprintf("Error reading file due to %v", err)})
		return registerResults, err
	}
	spec, err := unMarshalContents(ctx, fileContents, fileName)
	if err != nil {
		registerResult = Result{Name: fileName, Status: "Failed", Info: fmt.Sprintf("Error unmarshalling file due to %v", err)}
		registerResults = append(registerResults, registerResult)
		return registerResults, err
	}

	if err := hydrateSpec(spec, uploadLocation, config); err != nil {
		registerResult = Result{Name: fileName, Status: "Failed", Info: fmt.Sprintf("Error hydrating spec due to %v", err)}
		registerResults = append(registerResults, registerResult)
		return registerResults, err
	}

	logger.Debugf(ctx, "Hydrated spec : %v", getJSONSpec(spec))
	if err = validateSpec(ctx, spec, cmdCtx); err != nil {
		registerResult = Result{Name: fileName, Status: "Failed", Info: fmt.Sprintf("Error hydrating spec due to %v", err)}
		registerResults = append(registerResults, registerResult)
		return registerResults, err
	}
	if err := register(ctx, spec, cmdCtx, config.DryRun, config.EnableSchedule); err != nil {
		// If error is AlreadyExists then dont consider this to be an error but just a warning state
		if grpcError := status.Code(err); grpcError == codes.AlreadyExists {
			registerResult = Result{Name: fileName, Status: "Success", Info: fmt.Sprintf("%v", grpcError.String())}
			err = nil
		} else {
			registerResult = Result{Name: fileName, Status: "Failed", Info: fmt.Sprintf("Error registering file due to %v", err)}
		}
		registerResults = append(registerResults, registerResult)
		return registerResults, err
	}

	registerResult = Result{Name: fileName, Status: "Success", Info: "Successfully registered file"}
	logger.Debugf(ctx, "Successfully registered %v", fileName)
	registerResults = append(registerResults, registerResult)
	return registerResults, nil
}

func getArchiveReaderCloser(ctx context.Context, ref string) (io.ReadCloser, error) {
	dataRef := storage.DataReference(ref)
	scheme, _, key, err := dataRef.Split()
	if err != nil {
		return nil, err
	}
	var dataRefReaderCloser io.ReadCloser

	isValid, extension := checkSupportedExtensionForCompress(key)
	if !isValid {
		return nil, errors.New("only .tar, .tar.gz and .tgz extension archives are supported")
	}

	if scheme == "http" || scheme == "https" {
		dataRefReaderCloser, err = DownloadFileFromHTTP(ctx, dataRef)
	} else {
		dataRefReaderCloser, err = os.Open(dataRef.String())
	}
	if err != nil {
		return nil, err
	}

	for _, ext := range validGzipExtensions {
		if ext == extension {
			if dataRefReaderCloser, err = gzip.NewReader(dataRefReaderCloser); err != nil {
				return nil, err
			}
			break
		}
	}
	return dataRefReaderCloser, err
}

func getJSONSpec(message proto.Message) string {
	marshaller := jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}
	jsonSpec, _ := marshaller.MarshalToString(message)
	return jsonSpec
}

func filterExampleFromRelease(releases *github.RepositoryRelease) []*github.ReleaseAsset {
	var assets []*github.ReleaseAsset
	for _, v := range releases.Assets {
		isValid, _ := checkSupportedExtensionForCompress(*v.Name)
		if isValid {
			assets = append(assets, v)
		}
	}
	return assets
}

func getAllExample(repository, version string, repoService g.GHRepoService) ([]*github.ReleaseAsset, *github.RepositoryRelease, error) {
	if len(version) > 0 {
		release, err := g.GetReleaseByTag(repository, version, repoService)
		if err != nil {
			return nil, nil, err
		}
		return filterExampleFromRelease(release), release, nil
	}
	release, err := g.GetLatestRelease(repository, repoService)
	if err != nil {
		return nil, nil, err
	}
	return filterExampleFromRelease(release), release, nil
}

func getRemoteStoragePath(ctx context.Context, s *storage.DataStore, remoteLocation, file, identifier string) (storage.DataReference, error) {
	remotePath, err := s.ConstructReference(ctx, storage.DataReference(remoteLocation), fmt.Sprintf("%v-%v", identifier, file))
	if err != nil {
		return "", err
	}

	return remotePath, nil
}

func getTotalSize(reader io.Reader) (size int64, err error) {
	page := make([]byte, 512)
	size = 0

	n := 0
	for n, err = reader.Read(page); n > 0 && err == nil; n, err = reader.Read(page) {
		size += int64(n)
	}

	if err == io.EOF {
		return size + int64(n), nil
	}

	return size, err
}

func uploadFastRegisterArtifact(ctx context.Context, project, domain, sourceCodeFilePath, version string,
	dataProxyClient service.DataProxyServiceClient, deprecatedSourceUploadPath string) (uploadLocation storage.DataReference, err error) {

	fileHandle, err := os.Open(sourceCodeFilePath)
	if err != nil {
		return "", err
	}

	dataRefReaderCloser, err := gzip.NewReader(fileHandle)
	if err != nil {
		return "", err
	}

	/* #nosec */
	hash := md5.New()
	/* #nosec */
	size, err := io.Copy(hash, dataRefReaderCloser)
	if err != nil {
		return "", err
	}

	_, err = fileHandle.Seek(0, 0)
	if err != nil {
		return "", err
	}

	err = dataRefReaderCloser.Reset(fileHandle)
	if err != nil {
		return "", err
	}

	h := hash.Sum(nil)
	remotePath := storage.DataReference(deprecatedSourceUploadPath)
	_, fileName := filepath.Split(sourceCodeFilePath)
	resp, err := dataProxyClient.CreateUploadLocation(ctx, &service.CreateUploadLocationRequest{
		Project:    project,
		Domain:     domain,
		Filename:   fileName,
		ContentMd5: h,
	})

	if err != nil {
		if status.Code(err) == codes.Unimplemented {
			logger.Infof(ctx, "Using an older version of FlyteAdmin. Falling back to the configured storage client.")
		} else {
			return "", fmt.Errorf("failed to create an upload location. Error: %w", err)
		}
	}

	if resp != nil && len(resp.SignedUrl) > 0 {
		return storage.DataReference(resp.NativeUrl), DirectUpload(resp.SignedUrl, h, size, dataRefReaderCloser)
	}

	dataStore, err := getStorageClient(ctx)
	if err != nil {
		return "", err
	}

	if len(deprecatedSourceUploadPath) == 0 {
		remotePath, err = dataStore.ConstructReference(ctx, dataStore.GetBaseContainerFQN(ctx), "fast")
		if err != nil {
			return "", err
		}
	}

	remotePath, err = getRemoteStoragePath(ctx, dataStore, remotePath.String(), fileName, version)
	if err != nil {
		return "", err
	}

	if err := dataStore.ComposedProtobufStore.WriteRaw(ctx, remotePath, size, storage.Options{}, dataRefReaderCloser); err != nil {
		return "", err
	}

	return remotePath, nil
}

func DirectUpload(url string, contentMD5 []byte, size int64, data io.Reader) error {
	req, err := http.NewRequest(http.MethodPut, url, data)
	if err != nil {
		return err
	}

	req.ContentLength = size
	req.Header.Set("Content-Length", strconv.FormatInt(size, 10))
	req.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(contentMD5))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		raw, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("received response code [%v]. Failed to read response body. Error: %w", res.StatusCode, err)
		}

		return fmt.Errorf("failed uploading to [%v]. bad status: %s: %s", url, res.Status, string(raw))
	}

	return nil
}

func getStorageClient(ctx context.Context) (*storage.DataStore, error) {
	if Client != nil {
		return Client, nil
	}
	testScope := promutils.NewTestScope()
	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey)
	s, err := storage.NewDataStore(storage.GetConfig(), testScope.NewSubScope("flytectl"))
	if err != nil {
		logger.Errorf(ctx, "error while creating storage client %v", err)
		return Client, err
	}
	Client = s
	return Client, nil
}

func isFastRegister(file string) bool {
	_, f := filepath.Split(file)
	// Pyflyte always archive source code with a name that start with fast and have an extension .tar.gz
	if strings.HasPrefix(f, "fast") && strings.HasSuffix(f, sourceCodeExtension) {
		return true
	}
	return false
}

func segregateSourceAndProtos(dataRefs []string) (string, []string, []string) {
	var validProto, InvalidFiles []string
	var sourceCode string
	for _, v := range dataRefs {
		if isFastRegister(v) {
			sourceCode = v
		} else if strings.HasSuffix(v, ".pb") {
			validProto = append(validProto, v)
		} else {
			InvalidFiles = append(InvalidFiles, v)
		}
	}
	return sourceCode, validProto, InvalidFiles
}

func deprecatedCheck(ctx context.Context, k8sServiceAccount *string, k8ServiceAccount string) {
	if len(k8ServiceAccount) > 0 {
		logger.Warning(ctx, "--K8ServiceAccount is deprecated, Please use --K8sServiceAccount")
		*k8sServiceAccount = k8ServiceAccount
	}
}

func checkSupportedExtensionForCompress(file string) (bool, string) {
	for _, extension := range supportedExtensions {
		if strings.HasSuffix(file, extension) {
			return true, extension
		}
	}
	return false, ""
}
