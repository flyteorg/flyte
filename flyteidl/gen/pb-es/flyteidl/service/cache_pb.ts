// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/service/cache.proto (package flyteidl.service, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { TaskExecutionIdentifier } from "../core/identifier_pb.js";
import { CacheEvictionErrorList } from "../core/errors_pb.js";

/**
 * @generated from message flyteidl.service.EvictTaskExecutionCacheRequest
 */
export class EvictTaskExecutionCacheRequest extends Message<EvictTaskExecutionCacheRequest> {
  /**
   * Identifier of :ref:`ref_flyteidl.admin.TaskExecution` to evict cache for.
   *
   * @generated from field: flyteidl.core.TaskExecutionIdentifier task_execution_id = 1;
   */
  taskExecutionId?: TaskExecutionIdentifier;

  constructor(data?: PartialMessage<EvictTaskExecutionCacheRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.service.EvictTaskExecutionCacheRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "task_execution_id", kind: "message", T: TaskExecutionIdentifier },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EvictTaskExecutionCacheRequest {
    return new EvictTaskExecutionCacheRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EvictTaskExecutionCacheRequest {
    return new EvictTaskExecutionCacheRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EvictTaskExecutionCacheRequest {
    return new EvictTaskExecutionCacheRequest().fromJsonString(jsonString, options);
  }

  static equals(a: EvictTaskExecutionCacheRequest | PlainMessage<EvictTaskExecutionCacheRequest> | undefined, b: EvictTaskExecutionCacheRequest | PlainMessage<EvictTaskExecutionCacheRequest> | undefined): boolean {
    return proto3.util.equals(EvictTaskExecutionCacheRequest, a, b);
  }
}

/**
 * @generated from message flyteidl.service.EvictCacheResponse
 */
export class EvictCacheResponse extends Message<EvictCacheResponse> {
  /**
   * List of errors encountered during cache eviction (if any).
   *
   * @generated from field: flyteidl.core.CacheEvictionErrorList errors = 1;
   */
  errors?: CacheEvictionErrorList;

  constructor(data?: PartialMessage<EvictCacheResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.service.EvictCacheResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "errors", kind: "message", T: CacheEvictionErrorList },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EvictCacheResponse {
    return new EvictCacheResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EvictCacheResponse {
    return new EvictCacheResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EvictCacheResponse {
    return new EvictCacheResponse().fromJsonString(jsonString, options);
  }

  static equals(a: EvictCacheResponse | PlainMessage<EvictCacheResponse> | undefined, b: EvictCacheResponse | PlainMessage<EvictCacheResponse> | undefined): boolean {
    return proto3.util.equals(EvictCacheResponse, a, b);
  }
}

