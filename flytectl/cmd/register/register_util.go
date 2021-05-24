package register

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/flyteorg/flytectl/cmd/config"
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

const registrationProjectPattern = "{{ registration.project }}"
const registrationDomainPattern = "{{ registration.domain }}"
const registrationVersionPattern = "{{ registration.version }}"

type Result struct {
	Name   string
	Status string
	Info   string
}

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
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
				Version:      filesConfig.Version,
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
				Version:      filesConfig.Version,
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
				Version:      filesConfig.Version,
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
		identifier.Version = filesConfig.Version
	}
}

func hydrateSpec(message proto.Message) error {
	switch v := message.(type) {
	case *admin.LaunchPlan:
		launchPlan := message.(*admin.LaunchPlan)
		hydrateIdentifier(launchPlan.Spec.WorkflowId)
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
	default:
		return fmt.Errorf("Unknown type %T", v)
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
Get file list from the args list.
If the archive flag is on then download the archives to temp directory and extract it.
The o/p of this function would be sorted list of the file locations.
*/
func getSortedFileList(ctx context.Context, args []string) ([]string, string, error) {
	if !filesConfig.Archive {
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
	dataRefs := args
	var unarchivedFiles []string
	for i := 0; i < len(dataRefs); i++ {
		dataRefReaderCloser, err := getArchiveReaderCloser(ctx, dataRefs[i])
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

func registerFile(ctx context.Context, fileName string, registerResults []Result, cmdCtx cmdCore.CommandContext) ([]Result, error) {
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
	if err := hydrateSpec(spec); err != nil {
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
	if ext != "tar" && ext != "tgz" {
		return nil, errors.New("only .tar and .tgz extension archives are supported")
	}
	var dataRefReaderCloser io.ReadCloser
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
