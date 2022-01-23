package register

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flyteorg/flytectl/pkg/util/githubutil"

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
	"github.com/google/go-github/v37/github"

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
	if err := proto.Unmarshal(fileContents, workflowSpec); err == nil {
		return workflowSpec, nil
	}
	logger.Debugf(ctx, "Failed to unmarshal file %v for workflow type", fname)
	taskSpec := &admin.TaskSpec{}
	if err := proto.Unmarshal(fileContents, taskSpec); err == nil {
		return taskSpec, nil
	}
	logger.Debugf(ctx, "Failed to unmarshal  file %v for task type", fname)
	launchPlan := &admin.LaunchPlan{}
	if err := proto.Unmarshal(fileContents, launchPlan); err == nil {
		return launchPlan, nil
	}
	logger.Debugf(ctx, "Failed to unmarshal file %v for launch plan type", fname)
	return nil, fmt.Errorf("failed unmarshalling file %v", fname)

}

func register(ctx context.Context, message proto.Message, cmdCtx cmdCore.CommandContext, dryRun bool) error {
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
		return err
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

func hydrateTaskSpec(task *admin.TaskSpec, sourceCode, sourceUploadPath, version, destinationDir string) error {
	if task.Template.GetContainer() != nil {
		for k := range task.Template.GetContainer().Args {
			if task.Template.GetContainer().Args[k] == registrationRemotePackagePattern {
				remotePath, err := getRemoteStoragePath(context.Background(), Client, sourceUploadPath, sourceCode, version)
				if err != nil {
					return err
				}
				task.Template.GetContainer().Args[k] = string(remotePath)
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
					remotePath, err := getRemoteStoragePath(context.Background(), Client, sourceUploadPath, sourceCode, version)
					if err != nil {
						return err
					}
					podSpec.Containers[containerIdx].Args[argIdx] = string(remotePath)
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

func validateLaunchSpec(lpSpec *admin.LaunchPlanSpec) error {
	if lpSpec == nil {
		return nil
	}
	if lpSpec.EntityMetadata != nil && lpSpec.EntityMetadata.Schedule != nil {
		schedule := lpSpec.EntityMetadata.Schedule
		var scheduleFixedParams []string
		if lpSpec.DefaultInputs != nil {
			for paramKey := range lpSpec.DefaultInputs.Parameters {
				if paramKey != schedule.KickoffTimeInputArg {
					scheduleFixedParams = append(scheduleFixedParams, paramKey)
				}
			}
		}
		if (lpSpec.FixedInputs == nil && len(scheduleFixedParams) > 0) ||
			(len(scheduleFixedParams) > len(lpSpec.FixedInputs.Literals)) {
			fixedInputLen := 0
			if lpSpec.FixedInputs != nil {
				fixedInputLen = len(lpSpec.FixedInputs.Literals)
			}
			return fmt.Errorf("param values are missing on scheduled workflow."+
				"additional args other than %v on scheduled workflow are %v > %v fixed values", schedule.KickoffTimeInputArg,
				len(scheduleFixedParams), fixedInputLen)
		}
	}
	return nil
}

func hydrateLaunchPlanSpec(configAssumableIamRole string, configK8sServiceAccount string, configOutputLocationPrefix string, lpSpec *admin.LaunchPlanSpec) error {

	if err := validateLaunchSpec(lpSpec); err != nil {
		return err
	}
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

func hydrateSpec(message proto.Message, sourceCode string, config rconfig.FilesConfig) error {
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
		if err := hydrateTaskSpec(taskSpec, sourceCode, config.SourceUploadPath, config.Version, config.DestinationDirectory); err != nil {
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

		sort.Strings(args)
		return args, "", nil
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

func registerFile(ctx context.Context, fileName, sourceCode string, registerResults []Result, cmdCtx cmdCore.CommandContext, config rconfig.FilesConfig) ([]Result, error) {
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

	if err := hydrateSpec(spec, sourceCode, config); err != nil {
		registerResult = Result{Name: fileName, Status: "Failed", Info: fmt.Sprintf("Error hydrating spec due to %v", err)}
		registerResults = append(registerResults, registerResult)
		return registerResults, err
	}

	logger.Debugf(ctx, "Hydrated spec : %v", getJSONSpec(spec))

	if err := register(ctx, spec, cmdCtx, config.DryRun); err != nil {
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

func getAllExample(repository, version string) ([]*github.ReleaseAsset, *github.RepositoryRelease, error) {
	if len(version) > 0 {
		release, err := githubutil.CheckVersionExist(version, repository)
		if err != nil {
			return nil, nil, err
		}
		return filterExampleFromRelease(release), release, nil
	}
	releases, err := githubutil.GetListRelease(repository)
	if err != nil {
		return nil, nil, err
	}
	if len(releases) == 0 {
		return nil, nil, fmt.Errorf("repository doesn't have any release")
	}
	for _, v := range releases {
		if !*v.Prerelease {
			return filterExampleFromRelease(v), v, nil
		}
	}
	return nil, nil, nil

}

func getRemoteStoragePath(ctx context.Context, s *storage.DataStore, remoteLocation, file, identifier string) (storage.DataReference, error) {
	remotePath, err := s.ConstructReference(ctx, storage.DataReference(remoteLocation), fmt.Sprintf("%v-%v", identifier, file))
	if err != nil {
		return storage.DataReference(""), err
	}
	return remotePath, nil
}

func uploadFastRegisterArtifact(ctx context.Context, file, sourceCodeName, version string, sourceUploadPath *string) error {
	dataStore, err := getStorageClient(ctx)
	if err != nil {
		return err
	}
	var dataRefReaderCloser io.ReadCloser
	remotePath := storage.DataReference(*sourceUploadPath)
	if len(*sourceUploadPath) == 0 {
		remotePath, err = dataStore.ConstructReference(ctx, dataStore.GetBaseContainerFQN(ctx), "fast")
		if err != nil {
			return err
		}
	}
	*sourceUploadPath = string(remotePath)
	fullRemotePath, err := getRemoteStoragePath(ctx, dataStore, *sourceUploadPath, sourceCodeName, version)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(file)
	if err != nil {
		return err
	}
	dataRefReaderCloser, err = os.Open(file)
	if err != nil {
		return err
	}
	dataRefReaderCloser, err = gzip.NewReader(dataRefReaderCloser)
	if err != nil {
		return err
	}
	if err := dataStore.ComposedProtobufStore.WriteRaw(ctx, fullRemotePath, int64(len(raw)), storage.Options{}, dataRefReaderCloser); err != nil {
		return err
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
