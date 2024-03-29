// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/core/metrics.proto (package flyteidl.core, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Struct, Timestamp } from "@bufbuild/protobuf";
import { NodeExecutionIdentifier, TaskExecutionIdentifier, WorkflowExecutionIdentifier } from "./identifier_pb.js";

/**
 * Span represents a duration trace of Flyte execution. The id field denotes a Flyte execution entity or an operation
 * which uniquely identifies the Span. The spans attribute allows this Span to be further broken down into more
 * precise definitions.
 *
 * @generated from message flyteidl.core.Span
 */
export class Span extends Message<Span> {
  /**
   * start_time defines the instance this span began.
   *
   * @generated from field: google.protobuf.Timestamp start_time = 1;
   */
  startTime?: Timestamp;

  /**
   * end_time defines the instance this span completed.
   *
   * @generated from field: google.protobuf.Timestamp end_time = 2;
   */
  endTime?: Timestamp;

  /**
   * @generated from oneof flyteidl.core.Span.id
   */
  id: {
    /**
     * workflow_id is the id of the workflow execution this Span represents.
     *
     * @generated from field: flyteidl.core.WorkflowExecutionIdentifier workflow_id = 3;
     */
    value: WorkflowExecutionIdentifier;
    case: "workflowId";
  } | {
    /**
     * node_id is the id of the node execution this Span represents.
     *
     * @generated from field: flyteidl.core.NodeExecutionIdentifier node_id = 4;
     */
    value: NodeExecutionIdentifier;
    case: "nodeId";
  } | {
    /**
     * task_id is the id of the task execution this Span represents.
     *
     * @generated from field: flyteidl.core.TaskExecutionIdentifier task_id = 5;
     */
    value: TaskExecutionIdentifier;
    case: "taskId";
  } | {
    /**
     * operation_id is the id of a unique operation that this Span represents.
     *
     * @generated from field: string operation_id = 6;
     */
    value: string;
    case: "operationId";
  } | { case: undefined; value?: undefined } = { case: undefined };

  /**
   * spans defines a collection of Spans that breakdown this execution.
   *
   * @generated from field: repeated flyteidl.core.Span spans = 7;
   */
  spans: Span[] = [];

  constructor(data?: PartialMessage<Span>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.core.Span";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "start_time", kind: "message", T: Timestamp },
    { no: 2, name: "end_time", kind: "message", T: Timestamp },
    { no: 3, name: "workflow_id", kind: "message", T: WorkflowExecutionIdentifier, oneof: "id" },
    { no: 4, name: "node_id", kind: "message", T: NodeExecutionIdentifier, oneof: "id" },
    { no: 5, name: "task_id", kind: "message", T: TaskExecutionIdentifier, oneof: "id" },
    { no: 6, name: "operation_id", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "id" },
    { no: 7, name: "spans", kind: "message", T: Span, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Span {
    return new Span().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Span {
    return new Span().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Span {
    return new Span().fromJsonString(jsonString, options);
  }

  static equals(a: Span | PlainMessage<Span> | undefined, b: Span | PlainMessage<Span> | undefined): boolean {
    return proto3.util.equals(Span, a, b);
  }
}

/**
 * ExecutionMetrics is a collection of metrics that are collected during the execution of a Flyte task.
 *
 * @generated from message flyteidl.core.ExecutionMetricResult
 */
export class ExecutionMetricResult extends Message<ExecutionMetricResult> {
  /**
   * The metric this data represents. e.g. EXECUTION_METRIC_USED_CPU_AVG or EXECUTION_METRIC_USED_MEMORY_BYTES_AVG.
   *
   * @generated from field: string metric = 1;
   */
  metric = "";

  /**
   * The result data in prometheus range query result format
   * https://prometheus.io/docs/prometheus/latest/querying/api/#expression-query-result-formats.
   * This may include multiple time series, differentiated by their metric labels.
   * Start time is greater of (execution attempt start, 48h ago)
   * End time is lesser of (execution attempt end, now)
   *
   * @generated from field: google.protobuf.Struct data = 2;
   */
  data?: Struct;

  constructor(data?: PartialMessage<ExecutionMetricResult>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.core.ExecutionMetricResult";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "metric", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "data", kind: "message", T: Struct },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExecutionMetricResult {
    return new ExecutionMetricResult().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExecutionMetricResult {
    return new ExecutionMetricResult().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExecutionMetricResult {
    return new ExecutionMetricResult().fromJsonString(jsonString, options);
  }

  static equals(a: ExecutionMetricResult | PlainMessage<ExecutionMetricResult> | undefined, b: ExecutionMetricResult | PlainMessage<ExecutionMetricResult> | undefined): boolean {
    return proto3.util.equals(ExecutionMetricResult, a, b);
  }
}

