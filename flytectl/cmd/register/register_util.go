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

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/google/go-github/github"

	"github.com/flyteorg/flytectl/cmd/config"
	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/printer"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Variable define in serialized proto that needs to be replace in registration time
const registrationProjectPattern = "{{ registration.project }}"
const registrationDomainPattern = "{{ registration.domain }}"
const registrationVersionPattern = "{{ registration.version }}"

// Additional variable define in fast serialized proto that needs to be replace in registration time
const registrationRemotePackagePattern = "{{ .remote_package_path }}"

type Result struct {
	Name   string
	Status string
	Info   string
}

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var FlyteSnacksRelease []FlyteSnack
var Client *storage.DataStore

// FlyteSnack Defines flyte test manifest structure
type FlyteSnack struct {
	Name          string    `json:"name"`
	Priority      string    `json:"priority"`
	Path          string    `json:"path"`
	ExitCondition Condition `json:"exitCondition"`
}

type Condition struct {
	ExitSuccess bool   `json:"exit_success"`
	ExitMessage string `json:"exit_message"`
}

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

func register(ctx context.Context, message proto.Message, cmdCtx cmdCore.CommandContext) error {
	switch v := message.(type) {
	case *admin.LaunchPlan:
		launchPlan := message.(*admin.LaunchPlan)
		_, err := cmdCtx.AdminClient().CreateLaunchPlan(ctx, &admin.LaunchPlanCreateRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      config.GetConfig().Project,
				Domain:       config.GetConfig().Domain,
				Name:         launchPlan.Id.Name,
				Version:      rconfig.DefaultFilesConfig.Version,
			},
			Spec: launchPlan.Spec,
		})
		return err
	case *admin.WorkflowSpec:
		workflowSpec := message.(*admin.WorkflowSpec)
		_, err := cmdCtx.AdminClient().CreateWorkflow(ctx, &admin.WorkflowCreateRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_WORKFLOW,
				Project:      config.GetConfig().Project,
				Domain:       config.GetConfig().Domain,
				Name:         workflowSpec.Template.Id.Name,
				Version:      rconfig.DefaultFilesConfig.Version,
			},
			Spec: workflowSpec,
		})
		return err
	case *admin.TaskSpec:
		taskSpec := message.(*admin.TaskSpec)
		_, err := cmdCtx.AdminClient().CreateTask(ctx, &admin.TaskCreateRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      config.GetConfig().Project,
				Domain:       config.GetConfig().Domain,
				Name:         taskSpec.Template.Id.Name,
				Version:      rconfig.DefaultFilesConfig.Version,
			},
			Spec: taskSpec,
		})
		return err
	default:
		return fmt.Errorf("Failed registering unknown entity  %v", v)
	}
}

func hydrateNode(node *core.Node) error {
	targetNode := node.Target
	switch v := targetNode.(type) {
	case *core.Node_TaskNode:
		taskNodeWrapper := targetNode.(*core.Node_TaskNode)
		taskNodeReference := taskNodeWrapper.TaskNode.Reference.(*core.TaskNode_ReferenceId)
		hydrateIdentifier(taskNodeReference.ReferenceId)
	case *core.Node_WorkflowNode:
		workflowNodeWrapper := targetNode.(*core.Node_WorkflowNode)
		switch workflowNodeWrapper.WorkflowNode.Reference.(type) {
		case *core.WorkflowNode_SubWorkflowRef:
			subWorkflowNodeReference := workflowNodeWrapper.WorkflowNode.Reference.(*core.WorkflowNode_SubWorkflowRef)
			hydrateIdentifier(subWorkflowNodeReference.SubWorkflowRef)
		case *core.WorkflowNode_LaunchplanRef:
			launchPlanNodeReference := workflowNodeWrapper.WorkflowNode.Reference.(*core.WorkflowNode_LaunchplanRef)
			hydrateIdentifier(launchPlanNodeReference.LaunchplanRef)
		default:
			return fmt.Errorf("unknown type %T", workflowNodeWrapper.WorkflowNode.Reference)
		}
	case *core.Node_BranchNode:
		branchNodeWrapper := targetNode.(*core.Node_BranchNode)
		if err := hydrateNode(branchNodeWrapper.BranchNode.IfElse.Case.ThenNode); err != nil {
			return fmt.Errorf("failed to hydrateNode")
		}
		if len(branchNodeWrapper.BranchNode.IfElse.Other) > 0 {
			for _, ifBlock := range branchNodeWrapper.BranchNode.IfElse.Other {
				if err := hydrateNode(ifBlock.ThenNode); err != nil {
					return fmt.Errorf("failed to hydrateNode")
				}
			}
		}
		switch branchNodeWrapper.BranchNode.IfElse.Default.(type) {
		case *core.IfElseBlock_ElseNode:
			elseNodeReference := branchNodeWrapper.BranchNode.IfElse.Default.(*core.IfElseBlock_ElseNode)
			if err := hydrateNode(elseNodeReference.ElseNode); err != nil {
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

func hydrateIdentifier(identifier *core.Identifier) {
	if identifier.Project == "" || identifier.Project == registrationProjectPattern {
		identifier.Project = config.GetConfig().Project
	}
	if identifier.Domain == "" || identifier.Domain == registrationDomainPattern {
		identifier.Domain = config.GetConfig().Domain
	}
	if identifier.Version == "" || identifier.Version == registrationVersionPattern {
		identifier.Version = rconfig.DefaultFilesConfig.Version
	}
}

func hydrateTaskSpec(task *admin.TaskSpec, sourceCode string) error {
	if task.Template.GetContainer() != nil {
		for k := range task.Template.GetContainer().Args {
			if task.Template.GetContainer().Args[k] == "" || task.Template.GetContainer().Args[k] == registrationRemotePackagePattern {
				remotePath, err := getRemoteStoragePath(context.Background(), Client, rconfig.DefaultFilesConfig.SourceUploadPath, sourceCode, rconfig.DefaultFilesConfig.Version)
				if err != nil {
					return err
				}
				task.Template.GetContainer().Args[k] = string(remotePath)
			}
		}
	}
	return nil
}

func hydrateLaunchPlanSpec(lpSpec *admin.LaunchPlanSpec) {
	assumableIamRole := len(rconfig.DefaultFilesConfig.AssumableIamRole) > 0
	k8ServiceAcct := len(rconfig.DefaultFilesConfig.K8ServiceAccount) > 0
	outputLocationPrefix := len(rconfig.DefaultFilesConfig.OutputLocationPrefix) > 0
	if assumableIamRole || k8ServiceAcct {
		lpSpec.AuthRole = &admin.AuthRole{
			AssumableIamRole:         rconfig.DefaultFilesConfig.AssumableIamRole,
			KubernetesServiceAccount: rconfig.DefaultFilesConfig.K8ServiceAccount,
		}
	}
	if outputLocationPrefix {
		lpSpec.RawOutputDataConfig = &admin.RawOutputDataConfig{
			OutputLocationPrefix: rconfig.DefaultFilesConfig.OutputLocationPrefix,
		}
	}
}

func hydrateSpec(message proto.Message, sourceCode string) error {
	switch v := message.(type) {
	case *admin.LaunchPlan:
		launchPlan := message.(*admin.LaunchPlan)
		hydrateIdentifier(launchPlan.Spec.WorkflowId)
		hydrateLaunchPlanSpec(launchPlan.Spec)
	case *admin.WorkflowSpec:
		workflowSpec := message.(*admin.WorkflowSpec)
		for _, Noderef := range workflowSpec.Template.Nodes {
			if err := hydrateNode(Noderef); err != nil {
				return err
			}
		}
		hydrateIdentifier(workflowSpec.Template.Id)
		for _, subWorkflow := range workflowSpec.SubWorkflows {
			for _, Noderef := range subWorkflow.Nodes {
				if err := hydrateNode(Noderef); err != nil {
					return err
				}
			}
			hydrateIdentifier(subWorkflow.Id)
		}
	case *admin.TaskSpec:
		taskSpec := message.(*admin.TaskSpec)
		hydrateIdentifier(taskSpec.Template.Id)
		// In case of fast serialize input proto also have on additional variable to substitute i.e destination bucket for source code
		if err := hydrateTaskSpec(taskSpec, sourceCode); err != nil {
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
func getSerializeOutputFiles(ctx context.Context, args []string) ([]string, string, error) {
	if !rconfig.DefaultFilesConfig.Archive {
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

func registerFile(ctx context.Context, fileName, sourceCode string, registerResults []Result, cmdCtx cmdCore.CommandContext) ([]Result, error) {
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

	if err := hydrateSpec(spec, sourceCode); err != nil {
		registerResult = Result{Name: fileName, Status: "Failed", Info: fmt.Sprintf("Error hydrating spec due to %v", err)}
		registerResults = append(registerResults, registerResult)
		return registerResults, err
	}

	logger.Debugf(ctx, "Hydrated spec : %v", getJSONSpec(spec))

	if err := register(ctx, spec, cmdCtx); err != nil {
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
	segments := strings.Split(key, ".")
	ext := segments[len(segments)-1]
	if err != nil {
		return nil, err
	}
	var dataRefReaderCloser io.ReadCloser

	if ext != "tar" && ext != "tgz" {
		return nil, errors.New("only .tar and .tgz extension archives are supported")
	}

	if scheme == "http" || scheme == "https" {
		dataRefReaderCloser, err = DownloadFileFromHTTP(ctx, dataRef)
	} else {
		dataRefReaderCloser, err = os.Open(dataRef.String())
	}
	if err != nil {
		return nil, err
	}
	if ext == "tgz" {
		if dataRefReaderCloser, err = gzip.NewReader(dataRefReaderCloser); err != nil {
			return nil, err
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

func getFlyteTestManifest(org, repository string) ([]FlyteSnack, string, error) {
	c := github.NewClient(nil)
	opt := &github.ListOptions{Page: 1, PerPage: 1}
	releases, _, err := c.Repositories.ListReleases(context.Background(), org, repository, opt)
	if err != nil {
		return nil, "", err
	}
	if len(releases) == 0 {
		return nil, "", fmt.Errorf("Repository doesn't have any release")
	}
	response, err := http.Get(fmt.Sprintf(flyteManifest, *releases[0].TagName))
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	err = json.Unmarshal(data, &FlyteSnacksRelease)
	if err != nil {
		return nil, "", err
	}
	return FlyteSnacksRelease, *releases[0].TagName, nil

}

func getRemoteStoragePath(ctx context.Context, s *storage.DataStore, remoteLocation, file, identifier string) (storage.DataReference, error) {
	remotePath, err := s.ConstructReference(ctx, storage.DataReference(remoteLocation), fmt.Sprintf("%v-%v", identifier, file))
	if err != nil {
		return storage.DataReference(""), err
	}
	return remotePath, nil
}

func uploadFastRegisterArtifact(ctx context.Context, file, sourceCodeName, version string) error {
	dataStore, err := getStorageClient(ctx)
	if err != nil {
		return err
	}
	var dataRefReaderCloser io.ReadCloser
	remotePath := storage.DataReference(rconfig.DefaultFilesConfig.SourceUploadPath)
	if len(rconfig.DefaultFilesConfig.SourceUploadPath) == 0 {
		remotePath, err = dataStore.ConstructReference(ctx, dataStore.GetBaseContainerFQN(ctx), "fast")
		if err != nil {
			return err
		}
	}
	rconfig.DefaultFilesConfig.SourceUploadPath = string(remotePath)
	fullRemotePath, err := getRemoteStoragePath(ctx, dataStore, rconfig.DefaultFilesConfig.SourceUploadPath, sourceCodeName, version)
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
