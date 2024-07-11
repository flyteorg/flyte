// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/admin/configuration.proto (package flyteidl.admin, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { ClusterResourceAttributes, ExecutionClusterLabel, ExecutionQueueAttributes, ExternalResourceAttributes, PluginOverrides, TaskResourceAttributes, WorkflowExecutionConfig } from "./matchable_resource_pb.js";
import { QualityOfService } from "../core/execution_pb.js";
import { ClusterAssignment } from "./cluster_assignment_pb.js";

/**
 * The source of an attribute. We may have other sources in the future.
 *
 * @generated from enum flyteidl.admin.AttributesSource
 */
export enum AttributesSource {
  /**
   * The source is unspecified.
   *
   * @generated from enum value: SOURCE_UNSPECIFIED = 0;
   */
  SOURCE_UNSPECIFIED = 0,

  /**
   * The configuration is a global configuration.
   *
   * @generated from enum value: GLOBAL = 1;
   */
  GLOBAL = 1,

  /**
   * The configuration is a domain configuration.
   *
   * @generated from enum value: DOMAIN = 2;
   */
  DOMAIN = 2,

  /**
   * The configuration is a project configuration.
   *
   * @generated from enum value: PROJECT = 3;
   */
  PROJECT = 3,

  /**
   * The configuration is a project-domain configuration.
   *
   * @generated from enum value: PROJECT_DOMAIN = 4;
   */
  PROJECT_DOMAIN = 4,

  /**
   * The configuration is a org configuration.
   *
   * @generated from enum value: ORG = 5;
   */
  ORG = 5,
}
// Retrieve enum metadata with: proto3.getEnumType(AttributesSource)
proto3.util.setEnumType(AttributesSource, "flyteidl.admin.AttributesSource", [
  { no: 0, name: "SOURCE_UNSPECIFIED" },
  { no: 1, name: "GLOBAL" },
  { no: 2, name: "DOMAIN" },
  { no: 3, name: "PROJECT" },
  { no: 4, name: "PROJECT_DOMAIN" },
  { no: 5, name: "ORG" },
]);

/**
 * Identifier for a configuration.
 *
 * @generated from message flyteidl.admin.ConfigurationID
 */
export class ConfigurationID extends Message<ConfigurationID> {
  /**
   * Name of the org the configuration belongs to.
   * +optional
   *
   * @generated from field: string org = 1;
   */
  org = "";

  /**
   * Name of the domain the configuration belongs to.
   * +optional
   *
   * @generated from field: string domain = 2;
   */
  domain = "";

  /**
   * Name of the project the configuration belongs to.
   * +optional
   *
   * @generated from field: string project = 3;
   */
  project = "";

  /**
   * Name of the workflow the configuration belongs to.
   * +optional
   *
   * @generated from field: string workflow = 4;
   */
  workflow = "";

  /**
   * If it is a global configuration.
   * +optional
   *
   * @generated from field: bool global = 5;
   */
  global = false;

  constructor(data?: PartialMessage<ConfigurationID>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ConfigurationID";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "domain", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "project", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "workflow", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "global", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigurationID {
    return new ConfigurationID().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigurationID {
    return new ConfigurationID().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigurationID {
    return new ConfigurationID().fromJsonString(jsonString, options);
  }

  static equals(a: ConfigurationID | PlainMessage<ConfigurationID> | undefined, b: ConfigurationID | PlainMessage<ConfigurationID> | undefined): boolean {
    return proto3.util.equals(ConfigurationID, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.TaskResourceAttributesWithSource
 */
export class TaskResourceAttributesWithSource extends Message<TaskResourceAttributesWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.admin.TaskResourceAttributes value = 2;
   */
  value?: TaskResourceAttributes;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<TaskResourceAttributesWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.TaskResourceAttributesWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: TaskResourceAttributes },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TaskResourceAttributesWithSource {
    return new TaskResourceAttributesWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TaskResourceAttributesWithSource {
    return new TaskResourceAttributesWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TaskResourceAttributesWithSource {
    return new TaskResourceAttributesWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: TaskResourceAttributesWithSource | PlainMessage<TaskResourceAttributesWithSource> | undefined, b: TaskResourceAttributesWithSource | PlainMessage<TaskResourceAttributesWithSource> | undefined): boolean {
    return proto3.util.equals(TaskResourceAttributesWithSource, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.ClusterResourceAttributesWithSource
 */
export class ClusterResourceAttributesWithSource extends Message<ClusterResourceAttributesWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.admin.ClusterResourceAttributes value = 2;
   */
  value?: ClusterResourceAttributes;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<ClusterResourceAttributesWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ClusterResourceAttributesWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: ClusterResourceAttributes },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ClusterResourceAttributesWithSource {
    return new ClusterResourceAttributesWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ClusterResourceAttributesWithSource {
    return new ClusterResourceAttributesWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ClusterResourceAttributesWithSource {
    return new ClusterResourceAttributesWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: ClusterResourceAttributesWithSource | PlainMessage<ClusterResourceAttributesWithSource> | undefined, b: ClusterResourceAttributesWithSource | PlainMessage<ClusterResourceAttributesWithSource> | undefined): boolean {
    return proto3.util.equals(ClusterResourceAttributesWithSource, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.ExecutionQueueAttributesWithSource
 */
export class ExecutionQueueAttributesWithSource extends Message<ExecutionQueueAttributesWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.admin.ExecutionQueueAttributes value = 2;
   */
  value?: ExecutionQueueAttributes;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<ExecutionQueueAttributesWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ExecutionQueueAttributesWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: ExecutionQueueAttributes },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExecutionQueueAttributesWithSource {
    return new ExecutionQueueAttributesWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExecutionQueueAttributesWithSource {
    return new ExecutionQueueAttributesWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExecutionQueueAttributesWithSource {
    return new ExecutionQueueAttributesWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: ExecutionQueueAttributesWithSource | PlainMessage<ExecutionQueueAttributesWithSource> | undefined, b: ExecutionQueueAttributesWithSource | PlainMessage<ExecutionQueueAttributesWithSource> | undefined): boolean {
    return proto3.util.equals(ExecutionQueueAttributesWithSource, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.ExecutionClusterLabelWithSource
 */
export class ExecutionClusterLabelWithSource extends Message<ExecutionClusterLabelWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.admin.ExecutionClusterLabel value = 2;
   */
  value?: ExecutionClusterLabel;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<ExecutionClusterLabelWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ExecutionClusterLabelWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: ExecutionClusterLabel },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExecutionClusterLabelWithSource {
    return new ExecutionClusterLabelWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExecutionClusterLabelWithSource {
    return new ExecutionClusterLabelWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExecutionClusterLabelWithSource {
    return new ExecutionClusterLabelWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: ExecutionClusterLabelWithSource | PlainMessage<ExecutionClusterLabelWithSource> | undefined, b: ExecutionClusterLabelWithSource | PlainMessage<ExecutionClusterLabelWithSource> | undefined): boolean {
    return proto3.util.equals(ExecutionClusterLabelWithSource, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.QualityOfServiceWithSource
 */
export class QualityOfServiceWithSource extends Message<QualityOfServiceWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.core.QualityOfService value = 2;
   */
  value?: QualityOfService;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<QualityOfServiceWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.QualityOfServiceWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: QualityOfService },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): QualityOfServiceWithSource {
    return new QualityOfServiceWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): QualityOfServiceWithSource {
    return new QualityOfServiceWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): QualityOfServiceWithSource {
    return new QualityOfServiceWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: QualityOfServiceWithSource | PlainMessage<QualityOfServiceWithSource> | undefined, b: QualityOfServiceWithSource | PlainMessage<QualityOfServiceWithSource> | undefined): boolean {
    return proto3.util.equals(QualityOfServiceWithSource, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.PluginOverridesWithSource
 */
export class PluginOverridesWithSource extends Message<PluginOverridesWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.admin.PluginOverrides value = 2;
   */
  value?: PluginOverrides;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<PluginOverridesWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.PluginOverridesWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: PluginOverrides },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PluginOverridesWithSource {
    return new PluginOverridesWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PluginOverridesWithSource {
    return new PluginOverridesWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PluginOverridesWithSource {
    return new PluginOverridesWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: PluginOverridesWithSource | PlainMessage<PluginOverridesWithSource> | undefined, b: PluginOverridesWithSource | PlainMessage<PluginOverridesWithSource> | undefined): boolean {
    return proto3.util.equals(PluginOverridesWithSource, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.WorkflowExecutionConfigWithSource
 */
export class WorkflowExecutionConfigWithSource extends Message<WorkflowExecutionConfigWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.admin.WorkflowExecutionConfig value = 2;
   */
  value?: WorkflowExecutionConfig;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<WorkflowExecutionConfigWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.WorkflowExecutionConfigWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: WorkflowExecutionConfig },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): WorkflowExecutionConfigWithSource {
    return new WorkflowExecutionConfigWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): WorkflowExecutionConfigWithSource {
    return new WorkflowExecutionConfigWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): WorkflowExecutionConfigWithSource {
    return new WorkflowExecutionConfigWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: WorkflowExecutionConfigWithSource | PlainMessage<WorkflowExecutionConfigWithSource> | undefined, b: WorkflowExecutionConfigWithSource | PlainMessage<WorkflowExecutionConfigWithSource> | undefined): boolean {
    return proto3.util.equals(WorkflowExecutionConfigWithSource, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.ClusterAssignmentWithSource
 */
export class ClusterAssignmentWithSource extends Message<ClusterAssignmentWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.admin.ClusterAssignment value = 2;
   */
  value?: ClusterAssignment;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<ClusterAssignmentWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ClusterAssignmentWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: ClusterAssignment },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ClusterAssignmentWithSource {
    return new ClusterAssignmentWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ClusterAssignmentWithSource {
    return new ClusterAssignmentWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ClusterAssignmentWithSource {
    return new ClusterAssignmentWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: ClusterAssignmentWithSource | PlainMessage<ClusterAssignmentWithSource> | undefined, b: ClusterAssignmentWithSource | PlainMessage<ClusterAssignmentWithSource> | undefined): boolean {
    return proto3.util.equals(ClusterAssignmentWithSource, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.ExternalResourceAttributesWithSource
 */
export class ExternalResourceAttributesWithSource extends Message<ExternalResourceAttributesWithSource> {
  /**
   * @generated from field: flyteidl.admin.AttributesSource source = 1;
   */
  source = AttributesSource.SOURCE_UNSPECIFIED;

  /**
   * @generated from field: flyteidl.admin.ExternalResourceAttributes value = 2;
   */
  value?: ExternalResourceAttributes;

  /**
   * @generated from field: bool is_mutable = 3;
   */
  isMutable = false;

  constructor(data?: PartialMessage<ExternalResourceAttributesWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ExternalResourceAttributesWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "source", kind: "enum", T: proto3.getEnumType(AttributesSource) },
    { no: 2, name: "value", kind: "message", T: ExternalResourceAttributes },
    { no: 3, name: "is_mutable", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExternalResourceAttributesWithSource {
    return new ExternalResourceAttributesWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExternalResourceAttributesWithSource {
    return new ExternalResourceAttributesWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExternalResourceAttributesWithSource {
    return new ExternalResourceAttributesWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: ExternalResourceAttributesWithSource | PlainMessage<ExternalResourceAttributesWithSource> | undefined, b: ExternalResourceAttributesWithSource | PlainMessage<ExternalResourceAttributesWithSource> | undefined): boolean {
    return proto3.util.equals(ExternalResourceAttributesWithSource, a, b);
  }
}

/**
 * Configuration with source information.
 *
 * @generated from message flyteidl.admin.ConfigurationWithSource
 */
export class ConfigurationWithSource extends Message<ConfigurationWithSource> {
  /**
   * @generated from field: flyteidl.admin.TaskResourceAttributesWithSource task_resource_attributes = 1;
   */
  taskResourceAttributes?: TaskResourceAttributesWithSource;

  /**
   * @generated from field: flyteidl.admin.ClusterResourceAttributesWithSource cluster_resource_attributes = 2;
   */
  clusterResourceAttributes?: ClusterResourceAttributesWithSource;

  /**
   * @generated from field: flyteidl.admin.ExecutionQueueAttributesWithSource execution_queue_attributes = 3;
   */
  executionQueueAttributes?: ExecutionQueueAttributesWithSource;

  /**
   * @generated from field: flyteidl.admin.ExecutionClusterLabelWithSource execution_cluster_label = 4;
   */
  executionClusterLabel?: ExecutionClusterLabelWithSource;

  /**
   * @generated from field: flyteidl.admin.QualityOfServiceWithSource quality_of_service = 5;
   */
  qualityOfService?: QualityOfServiceWithSource;

  /**
   * @generated from field: flyteidl.admin.PluginOverridesWithSource plugin_overrides = 6;
   */
  pluginOverrides?: PluginOverridesWithSource;

  /**
   * @generated from field: flyteidl.admin.WorkflowExecutionConfigWithSource workflow_execution_config = 7;
   */
  workflowExecutionConfig?: WorkflowExecutionConfigWithSource;

  /**
   * @generated from field: flyteidl.admin.ClusterAssignmentWithSource cluster_assignment = 8;
   */
  clusterAssignment?: ClusterAssignmentWithSource;

  /**
   * @generated from field: flyteidl.admin.ExternalResourceAttributesWithSource external_resource_attributes = 9;
   */
  externalResourceAttributes?: ExternalResourceAttributesWithSource;

  constructor(data?: PartialMessage<ConfigurationWithSource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ConfigurationWithSource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "task_resource_attributes", kind: "message", T: TaskResourceAttributesWithSource },
    { no: 2, name: "cluster_resource_attributes", kind: "message", T: ClusterResourceAttributesWithSource },
    { no: 3, name: "execution_queue_attributes", kind: "message", T: ExecutionQueueAttributesWithSource },
    { no: 4, name: "execution_cluster_label", kind: "message", T: ExecutionClusterLabelWithSource },
    { no: 5, name: "quality_of_service", kind: "message", T: QualityOfServiceWithSource },
    { no: 6, name: "plugin_overrides", kind: "message", T: PluginOverridesWithSource },
    { no: 7, name: "workflow_execution_config", kind: "message", T: WorkflowExecutionConfigWithSource },
    { no: 8, name: "cluster_assignment", kind: "message", T: ClusterAssignmentWithSource },
    { no: 9, name: "external_resource_attributes", kind: "message", T: ExternalResourceAttributesWithSource },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigurationWithSource {
    return new ConfigurationWithSource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigurationWithSource {
    return new ConfigurationWithSource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigurationWithSource {
    return new ConfigurationWithSource().fromJsonString(jsonString, options);
  }

  static equals(a: ConfigurationWithSource | PlainMessage<ConfigurationWithSource> | undefined, b: ConfigurationWithSource | PlainMessage<ConfigurationWithSource> | undefined): boolean {
    return proto3.util.equals(ConfigurationWithSource, a, b);
  }
}

/**
 * Configuration is a collection of attributes that can be applied to a project or a workflow.
 *
 * @generated from message flyteidl.admin.Configuration
 */
export class Configuration extends Message<Configuration> {
  /**
   * @generated from field: flyteidl.admin.TaskResourceAttributes task_resource_attributes = 1;
   */
  taskResourceAttributes?: TaskResourceAttributes;

  /**
   * @generated from field: flyteidl.admin.ClusterResourceAttributes cluster_resource_attributes = 2;
   */
  clusterResourceAttributes?: ClusterResourceAttributes;

  /**
   * @generated from field: flyteidl.admin.ExecutionQueueAttributes execution_queue_attributes = 3;
   */
  executionQueueAttributes?: ExecutionQueueAttributes;

  /**
   * @generated from field: flyteidl.admin.ExecutionClusterLabel execution_cluster_label = 4;
   */
  executionClusterLabel?: ExecutionClusterLabel;

  /**
   * @generated from field: flyteidl.core.QualityOfService quality_of_service = 5;
   */
  qualityOfService?: QualityOfService;

  /**
   * @generated from field: flyteidl.admin.PluginOverrides plugin_overrides = 6;
   */
  pluginOverrides?: PluginOverrides;

  /**
   * @generated from field: flyteidl.admin.WorkflowExecutionConfig workflow_execution_config = 7;
   */
  workflowExecutionConfig?: WorkflowExecutionConfig;

  /**
   * @generated from field: flyteidl.admin.ClusterAssignment cluster_assignment = 8;
   */
  clusterAssignment?: ClusterAssignment;

  /**
   * @generated from field: flyteidl.admin.ExternalResourceAttributes external_resource_attributes = 9;
   */
  externalResourceAttributes?: ExternalResourceAttributes;

  constructor(data?: PartialMessage<Configuration>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.Configuration";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "task_resource_attributes", kind: "message", T: TaskResourceAttributes },
    { no: 2, name: "cluster_resource_attributes", kind: "message", T: ClusterResourceAttributes },
    { no: 3, name: "execution_queue_attributes", kind: "message", T: ExecutionQueueAttributes },
    { no: 4, name: "execution_cluster_label", kind: "message", T: ExecutionClusterLabel },
    { no: 5, name: "quality_of_service", kind: "message", T: QualityOfService },
    { no: 6, name: "plugin_overrides", kind: "message", T: PluginOverrides },
    { no: 7, name: "workflow_execution_config", kind: "message", T: WorkflowExecutionConfig },
    { no: 8, name: "cluster_assignment", kind: "message", T: ClusterAssignment },
    { no: 9, name: "external_resource_attributes", kind: "message", T: ExternalResourceAttributes },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Configuration {
    return new Configuration().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Configuration {
    return new Configuration().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Configuration {
    return new Configuration().fromJsonString(jsonString, options);
  }

  static equals(a: Configuration | PlainMessage<Configuration> | undefined, b: Configuration | PlainMessage<Configuration> | undefined): boolean {
    return proto3.util.equals(Configuration, a, b);
  }
}

/**
 * Request to get a configuration.
 *
 * @generated from message flyteidl.admin.ConfigurationGetRequest
 */
export class ConfigurationGetRequest extends Message<ConfigurationGetRequest> {
  /**
   * Identifier of the configuration to get.
   * +required
   *
   * @generated from field: flyteidl.admin.ConfigurationID id = 1;
   */
  id?: ConfigurationID;

  constructor(data?: PartialMessage<ConfigurationGetRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ConfigurationGetRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: ConfigurationID },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigurationGetRequest {
    return new ConfigurationGetRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigurationGetRequest {
    return new ConfigurationGetRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigurationGetRequest {
    return new ConfigurationGetRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ConfigurationGetRequest | PlainMessage<ConfigurationGetRequest> | undefined, b: ConfigurationGetRequest | PlainMessage<ConfigurationGetRequest> | undefined): boolean {
    return proto3.util.equals(ConfigurationGetRequest, a, b);
  }
}

/**
 * Response of a configuration get request.
 *
 * @generated from message flyteidl.admin.ConfigurationGetResponse
 */
export class ConfigurationGetResponse extends Message<ConfigurationGetResponse> {
  /**
   * Identifier of the configuration.
   *
   * @generated from field: flyteidl.admin.ConfigurationID id = 1;
   */
  id?: ConfigurationID;

  /**
   * Version of the configuration.
   *
   * @generated from field: string version = 2;
   */
  version = "";

  /**
   * Configuration with source information.
   *
   * @generated from field: flyteidl.admin.ConfigurationWithSource configuration = 3;
   */
  configuration?: ConfigurationWithSource;

  constructor(data?: PartialMessage<ConfigurationGetResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ConfigurationGetResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: ConfigurationID },
    { no: 2, name: "version", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "configuration", kind: "message", T: ConfigurationWithSource },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigurationGetResponse {
    return new ConfigurationGetResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigurationGetResponse {
    return new ConfigurationGetResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigurationGetResponse {
    return new ConfigurationGetResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ConfigurationGetResponse | PlainMessage<ConfigurationGetResponse> | undefined, b: ConfigurationGetResponse | PlainMessage<ConfigurationGetResponse> | undefined): boolean {
    return proto3.util.equals(ConfigurationGetResponse, a, b);
  }
}

/**
 * Request to update a configuration.
 *
 * @generated from message flyteidl.admin.ConfigurationUpdateRequest
 */
export class ConfigurationUpdateRequest extends Message<ConfigurationUpdateRequest> {
  /**
   * Identifier of the configuration to update.
   * +required
   *
   * @generated from field: flyteidl.admin.ConfigurationID id = 1;
   */
  id?: ConfigurationID;

  /**
   * Version of the configuration to update, which should to be the version of the currently active document.
   * +required
   *
   * @generated from field: string version_to_update = 2;
   */
  versionToUpdate = "";

  /**
   * Configuration to update.
   * +required
   *
   * @generated from field: flyteidl.admin.Configuration configuration = 3;
   */
  configuration?: Configuration;

  constructor(data?: PartialMessage<ConfigurationUpdateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ConfigurationUpdateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: ConfigurationID },
    { no: 2, name: "version_to_update", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "configuration", kind: "message", T: Configuration },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigurationUpdateRequest {
    return new ConfigurationUpdateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigurationUpdateRequest {
    return new ConfigurationUpdateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigurationUpdateRequest {
    return new ConfigurationUpdateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ConfigurationUpdateRequest | PlainMessage<ConfigurationUpdateRequest> | undefined, b: ConfigurationUpdateRequest | PlainMessage<ConfigurationUpdateRequest> | undefined): boolean {
    return proto3.util.equals(ConfigurationUpdateRequest, a, b);
  }
}

/**
 * Response of a configuration update request.
 *
 * @generated from message flyteidl.admin.ConfigurationUpdateResponse
 */
export class ConfigurationUpdateResponse extends Message<ConfigurationUpdateResponse> {
  /**
   * Identifier of the updated configuration.
   *
   * @generated from field: flyteidl.admin.ConfigurationID id = 1;
   */
  id?: ConfigurationID;

  /**
   * Version of the updated configuration.
   *
   * @generated from field: string version = 2;
   */
  version = "";

  /**
   * Updated configuration with source information.
   *
   * @generated from field: flyteidl.admin.ConfigurationWithSource configuration = 3;
   */
  configuration?: ConfigurationWithSource;

  constructor(data?: PartialMessage<ConfigurationUpdateResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ConfigurationUpdateResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: ConfigurationID },
    { no: 2, name: "version", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "configuration", kind: "message", T: ConfigurationWithSource },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigurationUpdateResponse {
    return new ConfigurationUpdateResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigurationUpdateResponse {
    return new ConfigurationUpdateResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigurationUpdateResponse {
    return new ConfigurationUpdateResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ConfigurationUpdateResponse | PlainMessage<ConfigurationUpdateResponse> | undefined, b: ConfigurationUpdateResponse | PlainMessage<ConfigurationUpdateResponse> | undefined): boolean {
    return proto3.util.equals(ConfigurationUpdateResponse, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.ConfigurationDocument
 */
export class ConfigurationDocument extends Message<ConfigurationDocument> {
  /**
   * Version of the configuration document.
   *
   * @generated from field: string version = 1;
   */
  version = "";

  /**
   * All configurations in the document.
   * The key is the string serialized ConfigurationID with (https://pkg.go.dev/github.com/golang/protobuf/jsonpb#Marshaler.Marshal).
   *
   * @generated from field: map<string, flyteidl.admin.Configuration> configurations = 3;
   */
  configurations: { [key: string]: Configuration } = {};

  constructor(data?: PartialMessage<ConfigurationDocument>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ConfigurationDocument";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "version", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "configurations", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: Configuration} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigurationDocument {
    return new ConfigurationDocument().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigurationDocument {
    return new ConfigurationDocument().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigurationDocument {
    return new ConfigurationDocument().fromJsonString(jsonString, options);
  }

  static equals(a: ConfigurationDocument | PlainMessage<ConfigurationDocument> | undefined, b: ConfigurationDocument | PlainMessage<ConfigurationDocument> | undefined): boolean {
    return proto3.util.equals(ConfigurationDocument, a, b);
  }
}

