// @generated by protoc-gen-es v1.7.1 with parameter "target=ts"
// @generated from file flyteidl/plugins/kubeflow/mpi.proto (package flyteidl.plugins.kubeflow, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { RestartPolicy, RunPolicy } from "./common_pb.js";
import { Resources } from "../../core/tasks_pb.js";

/**
 * Proto for plugin that enables distributed training using https://github.com/kubeflow/mpi-operator
 *
 * @generated from message flyteidl.plugins.kubeflow.DistributedMPITrainingTask
 */
export class DistributedMPITrainingTask extends Message<DistributedMPITrainingTask> {
  /**
   * Worker replicas spec
   *
   * @generated from field: flyteidl.plugins.kubeflow.DistributedMPITrainingReplicaSpec worker_replicas = 1;
   */
  workerReplicas?: DistributedMPITrainingReplicaSpec;

  /**
   * Master replicas spec
   *
   * @generated from field: flyteidl.plugins.kubeflow.DistributedMPITrainingReplicaSpec launcher_replicas = 2;
   */
  launcherReplicas?: DistributedMPITrainingReplicaSpec;

  /**
   * RunPolicy encapsulates various runtime policies of the distributed training
   * job, for example how to clean up resources and how long the job can stay
   * active.
   *
   * @generated from field: flyteidl.plugins.kubeflow.RunPolicy run_policy = 3;
   */
  runPolicy?: RunPolicy;

  /**
   * Number of slots per worker
   *
   * @generated from field: int32 slots = 4;
   */
  slots = 0;

  constructor(data?: PartialMessage<DistributedMPITrainingTask>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.kubeflow.DistributedMPITrainingTask";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "worker_replicas", kind: "message", T: DistributedMPITrainingReplicaSpec },
    { no: 2, name: "launcher_replicas", kind: "message", T: DistributedMPITrainingReplicaSpec },
    { no: 3, name: "run_policy", kind: "message", T: RunPolicy },
    { no: 4, name: "slots", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DistributedMPITrainingTask {
    return new DistributedMPITrainingTask().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DistributedMPITrainingTask {
    return new DistributedMPITrainingTask().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DistributedMPITrainingTask {
    return new DistributedMPITrainingTask().fromJsonString(jsonString, options);
  }

  static equals(a: DistributedMPITrainingTask | PlainMessage<DistributedMPITrainingTask> | undefined, b: DistributedMPITrainingTask | PlainMessage<DistributedMPITrainingTask> | undefined): boolean {
    return proto3.util.equals(DistributedMPITrainingTask, a, b);
  }
}

/**
 * Replica specification for distributed MPI training
 *
 * @generated from message flyteidl.plugins.kubeflow.DistributedMPITrainingReplicaSpec
 */
export class DistributedMPITrainingReplicaSpec extends Message<DistributedMPITrainingReplicaSpec> {
  /**
   * Number of replicas
   *
   * @generated from field: int32 replicas = 1;
   */
  replicas = 0;

  /**
   * Image used for the replica group
   *
   * @generated from field: string image = 2;
   */
  image = "";

  /**
   * Resources required for the replica group
   *
   * @generated from field: flyteidl.core.Resources resources = 3;
   */
  resources?: Resources;

  /**
   * Restart policy determines whether pods will be restarted when they exit
   *
   * @generated from field: flyteidl.plugins.kubeflow.RestartPolicy restart_policy = 4;
   */
  restartPolicy = RestartPolicy.NEVER;

  /**
   * MPI sometimes requires different command set for different replica groups
   *
   * @generated from field: repeated string command = 5;
   */
  command: string[] = [];

  constructor(data?: PartialMessage<DistributedMPITrainingReplicaSpec>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.kubeflow.DistributedMPITrainingReplicaSpec";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "replicas", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 2, name: "image", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "resources", kind: "message", T: Resources },
    { no: 4, name: "restart_policy", kind: "enum", T: proto3.getEnumType(RestartPolicy) },
    { no: 5, name: "command", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DistributedMPITrainingReplicaSpec {
    return new DistributedMPITrainingReplicaSpec().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DistributedMPITrainingReplicaSpec {
    return new DistributedMPITrainingReplicaSpec().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DistributedMPITrainingReplicaSpec {
    return new DistributedMPITrainingReplicaSpec().fromJsonString(jsonString, options);
  }

  static equals(a: DistributedMPITrainingReplicaSpec | PlainMessage<DistributedMPITrainingReplicaSpec> | undefined, b: DistributedMPITrainingReplicaSpec | PlainMessage<DistributedMPITrainingReplicaSpec> | undefined): boolean {
    return proto3.util.equals(DistributedMPITrainingReplicaSpec, a, b);
  }
}

