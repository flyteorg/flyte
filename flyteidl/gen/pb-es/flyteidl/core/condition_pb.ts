// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/core/condition.proto (package flyteidl.core, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Primitive, Scalar } from "./literals_pb.js";

/**
 * Defines a 2-level tree where the root is a comparison operator and Operands are primitives or known variables.
 * Each expression results in a boolean result.
 *
 * @generated from message flyteidl.core.ComparisonExpression
 */
export class ComparisonExpression extends Message<ComparisonExpression> {
  /**
   * @generated from field: flyteidl.core.ComparisonExpression.Operator operator = 1;
   */
  operator = ComparisonExpression_Operator.EQ;

  /**
   * @generated from field: flyteidl.core.Operand left_value = 2;
   */
  leftValue?: Operand;

  /**
   * @generated from field: flyteidl.core.Operand right_value = 3;
   */
  rightValue?: Operand;

  constructor(data?: PartialMessage<ComparisonExpression>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.core.ComparisonExpression";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "operator", kind: "enum", T: proto3.getEnumType(ComparisonExpression_Operator) },
    { no: 2, name: "left_value", kind: "message", T: Operand },
    { no: 3, name: "right_value", kind: "message", T: Operand },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ComparisonExpression {
    return new ComparisonExpression().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ComparisonExpression {
    return new ComparisonExpression().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ComparisonExpression {
    return new ComparisonExpression().fromJsonString(jsonString, options);
  }

  static equals(a: ComparisonExpression | PlainMessage<ComparisonExpression> | undefined, b: ComparisonExpression | PlainMessage<ComparisonExpression> | undefined): boolean {
    return proto3.util.equals(ComparisonExpression, a, b);
  }
}

/**
 * Binary Operator for each expression
 *
 * @generated from enum flyteidl.core.ComparisonExpression.Operator
 */
export enum ComparisonExpression_Operator {
  /**
   * @generated from enum value: EQ = 0;
   */
  EQ = 0,

  /**
   * @generated from enum value: NEQ = 1;
   */
  NEQ = 1,

  /**
   * Greater Than
   *
   * @generated from enum value: GT = 2;
   */
  GT = 2,

  /**
   * @generated from enum value: GTE = 3;
   */
  GTE = 3,

  /**
   * Less Than
   *
   * @generated from enum value: LT = 4;
   */
  LT = 4,

  /**
   * @generated from enum value: LTE = 5;
   */
  LTE = 5,
}
// Retrieve enum metadata with: proto3.getEnumType(ComparisonExpression_Operator)
proto3.util.setEnumType(ComparisonExpression_Operator, "flyteidl.core.ComparisonExpression.Operator", [
  { no: 0, name: "EQ" },
  { no: 1, name: "NEQ" },
  { no: 2, name: "GT" },
  { no: 3, name: "GTE" },
  { no: 4, name: "LT" },
  { no: 5, name: "LTE" },
]);

/**
 * Defines an operand to a comparison expression.
 *
 * @generated from message flyteidl.core.Operand
 */
export class Operand extends Message<Operand> {
  /**
   * @generated from oneof flyteidl.core.Operand.val
   */
  val: {
    /**
     * Can be a constant
     *
     * @generated from field: flyteidl.core.Primitive primitive = 1 [deprecated = true];
     * @deprecated
     */
    value: Primitive;
    case: "primitive";
  } | {
    /**
     * Or one of this node's input variables
     *
     * @generated from field: string var = 2;
     */
    value: string;
    case: "var";
  } | {
    /**
     * Replace the primitive field
     *
     * @generated from field: flyteidl.core.Scalar scalar = 3;
     */
    value: Scalar;
    case: "scalar";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Operand>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.core.Operand";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "primitive", kind: "message", T: Primitive, oneof: "val" },
    { no: 2, name: "var", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "val" },
    { no: 3, name: "scalar", kind: "message", T: Scalar, oneof: "val" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Operand {
    return new Operand().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Operand {
    return new Operand().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Operand {
    return new Operand().fromJsonString(jsonString, options);
  }

  static equals(a: Operand | PlainMessage<Operand> | undefined, b: Operand | PlainMessage<Operand> | undefined): boolean {
    return proto3.util.equals(Operand, a, b);
  }
}

/**
 * Defines a boolean expression tree. It can be a simple or a conjunction expression.
 * Multiple expressions can be combined using a conjunction or a disjunction to result in a final boolean result.
 *
 * @generated from message flyteidl.core.BooleanExpression
 */
export class BooleanExpression extends Message<BooleanExpression> {
  /**
   * @generated from oneof flyteidl.core.BooleanExpression.expr
   */
  expr: {
    /**
     * @generated from field: flyteidl.core.ConjunctionExpression conjunction = 1;
     */
    value: ConjunctionExpression;
    case: "conjunction";
  } | {
    /**
     * @generated from field: flyteidl.core.ComparisonExpression comparison = 2;
     */
    value: ComparisonExpression;
    case: "comparison";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<BooleanExpression>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.core.BooleanExpression";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "conjunction", kind: "message", T: ConjunctionExpression, oneof: "expr" },
    { no: 2, name: "comparison", kind: "message", T: ComparisonExpression, oneof: "expr" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): BooleanExpression {
    return new BooleanExpression().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): BooleanExpression {
    return new BooleanExpression().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): BooleanExpression {
    return new BooleanExpression().fromJsonString(jsonString, options);
  }

  static equals(a: BooleanExpression | PlainMessage<BooleanExpression> | undefined, b: BooleanExpression | PlainMessage<BooleanExpression> | undefined): boolean {
    return proto3.util.equals(BooleanExpression, a, b);
  }
}

/**
 * Defines a conjunction expression of two boolean expressions.
 *
 * @generated from message flyteidl.core.ConjunctionExpression
 */
export class ConjunctionExpression extends Message<ConjunctionExpression> {
  /**
   * @generated from field: flyteidl.core.ConjunctionExpression.LogicalOperator operator = 1;
   */
  operator = ConjunctionExpression_LogicalOperator.AND;

  /**
   * @generated from field: flyteidl.core.BooleanExpression left_expression = 2;
   */
  leftExpression?: BooleanExpression;

  /**
   * @generated from field: flyteidl.core.BooleanExpression right_expression = 3;
   */
  rightExpression?: BooleanExpression;

  constructor(data?: PartialMessage<ConjunctionExpression>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.core.ConjunctionExpression";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "operator", kind: "enum", T: proto3.getEnumType(ConjunctionExpression_LogicalOperator) },
    { no: 2, name: "left_expression", kind: "message", T: BooleanExpression },
    { no: 3, name: "right_expression", kind: "message", T: BooleanExpression },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConjunctionExpression {
    return new ConjunctionExpression().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConjunctionExpression {
    return new ConjunctionExpression().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConjunctionExpression {
    return new ConjunctionExpression().fromJsonString(jsonString, options);
  }

  static equals(a: ConjunctionExpression | PlainMessage<ConjunctionExpression> | undefined, b: ConjunctionExpression | PlainMessage<ConjunctionExpression> | undefined): boolean {
    return proto3.util.equals(ConjunctionExpression, a, b);
  }
}

/**
 * Nested conditions. They can be conjoined using AND / OR
 * Order of evaluation is not important as the operators are Commutative
 *
 * @generated from enum flyteidl.core.ConjunctionExpression.LogicalOperator
 */
export enum ConjunctionExpression_LogicalOperator {
  /**
   * Conjunction
   *
   * @generated from enum value: AND = 0;
   */
  AND = 0,

  /**
   * @generated from enum value: OR = 1;
   */
  OR = 1,
}
// Retrieve enum metadata with: proto3.getEnumType(ConjunctionExpression_LogicalOperator)
proto3.util.setEnumType(ConjunctionExpression_LogicalOperator, "flyteidl.core.ConjunctionExpression.LogicalOperator", [
  { no: 0, name: "AND" },
  { no: 1, name: "OR" },
]);

