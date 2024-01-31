// @generated by protoc-gen-es v1.7.1 with parameter "target=ts"
// @generated from file flyteidl/admin/cluster_assignment.proto (package flyteidl.admin, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * Encapsulates specifications for routing an execution onto a specific cluster.
 *
 * @generated from message flyteidl.admin.ClusterAssignment
 */
export class ClusterAssignment extends Message<ClusterAssignment> {
  /**
   * @generated from field: string cluster_pool_name = 3;
   */
  clusterPoolName = "";

  constructor(data?: PartialMessage<ClusterAssignment>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ClusterAssignment";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 3, name: "cluster_pool_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ClusterAssignment {
    return new ClusterAssignment().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ClusterAssignment {
    return new ClusterAssignment().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ClusterAssignment {
    return new ClusterAssignment().fromJsonString(jsonString, options);
  }

  static equals(a: ClusterAssignment | PlainMessage<ClusterAssignment> | undefined, b: ClusterAssignment | PlainMessage<ClusterAssignment> | undefined): boolean {
    return proto3.util.equals(ClusterAssignment, a, b);
  }
}

