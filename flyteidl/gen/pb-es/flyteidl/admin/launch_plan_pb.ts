// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/admin/launch_plan.proto (package flyteidl.admin, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Any, BoolValue, Message, proto3, Timestamp } from "@bufbuild/protobuf";
import { Identifier } from "../core/identifier_pb.js";
import { ParameterMap, VariableMap } from "../core/interface_pb.js";
import { LiteralMap } from "../core/literals_pb.js";
import { Annotations, AuthRole, Envs, Labels, NamedEntityIdentifier, Notification, RawOutputDataConfig, Sort } from "./common_pb.js";
import { SecurityContext } from "../core/security_pb.js";
import { QualityOfService } from "../core/execution_pb.js";
import { ExecutionEnvAssignment } from "../core/execution_envs_pb.js";
import { Schedule } from "./schedule_pb.js";
import { Node } from "../core/workflow_pb.js";

/**
 * By default any launch plan regardless of state can be used to launch a workflow execution.
 * However, at most one version of a launch plan
 * (e.g. a NamedEntityIdentifier set of shared project, domain and name values) can be
 * active at a time in regards to *schedules*. That is, at most one schedule in a NamedEntityIdentifier
 * group will be observed and trigger executions at a defined cadence.
 *
 * @generated from enum flyteidl.admin.LaunchPlanState
 */
export enum LaunchPlanState {
  /**
   * @generated from enum value: INACTIVE = 0;
   */
  INACTIVE = 0,

  /**
   * @generated from enum value: ACTIVE = 1;
   */
  ACTIVE = 1,
}
// Retrieve enum metadata with: proto3.getEnumType(LaunchPlanState)
proto3.util.setEnumType(LaunchPlanState, "flyteidl.admin.LaunchPlanState", [
  { no: 0, name: "INACTIVE" },
  { no: 1, name: "ACTIVE" },
]);

/**
 * Request to register a launch plan. The included LaunchPlanSpec may have a complete or incomplete set of inputs required
 * to launch a workflow execution. By default all launch plans are registered in state INACTIVE. If you wish to
 * set the state to ACTIVE, you must submit a LaunchPlanUpdateRequest, after you have successfully created a launch plan.
 *
 * @generated from message flyteidl.admin.LaunchPlanCreateRequest
 */
export class LaunchPlanCreateRequest extends Message<LaunchPlanCreateRequest> {
  /**
   * Uniquely identifies a launch plan entity.
   *
   * @generated from field: flyteidl.core.Identifier id = 1;
   */
  id?: Identifier;

  /**
   * User-provided launch plan details, including reference workflow, inputs and other metadata.
   *
   * @generated from field: flyteidl.admin.LaunchPlanSpec spec = 2;
   */
  spec?: LaunchPlanSpec;

  constructor(data?: PartialMessage<LaunchPlanCreateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlanCreateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: Identifier },
    { no: 2, name: "spec", kind: "message", T: LaunchPlanSpec },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlanCreateRequest {
    return new LaunchPlanCreateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlanCreateRequest {
    return new LaunchPlanCreateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlanCreateRequest {
    return new LaunchPlanCreateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlanCreateRequest | PlainMessage<LaunchPlanCreateRequest> | undefined, b: LaunchPlanCreateRequest | PlainMessage<LaunchPlanCreateRequest> | undefined): boolean {
    return proto3.util.equals(LaunchPlanCreateRequest, a, b);
  }
}

/**
 * Purposefully empty, may be populated in the future.
 *
 * @generated from message flyteidl.admin.LaunchPlanCreateResponse
 */
export class LaunchPlanCreateResponse extends Message<LaunchPlanCreateResponse> {
  constructor(data?: PartialMessage<LaunchPlanCreateResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlanCreateResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlanCreateResponse {
    return new LaunchPlanCreateResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlanCreateResponse {
    return new LaunchPlanCreateResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlanCreateResponse {
    return new LaunchPlanCreateResponse().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlanCreateResponse | PlainMessage<LaunchPlanCreateResponse> | undefined, b: LaunchPlanCreateResponse | PlainMessage<LaunchPlanCreateResponse> | undefined): boolean {
    return proto3.util.equals(LaunchPlanCreateResponse, a, b);
  }
}

/**
 * A LaunchPlan provides the capability to templatize workflow executions.
 * Launch plans simplify associating one or more schedules, inputs and notifications with your workflows.
 * Launch plans can be shared and used to trigger executions with predefined inputs even when a workflow
 * definition doesn't necessarily have a default value for said input.
 *
 * @generated from message flyteidl.admin.LaunchPlan
 */
export class LaunchPlan extends Message<LaunchPlan> {
  /**
   * Uniquely identifies a launch plan entity.
   *
   * @generated from field: flyteidl.core.Identifier id = 1;
   */
  id?: Identifier;

  /**
   * User-provided launch plan details, including reference workflow, inputs and other metadata.
   *
   * @generated from field: flyteidl.admin.LaunchPlanSpec spec = 2;
   */
  spec?: LaunchPlanSpec;

  /**
   * Values computed by the flyte platform after launch plan registration.
   *
   * @generated from field: flyteidl.admin.LaunchPlanClosure closure = 3;
   */
  closure?: LaunchPlanClosure;

  constructor(data?: PartialMessage<LaunchPlan>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlan";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: Identifier },
    { no: 2, name: "spec", kind: "message", T: LaunchPlanSpec },
    { no: 3, name: "closure", kind: "message", T: LaunchPlanClosure },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlan {
    return new LaunchPlan().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlan {
    return new LaunchPlan().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlan {
    return new LaunchPlan().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlan | PlainMessage<LaunchPlan> | undefined, b: LaunchPlan | PlainMessage<LaunchPlan> | undefined): boolean {
    return proto3.util.equals(LaunchPlan, a, b);
  }
}

/**
 * Response object for list launch plan requests.
 * See :ref:`ref_flyteidl.admin.LaunchPlan` for more details
 *
 * @generated from message flyteidl.admin.LaunchPlanList
 */
export class LaunchPlanList extends Message<LaunchPlanList> {
  /**
   * @generated from field: repeated flyteidl.admin.LaunchPlan launch_plans = 1;
   */
  launchPlans: LaunchPlan[] = [];

  /**
   * In the case of multiple pages of results, the server-provided token can be used to fetch the next page
   * in a query. If there are no more results, this value will be empty.
   *
   * @generated from field: string token = 2;
   */
  token = "";

  constructor(data?: PartialMessage<LaunchPlanList>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlanList";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "launch_plans", kind: "message", T: LaunchPlan, repeated: true },
    { no: 2, name: "token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlanList {
    return new LaunchPlanList().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlanList {
    return new LaunchPlanList().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlanList {
    return new LaunchPlanList().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlanList | PlainMessage<LaunchPlanList> | undefined, b: LaunchPlanList | PlainMessage<LaunchPlanList> | undefined): boolean {
    return proto3.util.equals(LaunchPlanList, a, b);
  }
}

/**
 * Defines permissions associated with executions created by this launch plan spec.
 * Use either of these roles when they have permissions required by your workflow execution.
 * Deprecated.
 *
 * @generated from message flyteidl.admin.Auth
 * @deprecated
 */
export class Auth extends Message<Auth> {
  /**
   * Defines an optional iam role which will be used for tasks run in executions created with this launch plan.
   *
   * @generated from field: string assumable_iam_role = 1;
   */
  assumableIamRole = "";

  /**
   * Defines an optional kubernetes service account which will be used for tasks run in executions created with this launch plan.
   *
   * @generated from field: string kubernetes_service_account = 2;
   */
  kubernetesServiceAccount = "";

  constructor(data?: PartialMessage<Auth>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.Auth";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "assumable_iam_role", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "kubernetes_service_account", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Auth {
    return new Auth().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Auth {
    return new Auth().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Auth {
    return new Auth().fromJsonString(jsonString, options);
  }

  static equals(a: Auth | PlainMessage<Auth> | undefined, b: Auth | PlainMessage<Auth> | undefined): boolean {
    return proto3.util.equals(Auth, a, b);
  }
}

/**
 * User-provided launch plan definition and configuration values.
 *
 * @generated from message flyteidl.admin.LaunchPlanSpec
 */
export class LaunchPlanSpec extends Message<LaunchPlanSpec> {
  /**
   * Reference to the Workflow template that the launch plan references
   *
   * @generated from field: flyteidl.core.Identifier workflow_id = 1;
   */
  workflowId?: Identifier;

  /**
   * Metadata for the Launch Plan
   *
   * @generated from field: flyteidl.admin.LaunchPlanMetadata entity_metadata = 2;
   */
  entityMetadata?: LaunchPlanMetadata;

  /**
   * Input values to be passed for the execution.
   * These can be overridden when an execution is created with this launch plan.
   *
   * @generated from field: flyteidl.core.ParameterMap default_inputs = 3;
   */
  defaultInputs?: ParameterMap;

  /**
   * Fixed, non-overridable inputs for the Launch Plan.
   * These can not be overridden when an execution is created with this launch plan.
   *
   * @generated from field: flyteidl.core.LiteralMap fixed_inputs = 4;
   */
  fixedInputs?: LiteralMap;

  /**
   * String to indicate the role to use to execute the workflow underneath
   *
   * @generated from field: string role = 5 [deprecated = true];
   * @deprecated
   */
  role = "";

  /**
   * Custom labels to be applied to the execution resource.
   *
   * @generated from field: flyteidl.admin.Labels labels = 6;
   */
  labels?: Labels;

  /**
   * Custom annotations to be applied to the execution resource.
   *
   * @generated from field: flyteidl.admin.Annotations annotations = 7;
   */
  annotations?: Annotations;

  /**
   * Indicates the permission associated with workflow executions triggered with this launch plan.
   *
   * @generated from field: flyteidl.admin.Auth auth = 8 [deprecated = true];
   * @deprecated
   */
  auth?: Auth;

  /**
   * @generated from field: flyteidl.admin.AuthRole auth_role = 9 [deprecated = true];
   * @deprecated
   */
  authRole?: AuthRole;

  /**
   * Indicates security context for permissions triggered with this launch plan
   *
   * @generated from field: flyteidl.core.SecurityContext security_context = 10;
   */
  securityContext?: SecurityContext;

  /**
   * Indicates the runtime priority of the execution.
   *
   * @generated from field: flyteidl.core.QualityOfService quality_of_service = 16;
   */
  qualityOfService?: QualityOfService;

  /**
   * Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).
   *
   * @generated from field: flyteidl.admin.RawOutputDataConfig raw_output_data_config = 17;
   */
  rawOutputDataConfig?: RawOutputDataConfig;

  /**
   * Controls the maximum number of tasknodes that can be run in parallel for the entire workflow.
   * This is useful to achieve fairness. Note: MapTasks are regarded as one unit,
   * and parallelism/concurrency of MapTasks is independent from this.
   *
   * @generated from field: int32 max_parallelism = 18;
   */
  maxParallelism = 0;

  /**
   * Allows for the interruptible flag of a workflow to be overwritten for a single execution.
   * Omitting this field uses the workflow's value as a default.
   * As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper
   * around the bool field.
   *
   * @generated from field: google.protobuf.BoolValue interruptible = 19;
   */
  interruptible?: boolean;

  /**
   * Allows for all cached values of a workflow and its tasks to be overwritten for a single execution.
   * If enabled, all calculations are performed even if cached results would be available, overwriting the stored
   * data once execution finishes successfully.
   *
   * @generated from field: bool overwrite_cache = 20;
   */
  overwriteCache = false;

  /**
   * Environment variables to be set for the execution.
   *
   * @generated from field: flyteidl.admin.Envs envs = 21;
   */
  envs?: Envs;

  /**
   * Execution environment assignments to be set for the execution.
   *
   * @generated from field: repeated flyteidl.core.ExecutionEnvAssignment execution_env_assignments = 22;
   */
  executionEnvAssignments: ExecutionEnvAssignment[] = [];

  constructor(data?: PartialMessage<LaunchPlanSpec>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlanSpec";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "workflow_id", kind: "message", T: Identifier },
    { no: 2, name: "entity_metadata", kind: "message", T: LaunchPlanMetadata },
    { no: 3, name: "default_inputs", kind: "message", T: ParameterMap },
    { no: 4, name: "fixed_inputs", kind: "message", T: LiteralMap },
    { no: 5, name: "role", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "labels", kind: "message", T: Labels },
    { no: 7, name: "annotations", kind: "message", T: Annotations },
    { no: 8, name: "auth", kind: "message", T: Auth },
    { no: 9, name: "auth_role", kind: "message", T: AuthRole },
    { no: 10, name: "security_context", kind: "message", T: SecurityContext },
    { no: 16, name: "quality_of_service", kind: "message", T: QualityOfService },
    { no: 17, name: "raw_output_data_config", kind: "message", T: RawOutputDataConfig },
    { no: 18, name: "max_parallelism", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 19, name: "interruptible", kind: "message", T: BoolValue },
    { no: 20, name: "overwrite_cache", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 21, name: "envs", kind: "message", T: Envs },
    { no: 22, name: "execution_env_assignments", kind: "message", T: ExecutionEnvAssignment, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlanSpec {
    return new LaunchPlanSpec().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlanSpec {
    return new LaunchPlanSpec().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlanSpec {
    return new LaunchPlanSpec().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlanSpec | PlainMessage<LaunchPlanSpec> | undefined, b: LaunchPlanSpec | PlainMessage<LaunchPlanSpec> | undefined): boolean {
    return proto3.util.equals(LaunchPlanSpec, a, b);
  }
}

/**
 * Values computed by the flyte platform after launch plan registration.
 * These include expected_inputs required to be present in a CreateExecutionRequest
 * to launch the reference workflow as well timestamp values associated with the launch plan.
 *
 * @generated from message flyteidl.admin.LaunchPlanClosure
 */
export class LaunchPlanClosure extends Message<LaunchPlanClosure> {
  /**
   * Indicate the Launch plan state. 
   *
   * @generated from field: flyteidl.admin.LaunchPlanState state = 1;
   */
  state = LaunchPlanState.INACTIVE;

  /**
   * Indicates the set of inputs expected when creating an execution with the Launch plan
   *
   * @generated from field: flyteidl.core.ParameterMap expected_inputs = 2;
   */
  expectedInputs?: ParameterMap;

  /**
   * Indicates the set of outputs expected to be produced by creating an execution with the Launch plan
   *
   * @generated from field: flyteidl.core.VariableMap expected_outputs = 3;
   */
  expectedOutputs?: VariableMap;

  /**
   * Time at which the launch plan was created.
   *
   * @generated from field: google.protobuf.Timestamp created_at = 4;
   */
  createdAt?: Timestamp;

  /**
   * Time at which the launch plan was last updated.
   *
   * @generated from field: google.protobuf.Timestamp updated_at = 5;
   */
  updatedAt?: Timestamp;

  constructor(data?: PartialMessage<LaunchPlanClosure>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlanClosure";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "state", kind: "enum", T: proto3.getEnumType(LaunchPlanState) },
    { no: 2, name: "expected_inputs", kind: "message", T: ParameterMap },
    { no: 3, name: "expected_outputs", kind: "message", T: VariableMap },
    { no: 4, name: "created_at", kind: "message", T: Timestamp },
    { no: 5, name: "updated_at", kind: "message", T: Timestamp },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlanClosure {
    return new LaunchPlanClosure().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlanClosure {
    return new LaunchPlanClosure().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlanClosure {
    return new LaunchPlanClosure().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlanClosure | PlainMessage<LaunchPlanClosure> | undefined, b: LaunchPlanClosure | PlainMessage<LaunchPlanClosure> | undefined): boolean {
    return proto3.util.equals(LaunchPlanClosure, a, b);
  }
}

/**
 * Additional launch plan attributes included in the LaunchPlanSpec not strictly required to launch
 * the reference workflow.
 *
 * @generated from message flyteidl.admin.LaunchPlanMetadata
 */
export class LaunchPlanMetadata extends Message<LaunchPlanMetadata> {
  /**
   * Schedule to execute the Launch Plan
   *
   * @generated from field: flyteidl.admin.Schedule schedule = 1;
   */
  schedule?: Schedule;

  /**
   * List of notifications based on Execution status transitions
   *
   * @generated from field: repeated flyteidl.admin.Notification notifications = 2;
   */
  notifications: Notification[] = [];

  /**
   * Additional metadata for how to launch the launch plan
   *
   * @generated from field: google.protobuf.Any launch_conditions = 3;
   */
  launchConditions?: Any;

  constructor(data?: PartialMessage<LaunchPlanMetadata>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlanMetadata";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "schedule", kind: "message", T: Schedule },
    { no: 2, name: "notifications", kind: "message", T: Notification, repeated: true },
    { no: 3, name: "launch_conditions", kind: "message", T: Any },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlanMetadata {
    return new LaunchPlanMetadata().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlanMetadata {
    return new LaunchPlanMetadata().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlanMetadata {
    return new LaunchPlanMetadata().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlanMetadata | PlainMessage<LaunchPlanMetadata> | undefined, b: LaunchPlanMetadata | PlainMessage<LaunchPlanMetadata> | undefined): boolean {
    return proto3.util.equals(LaunchPlanMetadata, a, b);
  }
}

/**
 * Request to set the referenced launch plan state to the configured value.
 * See :ref:`ref_flyteidl.admin.LaunchPlan` for more details
 *
 * @generated from message flyteidl.admin.LaunchPlanUpdateRequest
 */
export class LaunchPlanUpdateRequest extends Message<LaunchPlanUpdateRequest> {
  /**
   * Identifier of launch plan for which to change state.
   * +required.
   *
   * @generated from field: flyteidl.core.Identifier id = 1;
   */
  id?: Identifier;

  /**
   * Desired state to apply to the launch plan.
   * +required.
   *
   * @generated from field: flyteidl.admin.LaunchPlanState state = 2;
   */
  state = LaunchPlanState.INACTIVE;

  constructor(data?: PartialMessage<LaunchPlanUpdateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlanUpdateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: Identifier },
    { no: 2, name: "state", kind: "enum", T: proto3.getEnumType(LaunchPlanState) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlanUpdateRequest {
    return new LaunchPlanUpdateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlanUpdateRequest {
    return new LaunchPlanUpdateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlanUpdateRequest {
    return new LaunchPlanUpdateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlanUpdateRequest | PlainMessage<LaunchPlanUpdateRequest> | undefined, b: LaunchPlanUpdateRequest | PlainMessage<LaunchPlanUpdateRequest> | undefined): boolean {
    return proto3.util.equals(LaunchPlanUpdateRequest, a, b);
  }
}

/**
 * Purposefully empty, may be populated in the future.
 *
 * @generated from message flyteidl.admin.LaunchPlanUpdateResponse
 */
export class LaunchPlanUpdateResponse extends Message<LaunchPlanUpdateResponse> {
  constructor(data?: PartialMessage<LaunchPlanUpdateResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.LaunchPlanUpdateResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): LaunchPlanUpdateResponse {
    return new LaunchPlanUpdateResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): LaunchPlanUpdateResponse {
    return new LaunchPlanUpdateResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): LaunchPlanUpdateResponse {
    return new LaunchPlanUpdateResponse().fromJsonString(jsonString, options);
  }

  static equals(a: LaunchPlanUpdateResponse | PlainMessage<LaunchPlanUpdateResponse> | undefined, b: LaunchPlanUpdateResponse | PlainMessage<LaunchPlanUpdateResponse> | undefined): boolean {
    return proto3.util.equals(LaunchPlanUpdateResponse, a, b);
  }
}

/**
 * Represents a request struct for finding an active launch plan for a given NamedEntityIdentifier
 * See :ref:`ref_flyteidl.admin.LaunchPlan` for more details
 *
 * @generated from message flyteidl.admin.ActiveLaunchPlanRequest
 */
export class ActiveLaunchPlanRequest extends Message<ActiveLaunchPlanRequest> {
  /**
   * +required.
   *
   * @generated from field: flyteidl.admin.NamedEntityIdentifier id = 1;
   */
  id?: NamedEntityIdentifier;

  constructor(data?: PartialMessage<ActiveLaunchPlanRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ActiveLaunchPlanRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: NamedEntityIdentifier },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ActiveLaunchPlanRequest {
    return new ActiveLaunchPlanRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ActiveLaunchPlanRequest {
    return new ActiveLaunchPlanRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ActiveLaunchPlanRequest {
    return new ActiveLaunchPlanRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ActiveLaunchPlanRequest | PlainMessage<ActiveLaunchPlanRequest> | undefined, b: ActiveLaunchPlanRequest | PlainMessage<ActiveLaunchPlanRequest> | undefined): boolean {
    return proto3.util.equals(ActiveLaunchPlanRequest, a, b);
  }
}

/**
 * Represents a request structure to list active launch plans within a project/domain and optional org.
 * See :ref:`ref_flyteidl.admin.LaunchPlan` for more details
 *
 * @generated from message flyteidl.admin.ActiveLaunchPlanListRequest
 */
export class ActiveLaunchPlanListRequest extends Message<ActiveLaunchPlanListRequest> {
  /**
   * Name of the project that contains the identifiers.
   * +required.
   *
   * @generated from field: string project = 1;
   */
  project = "";

  /**
   * Name of the domain the identifiers belongs to within the project.
   * +required.
   *
   * @generated from field: string domain = 2;
   */
  domain = "";

  /**
   * Indicates the number of resources to be returned.
   * +required.
   *
   * @generated from field: uint32 limit = 3;
   */
  limit = 0;

  /**
   * In the case of multiple pages of results, the server-provided token can be used to fetch the next page
   * in a query.
   * +optional
   *
   * @generated from field: string token = 4;
   */
  token = "";

  /**
   * Sort ordering.
   * +optional
   *
   * @generated from field: flyteidl.admin.Sort sort_by = 5;
   */
  sortBy?: Sort;

  /**
   * Optional, org key applied to the resource.
   *
   * @generated from field: string org = 6;
   */
  org = "";

  constructor(data?: PartialMessage<ActiveLaunchPlanListRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ActiveLaunchPlanListRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "project", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "domain", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "limit", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 4, name: "token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "sort_by", kind: "message", T: Sort },
    { no: 6, name: "org", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ActiveLaunchPlanListRequest {
    return new ActiveLaunchPlanListRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ActiveLaunchPlanListRequest {
    return new ActiveLaunchPlanListRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ActiveLaunchPlanListRequest {
    return new ActiveLaunchPlanListRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ActiveLaunchPlanListRequest | PlainMessage<ActiveLaunchPlanListRequest> | undefined, b: ActiveLaunchPlanListRequest | PlainMessage<ActiveLaunchPlanListRequest> | undefined): boolean {
    return proto3.util.equals(ActiveLaunchPlanListRequest, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.SubNodeIdAsList
 */
export class SubNodeIdAsList extends Message<SubNodeIdAsList> {
  /**
   * subNodeID for a node within a workflow. If more then one subMode ID is provided,
   * then it is assumed to be a subNode within a subWorkflow
   *
   * @generated from field: repeated string sub_node_id = 1;
   */
  subNodeId: string[] = [];

  constructor(data?: PartialMessage<SubNodeIdAsList>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SubNodeIdAsList";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "sub_node_id", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SubNodeIdAsList {
    return new SubNodeIdAsList().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SubNodeIdAsList {
    return new SubNodeIdAsList().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SubNodeIdAsList {
    return new SubNodeIdAsList().fromJsonString(jsonString, options);
  }

  static equals(a: SubNodeIdAsList | PlainMessage<SubNodeIdAsList> | undefined, b: SubNodeIdAsList | PlainMessage<SubNodeIdAsList> | undefined): boolean {
    return proto3.util.equals(SubNodeIdAsList, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.SubNodeList
 */
export class SubNodeList extends Message<SubNodeList> {
  /**
   * List of sub node IDs to include in the execution
   *
   * @generated from field: repeated flyteidl.admin.SubNodeIdAsList sub_node_ids = 1;
   */
  subNodeIds: SubNodeIdAsList[] = [];

  constructor(data?: PartialMessage<SubNodeList>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SubNodeList";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "sub_node_ids", kind: "message", T: SubNodeIdAsList, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SubNodeList {
    return new SubNodeList().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SubNodeList {
    return new SubNodeList().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SubNodeList {
    return new SubNodeList().fromJsonString(jsonString, options);
  }

  static equals(a: SubNodeList | PlainMessage<SubNodeList> | undefined, b: SubNodeList | PlainMessage<SubNodeList> | undefined): boolean {
    return proto3.util.equals(SubNodeList, a, b);
  }
}

/**
 * Request to create or fetch a launch plan to re-run a node
 *
 * @generated from message flyteidl.admin.CreateLaunchPlanFromNodeRequest
 */
export class CreateLaunchPlanFromNodeRequest extends Message<CreateLaunchPlanFromNodeRequest> {
  /**
   * ID of the launch plan that executed the workflow that contains the node to be re-run
   * or the core.Identifier values aside from name used to create a new workflow when using a node spec
   * +required
   *
   * @generated from field: flyteidl.core.Identifier launch_plan_id = 1;
   */
  launchPlanId?: Identifier;

  /**
   * +required
   *
   * @generated from oneof flyteidl.admin.CreateLaunchPlanFromNodeRequest.sub_nodes
   */
  subNodes: {
    /**
     * List of sub node IDs to include in the execution. Utilized for re-running a node(s) within a workflow.
     *
     * @generated from field: flyteidl.admin.SubNodeList sub_node_ids = 2;
     */
    value: SubNodeList;
    case: "subNodeIds";
  } | {
    /**
     * Node spec to create a workflow from
     *
     * @generated from field: flyteidl.core.Node sub_node_spec = 3;
     */
    value: Node;
    case: "subNodeSpec";
  } | { case: undefined; value?: undefined } = { case: undefined };

  /**
   * Indicates security context for permissions triggered with this launch plan
   *
   * @generated from field: flyteidl.core.SecurityContext security_context = 4;
   */
  securityContext?: SecurityContext;

  /**
   * Optional name for the workflow & launch plan to be created. If not provided, a name will be generated.
   *
   * @generated from field: string name = 5;
   */
  name = "";

  constructor(data?: PartialMessage<CreateLaunchPlanFromNodeRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.CreateLaunchPlanFromNodeRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "launch_plan_id", kind: "message", T: Identifier },
    { no: 2, name: "sub_node_ids", kind: "message", T: SubNodeList, oneof: "sub_nodes" },
    { no: 3, name: "sub_node_spec", kind: "message", T: Node, oneof: "sub_nodes" },
    { no: 4, name: "security_context", kind: "message", T: SecurityContext },
    { no: 5, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateLaunchPlanFromNodeRequest {
    return new CreateLaunchPlanFromNodeRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateLaunchPlanFromNodeRequest {
    return new CreateLaunchPlanFromNodeRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateLaunchPlanFromNodeRequest {
    return new CreateLaunchPlanFromNodeRequest().fromJsonString(jsonString, options);
  }

  static equals(a: CreateLaunchPlanFromNodeRequest | PlainMessage<CreateLaunchPlanFromNodeRequest> | undefined, b: CreateLaunchPlanFromNodeRequest | PlainMessage<CreateLaunchPlanFromNodeRequest> | undefined): boolean {
    return proto3.util.equals(CreateLaunchPlanFromNodeRequest, a, b);
  }
}

/**
 * The launch plan that was created for or that already existed from re-running node(s) within a workflow
 *
 * @generated from message flyteidl.admin.CreateLaunchPlanFromNodeResponse
 */
export class CreateLaunchPlanFromNodeResponse extends Message<CreateLaunchPlanFromNodeResponse> {
  /**
   * @generated from field: flyteidl.admin.LaunchPlan launch_plan = 1;
   */
  launchPlan?: LaunchPlan;

  constructor(data?: PartialMessage<CreateLaunchPlanFromNodeResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.CreateLaunchPlanFromNodeResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "launch_plan", kind: "message", T: LaunchPlan },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateLaunchPlanFromNodeResponse {
    return new CreateLaunchPlanFromNodeResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateLaunchPlanFromNodeResponse {
    return new CreateLaunchPlanFromNodeResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateLaunchPlanFromNodeResponse {
    return new CreateLaunchPlanFromNodeResponse().fromJsonString(jsonString, options);
  }

  static equals(a: CreateLaunchPlanFromNodeResponse | PlainMessage<CreateLaunchPlanFromNodeResponse> | undefined, b: CreateLaunchPlanFromNodeResponse | PlainMessage<CreateLaunchPlanFromNodeResponse> | undefined): boolean {
    return proto3.util.equals(CreateLaunchPlanFromNodeResponse, a, b);
  }
}

