import { Admin, Core, Protobuf } from 'flyteidl';

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
export interface Notification extends Admin.INotification {}
export interface ParameterMap extends RequiredNonNullable<Core.IParameterMap> {}
export type ResourceType = Core.ResourceType;
export const ResourceType = Core.ResourceType;
export interface RetryStrategy extends Core.IRetryStrategy {}
export interface RuntimeMetadata extends Core.IRuntimeMetadata {}
export interface Schedule extends Admin.ISchedule {}
export interface TypedInterface extends Core.ITypedInterface {}
export interface UrlBlob extends Admin.IUrlBlob {}
export interface VariableMap extends RequiredNonNullable<Core.IVariableMap> {}

export type MessageFormat = Core.TaskLog.MessageFormat;
export type TaskLog = RequiredNonNullable<Core.ITaskLog>;

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

export interface Error extends RequiredNonNullable<Core.IError> {}

export interface Literal extends Core.Literal {
    value: keyof Core.ILiteral;
    scalar?: Scalar;
}

export interface LiteralCollection
    extends RequiredNonNullable<Core.ILiteralCollection> {}

export interface LiteralMap extends RequiredNonNullable<Core.ILiteralMap> {}
export const LiteralMap = Core.LiteralMap;
export interface LiteralMapBlob extends Admin.ILiteralMapBlob {
    values: LiteralMap;
}

export interface Scalar extends Core.IScalar {
    value: keyof Core.IScalar;
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
