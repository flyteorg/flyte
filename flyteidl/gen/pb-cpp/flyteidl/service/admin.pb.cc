// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: flyteidl/service/admin.proto

#include "flyteidl/service/admin.pb.h"

#include <algorithm>

#include <google/protobuf/stubs/common.h>
#include <google/protobuf/io/coded_stream.h>
#include <google/protobuf/extension_set.h>
#include <google/protobuf/wire_format_lite_inl.h>
#include <google/protobuf/descriptor.h>
#include <google/protobuf/generated_message_reflection.h>
#include <google/protobuf/reflection_ops.h>
#include <google/protobuf/wire_format.h>
// @@protoc_insertion_point(includes)
#include <google/protobuf/port_def.inc>

namespace flyteidl {
namespace service {
}  // namespace service
}  // namespace flyteidl
void InitDefaults_flyteidl_2fservice_2fadmin_2eproto() {
}

constexpr ::google::protobuf::Metadata* file_level_metadata_flyteidl_2fservice_2fadmin_2eproto = nullptr;
constexpr ::google::protobuf::EnumDescriptor const** file_level_enum_descriptors_flyteidl_2fservice_2fadmin_2eproto = nullptr;
constexpr ::google::protobuf::ServiceDescriptor const** file_level_service_descriptors_flyteidl_2fservice_2fadmin_2eproto = nullptr;
const ::google::protobuf::uint32 TableStruct_flyteidl_2fservice_2fadmin_2eproto::offsets[1] = {};
static constexpr ::google::protobuf::internal::MigrationSchema* schemas = nullptr;
static constexpr ::google::protobuf::Message* const* file_default_instances = nullptr;

::google::protobuf::internal::AssignDescriptorsTable assign_descriptors_table_flyteidl_2fservice_2fadmin_2eproto = {
  {}, AddDescriptors_flyteidl_2fservice_2fadmin_2eproto, "flyteidl/service/admin.proto", schemas,
  file_default_instances, TableStruct_flyteidl_2fservice_2fadmin_2eproto::offsets,
  file_level_metadata_flyteidl_2fservice_2fadmin_2eproto, 0, file_level_enum_descriptors_flyteidl_2fservice_2fadmin_2eproto, file_level_service_descriptors_flyteidl_2fservice_2fadmin_2eproto,
};

const char descriptor_table_protodef_flyteidl_2fservice_2fadmin_2eproto[] =
  "\n\034flyteidl/service/admin.proto\022\020flyteidl"
  ".service\032\034google/api/annotations.proto\032\034"
  "flyteidl/admin/project.proto\032.flyteidl/a"
  "dmin/project_domain_attributes.proto\032\'fl"
  "yteidl/admin/project_attributes.proto\032\031f"
  "lyteidl/admin/task.proto\032\035flyteidl/admin"
  "/workflow.proto\032(flyteidl/admin/workflow"
  "_attributes.proto\032 flyteidl/admin/launch"
  "_plan.proto\032\032flyteidl/admin/event.proto\032"
  "\036flyteidl/admin/execution.proto\032\'flyteid"
  "l/admin/matchable_resource.proto\032#flytei"
  "dl/admin/node_execution.proto\032#flyteidl/"
  "admin/task_execution.proto\032\034flyteidl/adm"
  "in/version.proto\032\033flyteidl/admin/common."
  "proto\032\'flyteidl/admin/description_entity"
  ".proto2\214x\n\014AdminService\022\216\001\n\nCreateTask\022!"
  ".flyteidl.admin.TaskCreateRequest\032\".flyt"
  "eidl.admin.TaskCreateResponse\"9\202\323\344\223\0023\"\r/"
  "api/v1/tasks:\001*Z\037\"\032/api/v1/tasks/org/{id"
  ".org}:\001*\022\330\001\n\007GetTask\022 .flyteidl.admin.Ob"
  "jectGetRequest\032\024.flyteidl.admin.Task\"\224\001\202"
  "\323\344\223\002\215\001\022=/api/v1/tasks/{id.project}/{id.d"
  "omain}/{id.name}/{id.version}ZL\022J/api/v1"
  "/tasks/org/{id.org}/{id.project}/{id.dom"
  "ain}/{id.name}/{id.version}\022\305\001\n\013ListTask"
  "Ids\0220.flyteidl.admin.NamedEntityIdentifi"
  "erListRequest\032).flyteidl.admin.NamedEnti"
  "tyIdentifierList\"Y\202\323\344\223\002S\022#/api/v1/task_i"
  "ds/{project}/{domain}Z,\022*/api/v1/tasks/o"
  "rg/{org}/{project}/{domain}\022\250\002\n\tListTask"
  "s\022#.flyteidl.admin.ResourceListRequest\032\030"
  ".flyteidl.admin.TaskList\"\333\001\202\323\344\223\002\324\001\0220/api"
  "/v1/tasks/{id.project}/{id.domain}/{id.n"
  "ame}Z\?\022=/api/v1/tasks/org/{id.org}/{id.p"
  "roject}/{id.domain}/{id.name}Z(\022&/api/v1"
  "/tasks/{id.project}/{id.domain}Z5\0223/api/"
  "v1/tasks/org/{id.org}/{id.project}/{id.d"
  "omain}\022\242\001\n\016CreateWorkflow\022%.flyteidl.adm"
  "in.WorkflowCreateRequest\032&.flyteidl.admi"
  "n.WorkflowCreateResponse\"A\202\323\344\223\002;\"\021/api/v"
  "1/workflows:\001*Z#\"\036/api/v1/workflows/org/"
  "{id.org}:\001*\022\350\001\n\013GetWorkflow\022 .flyteidl.a"
  "dmin.ObjectGetRequest\032\030.flyteidl.admin.W"
  "orkflow\"\234\001\202\323\344\223\002\225\001\022A/api/v1/workflows/{id"
  ".project}/{id.domain}/{id.name}/{id.vers"
  "ion}ZP\022N/api/v1/workflows/org/{id.org}/{"
  "id.project}/{id.domain}/{id.name}/{id.ve"
  "rsion}\022\362\001\n\026GetDynamicNodeWorkflow\022-.flyt"
  "eidl.admin.GetDynamicNodeWorkflowRequest"
  "\032+.flyteidl.admin.DynamicNodeWorkflowRes"
  "ponse\"|\202\323\344\223\002v\022t/api/v1/dynamic_node_work"
  "flow/{id.execution_id.project}/{id.execu"
  "tion_id.domain}/{id.execution_id.name}/{"
  "id.node_id}\022\321\001\n\017ListWorkflowIds\0220.flytei"
  "dl.admin.NamedEntityIdentifierListReques"
  "t\032).flyteidl.admin.NamedEntityIdentifier"
  "List\"a\202\323\344\223\002[\022\'/api/v1/workflow_ids/{proj"
  "ect}/{domain}Z0\022./api/v1/workflows/org/{"
  "org}/{project}/{domain}\022\300\002\n\rListWorkflow"
  "s\022#.flyteidl.admin.ResourceListRequest\032\034"
  ".flyteidl.admin.WorkflowList\"\353\001\202\323\344\223\002\344\001\0224"
  "/api/v1/workflows/{id.project}/{id.domai"
  "n}/{id.name}ZC\022A/api/v1/workflows/org/{i"
  "d.org}/{id.project}/{id.domain}/{id.name"
  "}Z,\022*/api/v1/workflows/{id.project}/{id."
  "domain}Z9\0227/api/v1/workflows/org/{id.org"
  "}/{id.project}/{id.domain}\022\256\001\n\020CreateLau"
  "nchPlan\022\'.flyteidl.admin.LaunchPlanCreat"
  "eRequest\032(.flyteidl.admin.LaunchPlanCrea"
  "teResponse\"G\202\323\344\223\002A\"\024/api/v1/launch_plans"
  ":\001*Z&\"!/api/v1/launch_plans/org/{id.org}"
  ":\001*\022\362\001\n\rGetLaunchPlan\022 .flyteidl.admin.O"
  "bjectGetRequest\032\032.flyteidl.admin.LaunchP"
  "lan\"\242\001\202\323\344\223\002\233\001\022D/api/v1/launch_plans/{id."
  "project}/{id.domain}/{id.name}/{id.versi"
  "on}ZS\022Q/api/v1/launch_plans/org/{id.org}"
  "/{id.project}/{id.domain}/{id.name}/{id."
  "version}\022\363\001\n\023GetActiveLaunchPlan\022\'.flyte"
  "idl.admin.ActiveLaunchPlanRequest\032\032.flyt"
  "eidl.admin.LaunchPlan\"\226\001\202\323\344\223\002\217\001\022>/api/v1"
  "/active_launch_plans/{id.project}/{id.do"
  "main}/{id.name}ZM\022K/api/v1/active_launch"
  "_plans/org/{id.org}/{id.project}/{id.dom"
  "ain}/{id.name}\022\330\001\n\025ListActiveLaunchPlans"
  "\022+.flyteidl.admin.ActiveLaunchPlanListRe"
  "quest\032\036.flyteidl.admin.LaunchPlanList\"r\202"
  "\323\344\223\002l\022./api/v1/active_launch_plans/{proj"
  "ect}/{domain}Z:\0228/api/v1/active_launch_p"
  "lans/org/{org}/{project}/{domain}\022\334\001\n\021Li"
  "stLaunchPlanIds\0220.flyteidl.admin.NamedEn"
  "tityIdentifierListRequest\032).flyteidl.adm"
  "in.NamedEntityIdentifierList\"j\202\323\344\223\002d\022*/a"
  "pi/v1/launch_plan_ids/{project}/{domain}"
  "Z6\0224/api/v1/launch_plan_ids/org/{org}/{p"
  "roject}/{domain}\022\320\002\n\017ListLaunchPlans\022#.f"
  "lyteidl.admin.ResourceListRequest\032\036.flyt"
  "eidl.admin.LaunchPlanList\"\367\001\202\323\344\223\002\360\001\0227/ap"
  "i/v1/launch_plans/{id.project}/{id.domai"
  "n}/{id.name}ZF\022D/api/v1/launch_plans/org"
  "/{id.org}/{id.project}/{id.domain}/{id.n"
  "ame}Z/\022-/api/v1/launch_plans/{id.project"
  "}/{id.domain}Z<\022:/api/v1/launch_plans/or"
  "g/{id.org}/{id.project}/{id.domain}\022\215\002\n\020"
  "UpdateLaunchPlan\022\'.flyteidl.admin.Launch"
  "PlanUpdateRequest\032(.flyteidl.admin.Launc"
  "hPlanUpdateResponse\"\245\001\202\323\344\223\002\236\001\032D/api/v1/l"
  "aunch_plans/{id.project}/{id.domain}/{id"
  ".name}/{id.version}:\001*ZS\032Q/api/v1/launch"
  "_plans/org/{id.org}/{id.project}/{id.dom"
  "ain}/{id.name}/{id.version}\022\244\001\n\017CreateEx"
  "ecution\022&.flyteidl.admin.ExecutionCreate"
  "Request\032\'.flyteidl.admin.ExecutionCreate"
  "Response\"@\202\323\344\223\002:\"\022/api/v1/executions:\001*Z"
  "!\032\034/api/v1/executions/org/{org}:\001*\022\275\001\n\021R"
  "elaunchExecution\022(.flyteidl.admin.Execut"
  "ionRelaunchRequest\032\'.flyteidl.admin.Exec"
  "utionCreateResponse\"U\202\323\344\223\002O\"\033/api/v1/exe"
  "cutions/relaunch:\001*Z-\"(/api/v1/execution"
  "s/org/{id.org}/relaunch:\001*\022\271\001\n\020RecoverEx"
  "ecution\022\'.flyteidl.admin.ExecutionRecove"
  "rRequest\032\'.flyteidl.admin.ExecutionCreat"
  "eResponse\"S\202\323\344\223\002M\"\032/api/v1/executions/re"
  "cover:\001*Z,\"\'/api/v1/executions/org/{id.o"
  "rg}/recover:\001*\022\334\001\n\014GetExecution\022+.flytei"
  "dl.admin.WorkflowExecutionGetRequest\032\031.f"
  "lyteidl.admin.Execution\"\203\001\202\323\344\223\002}\0225/api/v"
  "1/executions/{id.project}/{id.domain}/{i"
  "d.name}ZD\022B/api/v1/executions/org/{id.or"
  "g}/{id.project}/{id.domain}/{id.name}\022\357\001"
  "\n\017UpdateExecution\022&.flyteidl.admin.Execu"
  "tionUpdateRequest\032\'.flyteidl.admin.Execu"
  "tionUpdateResponse\"\212\001\202\323\344\223\002\203\001\0325/api/v1/ex"
  "ecutions/{id.project}/{id.domain}/{id.na"
  "me}:\001*ZG\032B/api/v1/executions/org/{id.org"
  "}/{id.project}/{id.domain}/{id.name}:\001*\022"
  "\206\002\n\020GetExecutionData\022/.flyteidl.admin.Wo"
  "rkflowExecutionGetDataRequest\0320.flyteidl"
  ".admin.WorkflowExecutionGetDataResponse\""
  "\216\001\202\323\344\223\002\207\001\022:/api/v1/data/executions/{id.p"
  "roject}/{id.domain}/{id.name}ZI\022G/api/v1"
  "/data/executions/org/{id.org}/{id.projec"
  "t}/{id.domain}/{id.name}\022\305\001\n\016ListExecuti"
  "ons\022#.flyteidl.admin.ResourceListRequest"
  "\032\035.flyteidl.admin.ExecutionList\"o\202\323\344\223\002i\022"
  "+/api/v1/executions/{id.project}/{id.dom"
  "ain}Z:\0228/api/v1/executions/org/{id.org}/"
  "{id.project}/{id.domain}\022\370\001\n\022TerminateEx"
  "ecution\022).flyteidl.admin.ExecutionTermin"
  "ateRequest\032*.flyteidl.admin.ExecutionTer"
  "minateResponse\"\212\001\202\323\344\223\002\203\001*5/api/v1/execut"
  "ions/{id.project}/{id.domain}/{id.name}:"
  "\001*ZG*B/api/v1/executions/org/{id.org}/{i"
  "d.project}/{id.domain}/{id.name}:\001*\022\342\002\n\020"
  "GetNodeExecution\022\'.flyteidl.admin.NodeEx"
  "ecutionGetRequest\032\035.flyteidl.admin.NodeE"
  "xecution\"\205\002\202\323\344\223\002\376\001\022n/api/v1/node_executi"
  "ons/{id.execution_id.project}/{id.execut"
  "ion_id.domain}/{id.execution_id.name}/{i"
  "d.node_id}Z\213\001\022\210\001/api/v1/node_executions/"
  "org/{id.execution_id.org}/{id.execution_"
  "id.project}/{id.execution_id.domain}/{id"
  ".execution_id.name}/{id.node_id}\022\371\002\n\022Lis"
  "tNodeExecutions\022(.flyteidl.admin.NodeExe"
  "cutionListRequest\032!.flyteidl.admin.NodeE"
  "xecutionList\"\225\002\202\323\344\223\002\216\002\022s/api/v1/node_exe"
  "cutions/{workflow_execution_id.project}/"
  "{workflow_execution_id.domain}/{workflow"
  "_execution_id.name}Z\226\001\022\223\001/api/v1/node_ex"
  "ecutions/org/{workflow_execution_id.org}"
  "/{workflow_execution_id.project}/{workfl"
  "ow_execution_id.domain}/{workflow_execut"
  "ion_id.name}\022\217\010\n\031ListNodeExecutionsForTa"
  "sk\022/.flyteidl.admin.NodeExecutionForTask"
  "ListRequest\032!.flyteidl.admin.NodeExecuti"
  "onList\"\235\007\202\323\344\223\002\226\007\022\251\003/api/v1/children/task"
  "_executions/{task_execution_id.node_exec"
  "ution_id.execution_id.project}/{task_exe"
  "cution_id.node_execution_id.execution_id"
  ".domain}/{task_execution_id.node_executi"
  "on_id.execution_id.name}/{task_execution"
  "_id.node_execution_id.node_id}/{task_exe"
  "cution_id.task_id.project}/{task_executi"
  "on_id.task_id.domain}/{task_execution_id"
  ".task_id.name}/{task_execution_id.task_i"
  "d.version}/{task_execution_id.retry_atte"
  "mpt}Z\347\003\022\344\003/api/v1/children/task_executio"
  "ns/org/{task_execution_id.node_execution"
  "_id.execution_id.org}/{task_execution_id"
  ".node_execution_id.execution_id.project}"
  "/{task_execution_id.node_execution_id.ex"
  "ecution_id.domain}/{task_execution_id.no"
  "de_execution_id.execution_id.name}/{task"
  "_execution_id.node_execution_id.node_id}"
  "/{task_execution_id.task_id.project}/{ta"
  "sk_execution_id.task_id.domain}/{task_ex"
  "ecution_id.task_id.name}/{task_execution"
  "_id.task_id.version}/{task_execution_id."
  "retry_attempt}\022\203\003\n\024GetNodeExecutionData\022"
  "+.flyteidl.admin.NodeExecutionGetDataReq"
  "uest\032,.flyteidl.admin.NodeExecutionGetDa"
  "taResponse\"\217\002\202\323\344\223\002\210\002\022s/api/v1/data/node_"
  "executions/{id.execution_id.project}/{id"
  ".execution_id.domain}/{id.execution_id.n"
  "ame}/{id.node_id}Z\220\001\022\215\001/api/v1/data/node"
  "_executions/org/{id.execution_id.org}/{i"
  "d.execution_id.project}/{id.execution_id"
  ".domain}/{id.execution_id.name}/{id.node"
  "_id}\022\250\001\n\017RegisterProject\022&.flyteidl.admi"
  "n.ProjectRegisterRequest\032\'.flyteidl.admi"
  "n.ProjectRegisterResponse\"D\202\323\344\223\002>\"\020/api/"
  "v1/projects:\001*Z\'\"\"/api/v1/projects/org/{"
  "project.org}:\001*\022\227\001\n\rUpdateProject\022\027.flyt"
  "eidl.admin.Project\032%.flyteidl.admin.Proj"
  "ectUpdateResponse\"F\202\323\344\223\002@\032\025/api/v1/proje"
  "cts/{id}:\001*Z$\032\037/api/v1/projects/org/{org"
  "}/{id}:\001*\022\204\001\n\014ListProjects\022\".flyteidl.ad"
  "min.ProjectListRequest\032\030.flyteidl.admin."
  "Projects\"6\202\323\344\223\0020\022\020/api/v1/projectsZ\034\022\032/a"
  "pi/v1/projects/org/{org}\022\325\001\n\023CreateWorkf"
  "lowEvent\022-.flyteidl.admin.WorkflowExecut"
  "ionEventRequest\032..flyteidl.admin.Workflo"
  "wExecutionEventResponse\"_\202\323\344\223\002Y\"\030/api/v1"
  "/events/workflows:\001*Z:\"5/api/v1/events/o"
  "rg/{event.execution_id.org}/workflows:\001*"
  "\022\304\001\n\017CreateNodeEvent\022).flyteidl.admin.No"
  "deExecutionEventRequest\032*.flyteidl.admin"
  ".NodeExecutionEventResponse\"Z\202\323\344\223\002T\"\024/ap"
  "i/v1/events/nodes:\001*Z9\"4/api/v1/events/o"
  "rg/{event.id.execution_id.org}/nodes:\001*\022"
  "\332\001\n\017CreateTaskEvent\022).flyteidl.admin.Tas"
  "kExecutionEventRequest\032*.flyteidl.admin."
  "TaskExecutionEventResponse\"p\202\323\344\223\002j\"\024/api"
  "/v1/events/tasks:\001*ZO\"J/api/v1/events/or"
  "g/{event.parent_node_execution_id.execut"
  "ion_id.org}/tasks:\001*\022\313\005\n\020GetTaskExecutio"
  "n\022\'.flyteidl.admin.TaskExecutionGetReque"
  "st\032\035.flyteidl.admin.TaskExecution\"\356\004\202\323\344\223"
  "\002\347\004\022\231\002/api/v1/task_executions/{id.node_e"
  "xecution_id.execution_id.project}/{id.no"
  "de_execution_id.execution_id.domain}/{id"
  ".node_execution_id.execution_id.name}/{i"
  "d.node_execution_id.node_id}/{id.task_id"
  ".project}/{id.task_id.domain}/{id.task_i"
  "d.name}/{id.task_id.version}/{id.retry_a"
  "ttempt}Z\310\002\022\305\002/api/v1/task_executions/org"
  "/{id.node_execution_id.execution_id.org}"
  "/{id.node_execution_id.execution_id.proj"
  "ect}/{id.node_execution_id.execution_id."
  "domain}/{id.node_execution_id.execution_"
  "id.name}/{id.node_execution_id.node_id}/"
  "{id.task_id.project}/{id.task_id.domain}"
  "/{id.task_id.name}/{id.task_id.version}/"
  "{id.retry_attempt}\022\361\003\n\022ListTaskExecution"
  "s\022(.flyteidl.admin.TaskExecutionListRequ"
  "est\032!.flyteidl.admin.TaskExecutionList\"\215"
  "\003\202\323\344\223\002\206\003\022\252\001/api/v1/task_executions/{node"
  "_execution_id.execution_id.project}/{nod"
  "e_execution_id.execution_id.domain}/{nod"
  "e_execution_id.execution_id.name}/{node_"
  "execution_id.node_id}Z\326\001\022\323\001/api/v1/task_"
  "executions/org/{node_execution_id.execut"
  "ion_id.org}/{node_execution_id.execution"
  "_id.project}/{node_execution_id.executio"
  "n_id.domain}/{node_execution_id.executio"
  "n_id.name}/{node_execution_id.node_id}\022\354"
  "\005\n\024GetTaskExecutionData\022+.flyteidl.admin"
  ".TaskExecutionGetDataRequest\032,.flyteidl."
  "admin.TaskExecutionGetDataResponse\"\370\004\202\323\344"
  "\223\002\361\004\022\236\002/api/v1/data/task_executions/{id."
  "node_execution_id.execution_id.project}/"
  "{id.node_execution_id.execution_id.domai"
  "n}/{id.node_execution_id.execution_id.na"
  "me}/{id.node_execution_id.node_id}/{id.t"
  "ask_id.project}/{id.task_id.domain}/{id."
  "task_id.name}/{id.task_id.version}/{id.r"
  "etry_attempt}Z\315\002\022\312\002/api/v1/data/task_exe"
  "cutions/org/{id.node_execution_id.execut"
  "ion_id.org}/{id.node_execution_id.execut"
  "ion_id.project}/{id.node_execution_id.ex"
  "ecution_id.domain}/{id.node_execution_id"
  ".execution_id.name}/{id.node_execution_i"
  "d.node_id}/{id.task_id.project}/{id.task"
  "_id.domain}/{id.task_id.name}/{id.task_i"
  "d.version}/{id.retry_attempt}\022\313\002\n\035Update"
  "ProjectDomainAttributes\0224.flyteidl.admin"
  ".ProjectDomainAttributesUpdateRequest\0325."
  "flyteidl.admin.ProjectDomainAttributesUp"
  "dateResponse\"\274\001\202\323\344\223\002\265\001\032J/api/v1/project_"
  "domain_attributes/{attributes.project}/{"
  "attributes.domain}:\001*Zd\032_/api/v1/project"
  "_domain_attributes/org/{attributes.org}/"
  "{attributes.project}/{attributes.domain}"
  ":\001*\022\203\002\n\032GetProjectDomainAttributes\0221.fly"
  "teidl.admin.ProjectDomainAttributesGetRe"
  "quest\0322.flyteidl.admin.ProjectDomainAttr"
  "ibutesGetResponse\"~\202\323\344\223\002x\0224/api/v1/proje"
  "ct_domain_attributes/{project}/{domain}Z"
  "@\022>/api/v1/project_domain_attributes/org"
  "/{org}/{project}/{domain}\022\223\002\n\035DeleteProj"
  "ectDomainAttributes\0224.flyteidl.admin.Pro"
  "jectDomainAttributesDeleteRequest\0325.flyt"
  "eidl.admin.ProjectDomainAttributesDelete"
  "Response\"\204\001\202\323\344\223\002~*4/api/v1/project_domai"
  "n_attributes/{project}/{domain}:\001*ZC*>/a"
  "pi/v1/project_domain_attributes/org/{org"
  "}/{project}/{domain}:\001*\022\212\002\n\027UpdateProjec"
  "tAttributes\022..flyteidl.admin.ProjectAttr"
  "ibutesUpdateRequest\032/.flyteidl.admin.Pro"
  "jectAttributesUpdateResponse\"\215\001\202\323\344\223\002\206\001\032/"
  "/api/v1/project_attributes/{attributes.p"
  "roject}:\001*ZP\032K/api/v1/project_domain_att"
  "ributes/org/{attributes.org}/{attributes"
  ".project}:\001*\022\330\001\n\024GetProjectAttributes\022+."
  "flyteidl.admin.ProjectAttributesGetReque"
  "st\032,.flyteidl.admin.ProjectAttributesGet"
  "Response\"e\202\323\344\223\002_\022$/api/v1/project_attrib"
  "utes/{project}Z7\0225/api/v1/project_domain"
  "_attributes/org/{org}/{project}\022\347\001\n\027Dele"
  "teProjectAttributes\022..flyteidl.admin.Pro"
  "jectAttributesDeleteRequest\032/.flyteidl.a"
  "dmin.ProjectAttributesDeleteResponse\"k\202\323"
  "\344\223\002e*$/api/v1/project_attributes/{projec"
  "t}:\001*Z:*5/api/v1/project_domain_attribut"
  "es/org/{org}/{project}:\001*\022\334\002\n\030UpdateWork"
  "flowAttributes\022/.flyteidl.admin.Workflow"
  "AttributesUpdateRequest\0320.flyteidl.admin"
  ".WorkflowAttributesUpdateResponse\"\334\001\202\323\344\223"
  "\002\325\001\032Z/api/v1/workflow_attributes/{attrib"
  "utes.project}/{attributes.domain}/{attri"
  "butes.workflow}:\001*Zt\032o/api/v1/workflow_a"
  "ttributes/org/{attributes.org}/{attribut"
  "es.project}/{attributes.domain}/{attribu"
  "tes.workflow}:\001*\022\200\002\n\025GetWorkflowAttribut"
  "es\022,.flyteidl.admin.WorkflowAttributesGe"
  "tRequest\032-.flyteidl.admin.WorkflowAttrib"
  "utesGetResponse\"\211\001\202\323\344\223\002\202\001\0229/api/v1/workf"
  "low_attributes/{project}/{domain}/{workf"
  "low}ZE\022C/api/v1/workflow_attributes/org/"
  "{org}/{project}/{domain}/{workflow}\022\217\002\n\030"
  "DeleteWorkflowAttributes\022/.flyteidl.admi"
  "n.WorkflowAttributesDeleteRequest\0320.flyt"
  "eidl.admin.WorkflowAttributesDeleteRespo"
  "nse\"\217\001\202\323\344\223\002\210\001*9/api/v1/workflow_attribut"
  "es/{project}/{domain}/{workflow}:\001*ZH*C/"
  "api/v1/workflow_attributes/org/{org}/{pr"
  "oject}/{domain}/{workflow}:\001*\022\312\001\n\027ListMa"
  "tchableAttributes\022..flyteidl.admin.ListM"
  "atchableAttributesRequest\032/.flyteidl.adm"
  "in.ListMatchableAttributesResponse\"N\202\323\344\223"
  "\002H\022\034/api/v1/matchable_attributesZ(\022&/api"
  "/v1/matchable_attributes/org/{org}\022\350\001\n\021L"
  "istNamedEntities\022&.flyteidl.admin.NamedE"
  "ntityListRequest\032\037.flyteidl.admin.NamedE"
  "ntityList\"\211\001\202\323\344\223\002\202\001\0229/api/v1/named_entit"
  "ies/{resource_type}/{project}/{domain}ZE"
  "\022C/api/v1/named_entities/{resource_type}"
  "/org/{org}/{project}/{domain}\022\203\002\n\016GetNam"
  "edEntity\022%.flyteidl.admin.NamedEntityGet"
  "Request\032\033.flyteidl.admin.NamedEntity\"\254\001\202"
  "\323\344\223\002\245\001\022I/api/v1/named_entities/{resource"
  "_type}/{id.project}/{id.domain}/{id.name"
  "}ZX\022V/api/v1/named_entities/{resource_ty"
  "pe}/org/{id.org}/{id.project}/{id.domain"
  "}/{id.name}\022\235\002\n\021UpdateNamedEntity\022(.flyt"
  "eidl.admin.NamedEntityUpdateRequest\032).fl"
  "yteidl.admin.NamedEntityUpdateResponse\"\262"
  "\001\202\323\344\223\002\253\001\032I/api/v1/named_entities/{resour"
  "ce_type}/{id.project}/{id.domain}/{id.na"
  "me}:\001*Z[\032V/api/v1/named_entities/{resour"
  "ce_type}/org/{id.org}/{id.project}/{id.d"
  "omain}/{id.name}:\001*\022l\n\nGetVersion\022!.flyt"
  "eidl.admin.GetVersionRequest\032\".flyteidl."
  "admin.GetVersionResponse\"\027\202\323\344\223\002\021\022\017/api/v"
  "1/version\022\266\002\n\024GetDescriptionEntity\022 .fly"
  "teidl.admin.ObjectGetRequest\032!.flyteidl."
  "admin.DescriptionEntity\"\330\001\202\323\344\223\002\321\001\022_/api/"
  "v1/description_entities/{id.resource_typ"
  "e}/{id.project}/{id.domain}/{id.name}/{i"
  "d.version}Zn\022l/api/v1/description_entiti"
  "es/org/{id.org}/{id.resource_type}/{id.p"
  "roject}/{id.domain}/{id.name}/{id.versio"
  "n}\022\310\003\n\027ListDescriptionEntities\022,.flyteid"
  "l.admin.DescriptionEntityListRequest\032%.f"
  "lyteidl.admin.DescriptionEntityList\"\327\002\202\323"
  "\344\223\002\320\002\022O/api/v1/description_entities/{res"
  "ource_type}/{id.project}/{id.domain}/{id"
  ".name}Z^\022\\/api/v1/description_entities/{"
  "resource_type}/org/{id.org}/{id.project}"
  "/{id.domain}/{id.name}ZG\022E/api/v1/descri"
  "ption_entities/{resource_type}/{id.proje"
  "ct}/{id.domain}ZT\022R/api/v1/description_e"
  "ntities/{resource_type}/org/{id.org}/{id"
  ".project}/{id.domain}\022\225\002\n\023GetExecutionMe"
  "trics\0222.flyteidl.admin.WorkflowExecution"
  "GetMetricsRequest\0323.flyteidl.admin.Workf"
  "lowExecutionGetMetricsResponse\"\224\001\202\323\344\223\002\215\001"
  "\022=/api/v1/metrics/executions/{id.project"
  "}/{id.domain}/{id.name}ZL\022J/api/v1/metri"
  "cs/executions/org/{id.org}/{id.project}/"
  "{id.domain}/{id.name}B\?Z=github.com/flyt"
  "eorg/flyte/flyteidl/gen/pb-go/flyteidl/s"
  "erviceb\006proto3"
  ;
::google::protobuf::internal::DescriptorTable descriptor_table_flyteidl_2fservice_2fadmin_2eproto = {
  false, InitDefaults_flyteidl_2fservice_2fadmin_2eproto, 
  descriptor_table_protodef_flyteidl_2fservice_2fadmin_2eproto,
  "flyteidl/service/admin.proto", &assign_descriptors_table_flyteidl_2fservice_2fadmin_2eproto, 16054,
};

void AddDescriptors_flyteidl_2fservice_2fadmin_2eproto() {
  static constexpr ::google::protobuf::internal::InitFunc deps[16] =
  {
    ::AddDescriptors_google_2fapi_2fannotations_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fproject_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fproject_5fdomain_5fattributes_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fproject_5fattributes_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2ftask_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fworkflow_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fworkflow_5fattributes_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2flaunch_5fplan_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fevent_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fexecution_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fmatchable_5fresource_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fnode_5fexecution_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2ftask_5fexecution_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fversion_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fcommon_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fdescription_5fentity_2eproto,
  };
 ::google::protobuf::internal::AddDescriptors(&descriptor_table_flyteidl_2fservice_2fadmin_2eproto, deps, 16);
}

// Force running AddDescriptors() at dynamic initialization time.
static bool dynamic_init_dummy_flyteidl_2fservice_2fadmin_2eproto = []() { AddDescriptors_flyteidl_2fservice_2fadmin_2eproto(); return true; }();
namespace flyteidl {
namespace service {

// @@protoc_insertion_point(namespace_scope)
}  // namespace service
}  // namespace flyteidl
namespace google {
namespace protobuf {
}  // namespace protobuf
}  // namespace google

// @@protoc_insertion_point(global_scope)
#include <google/protobuf/port_undef.inc>
