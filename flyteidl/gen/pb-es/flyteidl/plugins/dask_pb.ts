// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/plugins/dask.proto (package flyteidl.plugins, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Resources } from "../core/tasks_pb.js";

/**
 * Custom Proto for Dask Plugin.
 *
 * @generated from message flyteidl.plugins.DaskJob
 */
export class DaskJob extends Message<DaskJob> {
  /**
   * Spec for the scheduler pod.
   *
   * @generated from field: flyteidl.plugins.DaskScheduler scheduler = 1;
   */
  scheduler?: DaskScheduler;

  /**
   * Spec of the default worker group.
   *
   * @generated from field: flyteidl.plugins.DaskWorkerGroup workers = 2;
   */
  workers?: DaskWorkerGroup;

  constructor(data?: PartialMessage<DaskJob>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.DaskJob";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "scheduler", kind: "message", T: DaskScheduler },
    { no: 2, name: "workers", kind: "message", T: DaskWorkerGroup },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DaskJob {
    return new DaskJob().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DaskJob {
    return new DaskJob().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DaskJob {
    return new DaskJob().fromJsonString(jsonString, options);
  }

  static equals(a: DaskJob | PlainMessage<DaskJob> | undefined, b: DaskJob | PlainMessage<DaskJob> | undefined): boolean {
    return proto3.util.equals(DaskJob, a, b);
  }
}

/**
 * Specification for the scheduler pod.
 *
 * @generated from message flyteidl.plugins.DaskScheduler
 */
export class DaskScheduler extends Message<DaskScheduler> {
  /**
   * Optional image to use. If unset, will use the default image.
   *
   * @generated from field: string image = 1;
   */
  image = "";

  /**
   * Resources assigned to the scheduler pod.
   *
   * @generated from field: flyteidl.core.Resources resources = 2;
   */
  resources?: Resources;

  constructor(data?: PartialMessage<DaskScheduler>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.DaskScheduler";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "image", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "resources", kind: "message", T: Resources },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DaskScheduler {
    return new DaskScheduler().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DaskScheduler {
    return new DaskScheduler().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DaskScheduler {
    return new DaskScheduler().fromJsonString(jsonString, options);
  }

  static equals(a: DaskScheduler | PlainMessage<DaskScheduler> | undefined, b: DaskScheduler | PlainMessage<DaskScheduler> | undefined): boolean {
    return proto3.util.equals(DaskScheduler, a, b);
  }
}

/**
 * @generated from message flyteidl.plugins.DaskWorkerGroup
 */
export class DaskWorkerGroup extends Message<DaskWorkerGroup> {
  /**
   * Number of workers in the group.
   *
   * @generated from field: uint32 number_of_workers = 1;
   */
  numberOfWorkers = 0;

  /**
   * Optional image to use for the pods of the worker group. If unset, will use the default image.
   *
   * @generated from field: string image = 2;
   */
  image = "";

  /**
   * Resources assigned to the all pods of the worker group.
   * As per https://kubernetes.dask.org/en/latest/kubecluster.html?highlight=limit#best-practices 
   * it is advised to only set limits. If requests are not explicitly set, the plugin will make
   * sure to set requests==limits.
   * The plugin sets ` --memory-limit` as well as `--nthreads` for the workers according to the limit.
   *
   * @generated from field: flyteidl.core.Resources resources = 3;
   */
  resources?: Resources;

  constructor(data?: PartialMessage<DaskWorkerGroup>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.DaskWorkerGroup";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "number_of_workers", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 2, name: "image", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "resources", kind: "message", T: Resources },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DaskWorkerGroup {
    return new DaskWorkerGroup().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DaskWorkerGroup {
    return new DaskWorkerGroup().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DaskWorkerGroup {
    return new DaskWorkerGroup().fromJsonString(jsonString, options);
  }

  static equals(a: DaskWorkerGroup | PlainMessage<DaskWorkerGroup> | undefined, b: DaskWorkerGroup | PlainMessage<DaskWorkerGroup> | undefined): boolean {
    return proto3.util.equals(DaskWorkerGroup, a, b);
  }
}

