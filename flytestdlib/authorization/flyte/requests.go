// Package flyte maps reviewed Flyte API operations to Cedar permissions.
package flyte

import (
	"regexp"
	"strings"

	cedar "github.com/cedar-policy/cedar-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	auth "github.com/flyteorg/flyte/v2/flytestdlib/authorization"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

type rule struct {
	action string
	paths  []string
	// Broad operations require permission on Deployment, never a guessed scope.
	broad bool
}

// Each path is a protobuf field path to an identifier actually used by the
// handler's query. Alternative paths represent oneofs; repeated identifiers are
// checked individually. New RPCs are denied until explicitly reviewed here.
var rules = func() map[string]rule {
	r := map[string]rule{}
	add := func(service string, methods map[string]rule) {
		for method, v := range methods {
			r["/flyteidl2."+service+"/"+method] = v
		}
	}
	add("workflow.RunService", map[string]rule{
		"CreateRun":          {"RunCreate", []string{"run_id", "project_id"}, false},
		"AbortRun":           {"RunAbort", []string{"run_id"}, false},
		"AbortAction":        {"RunAbort", []string{"action_id.run"}, false},
		"SignalEvent":        {"RunSignal", []string{"action_id.run"}, false},
		"GetRunDetails":      {"RunRead", []string{"run_id"}, false},
		"WatchRunDetails":    {"RunRead", []string{"run_id"}, false},
		"GetActionDetails":   {"RunRead", []string{"action_id.run"}, false},
		"WatchActionDetails": {"RunRead", []string{"action_id.run"}, false},
		"ListActions":        {"RunRead", []string{"run_id"}, false},
		"WatchActions":       {"RunRead", []string{"run_id"}, false},
		// Task filters select task_project/task_domain, NOT the run's scope.
		"ListRuns": {"RunRead", []string{"project_id", "trigger_name"}, true},
		// WatchRuns currently only enforces its project filter (not task/cluster).
		"WatchRuns":           {"RunRead", []string{"project_id"}, true},
		"WatchGroups":         {"RunRead", []string{"project_id"}, false},
		"GetActionData":       {"DataRead", []string{"action_id.run"}, false},
		"GetActionDataURIs":   {"DataRead", []string{"action_id.run"}, false},
		"GetActionLogContext": {"LogsRead", []string{"action_id.run"}, false},
		"WatchClusterEvents":  {"LogsRead", []string{"id.run"}, false},
	})
	add("task.TaskService", map[string]rule{
		"DeployTask":     {"TaskDeploy", []string{"task_id"}, false},
		"GetTaskDetails": {"TaskRead", []string{"task_id"}, false},
		"ListTasks":      {"TaskRead", []string{"project_id"}, true},
		"ListVersions":   {"TaskRead", []string{"task_name"}, false},
	})
	add("trigger.TriggerService", map[string]rule{
		"DeployTrigger":             {"TriggerWrite", []string{"name"}, false},
		"GetTriggerDetails":         {"TriggerRead", []string{"name"}, false},
		"GetTriggerRevisionDetails": {"TriggerRead", []string{"id.name"}, false},
		"GetTriggerRevisionHistory": {"TriggerRead", []string{"name"}, false},
		"ListTriggers":              {"TriggerRead", []string{"project_id", "task_id", "task_name"}, true},
		"UpdateTriggers":            {"TriggerWrite", []string{"names"}, false},
		"DeleteTriggers":            {"TriggerWrite", []string{"names"}, false},
	})
	add("project.ProjectService", map[string]rule{
		"CreateProject": {"ProjectWrite", []string{"project.id"}, false},
		"UpdateProject": {"ProjectWrite", []string{"project.id"}, false},
		"GetProject":    {"ProjectRead", []string{"id"}, false},
		"ListProjects":  {"ProjectRead", nil, true},
	})
	add("settings.SettingsService", map[string]rule{
		"GetSettings":        {"SettingsRead", []string{"key"}, false},
		"GetSettingsForEdit": {"SettingsRead", []string{"key"}, false},
		"CreateSettings":     {"SettingsWrite", []string{"key"}, false},
		"UpdateSettings":     {"SettingsWrite", []string{"key"}, false},
	})
	add("dataproxy.DataProxyService", map[string]rule{
		"CreateUploadLocation": {"DataWrite", []string{"."}, false},
		"UploadInputs":         {"DataWrite", []string{"run_id", "project_id"}, false},
		"CreateDownloadLink":   {"DataRead", []string{"action_attempt_id.action_id.run"}, false},
		"GetActionData":        {"DataRead", []string{"action_id.run"}, false},
		"TailLogs":             {"LogsRead", []string{"action_id.run"}, false},
	})
	add("workflow.TranslatorService", map[string]rule{
		"LiteralsToLaunchFormJson": {"Translate", nil, true},
		"LaunchFormJsonToLiterals": {"Translate", nil, true},
		"TaskSpecToLaunchFormJson": {"Translate", nil, true},
		"JsonValuesToLiterals":     {"Translate", nil, true},
	})
	add("cluster.ClusterService", map[string]rule{"SelectCluster": {"ClusterRead", nil, true}})
	add("workflow.InternalRunService", map[string]rule{
		"RecordAction":             {"RuntimeWrite", nil, true},
		"RecordActionStream":       {"RuntimeWrite", nil, true},
		"RecordActionEvents":       {"RuntimeWrite", nil, true},
		"RecordActionEventStream":  {"RuntimeWrite", nil, true},
		"UpdateActionStatus":       {"RuntimeWrite", nil, true},
		"UpdateActionStatusStream": {"RuntimeWrite", nil, true},
	})
	return r
}()

// Resolve authorizes the complete request before any handler side effects. The
// first version intentionally supports scope permissions rather than per-row ACLs.
func Resolve(procedure string, message any) ([]auth.Requirement, error) {
	r, ok := rules[procedure]
	if !ok {
		return nil, auth.Denied()
	}
	m, ok := message.(proto.Message)
	if !ok || !m.ProtoReflect().IsValid() {
		return nil, auth.Denied()
	}
	parts := strings.Split(strings.TrimPrefix(procedure, "/"), "/")
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(parts[0]))
	if err != nil {
		return nil, auth.Denied()
	}
	service, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil, auth.Denied()
	}
	method := service.Methods().ByName(protoreflect.Name(parts[1]))
	if method == nil || method.Input().FullName() != m.ProtoReflect().Descriptor().FullName() {
		return nil, auth.Denied()
	}
	checks, err := resolvePaths(m, r)
	if err != nil {
		return nil, err
	}
	// Compound operations require authority over every referenced resource.
	add := func(action string, paths ...string) error {
		extra, err := resolvePaths(m, rule{action, paths, false})
		checks = append(checks, extra...)
		return err
	}
	switch req := message.(type) {
	case *workflow.CreateRunRequest:
		if req.GetTaskId() != nil {
			err = add("TaskExecute", "task_id")
		}
		if req.GetTriggerName() != nil {
			err = add("TriggerExecute", "trigger_name")
		}
		if req.GetRunSpec().GetRelation() != nil && err == nil {
			err = add("DataRead", "run_spec.relation.related_to")
		}
	case *task.DeployTaskRequest:
		// A nonempty set also prunes previously attached triggers.
		if len(req.GetTriggers()) > 0 {
			err = add("TriggerWrite", "task_id")
		}
	case *dataproxy.UploadInputsRequest:
		if req.GetTaskId() != nil {
			err = add("TaskRead", "task_id")
		}
		if req.GetTriggerName() != nil {
			err = add("TriggerRead", "trigger_name")
		}
		if req.GetBaseDir() != "" {
			checks = append(checks, deployment("StorageOverride"))
		}
		// org is used in a storage path, but is not a tenancy boundary in Runs.
		org := req.GetRunId().GetOrg()
		if req.GetProjectId() != nil {
			org = req.GetProjectId().GetOrganization()
		}
		if org != "" && !validSegment(org) {
			return nil, auth.Denied()
		}
	case *dataproxy.CreateUploadLocationRequest:
		if !safeRelativePath(req.GetFilename()) || !safeRelativePath(req.GetFilenameRoot()) {
			return nil, auth.Denied()
		}
	case *workflow.LiteralsToLaunchFormJsonRequest:
		if req.GetTriggerId() != nil {
			err = add("TriggerRead", "trigger_id.name")
		} else if req.GetLiteralsUri() != "" {
			err = add("DataRead", "action_id.run")
		}
	}
	if err != nil {
		return nil, err
	}
	// Caller-supplied storage locations cannot be treated as owned merely because
	// they accompany an authorized scope. Until storage ownership is modeled,
	// require explicit deployment-wide permission for these escape hatches.
	if r.action != "RuntimeWrite" && hasStorageOverride(m.ProtoReflect()) {
		checks = append(checks, deployment("StorageOverride"))
	}
	return checks, nil
}

func hasStorageOverride(m protoreflect.Message) bool {
	found := false
	m.Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if f.Name() == "offloaded_input_data" || (f.Name() == "run_base_dir" && f.Kind() == protoreflect.StringKind && v.String() != "") {
			found = true
			return false
		}
		if f.Kind() != protoreflect.MessageKind || f.IsMap() {
			return true
		}
		if f.IsList() {
			for idx := 0; idx < v.List().Len(); idx++ {
				if hasStorageOverride(v.List().Get(idx).Message()) {
					found = true
					return false
				}
			}
		} else {
			found = hasStorageOverride(v.Message())
		}
		return !found
	})
	return found
}

func resolvePaths(m proto.Message, r rule) ([]auth.Requirement, error) {
	var checks []auth.Requirement
	for _, path := range r.paths {
		for _, value := range valuesAt(m.ProtoReflect(), strings.Split(path, ".")) {
			check, err := resource(r.action, value)
			if err != nil {
				return nil, err
			}
			checks = append(checks, check)
		}
	}
	if len(checks) == 0 {
		if !r.broad {
			return nil, auth.Denied()
		}
		checks = append(checks, deployment(r.action))
	}
	return checks, nil
}

func valuesAt(m protoreflect.Message, path []string) []any {
	if !m.IsValid() {
		return nil
	}
	if len(path) == 0 || path[0] == "" {
		return []any{m.Interface()}
	}
	f := m.Descriptor().Fields().ByName(protoreflect.Name(path[0]))
	if f == nil || !m.Has(f) {
		return nil
	}
	v := m.Get(f)
	if f.IsList() {
		var result []any
		for idx := 0; idx < v.List().Len(); idx++ {
			result = append(result, valuesAt(v.List().Get(idx).Message(), path[1:])...)
		}
		return result
	}
	if f.Kind() == protoreflect.MessageKind {
		return valuesAt(v.Message(), path[1:])
	}
	if len(path) == 1 && f.Kind() == protoreflect.StringKind {
		return []any{v.String()}
	}
	return nil
}

var segment = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`)

func validSegment(s string) bool { return segment.MatchString(s) }
func safeRelativePath(s string) bool {
	// Storage constructs URLs before individual backends interpret their paths.
	// Reject URL escapes as well as literal traversal to avoid decode differences.
	if strings.ContainsAny(s, "\\%?#") || strings.HasPrefix(s, "/") {
		return false
	}
	for _, c := range s {
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	for _, part := range strings.Split(s, "/") {
		if part == ".." || part == "." {
			return false
		}
	}
	return true
}

func deployment(action string) auth.Requirement {
	id := cedar.NewEntityUID("Flyte::Deployment", "default")
	return auth.Requirement{Action: action, Resource: id, Entities: cedar.EntityMap{id: {UID: id}}}
}

func resource(action string, value any) (auth.Requirement, error) {
	var project, domain string
	projectOnly := false
	switch id := value.(type) {
	case string:
		project, projectOnly = id, true
	case *common.ProjectIdentifier:
		project, domain = id.GetName(), id.GetDomain()
	case *settings.SettingsKey:
		// Settings inheritance exposes parent records; keep administration global
		// until inherited settings receive a separately reviewed resource model.
		return deployment(action), nil
	case interface {
		GetProject() string
		GetDomain() string
	}:
		project, domain = id.GetProject(), id.GetDomain()
	default:
		return auth.Requirement{}, auth.Denied()
	}
	if !validSegment(project) || (!projectOnly && !validSegment(domain)) {
		return auth.Requirement{}, auth.Denied()
	}
	check := deployment(action)
	parent := check.Resource
	pid := cedar.NewEntityUID("Flyte::Project", cedar.String(project))
	check.Entities[pid] = cedar.Entity{UID: pid, Parents: cedar.NewEntityUIDSet(parent), Attributes: cedar.NewRecord(cedar.RecordMap{"project": cedar.String(project)})}
	check.Resource = pid
	if !projectOnly {
		id := cedar.NewEntityUID("Flyte::Scope", cedar.String(project+"/"+domain))
		check.Entities[id] = cedar.Entity{UID: id, Parents: cedar.NewEntityUIDSet(pid, parent), Attributes: cedar.NewRecord(cedar.RecordMap{"project": cedar.String(project), "domain": cedar.String(domain)})}
		check.Resource = id
	}
	return check, nil
}
