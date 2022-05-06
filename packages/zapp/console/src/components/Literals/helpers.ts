import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { Core, Protobuf } from 'flyteidl';
import * as Long from 'long';
import { BlobDimensionality, SchemaColumnType } from 'models/Common/types';

const DEFAULT_UNSUPPORTED = 'This type is not yet supported';

// PRIMITIVE
function processPrimitive(primitive?: (Core.IPrimitive & Pick<Core.Primitive, 'value'>) | null) {
  if (!primitive) {
    return 'invalid primitive';
  }

  const type = primitive.value;

  switch (type) {
    case 'datetime':
      return formatDateUTC(timestampToDate(primitive.datetime as Protobuf.Timestamp));
    case 'duration':
      return protobufDurationToHMS(primitive.duration as Protobuf.Duration);
    case 'integer': {
      return Long.fromValue(primitive.integer as Long).toNumber();
    }
    case 'boolean':
      return primitive.boolean;
    case 'floatValue':
      return primitive.floatValue;
    case 'stringValue':
      return `${primitive.stringValue}`;
    default:
      return 'unknown';
  }
}

// BLOB
const dimensionalityStrings: Record<BlobDimensionality, string> = {
  [BlobDimensionality.SINGLE]: 'single',
  [BlobDimensionality.MULTIPART]: 'multi-part',
};

function processBlobType(blobType?: Core.IBlobType | null) {
  if (!blobType) {
    return 'invalid blob type';
  }

  const formatString = blobType.format ? ` (${blobType.format})` : '';
  const dimensionality = blobType.dimensionality;
  const dimensionalityString =
    dimensionality !== null && dimensionality !== undefined
      ? dimensionalityStrings[dimensionality]
      : '';
  const typeString = `${dimensionalityString}${formatString}`;

  return `${typeString} blob`;
}

function processBlob(blob?: Core.IBlob | null) {
  if (!blob) {
    return 'invalid blob';
  }

  const type = blob.metadata?.type;

  return {
    type: processBlobType(type),
    uri: blob.uri,
  };
}

// BINARY
function processBinary(binary?: Core.IBinary | null) {
  const tag = binary?.tag;

  if (!tag) {
    return 'invalid binary';
  }

  return {
    tag: `${tag} (binary data not shown)`,
  };
}

// SCHEMA
export function columnTypeToString(type?: SchemaColumnType | null) {
  switch (type) {
    case SchemaColumnType.BOOLEAN:
      return 'boolean';
    case SchemaColumnType.DATETIME:
      return 'datetime';
    case SchemaColumnType.DURATION:
      return 'duration';
    case SchemaColumnType.FLOAT:
      return 'float';
    case SchemaColumnType.INTEGER:
      return 'integer';
    case SchemaColumnType.STRING:
      return 'string';
    default:
      return 'unknown';
  }
}

function processSchemaType(schemaType?: Core.ISchemaType | null, shortString = false) {
  const columns =
    schemaType?.columns?.length &&
    schemaType.columns.map((column) => {
      return shortString
        ? `${columnTypeToString(column.type)}`
        : `${column.name} (${columnTypeToString(column.type)})`;
    });

  return columns;
}

function processSchema(schema?: Core.ISchema | null) {
  if (!schema) {
    return 'invalid schema';
  }

  const uri = schema.uri;
  const columns = processSchemaType(schema.type);

  return {
    ...(uri && { uri }),
    ...(columns && { columns }),
  };
}

// NONE
/* eslint-disable @typescript-eslint/no-unused-vars */
function processNone(none?: Core.IVoid | null) {
  return '(empty)';
}

// TODO: FC#450 ass support for union types
function processUnionType(union?: Core.IUnionType | null, shortString = false) {
  return DEFAULT_UNSUPPORTED;
}

function processUnion(union: Core.IUnion) {
  return DEFAULT_UNSUPPORTED;
}
/* eslint-enable @typescript-eslint/no-unused-vars */

// ERROR
function processError(error?: Core.IError | null) {
  return {
    error: error?.message,
    nodeId: error?.failedNodeId,
  };
}

function processProtobufStructValue(struct?: Protobuf.IStruct | null) {
  if (!struct) {
    return 'invalid generic struct value';
  }

  const fields = struct?.fields;
  const res =
    fields &&
    Object.keys(fields)
      .map((v) => {
        return { [v]: processGenericValue(fields[v]) };
      })
      .reduce((acc, v) => ({ ...acc, ...v }), {});

  return res;
}

function processGenericValue(value: Protobuf.IValue & Pick<Protobuf.Value, 'kind'>) {
  const kind = value.kind;

  switch (kind) {
    case 'nullValue':
      return '(empty)';
    case 'listValue': {
      const list = value.listValue;
      return list?.values?.map((x) => processGenericValue(x));
    }
    case 'structValue':
      return processProtobufStructValue(value?.structValue);
    case 'numberValue':
    case 'stringValue':
    case 'boolValue':
      return value[kind];
    default:
      return 'unknown';
  }
}

function processGeneric(struct?: Protobuf.IStruct | null) {
  if (!struct || !struct?.fields) {
    return null;
  }

  const { fields } = struct;
  const mapContent = Object.keys(fields)
    .map((key) => {
      const value = fields[key];
      return { [key]: processGenericValue(value) };
    })
    .reduce((acc, v) => {
      return { ...acc, ...v };
    }, {});

  return mapContent;
}

// SIMPLE
export function processSimpleType(simpleType?: Core.SimpleType | null) {
  switch (simpleType) {
    case Core.SimpleType.NONE:
      return 'none';
    case Core.SimpleType.INTEGER:
      return 'integer';
    case Core.SimpleType.FLOAT:
      return 'float';
    case Core.SimpleType.STRING:
      return 'string';
    case Core.SimpleType.BOOLEAN:
      return 'booleam';
    case Core.SimpleType.DATETIME:
      return 'datetime';
    case Core.SimpleType.DURATION:
      return 'duration';
    case Core.SimpleType.BINARY:
      return 'binary';
    case Core.SimpleType.ERROR:
      return 'error';
    case Core.SimpleType.STRUCT:
      return 'struct';
    default:
      return 'unknown';
  }
}

function processEnumType(enumType?: Core.IEnumType | null) {
  return enumType?.values || [];
}

function processLiteralType(
  literalType?: (Core.ILiteralType & Pick<Core.LiteralType, 'type'>) | null,
) {
  const type = literalType?.type;

  switch (type) {
    case 'simple':
      return processSimpleType(literalType?.simple);
    case 'schema':
      return `schema (${processSchemaType(literalType?.schema, true)})`;
    case 'collectionType':
      return `collection of ${processLiteralType(literalType?.collectionType)}`;
    case 'mapValueType':
      return `map value of ${processLiteralType(literalType?.mapValueType)}`;
    case 'blob':
      return processBlobType(literalType?.blob);
    case 'enumType':
      return `enum (${processEnumType(literalType?.enumType)})`;
    case 'structuredDatasetType':
      return processStructuredDatasetType(literalType?.structuredDatasetType);
    case 'unionType':
      return processUnionType(literalType?.unionType, true);
    default:
      return DEFAULT_UNSUPPORTED;
  }
}

function processStructuredDatasetType(structuredDatasetType?: Core.IStructuredDatasetType | null) {
  if (!structuredDatasetType) {
    return {};
  }

  const { columns, format } = structuredDatasetType;
  const processedColumns =
    columns?.length &&
    columns
      .map(({ name, literalType }) => [name, processLiteralType(literalType)])
      .reduce((acc, v) => {
        acc[v[0]] = v[1];
        return acc;
      }, []);

  return {
    ...(format && { format }),
    ...(processedColumns && { columns: processedColumns }),
  };
}

function processStructuredDataset(structuredDataSet?: Core.IStructuredDataset | null) {
  if (!structuredDataSet) {
    return DEFAULT_UNSUPPORTED;
  }

  const retJson = {} as any;
  const { uri, metadata } = structuredDataSet;

  if (uri) {
    retJson.uri = uri;
  }

  const structuredDatasetType = processStructuredDatasetType(metadata?.structuredDatasetType);

  return {
    ...(uri && { uri }),
    ...structuredDatasetType,
  };
}

function processScalar(scalar?: (Core.IScalar & Pick<Core.Scalar, 'value'>) | null) {
  const type = scalar?.value;

  switch (type) {
    case 'primitive':
      return processPrimitive(scalar?.primitive);
    case 'blob':
      return processBlob(scalar?.blob);
    case 'binary':
      return processBinary(scalar?.binary);
    case 'schema':
      return processSchema(scalar?.schema);
    case 'noneType':
      return processNone(scalar?.noneType);
    case 'error':
      return processError(scalar?.error);
    case 'generic':
      return processGeneric(scalar?.generic);
    case 'structuredDataset':
      return processStructuredDataset(scalar?.structuredDataset);
    case 'union':
      return processUnion(scalar?.union as Core.IUnion);
    default:
      return DEFAULT_UNSUPPORTED;
  }
}

function processCollection(collection?: Core.ILiteralCollection | null) {
  const literals = collection?.literals;

  if (!literals) {
    return 'invalid collection';
  }

  return literals?.map((literal) => processLiteral(literal));
}

function processMap(map?: Core.ILiteralMap | null) {
  const literals = map?.literals;

  if (!literals) {
    return 'invalid map';
  }

  return transformLiterals(literals);
}

function processLiteral(literal?: Core.ILiteral & Pick<Core.Literal, 'value'>) {
  const type = literal?.value;

  if (!literal) {
    return 'invalid literal';
  }

  switch (type) {
    case 'scalar':
      return processScalar(literal.scalar);
    case 'collection':
      return processCollection(literal.collection);
    case 'map':
      return processMap(literal.map);
    default:
      return DEFAULT_UNSUPPORTED;
  }
}

export function transformLiterals(json: { [k: string]: Core.ILiteral }) {
  const obj = Object.entries(json)
    .map(([key, literal]) => ({
      [key]: processLiteral(literal),
    }))
    .reduce(
      (acc, cur) => ({
        ...acc,
        ...cur,
      }),
      {},
    );

  return obj;
}
