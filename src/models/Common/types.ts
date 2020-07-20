import { Admin, Core, Protobuf } from 'flyteidl';
import { Collection } from 'react-virtualized';

/* --- BEGIN flyteidl type aliases --- */
/** These are types shared across multiple sections of the data model. Most of
 * map to types found in `flyteidl.core`.
 */
export interface Alias extends Core.IAlias {}
export interface Binding extends Core.IBinding {}
export interface Container extends Core.IContainer {}
export type FixedRateUnit = Admin.FixedRateUnit;
export const FixedRateUnit = Admin.FixedRateUnit;
export interface Identifier extends Core.IIdentifier {
    resourceType?: ResourceType;
    project: string;
    domain: string;
    name: string;
    version: string;
}

export interface NamedEntityIdentifier
    extends RequiredNonNullable<Admin.INamedEntityIdentifier> {}
export interface NamedEntityMetadata
    extends RequiredNonNullable<Admin.INamedEntityMetadata> {}

export interface NamedEntity extends Admin.INamedEntity {
    resourceType: Core.ResourceType;
    id: NamedEntityIdentifier;
    metadata: NamedEntityMetadata;
}
export interface Notification extends Admin.INotification {}
export type ResourceType = Core.ResourceType;
export const ResourceType = Core.ResourceType;
export interface RetryStrategy extends Core.IRetryStrategy {}
export interface RuntimeMetadata extends Core.IRuntimeMetadata {}
export interface Schedule extends Admin.ISchedule {}
export type MessageFormat = Core.TaskLog.MessageFormat;
export interface TaskLog extends Core.ITaskLog {
    name: string;
    uri: string;
}

/*** Literals ****/
export interface Binary extends RequiredNonNullable<Core.IBinary> {}

export interface Blob extends Core.IBlob {
    metadata: BlobMetadata;
    uri: string;
}
export type BlobDimensionality = Core.BlobType.BlobDimensionality;
export const BlobDimensionality = Core.BlobType.BlobDimensionality;

export interface BlobMetadata extends Core.IBlobMetadata {
    type: BlobType;
}

export interface BlobType extends Core.IBlobType {
    dimensionality: BlobDimensionality;
}
export interface UrlBlob extends Admin.IUrlBlob {}

export interface Error extends RequiredNonNullable<Core.IError> {}

export interface Literal extends Core.Literal {
    value: keyof Core.ILiteral;
    collection?: Core.ILiteralCollection;
    map?: Core.ILiteralMap;
    scalar?: Scalar;
}

/** A Core.ILiteral guaranteed to have all subproperties necessary to specify
 * a Blob.
 */
export interface BlobLiteral extends Core.ILiteral {
    scalar: BlobScalar;
}

export interface LiteralCollection
    extends RequiredNonNullable<Core.ILiteralCollection> {}

export interface LiteralMap extends RequiredNonNullable<Core.ILiteralMap> {}
export const LiteralMap = Core.LiteralMap;
export interface LiteralMapBlob extends Admin.ILiteralMapBlob {
    values: LiteralMap;
}

export interface Scalar extends Core.IScalar {
    primitive?: Primitive;
    value: keyof Core.IScalar;
}
export interface BlobScalar extends Core.IScalar {
    blob: Blob;
}

export interface Schema extends Core.ISchema {
    uri: string;
    type: SchemaType;
}

export interface SchemaColumn extends Core.SchemaType.ISchemaColumn {
    name: string;
    type: SchemaColumnType;
}

export type SchemaColumnType = Core.SchemaType.SchemaColumn.SchemaColumnType;
export const SchemaColumnType = Core.SchemaType.SchemaColumn.SchemaColumnType;

export interface SchemaType extends Core.ISchemaType {
    columns: SchemaColumn[];
}

export interface Primitive extends Core.Primitive {
    value: keyof Core.IPrimitive;
}

export interface ProtobufListValue extends Protobuf.IListValue {
    values: ProtobufValue[];
}

export interface ProtobufStruct extends Protobuf.IStruct {
    fields: Dictionary<ProtobufValue>;
}

export interface ProtobufValue extends Protobuf.IValue {
    kind: keyof Protobuf.IValue;
}

export interface TypedInterface extends Core.ITypedInterface {
    inputs?: VariableMap;
    outputs?: VariableMap;
}

export interface LiteralType extends Core.ILiteralType {
    blob?: BlobType;
    collectionType?: LiteralType;
    mapValueType?: LiteralType;
    metadata?: ProtobufStruct;
    schema?: SchemaType;
    simple?: SimpleType;
}

export type SimpleType = Core.SimpleType;
export const SimpleType = Core.SimpleType;

export interface Variable extends Core.IVariable {
    type: LiteralType;
    description?: string;
}
export interface VariableMap extends Core.IVariableMap {
    variables: Record<string, Variable>;
}

export interface Parameter extends Core.IParameter {
    var: Variable;
    default?: Literal | null;
    required?: boolean;
}

export interface ParameterMap extends Core.IParameterMap {
    parameters: Record<string, Parameter>;
}

/* --- END flyteidl type aliases --- */

export interface ProjectIdentifierScope {
    project: string;
}

export interface DomainIdentifierScope extends ProjectIdentifierScope {
    domain: string;
}

export interface NameIdentifierScope extends DomainIdentifierScope {
    name: string;
}

export type IdentifierScope =
    | ProjectIdentifierScope
    | DomainIdentifierScope
    | NameIdentifierScope
    | Identifier;

export interface UserProfile {
    sub: string;
    name: string;
    preferredUsername: string;
    givenName: string;
    familyName: string;
    email: string;
    picture: string;
}

export type StatusString = 'normal' | 'degraded' | 'down';

export interface SystemStatus {
    message?: string;
    status: StatusString;
}
