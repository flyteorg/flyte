// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/admin/schedule.proto (package flyteidl.admin, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * Represents a frequency at which to run a schedule.
 *
 * @generated from enum flyteidl.admin.FixedRateUnit
 */
export enum FixedRateUnit {
  /**
   * @generated from enum value: MINUTE = 0;
   */
  MINUTE = 0,

  /**
   * @generated from enum value: HOUR = 1;
   */
  HOUR = 1,

  /**
   * @generated from enum value: DAY = 2;
   */
  DAY = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(FixedRateUnit)
proto3.util.setEnumType(FixedRateUnit, "flyteidl.admin.FixedRateUnit", [
  { no: 0, name: "MINUTE" },
  { no: 1, name: "HOUR" },
  { no: 2, name: "DAY" },
]);

/**
 * @generated from enum flyteidl.admin.ConcurrencyPolicy
 */
export enum ConcurrencyPolicy {
  /**
   * @generated from enum value: UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * wait for previous executions to terminate before starting a new one
   *
   * @generated from enum value: WAIT = 1;
   */
  WAIT = 1,

  /**
   * fail the CreateExecution request and do not permit the execution to start
   *
   * @generated from enum value: ABORT = 2;
   */
  ABORT = 2,

  /**
   * terminate the oldest execution when the concurrency limit is reached and immediately begin proceeding with the new execution
   *
   * @generated from enum value: REPLACE = 3;
   */
  REPLACE = 3,
}
// Retrieve enum metadata with: proto3.getEnumType(ConcurrencyPolicy)
proto3.util.setEnumType(ConcurrencyPolicy, "flyteidl.admin.ConcurrencyPolicy", [
  { no: 0, name: "UNSPECIFIED" },
  { no: 1, name: "WAIT" },
  { no: 2, name: "ABORT" },
  { no: 3, name: "REPLACE" },
]);

/**
 * @generated from enum flyteidl.admin.ConcurrencyLevel
 */
export enum ConcurrencyLevel {
  /**
   * Applies concurrency limits across all launch plan versions.
   *
   * @generated from enum value: LAUNCH_PLAN = 0;
   */
  LAUNCH_PLAN = 0,

  /**
   * Applies concurrency at the versioned launch plan level
   *
   * @generated from enum value: LAUNCH_PLAN_VERSION = 1;
   */
  LAUNCH_PLAN_VERSION = 1,
}
// Retrieve enum metadata with: proto3.getEnumType(ConcurrencyLevel)
proto3.util.setEnumType(ConcurrencyLevel, "flyteidl.admin.ConcurrencyLevel", [
  { no: 0, name: "LAUNCH_PLAN" },
  { no: 1, name: "LAUNCH_PLAN_VERSION" },
]);

/**
 * Option for schedules run at a certain frequency e.g. every 2 minutes.
 *
 * @generated from message flyteidl.admin.FixedRate
 */
export class FixedRate extends Message<FixedRate> {
  /**
   * @generated from field: uint32 value = 1;
   */
  value = 0;

  /**
   * @generated from field: flyteidl.admin.FixedRateUnit unit = 2;
   */
  unit = FixedRateUnit.MINUTE;

  constructor(data?: PartialMessage<FixedRate>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.FixedRate";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 2, name: "unit", kind: "enum", T: proto3.getEnumType(FixedRateUnit) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): FixedRate {
    return new FixedRate().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): FixedRate {
    return new FixedRate().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): FixedRate {
    return new FixedRate().fromJsonString(jsonString, options);
  }

  static equals(a: FixedRate | PlainMessage<FixedRate> | undefined, b: FixedRate | PlainMessage<FixedRate> | undefined): boolean {
    return proto3.util.equals(FixedRate, a, b);
  }
}

/**
 * Options for schedules to run according to a cron expression.
 *
 * @generated from message flyteidl.admin.CronSchedule
 */
export class CronSchedule extends Message<CronSchedule> {
  /**
   * Standard/default cron implementation as described by https://en.wikipedia.org/wiki/Cron#CRON_expression;
   * Also supports nonstandard predefined scheduling definitions
   * as described by https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions
   * except @reboot
   *
   * @generated from field: string schedule = 1;
   */
  schedule = "";

  /**
   * ISO 8601 duration as described by https://en.wikipedia.org/wiki/ISO_8601#Durations
   *
   * @generated from field: string offset = 2;
   */
  offset = "";

  constructor(data?: PartialMessage<CronSchedule>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.CronSchedule";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "schedule", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "offset", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CronSchedule {
    return new CronSchedule().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CronSchedule {
    return new CronSchedule().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CronSchedule {
    return new CronSchedule().fromJsonString(jsonString, options);
  }

  static equals(a: CronSchedule | PlainMessage<CronSchedule> | undefined, b: CronSchedule | PlainMessage<CronSchedule> | undefined): boolean {
    return proto3.util.equals(CronSchedule, a, b);
  }
}

/**
 * Defines complete set of information required to trigger an execution on a schedule.
 *
 * @generated from message flyteidl.admin.Schedule
 */
export class Schedule extends Message<Schedule> {
  /**
   * @generated from oneof flyteidl.admin.Schedule.ScheduleExpression
   */
  ScheduleExpression: {
    /**
     * Uses AWS syntax: Minutes Hours Day-of-month Month Day-of-week Year
     * e.g. for a schedule that runs every 15 minutes: 0/15 * * * ? *
     *
     * @generated from field: string cron_expression = 1 [deprecated = true];
     * @deprecated
     */
    value: string;
    case: "cronExpression";
  } | {
    /**
     * @generated from field: flyteidl.admin.FixedRate rate = 2;
     */
    value: FixedRate;
    case: "rate";
  } | {
    /**
     * @generated from field: flyteidl.admin.CronSchedule cron_schedule = 4;
     */
    value: CronSchedule;
    case: "cronSchedule";
  } | { case: undefined; value?: undefined } = { case: undefined };

  /**
   * Name of the input variable that the kickoff time will be supplied to when the workflow is kicked off.
   *
   * @generated from field: string kickoff_time_input_arg = 3;
   */
  kickoffTimeInputArg = "";

  /**
   * @generated from field: flyteidl.admin.SchedulerPolicy scheduler_policy = 5;
   */
  schedulerPolicy?: SchedulerPolicy;

  constructor(data?: PartialMessage<Schedule>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.Schedule";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "cron_expression", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "ScheduleExpression" },
    { no: 2, name: "rate", kind: "message", T: FixedRate, oneof: "ScheduleExpression" },
    { no: 4, name: "cron_schedule", kind: "message", T: CronSchedule, oneof: "ScheduleExpression" },
    { no: 3, name: "kickoff_time_input_arg", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "scheduler_policy", kind: "message", T: SchedulerPolicy },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Schedule {
    return new Schedule().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Schedule {
    return new Schedule().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Schedule {
    return new Schedule().fromJsonString(jsonString, options);
  }

  static equals(a: Schedule | PlainMessage<Schedule> | undefined, b: Schedule | PlainMessage<Schedule> | undefined): boolean {
    return proto3.util.equals(Schedule, a, b);
  }
}

/**
 * @generated from message flyteidl.admin.SchedulerPolicy
 */
export class SchedulerPolicy extends Message<SchedulerPolicy> {
  /**
   * Defines how many executions with this launch plan can run in parallel
   *
   * @generated from field: uint32 max = 1;
   */
  max = 0;

  /**
   * Defines how to handle the execution when the max concurrency is reached.
   *
   * @generated from field: flyteidl.admin.ConcurrencyPolicy policy = 2;
   */
  policy = ConcurrencyPolicy.UNSPECIFIED;

  /**
   * Defines the granularity to apply the concurrency policy to
   *
   * @generated from field: flyteidl.admin.ConcurrencyLevel level = 3;
   */
  level = ConcurrencyLevel.LAUNCH_PLAN;

  constructor(data?: PartialMessage<SchedulerPolicy>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SchedulerPolicy";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "max", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 2, name: "policy", kind: "enum", T: proto3.getEnumType(ConcurrencyPolicy) },
    { no: 3, name: "level", kind: "enum", T: proto3.getEnumType(ConcurrencyLevel) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SchedulerPolicy {
    return new SchedulerPolicy().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SchedulerPolicy {
    return new SchedulerPolicy().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SchedulerPolicy {
    return new SchedulerPolicy().fromJsonString(jsonString, options);
  }

  static equals(a: SchedulerPolicy | PlainMessage<SchedulerPolicy> | undefined, b: SchedulerPolicy | PlainMessage<SchedulerPolicy> | undefined): boolean {
    return proto3.util.equals(SchedulerPolicy, a, b);
  }
}

