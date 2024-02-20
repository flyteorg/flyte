// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/admin/signal.proto (package flyteidl.admin, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { SignalIdentifier, WorkflowExecutionIdentifier } from "../core/identifier_pb.js";
import { LiteralType } from "../core/types_pb.js";
import { Sort } from "./common_pb.js";
import { Literal } from "../core/literals_pb.js";

/**
 * SignalGetOrCreateRequest represents a request structure to retrieve or create a signal.
 * See :ref:`ref_flyteidl.admin.Signal` for more details
 *
 * @generated from message flyteidl.admin.SignalGetOrCreateRequest
 */
export class SignalGetOrCreateRequest extends Message<SignalGetOrCreateRequest> {
  /**
   * A unique identifier for the requested signal.
   *
   * @generated from field: flyteidl.core.SignalIdentifier id = 1;
   */
  id?: SignalIdentifier;

  /**
   * A type denoting the required value type for this signal.
   *
   * @generated from field: flyteidl.core.LiteralType type = 2;
   */
  type?: LiteralType;

  constructor(data?: PartialMessage<SignalGetOrCreateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SignalGetOrCreateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: SignalIdentifier },
    { no: 2, name: "type", kind: "message", T: LiteralType },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SignalGetOrCreateRequest {
    return new SignalGetOrCreateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SignalGetOrCreateRequest {
    return new SignalGetOrCreateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SignalGetOrCreateRequest {
    return new SignalGetOrCreateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: SignalGetOrCreateRequest | PlainMessage<SignalGetOrCreateRequest> | undefined, b: SignalGetOrCreateRequest | PlainMessage<SignalGetOrCreateRequest> | undefined): boolean {
    return proto3.util.equals(SignalGetOrCreateRequest, a, b);
  }
}

/**
 * SignalListRequest represents a request structure to retrieve a collection of signals.
 * See :ref:`ref_flyteidl.admin.Signal` for more details
 *
 * @generated from message flyteidl.admin.SignalListRequest
 */
export class SignalListRequest extends Message<SignalListRequest> {
  /**
   * Indicates the workflow execution to filter by.
   * +required
   *
   * @generated from field: flyteidl.core.WorkflowExecutionIdentifier workflow_execution_id = 1;
   */
  workflowExecutionId?: WorkflowExecutionIdentifier;

  /**
   * Indicates the number of resources to be returned.
   * +required
   *
   * @generated from field: uint32 limit = 2;
   */
  limit = 0;

  /**
   * In the case of multiple pages of results, the, server-provided token can be used to fetch the next page
   * in a query.
   * +optional
   *
   * @generated from field: string token = 3;
   */
  token = "";

  /**
   * Indicates a list of filters passed as string.
   * +optional
   *
   * @generated from field: string filters = 4;
   */
  filters = "";

  /**
   * Sort ordering.
   * +optional
   *
   * @generated from field: flyteidl.admin.Sort sort_by = 5;
   */
  sortBy?: Sort;

  constructor(data?: PartialMessage<SignalListRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SignalListRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "workflow_execution_id", kind: "message", T: WorkflowExecutionIdentifier },
    { no: 2, name: "limit", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 3, name: "token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "filters", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "sort_by", kind: "message", T: Sort },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SignalListRequest {
    return new SignalListRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SignalListRequest {
    return new SignalListRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SignalListRequest {
    return new SignalListRequest().fromJsonString(jsonString, options);
  }

  static equals(a: SignalListRequest | PlainMessage<SignalListRequest> | undefined, b: SignalListRequest | PlainMessage<SignalListRequest> | undefined): boolean {
    return proto3.util.equals(SignalListRequest, a, b);
  }
}

/**
 * SignalList represents collection of signals along with the token of the last result.
 * See :ref:`ref_flyteidl.admin.Signal` for more details
 *
 * @generated from message flyteidl.admin.SignalList
 */
export class SignalList extends Message<SignalList> {
  /**
   * A list of signals matching the input filters.
   *
   * @generated from field: repeated flyteidl.admin.Signal signals = 1;
   */
  signals: Signal[] = [];

  /**
   * In the case of multiple pages of results, the server-provided token can be used to fetch the next page
   * in a query. If there are no more results, this value will be empty.
   *
   * @generated from field: string token = 2;
   */
  token = "";

  constructor(data?: PartialMessage<SignalList>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SignalList";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "signals", kind: "message", T: Signal, repeated: true },
    { no: 2, name: "token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SignalList {
    return new SignalList().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SignalList {
    return new SignalList().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SignalList {
    return new SignalList().fromJsonString(jsonString, options);
  }

  static equals(a: SignalList | PlainMessage<SignalList> | undefined, b: SignalList | PlainMessage<SignalList> | undefined): boolean {
    return proto3.util.equals(SignalList, a, b);
  }
}

/**
 * SignalSetRequest represents a request structure to set the value on a signal. Setting a signal
 * effetively satisfies the signal condition within a Flyte workflow.
 * See :ref:`ref_flyteidl.admin.Signal` for more details
 *
 * @generated from message flyteidl.admin.SignalSetRequest
 */
export class SignalSetRequest extends Message<SignalSetRequest> {
  /**
   * A unique identifier for the requested signal.
   *
   * @generated from field: flyteidl.core.SignalIdentifier id = 1;
   */
  id?: SignalIdentifier;

  /**
   * The value of this signal, must match the defining signal type.
   *
   * @generated from field: flyteidl.core.Literal value = 2;
   */
  value?: Literal;

  constructor(data?: PartialMessage<SignalSetRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SignalSetRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: SignalIdentifier },
    { no: 2, name: "value", kind: "message", T: Literal },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SignalSetRequest {
    return new SignalSetRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SignalSetRequest {
    return new SignalSetRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SignalSetRequest {
    return new SignalSetRequest().fromJsonString(jsonString, options);
  }

  static equals(a: SignalSetRequest | PlainMessage<SignalSetRequest> | undefined, b: SignalSetRequest | PlainMessage<SignalSetRequest> | undefined): boolean {
    return proto3.util.equals(SignalSetRequest, a, b);
  }
}

/**
 * SignalSetResponse represents a response structure if signal setting succeeds.
 *
 * Purposefully empty, may be populated in the future.
 *
 * @generated from message flyteidl.admin.SignalSetResponse
 */
export class SignalSetResponse extends Message<SignalSetResponse> {
  constructor(data?: PartialMessage<SignalSetResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SignalSetResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SignalSetResponse {
    return new SignalSetResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SignalSetResponse {
    return new SignalSetResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SignalSetResponse {
    return new SignalSetResponse().fromJsonString(jsonString, options);
  }

  static equals(a: SignalSetResponse | PlainMessage<SignalSetResponse> | undefined, b: SignalSetResponse | PlainMessage<SignalSetResponse> | undefined): boolean {
    return proto3.util.equals(SignalSetResponse, a, b);
  }
}

/**
 * Signal encapsulates a unique identifier, associated metadata, and a value for a single Flyte
 * signal. Signals may exist either without a set value (representing a signal request) or with a
 * populated value (indicating the signal has been given).
 *
 * @generated from message flyteidl.admin.Signal
 */
export class Signal extends Message<Signal> {
  /**
   * A unique identifier for the requested signal.
   *
   * @generated from field: flyteidl.core.SignalIdentifier id = 1;
   */
  id?: SignalIdentifier;

  /**
   * A type denoting the required value type for this signal.
   *
   * @generated from field: flyteidl.core.LiteralType type = 2;
   */
  type?: LiteralType;

  /**
   * The value of the signal. This is only available if the signal has been "set" and must match
   * the defined the type.
   *
   * @generated from field: flyteidl.core.Literal value = 3;
   */
  value?: Literal;

  constructor(data?: PartialMessage<Signal>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.Signal";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: SignalIdentifier },
    { no: 2, name: "type", kind: "message", T: LiteralType },
    { no: 3, name: "value", kind: "message", T: Literal },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Signal {
    return new Signal().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Signal {
    return new Signal().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Signal {
    return new Signal().fromJsonString(jsonString, options);
  }

  static equals(a: Signal | PlainMessage<Signal> | undefined, b: Signal | PlainMessage<Signal> | undefined): boolean {
    return proto3.util.equals(Signal, a, b);
  }
}

