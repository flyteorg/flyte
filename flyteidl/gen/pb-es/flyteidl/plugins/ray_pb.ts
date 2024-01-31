// @generated by protoc-gen-es v1.7.1 with parameter "target=ts"
// @generated from file flyteidl/plugins/ray.proto (package flyteidl.plugins, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * RayJobSpec defines the desired state of RayJob
 *
 * @generated from message flyteidl.plugins.RayJob
 */
export class RayJob extends Message<RayJob> {
  /**
   * RayClusterSpec is the cluster template to run the job
   *
   * @generated from field: flyteidl.plugins.RayCluster ray_cluster = 1;
   */
  rayCluster?: RayCluster;

  /**
   * runtime_env is base64 encoded.
   * Ray runtime environments: https://docs.ray.io/en/latest/ray-core/handling-dependencies.html#runtime-environments
   *
   * @generated from field: string runtime_env = 2;
   */
  runtimeEnv = "";

  /**
   * shutdown_after_job_finishes specifies whether the RayCluster should be deleted after the RayJob finishes.
   *
   * @generated from field: bool shutdown_after_job_finishes = 3;
   */
  shutdownAfterJobFinishes = false;

  /**
   * ttl_seconds_after_finished specifies the number of seconds after which the RayCluster will be deleted after the RayJob finishes.
   *
   * @generated from field: int32 ttl_seconds_after_finished = 4;
   */
  ttlSecondsAfterFinished = 0;

  constructor(data?: PartialMessage<RayJob>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.RayJob";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ray_cluster", kind: "message", T: RayCluster },
    { no: 2, name: "runtime_env", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "shutdown_after_job_finishes", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 4, name: "ttl_seconds_after_finished", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RayJob {
    return new RayJob().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RayJob {
    return new RayJob().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RayJob {
    return new RayJob().fromJsonString(jsonString, options);
  }

  static equals(a: RayJob | PlainMessage<RayJob> | undefined, b: RayJob | PlainMessage<RayJob> | undefined): boolean {
    return proto3.util.equals(RayJob, a, b);
  }
}

/**
 * Define Ray cluster defines the desired state of RayCluster
 *
 * @generated from message flyteidl.plugins.RayCluster
 */
export class RayCluster extends Message<RayCluster> {
  /**
   * HeadGroupSpecs are the spec for the head pod
   *
   * @generated from field: flyteidl.plugins.HeadGroupSpec head_group_spec = 1;
   */
  headGroupSpec?: HeadGroupSpec;

  /**
   * WorkerGroupSpecs are the specs for the worker pods
   *
   * @generated from field: repeated flyteidl.plugins.WorkerGroupSpec worker_group_spec = 2;
   */
  workerGroupSpec: WorkerGroupSpec[] = [];

  /**
   * Whether to enable autoscaling.
   *
   * @generated from field: bool enable_autoscaling = 3;
   */
  enableAutoscaling = false;

  constructor(data?: PartialMessage<RayCluster>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.RayCluster";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "head_group_spec", kind: "message", T: HeadGroupSpec },
    { no: 2, name: "worker_group_spec", kind: "message", T: WorkerGroupSpec, repeated: true },
    { no: 3, name: "enable_autoscaling", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RayCluster {
    return new RayCluster().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RayCluster {
    return new RayCluster().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RayCluster {
    return new RayCluster().fromJsonString(jsonString, options);
  }

  static equals(a: RayCluster | PlainMessage<RayCluster> | undefined, b: RayCluster | PlainMessage<RayCluster> | undefined): boolean {
    return proto3.util.equals(RayCluster, a, b);
  }
}

/**
 * HeadGroupSpec are the spec for the head pod
 *
 * @generated from message flyteidl.plugins.HeadGroupSpec
 */
export class HeadGroupSpec extends Message<HeadGroupSpec> {
  /**
   * Optional. RayStartParams are the params of the start command: address, object-store-memory.
   * Refer to https://docs.ray.io/en/latest/ray-core/package-ref.html#ray-start
   *
   * @generated from field: map<string, string> ray_start_params = 1;
   */
  rayStartParams: { [key: string]: string } = {};

  constructor(data?: PartialMessage<HeadGroupSpec>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.HeadGroupSpec";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ray_start_params", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): HeadGroupSpec {
    return new HeadGroupSpec().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): HeadGroupSpec {
    return new HeadGroupSpec().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): HeadGroupSpec {
    return new HeadGroupSpec().fromJsonString(jsonString, options);
  }

  static equals(a: HeadGroupSpec | PlainMessage<HeadGroupSpec> | undefined, b: HeadGroupSpec | PlainMessage<HeadGroupSpec> | undefined): boolean {
    return proto3.util.equals(HeadGroupSpec, a, b);
  }
}

/**
 * WorkerGroupSpec are the specs for the worker pods
 *
 * @generated from message flyteidl.plugins.WorkerGroupSpec
 */
export class WorkerGroupSpec extends Message<WorkerGroupSpec> {
  /**
   * Required. RayCluster can have multiple worker groups, and it distinguishes them by name
   *
   * @generated from field: string group_name = 1;
   */
  groupName = "";

  /**
   * Required. Desired replicas of the worker group. Defaults to 1.
   *
   * @generated from field: int32 replicas = 2;
   */
  replicas = 0;

  /**
   * Optional. Min replicas of the worker group. MinReplicas defaults to 1.
   *
   * @generated from field: int32 min_replicas = 3;
   */
  minReplicas = 0;

  /**
   * Optional. Max replicas of the worker group. MaxReplicas defaults to maxInt32
   *
   * @generated from field: int32 max_replicas = 4;
   */
  maxReplicas = 0;

  /**
   * Optional. RayStartParams are the params of the start command: address, object-store-memory.
   * Refer to https://docs.ray.io/en/latest/ray-core/package-ref.html#ray-start
   *
   * @generated from field: map<string, string> ray_start_params = 5;
   */
  rayStartParams: { [key: string]: string } = {};

  constructor(data?: PartialMessage<WorkerGroupSpec>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.WorkerGroupSpec";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "group_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "replicas", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 3, name: "min_replicas", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 4, name: "max_replicas", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 5, name: "ray_start_params", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): WorkerGroupSpec {
    return new WorkerGroupSpec().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): WorkerGroupSpec {
    return new WorkerGroupSpec().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): WorkerGroupSpec {
    return new WorkerGroupSpec().fromJsonString(jsonString, options);
  }

  static equals(a: WorkerGroupSpec | PlainMessage<WorkerGroupSpec> | undefined, b: WorkerGroupSpec | PlainMessage<WorkerGroupSpec> | undefined): boolean {
    return proto3.util.equals(WorkerGroupSpec, a, b);
  }
}

