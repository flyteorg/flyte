// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/admin/description_entity.proto (package flyteidl.admin, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Identifier, ResourceType } from "../core/identifier_pb.js";
import { NamedEntityIdentifier, Sort } from "./common_pb.js";

/**
 * The format of the long description
 *
 * @generated from enum flyteidl.admin.DescriptionFormat
 */
export enum DescriptionFormat {
  /**
   * @generated from enum value: DESCRIPTION_FORMAT_UNKNOWN = 0;
   */
  UNKNOWN = 0,

  /**
   * @generated from enum value: DESCRIPTION_FORMAT_MARKDOWN = 1;
   */
  MARKDOWN = 1,

  /**
   * @generated from enum value: DESCRIPTION_FORMAT_HTML = 2;
   */
  HTML = 2,

  /**
   * python default documentation - comments is rst
   *
   * @generated from enum value: DESCRIPTION_FORMAT_RST = 3;
   */
  RST = 3,
}
// Retrieve enum metadata with: proto3.getEnumType(DescriptionFormat)
proto3.util.setEnumType(DescriptionFormat, "flyteidl.admin.DescriptionFormat", [
  { no: 0, name: "DESCRIPTION_FORMAT_UNKNOWN" },
  { no: 1, name: "DESCRIPTION_FORMAT_MARKDOWN" },
  { no: 2, name: "DESCRIPTION_FORMAT_HTML" },
  { no: 3, name: "DESCRIPTION_FORMAT_RST" },
]);

/**
 * DescriptionEntity contains detailed description for the task/workflow.
 * Documentation could provide insight into the algorithms, business use case, etc.
 *
 * @generated from message flyteidl.admin.DescriptionEntity
 */
export class DescriptionEntity extends Message<DescriptionEntity> {
  /**
   * id represents the unique identifier of the description entity.
   *
   * @generated from field: flyteidl.core.Identifier id = 1;
   */
  id?: Identifier;

  /**
   * One-liner overview of the entity.
   *
   * @generated from field: string short_description = 2;
   */
  shortDescription = "";

  /**
   * Full user description with formatting preserved.
   *
   * @generated from field: flyteidl.admin.Description long_description = 3;
   */
  longDescription?: Description;

  /**
   * Optional link to source code used to define this entity.
   *
   * @generated from field: flyteidl.admin.SourceCode source_code = 4;
   */
  sourceCode?: SourceCode;

  /**
   * User-specified tags. These are arbitrary and can be used for searching
   * filtering and discovering tasks.
   *
   * @generated from field: repeated string tags = 5;
   */
  tags: string[] = [];

  constructor(data?: PartialMessage<DescriptionEntity>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.DescriptionEntity";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "message", T: Identifier },
    { no: 2, name: "short_description", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "long_description", kind: "message", T: Description },
    { no: 4, name: "source_code", kind: "message", T: SourceCode },
    { no: 5, name: "tags", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DescriptionEntity {
    return new DescriptionEntity().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DescriptionEntity {
    return new DescriptionEntity().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DescriptionEntity {
    return new DescriptionEntity().fromJsonString(jsonString, options);
  }

  static equals(a: DescriptionEntity | PlainMessage<DescriptionEntity> | undefined, b: DescriptionEntity | PlainMessage<DescriptionEntity> | undefined): boolean {
    return proto3.util.equals(DescriptionEntity, a, b);
  }
}

/**
 * Full user description with formatting preserved. This can be rendered
 * by clients, such as the console or command line tools with in-tact
 * formatting.
 *
 * @generated from message flyteidl.admin.Description
 */
export class Description extends Message<Description> {
  /**
   * @generated from oneof flyteidl.admin.Description.content
   */
  content: {
    /**
     * long description - no more than 4KB
     *
     * @generated from field: string value = 1;
     */
    value: string;
    case: "value";
  } | {
    /**
     * if the description sizes exceed some threshold we can offload the entire
     * description proto altogether to an external data store, like S3 rather than store inline in the db
     *
     * @generated from field: string uri = 2;
     */
    value: string;
    case: "uri";
  } | { case: undefined; value?: undefined } = { case: undefined };

  /**
   * Format of the long description
   *
   * @generated from field: flyteidl.admin.DescriptionFormat format = 3;
   */
  format = DescriptionFormat.UNKNOWN;

  /**
   * Optional link to an icon for the entity
   *
   * @generated from field: string icon_link = 4;
   */
  iconLink = "";

  constructor(data?: PartialMessage<Description>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.Description";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "content" },
    { no: 2, name: "uri", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "content" },
    { no: 3, name: "format", kind: "enum", T: proto3.getEnumType(DescriptionFormat) },
    { no: 4, name: "icon_link", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Description {
    return new Description().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Description {
    return new Description().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Description {
    return new Description().fromJsonString(jsonString, options);
  }

  static equals(a: Description | PlainMessage<Description> | undefined, b: Description | PlainMessage<Description> | undefined): boolean {
    return proto3.util.equals(Description, a, b);
  }
}

/**
 * Link to source code used to define this entity
 *
 * @generated from message flyteidl.admin.SourceCode
 */
export class SourceCode extends Message<SourceCode> {
  /**
   * @generated from field: string link = 1;
   */
  link = "";

  constructor(data?: PartialMessage<SourceCode>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.SourceCode";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "link", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SourceCode {
    return new SourceCode().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SourceCode {
    return new SourceCode().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SourceCode {
    return new SourceCode().fromJsonString(jsonString, options);
  }

  static equals(a: SourceCode | PlainMessage<SourceCode> | undefined, b: SourceCode | PlainMessage<SourceCode> | undefined): boolean {
    return proto3.util.equals(SourceCode, a, b);
  }
}

/**
 * Represents a list of DescriptionEntities returned from the admin.
 * See :ref:`ref_flyteidl.admin.DescriptionEntity` for more details
 *
 * @generated from message flyteidl.admin.DescriptionEntityList
 */
export class DescriptionEntityList extends Message<DescriptionEntityList> {
  /**
   * A list of DescriptionEntities returned based on the request.
   *
   * @generated from field: repeated flyteidl.admin.DescriptionEntity descriptionEntities = 1;
   */
  descriptionEntities: DescriptionEntity[] = [];

  /**
   * In the case of multiple pages of results, the server-provided token can be used to fetch the next page
   * in a query. If there are no more results, this value will be empty.
   *
   * @generated from field: string token = 2;
   */
  token = "";

  constructor(data?: PartialMessage<DescriptionEntityList>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.DescriptionEntityList";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "descriptionEntities", kind: "message", T: DescriptionEntity, repeated: true },
    { no: 2, name: "token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DescriptionEntityList {
    return new DescriptionEntityList().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DescriptionEntityList {
    return new DescriptionEntityList().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DescriptionEntityList {
    return new DescriptionEntityList().fromJsonString(jsonString, options);
  }

  static equals(a: DescriptionEntityList | PlainMessage<DescriptionEntityList> | undefined, b: DescriptionEntityList | PlainMessage<DescriptionEntityList> | undefined): boolean {
    return proto3.util.equals(DescriptionEntityList, a, b);
  }
}

/**
 * Represents a request structure to retrieve a list of DescriptionEntities.
 * See :ref:`ref_flyteidl.admin.DescriptionEntity` for more details
 *
 * @generated from message flyteidl.admin.DescriptionEntityListRequest
 */
export class DescriptionEntityListRequest extends Message<DescriptionEntityListRequest> {
  /**
   * Identifies the specific type of resource that this identifier corresponds to.
   *
   * @generated from field: flyteidl.core.ResourceType resource_type = 1;
   */
  resourceType = ResourceType.UNSPECIFIED;

  /**
   * The identifier for the description entity.
   * +required
   *
   * @generated from field: flyteidl.admin.NamedEntityIdentifier id = 2;
   */
  id?: NamedEntityIdentifier;

  /**
   * Indicates the number of resources to be returned.
   * +required
   *
   * @generated from field: uint32 limit = 3;
   */
  limit = 0;

  /**
   * In the case of multiple pages of results, the server-provided token can be used to fetch the next page
   * in a query.
   * +optional
   *
   * @generated from field: string token = 4;
   */
  token = "";

  /**
   * Indicates a list of filters passed as string.
   * More info on constructing filters : <Link>
   * +optional
   *
   * @generated from field: string filters = 5;
   */
  filters = "";

  /**
   * Sort ordering for returned list.
   * +optional
   *
   * @generated from field: flyteidl.admin.Sort sort_by = 6;
   */
  sortBy?: Sort;

  constructor(data?: PartialMessage<DescriptionEntityListRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.DescriptionEntityListRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "resource_type", kind: "enum", T: proto3.getEnumType(ResourceType) },
    { no: 2, name: "id", kind: "message", T: NamedEntityIdentifier },
    { no: 3, name: "limit", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 4, name: "token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "filters", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "sort_by", kind: "message", T: Sort },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DescriptionEntityListRequest {
    return new DescriptionEntityListRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DescriptionEntityListRequest {
    return new DescriptionEntityListRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DescriptionEntityListRequest {
    return new DescriptionEntityListRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DescriptionEntityListRequest | PlainMessage<DescriptionEntityListRequest> | undefined, b: DescriptionEntityListRequest | PlainMessage<DescriptionEntityListRequest> | undefined): boolean {
    return proto3.util.equals(DescriptionEntityListRequest, a, b);
  }
}

