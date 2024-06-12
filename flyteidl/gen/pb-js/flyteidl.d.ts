import * as $protobuf from "protobufjs";
/** Namespace flyteidl. */
export namespace flyteidl {

    /** Namespace core. */
    namespace core {

        /** Properties of an ArtifactKey. */
        interface IArtifactKey {

            /** ArtifactKey project */
            project?: (string|null);

            /** ArtifactKey domain */
            domain?: (string|null);

            /** ArtifactKey name */
            name?: (string|null);

            /** ArtifactKey org */
            org?: (string|null);
        }

        /** Represents an ArtifactKey. */
        class ArtifactKey implements IArtifactKey {

            /**
             * Constructs a new ArtifactKey.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IArtifactKey);

            /** ArtifactKey project. */
            public project: string;

            /** ArtifactKey domain. */
            public domain: string;

            /** ArtifactKey name. */
            public name: string;

            /** ArtifactKey org. */
            public org: string;

            /**
             * Creates a new ArtifactKey instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ArtifactKey instance
             */
            public static create(properties?: flyteidl.core.IArtifactKey): flyteidl.core.ArtifactKey;

            /**
             * Encodes the specified ArtifactKey message. Does not implicitly {@link flyteidl.core.ArtifactKey.verify|verify} messages.
             * @param message ArtifactKey message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IArtifactKey, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ArtifactKey message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ArtifactKey
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ArtifactKey;

            /**
             * Verifies an ArtifactKey message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ArtifactBindingData. */
        interface IArtifactBindingData {

            /** ArtifactBindingData partitionKey */
            partitionKey?: (string|null);

            /** ArtifactBindingData bindToTimePartition */
            bindToTimePartition?: (boolean|null);

            /** ArtifactBindingData timeTransform */
            timeTransform?: (flyteidl.core.ITimeTransform|null);
        }

        /** Represents an ArtifactBindingData. */
        class ArtifactBindingData implements IArtifactBindingData {

            /**
             * Constructs a new ArtifactBindingData.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IArtifactBindingData);

            /** ArtifactBindingData partitionKey. */
            public partitionKey: string;

            /** ArtifactBindingData bindToTimePartition. */
            public bindToTimePartition: boolean;

            /** ArtifactBindingData timeTransform. */
            public timeTransform?: (flyteidl.core.ITimeTransform|null);

            /** ArtifactBindingData partitionData. */
            public partitionData?: ("partitionKey"|"bindToTimePartition");

            /**
             * Creates a new ArtifactBindingData instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ArtifactBindingData instance
             */
            public static create(properties?: flyteidl.core.IArtifactBindingData): flyteidl.core.ArtifactBindingData;

            /**
             * Encodes the specified ArtifactBindingData message. Does not implicitly {@link flyteidl.core.ArtifactBindingData.verify|verify} messages.
             * @param message ArtifactBindingData message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IArtifactBindingData, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ArtifactBindingData message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ArtifactBindingData
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ArtifactBindingData;

            /**
             * Verifies an ArtifactBindingData message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Granularity enum. */
        enum Granularity {
            UNSET = 0,
            MINUTE = 1,
            HOUR = 2,
            DAY = 3,
            MONTH = 4
        }

        /** Operator enum. */
        enum Operator {
            MINUS = 0,
            PLUS = 1
        }

        /** Properties of a TimeTransform. */
        interface ITimeTransform {

            /** TimeTransform transform */
            transform?: (string|null);

            /** TimeTransform op */
            op?: (flyteidl.core.Operator|null);
        }

        /** Represents a TimeTransform. */
        class TimeTransform implements ITimeTransform {

            /**
             * Constructs a new TimeTransform.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITimeTransform);

            /** TimeTransform transform. */
            public transform: string;

            /** TimeTransform op. */
            public op: flyteidl.core.Operator;

            /**
             * Creates a new TimeTransform instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TimeTransform instance
             */
            public static create(properties?: flyteidl.core.ITimeTransform): flyteidl.core.TimeTransform;

            /**
             * Encodes the specified TimeTransform message. Does not implicitly {@link flyteidl.core.TimeTransform.verify|verify} messages.
             * @param message TimeTransform message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITimeTransform, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TimeTransform message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TimeTransform
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TimeTransform;

            /**
             * Verifies a TimeTransform message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an InputBindingData. */
        interface IInputBindingData {

            /** InputBindingData var */
            "var"?: (string|null);
        }

        /** Represents an InputBindingData. */
        class InputBindingData implements IInputBindingData {

            /**
             * Constructs a new InputBindingData.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IInputBindingData);

            /** InputBindingData var. */
            public var: string;

            /**
             * Creates a new InputBindingData instance using the specified properties.
             * @param [properties] Properties to set
             * @returns InputBindingData instance
             */
            public static create(properties?: flyteidl.core.IInputBindingData): flyteidl.core.InputBindingData;

            /**
             * Encodes the specified InputBindingData message. Does not implicitly {@link flyteidl.core.InputBindingData.verify|verify} messages.
             * @param message InputBindingData message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IInputBindingData, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an InputBindingData message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns InputBindingData
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.InputBindingData;

            /**
             * Verifies an InputBindingData message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a RuntimeBinding. */
        interface IRuntimeBinding {
        }

        /** Represents a RuntimeBinding. */
        class RuntimeBinding implements IRuntimeBinding {

            /**
             * Constructs a new RuntimeBinding.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IRuntimeBinding);

            /**
             * Creates a new RuntimeBinding instance using the specified properties.
             * @param [properties] Properties to set
             * @returns RuntimeBinding instance
             */
            public static create(properties?: flyteidl.core.IRuntimeBinding): flyteidl.core.RuntimeBinding;

            /**
             * Encodes the specified RuntimeBinding message. Does not implicitly {@link flyteidl.core.RuntimeBinding.verify|verify} messages.
             * @param message RuntimeBinding message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IRuntimeBinding, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a RuntimeBinding message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns RuntimeBinding
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.RuntimeBinding;

            /**
             * Verifies a RuntimeBinding message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LabelValue. */
        interface ILabelValue {

            /** LabelValue staticValue */
            staticValue?: (string|null);

            /** LabelValue timeValue */
            timeValue?: (google.protobuf.ITimestamp|null);

            /** LabelValue triggeredBinding */
            triggeredBinding?: (flyteidl.core.IArtifactBindingData|null);

            /** LabelValue inputBinding */
            inputBinding?: (flyteidl.core.IInputBindingData|null);

            /** LabelValue runtimeBinding */
            runtimeBinding?: (flyteidl.core.IRuntimeBinding|null);
        }

        /** Represents a LabelValue. */
        class LabelValue implements ILabelValue {

            /**
             * Constructs a new LabelValue.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ILabelValue);

            /** LabelValue staticValue. */
            public staticValue: string;

            /** LabelValue timeValue. */
            public timeValue?: (google.protobuf.ITimestamp|null);

            /** LabelValue triggeredBinding. */
            public triggeredBinding?: (flyteidl.core.IArtifactBindingData|null);

            /** LabelValue inputBinding. */
            public inputBinding?: (flyteidl.core.IInputBindingData|null);

            /** LabelValue runtimeBinding. */
            public runtimeBinding?: (flyteidl.core.IRuntimeBinding|null);

            /** LabelValue value. */
            public value?: ("staticValue"|"timeValue"|"triggeredBinding"|"inputBinding"|"runtimeBinding");

            /**
             * Creates a new LabelValue instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LabelValue instance
             */
            public static create(properties?: flyteidl.core.ILabelValue): flyteidl.core.LabelValue;

            /**
             * Encodes the specified LabelValue message. Does not implicitly {@link flyteidl.core.LabelValue.verify|verify} messages.
             * @param message LabelValue message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ILabelValue, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LabelValue message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LabelValue
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.LabelValue;

            /**
             * Verifies a LabelValue message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Partitions. */
        interface IPartitions {

            /** Partitions value */
            value?: ({ [k: string]: flyteidl.core.ILabelValue }|null);
        }

        /** Represents a Partitions. */
        class Partitions implements IPartitions {

            /**
             * Constructs a new Partitions.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IPartitions);

            /** Partitions value. */
            public value: { [k: string]: flyteidl.core.ILabelValue };

            /**
             * Creates a new Partitions instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Partitions instance
             */
            public static create(properties?: flyteidl.core.IPartitions): flyteidl.core.Partitions;

            /**
             * Encodes the specified Partitions message. Does not implicitly {@link flyteidl.core.Partitions.verify|verify} messages.
             * @param message Partitions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IPartitions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Partitions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Partitions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Partitions;

            /**
             * Verifies a Partitions message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TimePartition. */
        interface ITimePartition {

            /** TimePartition value */
            value?: (flyteidl.core.ILabelValue|null);

            /** TimePartition granularity */
            granularity?: (flyteidl.core.Granularity|null);
        }

        /** Represents a TimePartition. */
        class TimePartition implements ITimePartition {

            /**
             * Constructs a new TimePartition.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITimePartition);

            /** TimePartition value. */
            public value?: (flyteidl.core.ILabelValue|null);

            /** TimePartition granularity. */
            public granularity: flyteidl.core.Granularity;

            /**
             * Creates a new TimePartition instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TimePartition instance
             */
            public static create(properties?: flyteidl.core.ITimePartition): flyteidl.core.TimePartition;

            /**
             * Encodes the specified TimePartition message. Does not implicitly {@link flyteidl.core.TimePartition.verify|verify} messages.
             * @param message TimePartition message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITimePartition, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TimePartition message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TimePartition
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TimePartition;

            /**
             * Verifies a TimePartition message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ArtifactID. */
        interface IArtifactID {

            /** ArtifactID artifactKey */
            artifactKey?: (flyteidl.core.IArtifactKey|null);

            /** ArtifactID version */
            version?: (string|null);

            /** ArtifactID partitions */
            partitions?: (flyteidl.core.IPartitions|null);

            /** ArtifactID timePartition */
            timePartition?: (flyteidl.core.ITimePartition|null);
        }

        /** Represents an ArtifactID. */
        class ArtifactID implements IArtifactID {

            /**
             * Constructs a new ArtifactID.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IArtifactID);

            /** ArtifactID artifactKey. */
            public artifactKey?: (flyteidl.core.IArtifactKey|null);

            /** ArtifactID version. */
            public version: string;

            /** ArtifactID partitions. */
            public partitions?: (flyteidl.core.IPartitions|null);

            /** ArtifactID timePartition. */
            public timePartition?: (flyteidl.core.ITimePartition|null);

            /**
             * Creates a new ArtifactID instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ArtifactID instance
             */
            public static create(properties?: flyteidl.core.IArtifactID): flyteidl.core.ArtifactID;

            /**
             * Encodes the specified ArtifactID message. Does not implicitly {@link flyteidl.core.ArtifactID.verify|verify} messages.
             * @param message ArtifactID message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IArtifactID, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ArtifactID message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ArtifactID
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ArtifactID;

            /**
             * Verifies an ArtifactID message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ArtifactTag. */
        interface IArtifactTag {

            /** ArtifactTag artifactKey */
            artifactKey?: (flyteidl.core.IArtifactKey|null);

            /** ArtifactTag value */
            value?: (flyteidl.core.ILabelValue|null);
        }

        /** Represents an ArtifactTag. */
        class ArtifactTag implements IArtifactTag {

            /**
             * Constructs a new ArtifactTag.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IArtifactTag);

            /** ArtifactTag artifactKey. */
            public artifactKey?: (flyteidl.core.IArtifactKey|null);

            /** ArtifactTag value. */
            public value?: (flyteidl.core.ILabelValue|null);

            /**
             * Creates a new ArtifactTag instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ArtifactTag instance
             */
            public static create(properties?: flyteidl.core.IArtifactTag): flyteidl.core.ArtifactTag;

            /**
             * Encodes the specified ArtifactTag message. Does not implicitly {@link flyteidl.core.ArtifactTag.verify|verify} messages.
             * @param message ArtifactTag message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IArtifactTag, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ArtifactTag message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ArtifactTag
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ArtifactTag;

            /**
             * Verifies an ArtifactTag message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ArtifactQuery. */
        interface IArtifactQuery {

            /** ArtifactQuery artifactId */
            artifactId?: (flyteidl.core.IArtifactID|null);

            /** ArtifactQuery artifactTag */
            artifactTag?: (flyteidl.core.IArtifactTag|null);

            /** ArtifactQuery uri */
            uri?: (string|null);

            /** ArtifactQuery binding */
            binding?: (flyteidl.core.IArtifactBindingData|null);
        }

        /** Represents an ArtifactQuery. */
        class ArtifactQuery implements IArtifactQuery {

            /**
             * Constructs a new ArtifactQuery.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IArtifactQuery);

            /** ArtifactQuery artifactId. */
            public artifactId?: (flyteidl.core.IArtifactID|null);

            /** ArtifactQuery artifactTag. */
            public artifactTag?: (flyteidl.core.IArtifactTag|null);

            /** ArtifactQuery uri. */
            public uri: string;

            /** ArtifactQuery binding. */
            public binding?: (flyteidl.core.IArtifactBindingData|null);

            /** ArtifactQuery identifier. */
            public identifier?: ("artifactId"|"artifactTag"|"uri"|"binding");

            /**
             * Creates a new ArtifactQuery instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ArtifactQuery instance
             */
            public static create(properties?: flyteidl.core.IArtifactQuery): flyteidl.core.ArtifactQuery;

            /**
             * Encodes the specified ArtifactQuery message. Does not implicitly {@link flyteidl.core.ArtifactQuery.verify|verify} messages.
             * @param message ArtifactQuery message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IArtifactQuery, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ArtifactQuery message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ArtifactQuery
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ArtifactQuery;

            /**
             * Verifies an ArtifactQuery message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** ResourceType enum. */
        enum ResourceType {
            UNSPECIFIED = 0,
            TASK = 1,
            WORKFLOW = 2,
            LAUNCH_PLAN = 3,
            DATASET = 4
        }

        /** Properties of an Identifier. */
        interface IIdentifier {

            /** Identifier resourceType */
            resourceType?: (flyteidl.core.ResourceType|null);

            /** Identifier project */
            project?: (string|null);

            /** Identifier domain */
            domain?: (string|null);

            /** Identifier name */
            name?: (string|null);

            /** Identifier version */
            version?: (string|null);

            /** Identifier org */
            org?: (string|null);
        }

        /** Represents an Identifier. */
        class Identifier implements IIdentifier {

            /**
             * Constructs a new Identifier.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IIdentifier);

            /** Identifier resourceType. */
            public resourceType: flyteidl.core.ResourceType;

            /** Identifier project. */
            public project: string;

            /** Identifier domain. */
            public domain: string;

            /** Identifier name. */
            public name: string;

            /** Identifier version. */
            public version: string;

            /** Identifier org. */
            public org: string;

            /**
             * Creates a new Identifier instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Identifier instance
             */
            public static create(properties?: flyteidl.core.IIdentifier): flyteidl.core.Identifier;

            /**
             * Encodes the specified Identifier message. Does not implicitly {@link flyteidl.core.Identifier.verify|verify} messages.
             * @param message Identifier message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IIdentifier, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Identifier message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Identifier
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Identifier;

            /**
             * Verifies an Identifier message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionIdentifier. */
        interface IWorkflowExecutionIdentifier {

            /** WorkflowExecutionIdentifier project */
            project?: (string|null);

            /** WorkflowExecutionIdentifier domain */
            domain?: (string|null);

            /** WorkflowExecutionIdentifier name */
            name?: (string|null);

            /** WorkflowExecutionIdentifier org */
            org?: (string|null);
        }

        /** Represents a WorkflowExecutionIdentifier. */
        class WorkflowExecutionIdentifier implements IWorkflowExecutionIdentifier {

            /**
             * Constructs a new WorkflowExecutionIdentifier.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IWorkflowExecutionIdentifier);

            /** WorkflowExecutionIdentifier project. */
            public project: string;

            /** WorkflowExecutionIdentifier domain. */
            public domain: string;

            /** WorkflowExecutionIdentifier name. */
            public name: string;

            /** WorkflowExecutionIdentifier org. */
            public org: string;

            /**
             * Creates a new WorkflowExecutionIdentifier instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionIdentifier instance
             */
            public static create(properties?: flyteidl.core.IWorkflowExecutionIdentifier): flyteidl.core.WorkflowExecutionIdentifier;

            /**
             * Encodes the specified WorkflowExecutionIdentifier message. Does not implicitly {@link flyteidl.core.WorkflowExecutionIdentifier.verify|verify} messages.
             * @param message WorkflowExecutionIdentifier message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IWorkflowExecutionIdentifier, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionIdentifier message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionIdentifier
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.WorkflowExecutionIdentifier;

            /**
             * Verifies a WorkflowExecutionIdentifier message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionIdentifier. */
        interface INodeExecutionIdentifier {

            /** NodeExecutionIdentifier nodeId */
            nodeId?: (string|null);

            /** NodeExecutionIdentifier executionId */
            executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);
        }

        /** Represents a NodeExecutionIdentifier. */
        class NodeExecutionIdentifier implements INodeExecutionIdentifier {

            /**
             * Constructs a new NodeExecutionIdentifier.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.INodeExecutionIdentifier);

            /** NodeExecutionIdentifier nodeId. */
            public nodeId: string;

            /** NodeExecutionIdentifier executionId. */
            public executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /**
             * Creates a new NodeExecutionIdentifier instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionIdentifier instance
             */
            public static create(properties?: flyteidl.core.INodeExecutionIdentifier): flyteidl.core.NodeExecutionIdentifier;

            /**
             * Encodes the specified NodeExecutionIdentifier message. Does not implicitly {@link flyteidl.core.NodeExecutionIdentifier.verify|verify} messages.
             * @param message NodeExecutionIdentifier message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.INodeExecutionIdentifier, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionIdentifier message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionIdentifier
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.NodeExecutionIdentifier;

            /**
             * Verifies a NodeExecutionIdentifier message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskExecutionIdentifier. */
        interface ITaskExecutionIdentifier {

            /** TaskExecutionIdentifier taskId */
            taskId?: (flyteidl.core.IIdentifier|null);

            /** TaskExecutionIdentifier nodeExecutionId */
            nodeExecutionId?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** TaskExecutionIdentifier retryAttempt */
            retryAttempt?: (number|null);
        }

        /** Represents a TaskExecutionIdentifier. */
        class TaskExecutionIdentifier implements ITaskExecutionIdentifier {

            /**
             * Constructs a new TaskExecutionIdentifier.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITaskExecutionIdentifier);

            /** TaskExecutionIdentifier taskId. */
            public taskId?: (flyteidl.core.IIdentifier|null);

            /** TaskExecutionIdentifier nodeExecutionId. */
            public nodeExecutionId?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** TaskExecutionIdentifier retryAttempt. */
            public retryAttempt: number;

            /**
             * Creates a new TaskExecutionIdentifier instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskExecutionIdentifier instance
             */
            public static create(properties?: flyteidl.core.ITaskExecutionIdentifier): flyteidl.core.TaskExecutionIdentifier;

            /**
             * Encodes the specified TaskExecutionIdentifier message. Does not implicitly {@link flyteidl.core.TaskExecutionIdentifier.verify|verify} messages.
             * @param message TaskExecutionIdentifier message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITaskExecutionIdentifier, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskExecutionIdentifier message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskExecutionIdentifier
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TaskExecutionIdentifier;

            /**
             * Verifies a TaskExecutionIdentifier message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a SignalIdentifier. */
        interface ISignalIdentifier {

            /** SignalIdentifier signalId */
            signalId?: (string|null);

            /** SignalIdentifier executionId */
            executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);
        }

        /** Represents a SignalIdentifier. */
        class SignalIdentifier implements ISignalIdentifier {

            /**
             * Constructs a new SignalIdentifier.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISignalIdentifier);

            /** SignalIdentifier signalId. */
            public signalId: string;

            /** SignalIdentifier executionId. */
            public executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /**
             * Creates a new SignalIdentifier instance using the specified properties.
             * @param [properties] Properties to set
             * @returns SignalIdentifier instance
             */
            public static create(properties?: flyteidl.core.ISignalIdentifier): flyteidl.core.SignalIdentifier;

            /**
             * Encodes the specified SignalIdentifier message. Does not implicitly {@link flyteidl.core.SignalIdentifier.verify|verify} messages.
             * @param message SignalIdentifier message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISignalIdentifier, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SignalIdentifier message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SignalIdentifier
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.SignalIdentifier;

            /**
             * Verifies a SignalIdentifier message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** CatalogCacheStatus enum. */
        enum CatalogCacheStatus {
            CACHE_DISABLED = 0,
            CACHE_MISS = 1,
            CACHE_HIT = 2,
            CACHE_POPULATED = 3,
            CACHE_LOOKUP_FAILURE = 4,
            CACHE_PUT_FAILURE = 5,
            CACHE_SKIPPED = 6,
            CACHE_EVICTED = 7
        }

        /** Properties of a CatalogArtifactTag. */
        interface ICatalogArtifactTag {

            /** CatalogArtifactTag artifactId */
            artifactId?: (string|null);

            /** CatalogArtifactTag name */
            name?: (string|null);
        }

        /** Represents a CatalogArtifactTag. */
        class CatalogArtifactTag implements ICatalogArtifactTag {

            /**
             * Constructs a new CatalogArtifactTag.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ICatalogArtifactTag);

            /** CatalogArtifactTag artifactId. */
            public artifactId: string;

            /** CatalogArtifactTag name. */
            public name: string;

            /**
             * Creates a new CatalogArtifactTag instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CatalogArtifactTag instance
             */
            public static create(properties?: flyteidl.core.ICatalogArtifactTag): flyteidl.core.CatalogArtifactTag;

            /**
             * Encodes the specified CatalogArtifactTag message. Does not implicitly {@link flyteidl.core.CatalogArtifactTag.verify|verify} messages.
             * @param message CatalogArtifactTag message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ICatalogArtifactTag, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CatalogArtifactTag message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CatalogArtifactTag
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.CatalogArtifactTag;

            /**
             * Verifies a CatalogArtifactTag message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CatalogMetadata. */
        interface ICatalogMetadata {

            /** CatalogMetadata datasetId */
            datasetId?: (flyteidl.core.IIdentifier|null);

            /** CatalogMetadata artifactTag */
            artifactTag?: (flyteidl.core.ICatalogArtifactTag|null);

            /** CatalogMetadata sourceTaskExecution */
            sourceTaskExecution?: (flyteidl.core.ITaskExecutionIdentifier|null);
        }

        /** Represents a CatalogMetadata. */
        class CatalogMetadata implements ICatalogMetadata {

            /**
             * Constructs a new CatalogMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ICatalogMetadata);

            /** CatalogMetadata datasetId. */
            public datasetId?: (flyteidl.core.IIdentifier|null);

            /** CatalogMetadata artifactTag. */
            public artifactTag?: (flyteidl.core.ICatalogArtifactTag|null);

            /** CatalogMetadata sourceTaskExecution. */
            public sourceTaskExecution?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** CatalogMetadata sourceExecution. */
            public sourceExecution?: "sourceTaskExecution";

            /**
             * Creates a new CatalogMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CatalogMetadata instance
             */
            public static create(properties?: flyteidl.core.ICatalogMetadata): flyteidl.core.CatalogMetadata;

            /**
             * Encodes the specified CatalogMetadata message. Does not implicitly {@link flyteidl.core.CatalogMetadata.verify|verify} messages.
             * @param message CatalogMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ICatalogMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CatalogMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CatalogMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.CatalogMetadata;

            /**
             * Verifies a CatalogMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CatalogReservation. */
        interface ICatalogReservation {
        }

        /** Represents a CatalogReservation. */
        class CatalogReservation implements ICatalogReservation {

            /**
             * Constructs a new CatalogReservation.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ICatalogReservation);

            /**
             * Creates a new CatalogReservation instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CatalogReservation instance
             */
            public static create(properties?: flyteidl.core.ICatalogReservation): flyteidl.core.CatalogReservation;

            /**
             * Encodes the specified CatalogReservation message. Does not implicitly {@link flyteidl.core.CatalogReservation.verify|verify} messages.
             * @param message CatalogReservation message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ICatalogReservation, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CatalogReservation message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CatalogReservation
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.CatalogReservation;

            /**
             * Verifies a CatalogReservation message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace CatalogReservation {

            /** Status enum. */
            enum Status {
                RESERVATION_DISABLED = 0,
                RESERVATION_ACQUIRED = 1,
                RESERVATION_EXISTS = 2,
                RESERVATION_RELEASED = 3,
                RESERVATION_FAILURE = 4
            }
        }

        /** Properties of a ConnectionSet. */
        interface IConnectionSet {

            /** ConnectionSet downstream */
            downstream?: ({ [k: string]: flyteidl.core.ConnectionSet.IIdList }|null);

            /** ConnectionSet upstream */
            upstream?: ({ [k: string]: flyteidl.core.ConnectionSet.IIdList }|null);
        }

        /** Represents a ConnectionSet. */
        class ConnectionSet implements IConnectionSet {

            /**
             * Constructs a new ConnectionSet.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IConnectionSet);

            /** ConnectionSet downstream. */
            public downstream: { [k: string]: flyteidl.core.ConnectionSet.IIdList };

            /** ConnectionSet upstream. */
            public upstream: { [k: string]: flyteidl.core.ConnectionSet.IIdList };

            /**
             * Creates a new ConnectionSet instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ConnectionSet instance
             */
            public static create(properties?: flyteidl.core.IConnectionSet): flyteidl.core.ConnectionSet;

            /**
             * Encodes the specified ConnectionSet message. Does not implicitly {@link flyteidl.core.ConnectionSet.verify|verify} messages.
             * @param message ConnectionSet message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IConnectionSet, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ConnectionSet message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ConnectionSet
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ConnectionSet;

            /**
             * Verifies a ConnectionSet message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace ConnectionSet {

            /** Properties of an IdList. */
            interface IIdList {

                /** IdList ids */
                ids?: (string[]|null);
            }

            /** Represents an IdList. */
            class IdList implements IIdList {

                /**
                 * Constructs a new IdList.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: flyteidl.core.ConnectionSet.IIdList);

                /** IdList ids. */
                public ids: string[];

                /**
                 * Creates a new IdList instance using the specified properties.
                 * @param [properties] Properties to set
                 * @returns IdList instance
                 */
                public static create(properties?: flyteidl.core.ConnectionSet.IIdList): flyteidl.core.ConnectionSet.IdList;

                /**
                 * Encodes the specified IdList message. Does not implicitly {@link flyteidl.core.ConnectionSet.IdList.verify|verify} messages.
                 * @param message IdList message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: flyteidl.core.ConnectionSet.IIdList, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes an IdList message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns IdList
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ConnectionSet.IdList;

                /**
                 * Verifies an IdList message.
                 * @param message Plain object to verify
                 * @returns `null` if valid, otherwise the reason why it is not
                 */
                public static verify(message: { [k: string]: any }): (string|null);
            }
        }

        /** Properties of a CompiledWorkflow. */
        interface ICompiledWorkflow {

            /** CompiledWorkflow template */
            template?: (flyteidl.core.IWorkflowTemplate|null);

            /** CompiledWorkflow connections */
            connections?: (flyteidl.core.IConnectionSet|null);
        }

        /** Represents a CompiledWorkflow. */
        class CompiledWorkflow implements ICompiledWorkflow {

            /**
             * Constructs a new CompiledWorkflow.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ICompiledWorkflow);

            /** CompiledWorkflow template. */
            public template?: (flyteidl.core.IWorkflowTemplate|null);

            /** CompiledWorkflow connections. */
            public connections?: (flyteidl.core.IConnectionSet|null);

            /**
             * Creates a new CompiledWorkflow instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CompiledWorkflow instance
             */
            public static create(properties?: flyteidl.core.ICompiledWorkflow): flyteidl.core.CompiledWorkflow;

            /**
             * Encodes the specified CompiledWorkflow message. Does not implicitly {@link flyteidl.core.CompiledWorkflow.verify|verify} messages.
             * @param message CompiledWorkflow message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ICompiledWorkflow, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CompiledWorkflow message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CompiledWorkflow
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.CompiledWorkflow;

            /**
             * Verifies a CompiledWorkflow message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CompiledLaunchPlan. */
        interface ICompiledLaunchPlan {

            /** CompiledLaunchPlan template */
            template?: (flyteidl.core.ILaunchPlanTemplate|null);
        }

        /** Represents a CompiledLaunchPlan. */
        class CompiledLaunchPlan implements ICompiledLaunchPlan {

            /**
             * Constructs a new CompiledLaunchPlan.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ICompiledLaunchPlan);

            /** CompiledLaunchPlan template. */
            public template?: (flyteidl.core.ILaunchPlanTemplate|null);

            /**
             * Creates a new CompiledLaunchPlan instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CompiledLaunchPlan instance
             */
            public static create(properties?: flyteidl.core.ICompiledLaunchPlan): flyteidl.core.CompiledLaunchPlan;

            /**
             * Encodes the specified CompiledLaunchPlan message. Does not implicitly {@link flyteidl.core.CompiledLaunchPlan.verify|verify} messages.
             * @param message CompiledLaunchPlan message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ICompiledLaunchPlan, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CompiledLaunchPlan message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CompiledLaunchPlan
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.CompiledLaunchPlan;

            /**
             * Verifies a CompiledLaunchPlan message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CompiledTask. */
        interface ICompiledTask {

            /** CompiledTask template */
            template?: (flyteidl.core.ITaskTemplate|null);
        }

        /** Represents a CompiledTask. */
        class CompiledTask implements ICompiledTask {

            /**
             * Constructs a new CompiledTask.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ICompiledTask);

            /** CompiledTask template. */
            public template?: (flyteidl.core.ITaskTemplate|null);

            /**
             * Creates a new CompiledTask instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CompiledTask instance
             */
            public static create(properties?: flyteidl.core.ICompiledTask): flyteidl.core.CompiledTask;

            /**
             * Encodes the specified CompiledTask message. Does not implicitly {@link flyteidl.core.CompiledTask.verify|verify} messages.
             * @param message CompiledTask message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ICompiledTask, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CompiledTask message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CompiledTask
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.CompiledTask;

            /**
             * Verifies a CompiledTask message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CompiledWorkflowClosure. */
        interface ICompiledWorkflowClosure {

            /** CompiledWorkflowClosure primary */
            primary?: (flyteidl.core.ICompiledWorkflow|null);

            /** CompiledWorkflowClosure subWorkflows */
            subWorkflows?: (flyteidl.core.ICompiledWorkflow[]|null);

            /** CompiledWorkflowClosure tasks */
            tasks?: (flyteidl.core.ICompiledTask[]|null);

            /** CompiledWorkflowClosure launchPlans */
            launchPlans?: (flyteidl.core.ICompiledLaunchPlan[]|null);
        }

        /** Represents a CompiledWorkflowClosure. */
        class CompiledWorkflowClosure implements ICompiledWorkflowClosure {

            /**
             * Constructs a new CompiledWorkflowClosure.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ICompiledWorkflowClosure);

            /** CompiledWorkflowClosure primary. */
            public primary?: (flyteidl.core.ICompiledWorkflow|null);

            /** CompiledWorkflowClosure subWorkflows. */
            public subWorkflows: flyteidl.core.ICompiledWorkflow[];

            /** CompiledWorkflowClosure tasks. */
            public tasks: flyteidl.core.ICompiledTask[];

            /** CompiledWorkflowClosure launchPlans. */
            public launchPlans: flyteidl.core.ICompiledLaunchPlan[];

            /**
             * Creates a new CompiledWorkflowClosure instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CompiledWorkflowClosure instance
             */
            public static create(properties?: flyteidl.core.ICompiledWorkflowClosure): flyteidl.core.CompiledWorkflowClosure;

            /**
             * Encodes the specified CompiledWorkflowClosure message. Does not implicitly {@link flyteidl.core.CompiledWorkflowClosure.verify|verify} messages.
             * @param message CompiledWorkflowClosure message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ICompiledWorkflowClosure, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CompiledWorkflowClosure message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CompiledWorkflowClosure
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.CompiledWorkflowClosure;

            /**
             * Verifies a CompiledWorkflowClosure message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Variable. */
        interface IVariable {

            /** Variable type */
            type?: (flyteidl.core.ILiteralType|null);

            /** Variable description */
            description?: (string|null);

            /** Variable artifactPartialId */
            artifactPartialId?: (flyteidl.core.IArtifactID|null);

            /** Variable artifactTag */
            artifactTag?: (flyteidl.core.IArtifactTag|null);
        }

        /** Represents a Variable. */
        class Variable implements IVariable {

            /**
             * Constructs a new Variable.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IVariable);

            /** Variable type. */
            public type?: (flyteidl.core.ILiteralType|null);

            /** Variable description. */
            public description: string;

            /** Variable artifactPartialId. */
            public artifactPartialId?: (flyteidl.core.IArtifactID|null);

            /** Variable artifactTag. */
            public artifactTag?: (flyteidl.core.IArtifactTag|null);

            /**
             * Creates a new Variable instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Variable instance
             */
            public static create(properties?: flyteidl.core.IVariable): flyteidl.core.Variable;

            /**
             * Encodes the specified Variable message. Does not implicitly {@link flyteidl.core.Variable.verify|verify} messages.
             * @param message Variable message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IVariable, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Variable message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Variable
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Variable;

            /**
             * Verifies a Variable message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a VariableMap. */
        interface IVariableMap {

            /** VariableMap variables */
            variables?: ({ [k: string]: flyteidl.core.IVariable }|null);
        }

        /** Represents a VariableMap. */
        class VariableMap implements IVariableMap {

            /**
             * Constructs a new VariableMap.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IVariableMap);

            /** VariableMap variables. */
            public variables: { [k: string]: flyteidl.core.IVariable };

            /**
             * Creates a new VariableMap instance using the specified properties.
             * @param [properties] Properties to set
             * @returns VariableMap instance
             */
            public static create(properties?: flyteidl.core.IVariableMap): flyteidl.core.VariableMap;

            /**
             * Encodes the specified VariableMap message. Does not implicitly {@link flyteidl.core.VariableMap.verify|verify} messages.
             * @param message VariableMap message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IVariableMap, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a VariableMap message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns VariableMap
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.VariableMap;

            /**
             * Verifies a VariableMap message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TypedInterface. */
        interface ITypedInterface {

            /** TypedInterface inputs */
            inputs?: (flyteidl.core.IVariableMap|null);

            /** TypedInterface outputs */
            outputs?: (flyteidl.core.IVariableMap|null);
        }

        /** Represents a TypedInterface. */
        class TypedInterface implements ITypedInterface {

            /**
             * Constructs a new TypedInterface.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITypedInterface);

            /** TypedInterface inputs. */
            public inputs?: (flyteidl.core.IVariableMap|null);

            /** TypedInterface outputs. */
            public outputs?: (flyteidl.core.IVariableMap|null);

            /**
             * Creates a new TypedInterface instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TypedInterface instance
             */
            public static create(properties?: flyteidl.core.ITypedInterface): flyteidl.core.TypedInterface;

            /**
             * Encodes the specified TypedInterface message. Does not implicitly {@link flyteidl.core.TypedInterface.verify|verify} messages.
             * @param message TypedInterface message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITypedInterface, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TypedInterface message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TypedInterface
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TypedInterface;

            /**
             * Verifies a TypedInterface message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Parameter. */
        interface IParameter {

            /** Parameter var */
            "var"?: (flyteidl.core.IVariable|null);

            /** Parameter default */
            "default"?: (flyteidl.core.ILiteral|null);

            /** Parameter required */
            required?: (boolean|null);

            /** Parameter artifactQuery */
            artifactQuery?: (flyteidl.core.IArtifactQuery|null);

            /** Parameter artifactId */
            artifactId?: (flyteidl.core.IArtifactID|null);
        }

        /** Represents a Parameter. */
        class Parameter implements IParameter {

            /**
             * Constructs a new Parameter.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IParameter);

            /** Parameter var. */
            public var?: (flyteidl.core.IVariable|null);

            /** Parameter default. */
            public default?: (flyteidl.core.ILiteral|null);

            /** Parameter required. */
            public required: boolean;

            /** Parameter artifactQuery. */
            public artifactQuery?: (flyteidl.core.IArtifactQuery|null);

            /** Parameter artifactId. */
            public artifactId?: (flyteidl.core.IArtifactID|null);

            /** Parameter behavior. */
            public behavior?: ("default"|"required"|"artifactQuery"|"artifactId");

            /**
             * Creates a new Parameter instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Parameter instance
             */
            public static create(properties?: flyteidl.core.IParameter): flyteidl.core.Parameter;

            /**
             * Encodes the specified Parameter message. Does not implicitly {@link flyteidl.core.Parameter.verify|verify} messages.
             * @param message Parameter message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IParameter, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Parameter message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Parameter
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Parameter;

            /**
             * Verifies a Parameter message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ParameterMap. */
        interface IParameterMap {

            /** ParameterMap parameters */
            parameters?: ({ [k: string]: flyteidl.core.IParameter }|null);
        }

        /** Represents a ParameterMap. */
        class ParameterMap implements IParameterMap {

            /**
             * Constructs a new ParameterMap.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IParameterMap);

            /** ParameterMap parameters. */
            public parameters: { [k: string]: flyteidl.core.IParameter };

            /**
             * Creates a new ParameterMap instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ParameterMap instance
             */
            public static create(properties?: flyteidl.core.IParameterMap): flyteidl.core.ParameterMap;

            /**
             * Encodes the specified ParameterMap message. Does not implicitly {@link flyteidl.core.ParameterMap.verify|verify} messages.
             * @param message ParameterMap message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IParameterMap, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ParameterMap message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ParameterMap
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ParameterMap;

            /**
             * Verifies a ParameterMap message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** SimpleType enum. */
        enum SimpleType {
            NONE = 0,
            INTEGER = 1,
            FLOAT = 2,
            STRING = 3,
            BOOLEAN = 4,
            DATETIME = 5,
            DURATION = 6,
            BINARY = 7,
            ERROR = 8,
            STRUCT = 9
        }

        /** Properties of a SchemaType. */
        interface ISchemaType {

            /** SchemaType columns */
            columns?: (flyteidl.core.SchemaType.ISchemaColumn[]|null);
        }

        /** Represents a SchemaType. */
        class SchemaType implements ISchemaType {

            /**
             * Constructs a new SchemaType.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISchemaType);

            /** SchemaType columns. */
            public columns: flyteidl.core.SchemaType.ISchemaColumn[];

            /**
             * Creates a new SchemaType instance using the specified properties.
             * @param [properties] Properties to set
             * @returns SchemaType instance
             */
            public static create(properties?: flyteidl.core.ISchemaType): flyteidl.core.SchemaType;

            /**
             * Encodes the specified SchemaType message. Does not implicitly {@link flyteidl.core.SchemaType.verify|verify} messages.
             * @param message SchemaType message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISchemaType, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SchemaType message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SchemaType
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.SchemaType;

            /**
             * Verifies a SchemaType message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace SchemaType {

            /** Properties of a SchemaColumn. */
            interface ISchemaColumn {

                /** SchemaColumn name */
                name?: (string|null);

                /** SchemaColumn type */
                type?: (flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType|null);
            }

            /** Represents a SchemaColumn. */
            class SchemaColumn implements ISchemaColumn {

                /**
                 * Constructs a new SchemaColumn.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: flyteidl.core.SchemaType.ISchemaColumn);

                /** SchemaColumn name. */
                public name: string;

                /** SchemaColumn type. */
                public type: flyteidl.core.SchemaType.SchemaColumn.SchemaColumnType;

                /**
                 * Creates a new SchemaColumn instance using the specified properties.
                 * @param [properties] Properties to set
                 * @returns SchemaColumn instance
                 */
                public static create(properties?: flyteidl.core.SchemaType.ISchemaColumn): flyteidl.core.SchemaType.SchemaColumn;

                /**
                 * Encodes the specified SchemaColumn message. Does not implicitly {@link flyteidl.core.SchemaType.SchemaColumn.verify|verify} messages.
                 * @param message SchemaColumn message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: flyteidl.core.SchemaType.ISchemaColumn, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes a SchemaColumn message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns SchemaColumn
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.SchemaType.SchemaColumn;

                /**
                 * Verifies a SchemaColumn message.
                 * @param message Plain object to verify
                 * @returns `null` if valid, otherwise the reason why it is not
                 */
                public static verify(message: { [k: string]: any }): (string|null);
            }

            namespace SchemaColumn {

                /** SchemaColumnType enum. */
                enum SchemaColumnType {
                    INTEGER = 0,
                    FLOAT = 1,
                    STRING = 2,
                    BOOLEAN = 3,
                    DATETIME = 4,
                    DURATION = 5
                }
            }
        }

        /** Properties of a StructuredDatasetType. */
        interface IStructuredDatasetType {

            /** StructuredDatasetType columns */
            columns?: (flyteidl.core.StructuredDatasetType.IDatasetColumn[]|null);

            /** StructuredDatasetType format */
            format?: (string|null);

            /** StructuredDatasetType externalSchemaType */
            externalSchemaType?: (string|null);

            /** StructuredDatasetType externalSchemaBytes */
            externalSchemaBytes?: (Uint8Array|null);
        }

        /** Represents a StructuredDatasetType. */
        class StructuredDatasetType implements IStructuredDatasetType {

            /**
             * Constructs a new StructuredDatasetType.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IStructuredDatasetType);

            /** StructuredDatasetType columns. */
            public columns: flyteidl.core.StructuredDatasetType.IDatasetColumn[];

            /** StructuredDatasetType format. */
            public format: string;

            /** StructuredDatasetType externalSchemaType. */
            public externalSchemaType: string;

            /** StructuredDatasetType externalSchemaBytes. */
            public externalSchemaBytes: Uint8Array;

            /**
             * Creates a new StructuredDatasetType instance using the specified properties.
             * @param [properties] Properties to set
             * @returns StructuredDatasetType instance
             */
            public static create(properties?: flyteidl.core.IStructuredDatasetType): flyteidl.core.StructuredDatasetType;

            /**
             * Encodes the specified StructuredDatasetType message. Does not implicitly {@link flyteidl.core.StructuredDatasetType.verify|verify} messages.
             * @param message StructuredDatasetType message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IStructuredDatasetType, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a StructuredDatasetType message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns StructuredDatasetType
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.StructuredDatasetType;

            /**
             * Verifies a StructuredDatasetType message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace StructuredDatasetType {

            /** Properties of a DatasetColumn. */
            interface IDatasetColumn {

                /** DatasetColumn name */
                name?: (string|null);

                /** DatasetColumn literalType */
                literalType?: (flyteidl.core.ILiteralType|null);
            }

            /** Represents a DatasetColumn. */
            class DatasetColumn implements IDatasetColumn {

                /**
                 * Constructs a new DatasetColumn.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: flyteidl.core.StructuredDatasetType.IDatasetColumn);

                /** DatasetColumn name. */
                public name: string;

                /** DatasetColumn literalType. */
                public literalType?: (flyteidl.core.ILiteralType|null);

                /**
                 * Creates a new DatasetColumn instance using the specified properties.
                 * @param [properties] Properties to set
                 * @returns DatasetColumn instance
                 */
                public static create(properties?: flyteidl.core.StructuredDatasetType.IDatasetColumn): flyteidl.core.StructuredDatasetType.DatasetColumn;

                /**
                 * Encodes the specified DatasetColumn message. Does not implicitly {@link flyteidl.core.StructuredDatasetType.DatasetColumn.verify|verify} messages.
                 * @param message DatasetColumn message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: flyteidl.core.StructuredDatasetType.IDatasetColumn, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes a DatasetColumn message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns DatasetColumn
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.StructuredDatasetType.DatasetColumn;

                /**
                 * Verifies a DatasetColumn message.
                 * @param message Plain object to verify
                 * @returns `null` if valid, otherwise the reason why it is not
                 */
                public static verify(message: { [k: string]: any }): (string|null);
            }
        }

        /** Properties of a BlobType. */
        interface IBlobType {

            /** BlobType format */
            format?: (string|null);

            /** BlobType dimensionality */
            dimensionality?: (flyteidl.core.BlobType.BlobDimensionality|null);
        }

        /** Represents a BlobType. */
        class BlobType implements IBlobType {

            /**
             * Constructs a new BlobType.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBlobType);

            /** BlobType format. */
            public format: string;

            /** BlobType dimensionality. */
            public dimensionality: flyteidl.core.BlobType.BlobDimensionality;

            /**
             * Creates a new BlobType instance using the specified properties.
             * @param [properties] Properties to set
             * @returns BlobType instance
             */
            public static create(properties?: flyteidl.core.IBlobType): flyteidl.core.BlobType;

            /**
             * Encodes the specified BlobType message. Does not implicitly {@link flyteidl.core.BlobType.verify|verify} messages.
             * @param message BlobType message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBlobType, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a BlobType message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns BlobType
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.BlobType;

            /**
             * Verifies a BlobType message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace BlobType {

            /** BlobDimensionality enum. */
            enum BlobDimensionality {
                SINGLE = 0,
                MULTIPART = 1
            }
        }

        /** Properties of an EnumType. */
        interface IEnumType {

            /** EnumType values */
            values?: (string[]|null);
        }

        /** Represents an EnumType. */
        class EnumType implements IEnumType {

            /**
             * Constructs a new EnumType.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IEnumType);

            /** EnumType values. */
            public values: string[];

            /**
             * Creates a new EnumType instance using the specified properties.
             * @param [properties] Properties to set
             * @returns EnumType instance
             */
            public static create(properties?: flyteidl.core.IEnumType): flyteidl.core.EnumType;

            /**
             * Encodes the specified EnumType message. Does not implicitly {@link flyteidl.core.EnumType.verify|verify} messages.
             * @param message EnumType message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IEnumType, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EnumType message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EnumType
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.EnumType;

            /**
             * Verifies an EnumType message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an UnionType. */
        interface IUnionType {

            /** UnionType variants */
            variants?: (flyteidl.core.ILiteralType[]|null);
        }

        /** Represents an UnionType. */
        class UnionType implements IUnionType {

            /**
             * Constructs a new UnionType.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IUnionType);

            /** UnionType variants. */
            public variants: flyteidl.core.ILiteralType[];

            /**
             * Creates a new UnionType instance using the specified properties.
             * @param [properties] Properties to set
             * @returns UnionType instance
             */
            public static create(properties?: flyteidl.core.IUnionType): flyteidl.core.UnionType;

            /**
             * Encodes the specified UnionType message. Does not implicitly {@link flyteidl.core.UnionType.verify|verify} messages.
             * @param message UnionType message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IUnionType, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an UnionType message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns UnionType
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.UnionType;

            /**
             * Verifies an UnionType message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TypeStructure. */
        interface ITypeStructure {

            /** TypeStructure tag */
            tag?: (string|null);

            /** TypeStructure dataclassType */
            dataclassType?: ({ [k: string]: flyteidl.core.ILiteralType }|null);
        }

        /** Represents a TypeStructure. */
        class TypeStructure implements ITypeStructure {

            /**
             * Constructs a new TypeStructure.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITypeStructure);

            /** TypeStructure tag. */
            public tag: string;

            /** TypeStructure dataclassType. */
            public dataclassType: { [k: string]: flyteidl.core.ILiteralType };

            /**
             * Creates a new TypeStructure instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TypeStructure instance
             */
            public static create(properties?: flyteidl.core.ITypeStructure): flyteidl.core.TypeStructure;

            /**
             * Encodes the specified TypeStructure message. Does not implicitly {@link flyteidl.core.TypeStructure.verify|verify} messages.
             * @param message TypeStructure message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITypeStructure, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TypeStructure message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TypeStructure
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TypeStructure;

            /**
             * Verifies a TypeStructure message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TypeAnnotation. */
        interface ITypeAnnotation {

            /** TypeAnnotation annotations */
            annotations?: (google.protobuf.IStruct|null);
        }

        /** Represents a TypeAnnotation. */
        class TypeAnnotation implements ITypeAnnotation {

            /**
             * Constructs a new TypeAnnotation.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITypeAnnotation);

            /** TypeAnnotation annotations. */
            public annotations?: (google.protobuf.IStruct|null);

            /**
             * Creates a new TypeAnnotation instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TypeAnnotation instance
             */
            public static create(properties?: flyteidl.core.ITypeAnnotation): flyteidl.core.TypeAnnotation;

            /**
             * Encodes the specified TypeAnnotation message. Does not implicitly {@link flyteidl.core.TypeAnnotation.verify|verify} messages.
             * @param message TypeAnnotation message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITypeAnnotation, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TypeAnnotation message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TypeAnnotation
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TypeAnnotation;

            /**
             * Verifies a TypeAnnotation message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LiteralType. */
        interface ILiteralType {

            /** LiteralType simple */
            simple?: (flyteidl.core.SimpleType|null);

            /** LiteralType schema */
            schema?: (flyteidl.core.ISchemaType|null);

            /** LiteralType collectionType */
            collectionType?: (flyteidl.core.ILiteralType|null);

            /** LiteralType mapValueType */
            mapValueType?: (flyteidl.core.ILiteralType|null);

            /** LiteralType blob */
            blob?: (flyteidl.core.IBlobType|null);

            /** LiteralType enumType */
            enumType?: (flyteidl.core.IEnumType|null);

            /** LiteralType structuredDatasetType */
            structuredDatasetType?: (flyteidl.core.IStructuredDatasetType|null);

            /** LiteralType unionType */
            unionType?: (flyteidl.core.IUnionType|null);

            /** LiteralType metadata */
            metadata?: (google.protobuf.IStruct|null);

            /** LiteralType annotation */
            annotation?: (flyteidl.core.ITypeAnnotation|null);

            /** LiteralType structure */
            structure?: (flyteidl.core.ITypeStructure|null);
        }

        /** Represents a LiteralType. */
        class LiteralType implements ILiteralType {

            /**
             * Constructs a new LiteralType.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ILiteralType);

            /** LiteralType simple. */
            public simple: flyteidl.core.SimpleType;

            /** LiteralType schema. */
            public schema?: (flyteidl.core.ISchemaType|null);

            /** LiteralType collectionType. */
            public collectionType?: (flyteidl.core.ILiteralType|null);

            /** LiteralType mapValueType. */
            public mapValueType?: (flyteidl.core.ILiteralType|null);

            /** LiteralType blob. */
            public blob?: (flyteidl.core.IBlobType|null);

            /** LiteralType enumType. */
            public enumType?: (flyteidl.core.IEnumType|null);

            /** LiteralType structuredDatasetType. */
            public structuredDatasetType?: (flyteidl.core.IStructuredDatasetType|null);

            /** LiteralType unionType. */
            public unionType?: (flyteidl.core.IUnionType|null);

            /** LiteralType metadata. */
            public metadata?: (google.protobuf.IStruct|null);

            /** LiteralType annotation. */
            public annotation?: (flyteidl.core.ITypeAnnotation|null);

            /** LiteralType structure. */
            public structure?: (flyteidl.core.ITypeStructure|null);

            /** LiteralType type. */
            public type?: ("simple"|"schema"|"collectionType"|"mapValueType"|"blob"|"enumType"|"structuredDatasetType"|"unionType");

            /**
             * Creates a new LiteralType instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LiteralType instance
             */
            public static create(properties?: flyteidl.core.ILiteralType): flyteidl.core.LiteralType;

            /**
             * Encodes the specified LiteralType message. Does not implicitly {@link flyteidl.core.LiteralType.verify|verify} messages.
             * @param message LiteralType message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ILiteralType, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LiteralType message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LiteralType
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.LiteralType;

            /**
             * Verifies a LiteralType message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an OutputReference. */
        interface IOutputReference {

            /** OutputReference nodeId */
            nodeId?: (string|null);

            /** OutputReference var */
            "var"?: (string|null);

            /** OutputReference attrPath */
            attrPath?: (flyteidl.core.IPromiseAttribute[]|null);
        }

        /** Represents an OutputReference. */
        class OutputReference implements IOutputReference {

            /**
             * Constructs a new OutputReference.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IOutputReference);

            /** OutputReference nodeId. */
            public nodeId: string;

            /** OutputReference var. */
            public var: string;

            /** OutputReference attrPath. */
            public attrPath: flyteidl.core.IPromiseAttribute[];

            /**
             * Creates a new OutputReference instance using the specified properties.
             * @param [properties] Properties to set
             * @returns OutputReference instance
             */
            public static create(properties?: flyteidl.core.IOutputReference): flyteidl.core.OutputReference;

            /**
             * Encodes the specified OutputReference message. Does not implicitly {@link flyteidl.core.OutputReference.verify|verify} messages.
             * @param message OutputReference message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IOutputReference, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an OutputReference message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns OutputReference
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.OutputReference;

            /**
             * Verifies an OutputReference message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a PromiseAttribute. */
        interface IPromiseAttribute {

            /** PromiseAttribute stringValue */
            stringValue?: (string|null);

            /** PromiseAttribute intValue */
            intValue?: (number|null);
        }

        /** Represents a PromiseAttribute. */
        class PromiseAttribute implements IPromiseAttribute {

            /**
             * Constructs a new PromiseAttribute.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IPromiseAttribute);

            /** PromiseAttribute stringValue. */
            public stringValue: string;

            /** PromiseAttribute intValue. */
            public intValue: number;

            /** PromiseAttribute value. */
            public value?: ("stringValue"|"intValue");

            /**
             * Creates a new PromiseAttribute instance using the specified properties.
             * @param [properties] Properties to set
             * @returns PromiseAttribute instance
             */
            public static create(properties?: flyteidl.core.IPromiseAttribute): flyteidl.core.PromiseAttribute;

            /**
             * Encodes the specified PromiseAttribute message. Does not implicitly {@link flyteidl.core.PromiseAttribute.verify|verify} messages.
             * @param message PromiseAttribute message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IPromiseAttribute, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a PromiseAttribute message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns PromiseAttribute
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.PromiseAttribute;

            /**
             * Verifies a PromiseAttribute message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Error. */
        interface IError {

            /** Error failedNodeId */
            failedNodeId?: (string|null);

            /** Error message */
            message?: (string|null);
        }

        /** Represents an Error. */
        class Error implements IError {

            /**
             * Constructs a new Error.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IError);

            /** Error failedNodeId. */
            public failedNodeId: string;

            /** Error message. */
            public message: string;

            /**
             * Creates a new Error instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Error instance
             */
            public static create(properties?: flyteidl.core.IError): flyteidl.core.Error;

            /**
             * Encodes the specified Error message. Does not implicitly {@link flyteidl.core.Error.verify|verify} messages.
             * @param message Error message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IError, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Error message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Error
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Error;

            /**
             * Verifies an Error message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Primitive. */
        interface IPrimitive {

            /** Primitive integer */
            integer?: (Long|null);

            /** Primitive floatValue */
            floatValue?: (number|null);

            /** Primitive stringValue */
            stringValue?: (string|null);

            /** Primitive boolean */
            boolean?: (boolean|null);

            /** Primitive datetime */
            datetime?: (google.protobuf.ITimestamp|null);

            /** Primitive duration */
            duration?: (google.protobuf.IDuration|null);
        }

        /** Represents a Primitive. */
        class Primitive implements IPrimitive {

            /**
             * Constructs a new Primitive.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IPrimitive);

            /** Primitive integer. */
            public integer: Long;

            /** Primitive floatValue. */
            public floatValue: number;

            /** Primitive stringValue. */
            public stringValue: string;

            /** Primitive boolean. */
            public boolean: boolean;

            /** Primitive datetime. */
            public datetime?: (google.protobuf.ITimestamp|null);

            /** Primitive duration. */
            public duration?: (google.protobuf.IDuration|null);

            /** Primitive value. */
            public value?: ("integer"|"floatValue"|"stringValue"|"boolean"|"datetime"|"duration");

            /**
             * Creates a new Primitive instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Primitive instance
             */
            public static create(properties?: flyteidl.core.IPrimitive): flyteidl.core.Primitive;

            /**
             * Encodes the specified Primitive message. Does not implicitly {@link flyteidl.core.Primitive.verify|verify} messages.
             * @param message Primitive message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IPrimitive, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Primitive message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Primitive
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Primitive;

            /**
             * Verifies a Primitive message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Void. */
        interface IVoid {
        }

        /** Represents a Void. */
        class Void implements IVoid {

            /**
             * Constructs a new Void.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IVoid);

            /**
             * Creates a new Void instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Void instance
             */
            public static create(properties?: flyteidl.core.IVoid): flyteidl.core.Void;

            /**
             * Encodes the specified Void message. Does not implicitly {@link flyteidl.core.Void.verify|verify} messages.
             * @param message Void message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IVoid, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Void message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Void
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Void;

            /**
             * Verifies a Void message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Blob. */
        interface IBlob {

            /** Blob metadata */
            metadata?: (flyteidl.core.IBlobMetadata|null);

            /** Blob uri */
            uri?: (string|null);
        }

        /** Represents a Blob. */
        class Blob implements IBlob {

            /**
             * Constructs a new Blob.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBlob);

            /** Blob metadata. */
            public metadata?: (flyteidl.core.IBlobMetadata|null);

            /** Blob uri. */
            public uri: string;

            /**
             * Creates a new Blob instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Blob instance
             */
            public static create(properties?: flyteidl.core.IBlob): flyteidl.core.Blob;

            /**
             * Encodes the specified Blob message. Does not implicitly {@link flyteidl.core.Blob.verify|verify} messages.
             * @param message Blob message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBlob, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Blob message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Blob
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Blob;

            /**
             * Verifies a Blob message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a BlobMetadata. */
        interface IBlobMetadata {

            /** BlobMetadata type */
            type?: (flyteidl.core.IBlobType|null);
        }

        /** Represents a BlobMetadata. */
        class BlobMetadata implements IBlobMetadata {

            /**
             * Constructs a new BlobMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBlobMetadata);

            /** BlobMetadata type. */
            public type?: (flyteidl.core.IBlobType|null);

            /**
             * Creates a new BlobMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns BlobMetadata instance
             */
            public static create(properties?: flyteidl.core.IBlobMetadata): flyteidl.core.BlobMetadata;

            /**
             * Encodes the specified BlobMetadata message. Does not implicitly {@link flyteidl.core.BlobMetadata.verify|verify} messages.
             * @param message BlobMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBlobMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a BlobMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns BlobMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.BlobMetadata;

            /**
             * Verifies a BlobMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Binary. */
        interface IBinary {

            /** Binary value */
            value?: (Uint8Array|null);

            /** Binary tag */
            tag?: (string|null);
        }

        /** Represents a Binary. */
        class Binary implements IBinary {

            /**
             * Constructs a new Binary.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBinary);

            /** Binary value. */
            public value: Uint8Array;

            /** Binary tag. */
            public tag: string;

            /**
             * Creates a new Binary instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Binary instance
             */
            public static create(properties?: flyteidl.core.IBinary): flyteidl.core.Binary;

            /**
             * Encodes the specified Binary message. Does not implicitly {@link flyteidl.core.Binary.verify|verify} messages.
             * @param message Binary message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBinary, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Binary message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Binary
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Binary;

            /**
             * Verifies a Binary message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Schema. */
        interface ISchema {

            /** Schema uri */
            uri?: (string|null);

            /** Schema type */
            type?: (flyteidl.core.ISchemaType|null);
        }

        /** Represents a Schema. */
        class Schema implements ISchema {

            /**
             * Constructs a new Schema.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISchema);

            /** Schema uri. */
            public uri: string;

            /** Schema type. */
            public type?: (flyteidl.core.ISchemaType|null);

            /**
             * Creates a new Schema instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Schema instance
             */
            public static create(properties?: flyteidl.core.ISchema): flyteidl.core.Schema;

            /**
             * Encodes the specified Schema message. Does not implicitly {@link flyteidl.core.Schema.verify|verify} messages.
             * @param message Schema message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISchema, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Schema message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Schema
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Schema;

            /**
             * Verifies a Schema message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Union. */
        interface IUnion {

            /** Union value */
            value?: (flyteidl.core.ILiteral|null);

            /** Union type */
            type?: (flyteidl.core.ILiteralType|null);
        }

        /** Represents an Union. */
        class Union implements IUnion {

            /**
             * Constructs a new Union.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IUnion);

            /** Union value. */
            public value?: (flyteidl.core.ILiteral|null);

            /** Union type. */
            public type?: (flyteidl.core.ILiteralType|null);

            /**
             * Creates a new Union instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Union instance
             */
            public static create(properties?: flyteidl.core.IUnion): flyteidl.core.Union;

            /**
             * Encodes the specified Union message. Does not implicitly {@link flyteidl.core.Union.verify|verify} messages.
             * @param message Union message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IUnion, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Union message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Union
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Union;

            /**
             * Verifies an Union message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a StructuredDatasetMetadata. */
        interface IStructuredDatasetMetadata {

            /** StructuredDatasetMetadata structuredDatasetType */
            structuredDatasetType?: (flyteidl.core.IStructuredDatasetType|null);
        }

        /** Represents a StructuredDatasetMetadata. */
        class StructuredDatasetMetadata implements IStructuredDatasetMetadata {

            /**
             * Constructs a new StructuredDatasetMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IStructuredDatasetMetadata);

            /** StructuredDatasetMetadata structuredDatasetType. */
            public structuredDatasetType?: (flyteidl.core.IStructuredDatasetType|null);

            /**
             * Creates a new StructuredDatasetMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns StructuredDatasetMetadata instance
             */
            public static create(properties?: flyteidl.core.IStructuredDatasetMetadata): flyteidl.core.StructuredDatasetMetadata;

            /**
             * Encodes the specified StructuredDatasetMetadata message. Does not implicitly {@link flyteidl.core.StructuredDatasetMetadata.verify|verify} messages.
             * @param message StructuredDatasetMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IStructuredDatasetMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a StructuredDatasetMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns StructuredDatasetMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.StructuredDatasetMetadata;

            /**
             * Verifies a StructuredDatasetMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a StructuredDataset. */
        interface IStructuredDataset {

            /** StructuredDataset uri */
            uri?: (string|null);

            /** StructuredDataset metadata */
            metadata?: (flyteidl.core.IStructuredDatasetMetadata|null);
        }

        /** Represents a StructuredDataset. */
        class StructuredDataset implements IStructuredDataset {

            /**
             * Constructs a new StructuredDataset.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IStructuredDataset);

            /** StructuredDataset uri. */
            public uri: string;

            /** StructuredDataset metadata. */
            public metadata?: (flyteidl.core.IStructuredDatasetMetadata|null);

            /**
             * Creates a new StructuredDataset instance using the specified properties.
             * @param [properties] Properties to set
             * @returns StructuredDataset instance
             */
            public static create(properties?: flyteidl.core.IStructuredDataset): flyteidl.core.StructuredDataset;

            /**
             * Encodes the specified StructuredDataset message. Does not implicitly {@link flyteidl.core.StructuredDataset.verify|verify} messages.
             * @param message StructuredDataset message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IStructuredDataset, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a StructuredDataset message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns StructuredDataset
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.StructuredDataset;

            /**
             * Verifies a StructuredDataset message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Scalar. */
        interface IScalar {

            /** Scalar primitive */
            primitive?: (flyteidl.core.IPrimitive|null);

            /** Scalar blob */
            blob?: (flyteidl.core.IBlob|null);

            /** Scalar binary */
            binary?: (flyteidl.core.IBinary|null);

            /** Scalar schema */
            schema?: (flyteidl.core.ISchema|null);

            /** Scalar noneType */
            noneType?: (flyteidl.core.IVoid|null);

            /** Scalar error */
            error?: (flyteidl.core.IError|null);

            /** Scalar generic */
            generic?: (google.protobuf.IStruct|null);

            /** Scalar structuredDataset */
            structuredDataset?: (flyteidl.core.IStructuredDataset|null);

            /** Scalar union */
            union?: (flyteidl.core.IUnion|null);
        }

        /** Represents a Scalar. */
        class Scalar implements IScalar {

            /**
             * Constructs a new Scalar.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IScalar);

            /** Scalar primitive. */
            public primitive?: (flyteidl.core.IPrimitive|null);

            /** Scalar blob. */
            public blob?: (flyteidl.core.IBlob|null);

            /** Scalar binary. */
            public binary?: (flyteidl.core.IBinary|null);

            /** Scalar schema. */
            public schema?: (flyteidl.core.ISchema|null);

            /** Scalar noneType. */
            public noneType?: (flyteidl.core.IVoid|null);

            /** Scalar error. */
            public error?: (flyteidl.core.IError|null);

            /** Scalar generic. */
            public generic?: (google.protobuf.IStruct|null);

            /** Scalar structuredDataset. */
            public structuredDataset?: (flyteidl.core.IStructuredDataset|null);

            /** Scalar union. */
            public union?: (flyteidl.core.IUnion|null);

            /** Scalar value. */
            public value?: ("primitive"|"blob"|"binary"|"schema"|"noneType"|"error"|"generic"|"structuredDataset"|"union");

            /**
             * Creates a new Scalar instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Scalar instance
             */
            public static create(properties?: flyteidl.core.IScalar): flyteidl.core.Scalar;

            /**
             * Encodes the specified Scalar message. Does not implicitly {@link flyteidl.core.Scalar.verify|verify} messages.
             * @param message Scalar message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IScalar, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Scalar message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Scalar
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Scalar;

            /**
             * Verifies a Scalar message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Literal. */
        interface ILiteral {

            /** Literal scalar */
            scalar?: (flyteidl.core.IScalar|null);

            /** Literal collection */
            collection?: (flyteidl.core.ILiteralCollection|null);

            /** Literal map */
            map?: (flyteidl.core.ILiteralMap|null);

            /** Literal hash */
            hash?: (string|null);

            /** Literal metadata */
            metadata?: ({ [k: string]: string }|null);
        }

        /** Represents a Literal. */
        class Literal implements ILiteral {

            /**
             * Constructs a new Literal.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ILiteral);

            /** Literal scalar. */
            public scalar?: (flyteidl.core.IScalar|null);

            /** Literal collection. */
            public collection?: (flyteidl.core.ILiteralCollection|null);

            /** Literal map. */
            public map?: (flyteidl.core.ILiteralMap|null);

            /** Literal hash. */
            public hash: string;

            /** Literal metadata. */
            public metadata: { [k: string]: string };

            /** Literal value. */
            public value?: ("scalar"|"collection"|"map");

            /**
             * Creates a new Literal instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Literal instance
             */
            public static create(properties?: flyteidl.core.ILiteral): flyteidl.core.Literal;

            /**
             * Encodes the specified Literal message. Does not implicitly {@link flyteidl.core.Literal.verify|verify} messages.
             * @param message Literal message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ILiteral, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Literal message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Literal
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Literal;

            /**
             * Verifies a Literal message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LiteralCollection. */
        interface ILiteralCollection {

            /** LiteralCollection literals */
            literals?: (flyteidl.core.ILiteral[]|null);
        }

        /** Represents a LiteralCollection. */
        class LiteralCollection implements ILiteralCollection {

            /**
             * Constructs a new LiteralCollection.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ILiteralCollection);

            /** LiteralCollection literals. */
            public literals: flyteidl.core.ILiteral[];

            /**
             * Creates a new LiteralCollection instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LiteralCollection instance
             */
            public static create(properties?: flyteidl.core.ILiteralCollection): flyteidl.core.LiteralCollection;

            /**
             * Encodes the specified LiteralCollection message. Does not implicitly {@link flyteidl.core.LiteralCollection.verify|verify} messages.
             * @param message LiteralCollection message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ILiteralCollection, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LiteralCollection message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LiteralCollection
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.LiteralCollection;

            /**
             * Verifies a LiteralCollection message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LiteralMap. */
        interface ILiteralMap {

            /** LiteralMap literals */
            literals?: ({ [k: string]: flyteidl.core.ILiteral }|null);
        }

        /** Represents a LiteralMap. */
        class LiteralMap implements ILiteralMap {

            /**
             * Constructs a new LiteralMap.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ILiteralMap);

            /** LiteralMap literals. */
            public literals: { [k: string]: flyteidl.core.ILiteral };

            /**
             * Creates a new LiteralMap instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LiteralMap instance
             */
            public static create(properties?: flyteidl.core.ILiteralMap): flyteidl.core.LiteralMap;

            /**
             * Encodes the specified LiteralMap message. Does not implicitly {@link flyteidl.core.LiteralMap.verify|verify} messages.
             * @param message LiteralMap message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ILiteralMap, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LiteralMap message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LiteralMap
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.LiteralMap;

            /**
             * Verifies a LiteralMap message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a BindingDataCollection. */
        interface IBindingDataCollection {

            /** BindingDataCollection bindings */
            bindings?: (flyteidl.core.IBindingData[]|null);
        }

        /** Represents a BindingDataCollection. */
        class BindingDataCollection implements IBindingDataCollection {

            /**
             * Constructs a new BindingDataCollection.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBindingDataCollection);

            /** BindingDataCollection bindings. */
            public bindings: flyteidl.core.IBindingData[];

            /**
             * Creates a new BindingDataCollection instance using the specified properties.
             * @param [properties] Properties to set
             * @returns BindingDataCollection instance
             */
            public static create(properties?: flyteidl.core.IBindingDataCollection): flyteidl.core.BindingDataCollection;

            /**
             * Encodes the specified BindingDataCollection message. Does not implicitly {@link flyteidl.core.BindingDataCollection.verify|verify} messages.
             * @param message BindingDataCollection message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBindingDataCollection, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a BindingDataCollection message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns BindingDataCollection
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.BindingDataCollection;

            /**
             * Verifies a BindingDataCollection message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a BindingDataMap. */
        interface IBindingDataMap {

            /** BindingDataMap bindings */
            bindings?: ({ [k: string]: flyteidl.core.IBindingData }|null);
        }

        /** Represents a BindingDataMap. */
        class BindingDataMap implements IBindingDataMap {

            /**
             * Constructs a new BindingDataMap.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBindingDataMap);

            /** BindingDataMap bindings. */
            public bindings: { [k: string]: flyteidl.core.IBindingData };

            /**
             * Creates a new BindingDataMap instance using the specified properties.
             * @param [properties] Properties to set
             * @returns BindingDataMap instance
             */
            public static create(properties?: flyteidl.core.IBindingDataMap): flyteidl.core.BindingDataMap;

            /**
             * Encodes the specified BindingDataMap message. Does not implicitly {@link flyteidl.core.BindingDataMap.verify|verify} messages.
             * @param message BindingDataMap message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBindingDataMap, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a BindingDataMap message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns BindingDataMap
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.BindingDataMap;

            /**
             * Verifies a BindingDataMap message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an UnionInfo. */
        interface IUnionInfo {

            /** UnionInfo targetType */
            targetType?: (flyteidl.core.ILiteralType|null);
        }

        /** Represents an UnionInfo. */
        class UnionInfo implements IUnionInfo {

            /**
             * Constructs a new UnionInfo.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IUnionInfo);

            /** UnionInfo targetType. */
            public targetType?: (flyteidl.core.ILiteralType|null);

            /**
             * Creates a new UnionInfo instance using the specified properties.
             * @param [properties] Properties to set
             * @returns UnionInfo instance
             */
            public static create(properties?: flyteidl.core.IUnionInfo): flyteidl.core.UnionInfo;

            /**
             * Encodes the specified UnionInfo message. Does not implicitly {@link flyteidl.core.UnionInfo.verify|verify} messages.
             * @param message UnionInfo message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IUnionInfo, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an UnionInfo message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns UnionInfo
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.UnionInfo;

            /**
             * Verifies an UnionInfo message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a BindingData. */
        interface IBindingData {

            /** BindingData scalar */
            scalar?: (flyteidl.core.IScalar|null);

            /** BindingData collection */
            collection?: (flyteidl.core.IBindingDataCollection|null);

            /** BindingData promise */
            promise?: (flyteidl.core.IOutputReference|null);

            /** BindingData map */
            map?: (flyteidl.core.IBindingDataMap|null);

            /** BindingData union */
            union?: (flyteidl.core.IUnionInfo|null);
        }

        /** Represents a BindingData. */
        class BindingData implements IBindingData {

            /**
             * Constructs a new BindingData.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBindingData);

            /** BindingData scalar. */
            public scalar?: (flyteidl.core.IScalar|null);

            /** BindingData collection. */
            public collection?: (flyteidl.core.IBindingDataCollection|null);

            /** BindingData promise. */
            public promise?: (flyteidl.core.IOutputReference|null);

            /** BindingData map. */
            public map?: (flyteidl.core.IBindingDataMap|null);

            /** BindingData union. */
            public union?: (flyteidl.core.IUnionInfo|null);

            /** BindingData value. */
            public value?: ("scalar"|"collection"|"promise"|"map");

            /**
             * Creates a new BindingData instance using the specified properties.
             * @param [properties] Properties to set
             * @returns BindingData instance
             */
            public static create(properties?: flyteidl.core.IBindingData): flyteidl.core.BindingData;

            /**
             * Encodes the specified BindingData message. Does not implicitly {@link flyteidl.core.BindingData.verify|verify} messages.
             * @param message BindingData message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBindingData, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a BindingData message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns BindingData
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.BindingData;

            /**
             * Verifies a BindingData message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Binding. */
        interface IBinding {

            /** Binding var */
            "var"?: (string|null);

            /** Binding binding */
            binding?: (flyteidl.core.IBindingData|null);
        }

        /** Represents a Binding. */
        class Binding implements IBinding {

            /**
             * Constructs a new Binding.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBinding);

            /** Binding var. */
            public var: string;

            /** Binding binding. */
            public binding?: (flyteidl.core.IBindingData|null);

            /**
             * Creates a new Binding instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Binding instance
             */
            public static create(properties?: flyteidl.core.IBinding): flyteidl.core.Binding;

            /**
             * Encodes the specified Binding message. Does not implicitly {@link flyteidl.core.Binding.verify|verify} messages.
             * @param message Binding message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBinding, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Binding message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Binding
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Binding;

            /**
             * Verifies a Binding message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a KeyValuePair. */
        interface IKeyValuePair {

            /** KeyValuePair key */
            key?: (string|null);

            /** KeyValuePair value */
            value?: (string|null);
        }

        /** Represents a KeyValuePair. */
        class KeyValuePair implements IKeyValuePair {

            /**
             * Constructs a new KeyValuePair.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IKeyValuePair);

            /** KeyValuePair key. */
            public key: string;

            /** KeyValuePair value. */
            public value: string;

            /**
             * Creates a new KeyValuePair instance using the specified properties.
             * @param [properties] Properties to set
             * @returns KeyValuePair instance
             */
            public static create(properties?: flyteidl.core.IKeyValuePair): flyteidl.core.KeyValuePair;

            /**
             * Encodes the specified KeyValuePair message. Does not implicitly {@link flyteidl.core.KeyValuePair.verify|verify} messages.
             * @param message KeyValuePair message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IKeyValuePair, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a KeyValuePair message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns KeyValuePair
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.KeyValuePair;

            /**
             * Verifies a KeyValuePair message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a RetryStrategy. */
        interface IRetryStrategy {

            /** RetryStrategy retries */
            retries?: (number|null);
        }

        /** Represents a RetryStrategy. */
        class RetryStrategy implements IRetryStrategy {

            /**
             * Constructs a new RetryStrategy.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IRetryStrategy);

            /** RetryStrategy retries. */
            public retries: number;

            /**
             * Creates a new RetryStrategy instance using the specified properties.
             * @param [properties] Properties to set
             * @returns RetryStrategy instance
             */
            public static create(properties?: flyteidl.core.IRetryStrategy): flyteidl.core.RetryStrategy;

            /**
             * Encodes the specified RetryStrategy message. Does not implicitly {@link flyteidl.core.RetryStrategy.verify|verify} messages.
             * @param message RetryStrategy message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IRetryStrategy, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a RetryStrategy message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns RetryStrategy
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.RetryStrategy;

            /**
             * Verifies a RetryStrategy message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an IfBlock. */
        interface IIfBlock {

            /** IfBlock condition */
            condition?: (flyteidl.core.IBooleanExpression|null);

            /** IfBlock thenNode */
            thenNode?: (flyteidl.core.INode|null);
        }

        /** Represents an IfBlock. */
        class IfBlock implements IIfBlock {

            /**
             * Constructs a new IfBlock.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IIfBlock);

            /** IfBlock condition. */
            public condition?: (flyteidl.core.IBooleanExpression|null);

            /** IfBlock thenNode. */
            public thenNode?: (flyteidl.core.INode|null);

            /**
             * Creates a new IfBlock instance using the specified properties.
             * @param [properties] Properties to set
             * @returns IfBlock instance
             */
            public static create(properties?: flyteidl.core.IIfBlock): flyteidl.core.IfBlock;

            /**
             * Encodes the specified IfBlock message. Does not implicitly {@link flyteidl.core.IfBlock.verify|verify} messages.
             * @param message IfBlock message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IIfBlock, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an IfBlock message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns IfBlock
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.IfBlock;

            /**
             * Verifies an IfBlock message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an IfElseBlock. */
        interface IIfElseBlock {

            /** IfElseBlock case */
            "case"?: (flyteidl.core.IIfBlock|null);

            /** IfElseBlock other */
            other?: (flyteidl.core.IIfBlock[]|null);

            /** IfElseBlock elseNode */
            elseNode?: (flyteidl.core.INode|null);

            /** IfElseBlock error */
            error?: (flyteidl.core.IError|null);
        }

        /** Represents an IfElseBlock. */
        class IfElseBlock implements IIfElseBlock {

            /**
             * Constructs a new IfElseBlock.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IIfElseBlock);

            /** IfElseBlock case. */
            public case?: (flyteidl.core.IIfBlock|null);

            /** IfElseBlock other. */
            public other: flyteidl.core.IIfBlock[];

            /** IfElseBlock elseNode. */
            public elseNode?: (flyteidl.core.INode|null);

            /** IfElseBlock error. */
            public error?: (flyteidl.core.IError|null);

            /** IfElseBlock default. */
            public default_?: ("elseNode"|"error");

            /**
             * Creates a new IfElseBlock instance using the specified properties.
             * @param [properties] Properties to set
             * @returns IfElseBlock instance
             */
            public static create(properties?: flyteidl.core.IIfElseBlock): flyteidl.core.IfElseBlock;

            /**
             * Encodes the specified IfElseBlock message. Does not implicitly {@link flyteidl.core.IfElseBlock.verify|verify} messages.
             * @param message IfElseBlock message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IIfElseBlock, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an IfElseBlock message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns IfElseBlock
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.IfElseBlock;

            /**
             * Verifies an IfElseBlock message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a BranchNode. */
        interface IBranchNode {

            /** BranchNode ifElse */
            ifElse?: (flyteidl.core.IIfElseBlock|null);
        }

        /** Represents a BranchNode. */
        class BranchNode implements IBranchNode {

            /**
             * Constructs a new BranchNode.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBranchNode);

            /** BranchNode ifElse. */
            public ifElse?: (flyteidl.core.IIfElseBlock|null);

            /**
             * Creates a new BranchNode instance using the specified properties.
             * @param [properties] Properties to set
             * @returns BranchNode instance
             */
            public static create(properties?: flyteidl.core.IBranchNode): flyteidl.core.BranchNode;

            /**
             * Encodes the specified BranchNode message. Does not implicitly {@link flyteidl.core.BranchNode.verify|verify} messages.
             * @param message BranchNode message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBranchNode, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a BranchNode message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns BranchNode
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.BranchNode;

            /**
             * Verifies a BranchNode message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskNode. */
        interface ITaskNode {

            /** TaskNode referenceId */
            referenceId?: (flyteidl.core.IIdentifier|null);

            /** TaskNode overrides */
            overrides?: (flyteidl.core.ITaskNodeOverrides|null);
        }

        /** Represents a TaskNode. */
        class TaskNode implements ITaskNode {

            /**
             * Constructs a new TaskNode.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITaskNode);

            /** TaskNode referenceId. */
            public referenceId?: (flyteidl.core.IIdentifier|null);

            /** TaskNode overrides. */
            public overrides?: (flyteidl.core.ITaskNodeOverrides|null);

            /** TaskNode reference. */
            public reference?: "referenceId";

            /**
             * Creates a new TaskNode instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskNode instance
             */
            public static create(properties?: flyteidl.core.ITaskNode): flyteidl.core.TaskNode;

            /**
             * Encodes the specified TaskNode message. Does not implicitly {@link flyteidl.core.TaskNode.verify|verify} messages.
             * @param message TaskNode message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITaskNode, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskNode message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskNode
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TaskNode;

            /**
             * Verifies a TaskNode message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowNode. */
        interface IWorkflowNode {

            /** WorkflowNode launchplanRef */
            launchplanRef?: (flyteidl.core.IIdentifier|null);

            /** WorkflowNode subWorkflowRef */
            subWorkflowRef?: (flyteidl.core.IIdentifier|null);
        }

        /** Represents a WorkflowNode. */
        class WorkflowNode implements IWorkflowNode {

            /**
             * Constructs a new WorkflowNode.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IWorkflowNode);

            /** WorkflowNode launchplanRef. */
            public launchplanRef?: (flyteidl.core.IIdentifier|null);

            /** WorkflowNode subWorkflowRef. */
            public subWorkflowRef?: (flyteidl.core.IIdentifier|null);

            /** WorkflowNode reference. */
            public reference?: ("launchplanRef"|"subWorkflowRef");

            /**
             * Creates a new WorkflowNode instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowNode instance
             */
            public static create(properties?: flyteidl.core.IWorkflowNode): flyteidl.core.WorkflowNode;

            /**
             * Encodes the specified WorkflowNode message. Does not implicitly {@link flyteidl.core.WorkflowNode.verify|verify} messages.
             * @param message WorkflowNode message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IWorkflowNode, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowNode message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowNode
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.WorkflowNode;

            /**
             * Verifies a WorkflowNode message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ApproveCondition. */
        interface IApproveCondition {

            /** ApproveCondition signalId */
            signalId?: (string|null);
        }

        /** Represents an ApproveCondition. */
        class ApproveCondition implements IApproveCondition {

            /**
             * Constructs a new ApproveCondition.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IApproveCondition);

            /** ApproveCondition signalId. */
            public signalId: string;

            /**
             * Creates a new ApproveCondition instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ApproveCondition instance
             */
            public static create(properties?: flyteidl.core.IApproveCondition): flyteidl.core.ApproveCondition;

            /**
             * Encodes the specified ApproveCondition message. Does not implicitly {@link flyteidl.core.ApproveCondition.verify|verify} messages.
             * @param message ApproveCondition message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IApproveCondition, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ApproveCondition message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ApproveCondition
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ApproveCondition;

            /**
             * Verifies an ApproveCondition message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a SignalCondition. */
        interface ISignalCondition {

            /** SignalCondition signalId */
            signalId?: (string|null);

            /** SignalCondition type */
            type?: (flyteidl.core.ILiteralType|null);

            /** SignalCondition outputVariableName */
            outputVariableName?: (string|null);
        }

        /** Represents a SignalCondition. */
        class SignalCondition implements ISignalCondition {

            /**
             * Constructs a new SignalCondition.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISignalCondition);

            /** SignalCondition signalId. */
            public signalId: string;

            /** SignalCondition type. */
            public type?: (flyteidl.core.ILiteralType|null);

            /** SignalCondition outputVariableName. */
            public outputVariableName: string;

            /**
             * Creates a new SignalCondition instance using the specified properties.
             * @param [properties] Properties to set
             * @returns SignalCondition instance
             */
            public static create(properties?: flyteidl.core.ISignalCondition): flyteidl.core.SignalCondition;

            /**
             * Encodes the specified SignalCondition message. Does not implicitly {@link flyteidl.core.SignalCondition.verify|verify} messages.
             * @param message SignalCondition message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISignalCondition, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SignalCondition message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SignalCondition
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.SignalCondition;

            /**
             * Verifies a SignalCondition message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a SleepCondition. */
        interface ISleepCondition {

            /** SleepCondition duration */
            duration?: (google.protobuf.IDuration|null);
        }

        /** Represents a SleepCondition. */
        class SleepCondition implements ISleepCondition {

            /**
             * Constructs a new SleepCondition.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISleepCondition);

            /** SleepCondition duration. */
            public duration?: (google.protobuf.IDuration|null);

            /**
             * Creates a new SleepCondition instance using the specified properties.
             * @param [properties] Properties to set
             * @returns SleepCondition instance
             */
            public static create(properties?: flyteidl.core.ISleepCondition): flyteidl.core.SleepCondition;

            /**
             * Encodes the specified SleepCondition message. Does not implicitly {@link flyteidl.core.SleepCondition.verify|verify} messages.
             * @param message SleepCondition message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISleepCondition, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SleepCondition message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SleepCondition
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.SleepCondition;

            /**
             * Verifies a SleepCondition message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GateNode. */
        interface IGateNode {

            /** GateNode approve */
            approve?: (flyteidl.core.IApproveCondition|null);

            /** GateNode signal */
            signal?: (flyteidl.core.ISignalCondition|null);

            /** GateNode sleep */
            sleep?: (flyteidl.core.ISleepCondition|null);
        }

        /** Represents a GateNode. */
        class GateNode implements IGateNode {

            /**
             * Constructs a new GateNode.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IGateNode);

            /** GateNode approve. */
            public approve?: (flyteidl.core.IApproveCondition|null);

            /** GateNode signal. */
            public signal?: (flyteidl.core.ISignalCondition|null);

            /** GateNode sleep. */
            public sleep?: (flyteidl.core.ISleepCondition|null);

            /** GateNode condition. */
            public condition?: ("approve"|"signal"|"sleep");

            /**
             * Creates a new GateNode instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GateNode instance
             */
            public static create(properties?: flyteidl.core.IGateNode): flyteidl.core.GateNode;

            /**
             * Encodes the specified GateNode message. Does not implicitly {@link flyteidl.core.GateNode.verify|verify} messages.
             * @param message GateNode message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IGateNode, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GateNode message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GateNode
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.GateNode;

            /**
             * Verifies a GateNode message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ArrayNode. */
        interface IArrayNode {

            /** ArrayNode node */
            node?: (flyteidl.core.INode|null);

            /** ArrayNode parallelism */
            parallelism?: (number|null);

            /** ArrayNode minSuccesses */
            minSuccesses?: (number|null);

            /** ArrayNode minSuccessRatio */
            minSuccessRatio?: (number|null);
        }

        /** Represents an ArrayNode. */
        class ArrayNode implements IArrayNode {

            /**
             * Constructs a new ArrayNode.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IArrayNode);

            /** ArrayNode node. */
            public node?: (flyteidl.core.INode|null);

            /** ArrayNode parallelism. */
            public parallelism: number;

            /** ArrayNode minSuccesses. */
            public minSuccesses: number;

            /** ArrayNode minSuccessRatio. */
            public minSuccessRatio: number;

            /** ArrayNode parallelismOption. */
            public parallelismOption?: "parallelism";

            /** ArrayNode successCriteria. */
            public successCriteria?: ("minSuccesses"|"minSuccessRatio");

            /**
             * Creates a new ArrayNode instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ArrayNode instance
             */
            public static create(properties?: flyteidl.core.IArrayNode): flyteidl.core.ArrayNode;

            /**
             * Encodes the specified ArrayNode message. Does not implicitly {@link flyteidl.core.ArrayNode.verify|verify} messages.
             * @param message ArrayNode message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IArrayNode, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ArrayNode message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ArrayNode
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ArrayNode;

            /**
             * Verifies an ArrayNode message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeMetadata. */
        interface INodeMetadata {

            /** NodeMetadata name */
            name?: (string|null);

            /** NodeMetadata timeout */
            timeout?: (google.protobuf.IDuration|null);

            /** NodeMetadata retries */
            retries?: (flyteidl.core.IRetryStrategy|null);

            /** NodeMetadata interruptible */
            interruptible?: (boolean|null);

            /** NodeMetadata cacheable */
            cacheable?: (boolean|null);

            /** NodeMetadata cacheVersion */
            cacheVersion?: (string|null);

            /** NodeMetadata cacheSerializable */
            cacheSerializable?: (boolean|null);
        }

        /** Represents a NodeMetadata. */
        class NodeMetadata implements INodeMetadata {

            /**
             * Constructs a new NodeMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.INodeMetadata);

            /** NodeMetadata name. */
            public name: string;

            /** NodeMetadata timeout. */
            public timeout?: (google.protobuf.IDuration|null);

            /** NodeMetadata retries. */
            public retries?: (flyteidl.core.IRetryStrategy|null);

            /** NodeMetadata interruptible. */
            public interruptible: boolean;

            /** NodeMetadata cacheable. */
            public cacheable: boolean;

            /** NodeMetadata cacheVersion. */
            public cacheVersion: string;

            /** NodeMetadata cacheSerializable. */
            public cacheSerializable: boolean;

            /** NodeMetadata interruptibleValue. */
            public interruptibleValue?: "interruptible";

            /** NodeMetadata cacheableValue. */
            public cacheableValue?: "cacheable";

            /** NodeMetadata cacheVersionValue. */
            public cacheVersionValue?: "cacheVersion";

            /** NodeMetadata cacheSerializableValue. */
            public cacheSerializableValue?: "cacheSerializable";

            /**
             * Creates a new NodeMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeMetadata instance
             */
            public static create(properties?: flyteidl.core.INodeMetadata): flyteidl.core.NodeMetadata;

            /**
             * Encodes the specified NodeMetadata message. Does not implicitly {@link flyteidl.core.NodeMetadata.verify|verify} messages.
             * @param message NodeMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.INodeMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.NodeMetadata;

            /**
             * Verifies a NodeMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Alias. */
        interface IAlias {

            /** Alias var */
            "var"?: (string|null);

            /** Alias alias */
            alias?: (string|null);
        }

        /** Represents an Alias. */
        class Alias implements IAlias {

            /**
             * Constructs a new Alias.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IAlias);

            /** Alias var. */
            public var: string;

            /** Alias alias. */
            public alias: string;

            /**
             * Creates a new Alias instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Alias instance
             */
            public static create(properties?: flyteidl.core.IAlias): flyteidl.core.Alias;

            /**
             * Encodes the specified Alias message. Does not implicitly {@link flyteidl.core.Alias.verify|verify} messages.
             * @param message Alias message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IAlias, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Alias message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Alias
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Alias;

            /**
             * Verifies an Alias message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Node. */
        interface INode {

            /** Node id */
            id?: (string|null);

            /** Node metadata */
            metadata?: (flyteidl.core.INodeMetadata|null);

            /** Node inputs */
            inputs?: (flyteidl.core.IBinding[]|null);

            /** Node upstreamNodeIds */
            upstreamNodeIds?: (string[]|null);

            /** Node outputAliases */
            outputAliases?: (flyteidl.core.IAlias[]|null);

            /** Node taskNode */
            taskNode?: (flyteidl.core.ITaskNode|null);

            /** Node workflowNode */
            workflowNode?: (flyteidl.core.IWorkflowNode|null);

            /** Node branchNode */
            branchNode?: (flyteidl.core.IBranchNode|null);

            /** Node gateNode */
            gateNode?: (flyteidl.core.IGateNode|null);

            /** Node arrayNode */
            arrayNode?: (flyteidl.core.IArrayNode|null);
        }

        /** Represents a Node. */
        class Node implements INode {

            /**
             * Constructs a new Node.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.INode);

            /** Node id. */
            public id: string;

            /** Node metadata. */
            public metadata?: (flyteidl.core.INodeMetadata|null);

            /** Node inputs. */
            public inputs: flyteidl.core.IBinding[];

            /** Node upstreamNodeIds. */
            public upstreamNodeIds: string[];

            /** Node outputAliases. */
            public outputAliases: flyteidl.core.IAlias[];

            /** Node taskNode. */
            public taskNode?: (flyteidl.core.ITaskNode|null);

            /** Node workflowNode. */
            public workflowNode?: (flyteidl.core.IWorkflowNode|null);

            /** Node branchNode. */
            public branchNode?: (flyteidl.core.IBranchNode|null);

            /** Node gateNode. */
            public gateNode?: (flyteidl.core.IGateNode|null);

            /** Node arrayNode. */
            public arrayNode?: (flyteidl.core.IArrayNode|null);

            /** Node target. */
            public target?: ("taskNode"|"workflowNode"|"branchNode"|"gateNode"|"arrayNode");

            /**
             * Creates a new Node instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Node instance
             */
            public static create(properties?: flyteidl.core.INode): flyteidl.core.Node;

            /**
             * Encodes the specified Node message. Does not implicitly {@link flyteidl.core.Node.verify|verify} messages.
             * @param message Node message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.INode, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Node message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Node
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Node;

            /**
             * Verifies a Node message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowMetadata. */
        interface IWorkflowMetadata {

            /** WorkflowMetadata qualityOfService */
            qualityOfService?: (flyteidl.core.IQualityOfService|null);

            /** WorkflowMetadata onFailure */
            onFailure?: (flyteidl.core.WorkflowMetadata.OnFailurePolicy|null);

            /** WorkflowMetadata tags */
            tags?: ({ [k: string]: string }|null);
        }

        /** Represents a WorkflowMetadata. */
        class WorkflowMetadata implements IWorkflowMetadata {

            /**
             * Constructs a new WorkflowMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IWorkflowMetadata);

            /** WorkflowMetadata qualityOfService. */
            public qualityOfService?: (flyteidl.core.IQualityOfService|null);

            /** WorkflowMetadata onFailure. */
            public onFailure: flyteidl.core.WorkflowMetadata.OnFailurePolicy;

            /** WorkflowMetadata tags. */
            public tags: { [k: string]: string };

            /**
             * Creates a new WorkflowMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowMetadata instance
             */
            public static create(properties?: flyteidl.core.IWorkflowMetadata): flyteidl.core.WorkflowMetadata;

            /**
             * Encodes the specified WorkflowMetadata message. Does not implicitly {@link flyteidl.core.WorkflowMetadata.verify|verify} messages.
             * @param message WorkflowMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IWorkflowMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.WorkflowMetadata;

            /**
             * Verifies a WorkflowMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace WorkflowMetadata {

            /** OnFailurePolicy enum. */
            enum OnFailurePolicy {
                FAIL_IMMEDIATELY = 0,
                FAIL_AFTER_EXECUTABLE_NODES_COMPLETE = 1
            }
        }

        /** Properties of a WorkflowMetadataDefaults. */
        interface IWorkflowMetadataDefaults {

            /** WorkflowMetadataDefaults interruptible */
            interruptible?: (boolean|null);
        }

        /** Represents a WorkflowMetadataDefaults. */
        class WorkflowMetadataDefaults implements IWorkflowMetadataDefaults {

            /**
             * Constructs a new WorkflowMetadataDefaults.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IWorkflowMetadataDefaults);

            /** WorkflowMetadataDefaults interruptible. */
            public interruptible: boolean;

            /**
             * Creates a new WorkflowMetadataDefaults instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowMetadataDefaults instance
             */
            public static create(properties?: flyteidl.core.IWorkflowMetadataDefaults): flyteidl.core.WorkflowMetadataDefaults;

            /**
             * Encodes the specified WorkflowMetadataDefaults message. Does not implicitly {@link flyteidl.core.WorkflowMetadataDefaults.verify|verify} messages.
             * @param message WorkflowMetadataDefaults message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IWorkflowMetadataDefaults, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowMetadataDefaults message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowMetadataDefaults
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.WorkflowMetadataDefaults;

            /**
             * Verifies a WorkflowMetadataDefaults message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowTemplate. */
        interface IWorkflowTemplate {

            /** WorkflowTemplate id */
            id?: (flyteidl.core.IIdentifier|null);

            /** WorkflowTemplate metadata */
            metadata?: (flyteidl.core.IWorkflowMetadata|null);

            /** WorkflowTemplate interface */
            "interface"?: (flyteidl.core.ITypedInterface|null);

            /** WorkflowTemplate nodes */
            nodes?: (flyteidl.core.INode[]|null);

            /** WorkflowTemplate outputs */
            outputs?: (flyteidl.core.IBinding[]|null);

            /** WorkflowTemplate failureNode */
            failureNode?: (flyteidl.core.INode|null);

            /** WorkflowTemplate metadataDefaults */
            metadataDefaults?: (flyteidl.core.IWorkflowMetadataDefaults|null);
        }

        /** Represents a WorkflowTemplate. */
        class WorkflowTemplate implements IWorkflowTemplate {

            /**
             * Constructs a new WorkflowTemplate.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IWorkflowTemplate);

            /** WorkflowTemplate id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** WorkflowTemplate metadata. */
            public metadata?: (flyteidl.core.IWorkflowMetadata|null);

            /** WorkflowTemplate interface. */
            public interface?: (flyteidl.core.ITypedInterface|null);

            /** WorkflowTemplate nodes. */
            public nodes: flyteidl.core.INode[];

            /** WorkflowTemplate outputs. */
            public outputs: flyteidl.core.IBinding[];

            /** WorkflowTemplate failureNode. */
            public failureNode?: (flyteidl.core.INode|null);

            /** WorkflowTemplate metadataDefaults. */
            public metadataDefaults?: (flyteidl.core.IWorkflowMetadataDefaults|null);

            /**
             * Creates a new WorkflowTemplate instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowTemplate instance
             */
            public static create(properties?: flyteidl.core.IWorkflowTemplate): flyteidl.core.WorkflowTemplate;

            /**
             * Encodes the specified WorkflowTemplate message. Does not implicitly {@link flyteidl.core.WorkflowTemplate.verify|verify} messages.
             * @param message WorkflowTemplate message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IWorkflowTemplate, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowTemplate message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowTemplate
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.WorkflowTemplate;

            /**
             * Verifies a WorkflowTemplate message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskNodeOverrides. */
        interface ITaskNodeOverrides {

            /** TaskNodeOverrides resources */
            resources?: (flyteidl.core.IResources|null);

            /** TaskNodeOverrides extendedResources */
            extendedResources?: (flyteidl.core.IExtendedResources|null);

            /** TaskNodeOverrides containerImage */
            containerImage?: (string|null);
        }

        /** Represents a TaskNodeOverrides. */
        class TaskNodeOverrides implements ITaskNodeOverrides {

            /**
             * Constructs a new TaskNodeOverrides.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITaskNodeOverrides);

            /** TaskNodeOverrides resources. */
            public resources?: (flyteidl.core.IResources|null);

            /** TaskNodeOverrides extendedResources. */
            public extendedResources?: (flyteidl.core.IExtendedResources|null);

            /** TaskNodeOverrides containerImage. */
            public containerImage: string;

            /**
             * Creates a new TaskNodeOverrides instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskNodeOverrides instance
             */
            public static create(properties?: flyteidl.core.ITaskNodeOverrides): flyteidl.core.TaskNodeOverrides;

            /**
             * Encodes the specified TaskNodeOverrides message. Does not implicitly {@link flyteidl.core.TaskNodeOverrides.verify|verify} messages.
             * @param message TaskNodeOverrides message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITaskNodeOverrides, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskNodeOverrides message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskNodeOverrides
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TaskNodeOverrides;

            /**
             * Verifies a TaskNodeOverrides message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanTemplate. */
        interface ILaunchPlanTemplate {

            /** LaunchPlanTemplate id */
            id?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlanTemplate interface */
            "interface"?: (flyteidl.core.ITypedInterface|null);

            /** LaunchPlanTemplate fixedInputs */
            fixedInputs?: (flyteidl.core.ILiteralMap|null);
        }

        /** Represents a LaunchPlanTemplate. */
        class LaunchPlanTemplate implements ILaunchPlanTemplate {

            /**
             * Constructs a new LaunchPlanTemplate.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ILaunchPlanTemplate);

            /** LaunchPlanTemplate id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlanTemplate interface. */
            public interface?: (flyteidl.core.ITypedInterface|null);

            /** LaunchPlanTemplate fixedInputs. */
            public fixedInputs?: (flyteidl.core.ILiteralMap|null);

            /**
             * Creates a new LaunchPlanTemplate instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanTemplate instance
             */
            public static create(properties?: flyteidl.core.ILaunchPlanTemplate): flyteidl.core.LaunchPlanTemplate;

            /**
             * Encodes the specified LaunchPlanTemplate message. Does not implicitly {@link flyteidl.core.LaunchPlanTemplate.verify|verify} messages.
             * @param message LaunchPlanTemplate message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ILaunchPlanTemplate, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanTemplate message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanTemplate
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.LaunchPlanTemplate;

            /**
             * Verifies a LaunchPlanTemplate message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ComparisonExpression. */
        interface IComparisonExpression {

            /** ComparisonExpression operator */
            operator?: (flyteidl.core.ComparisonExpression.Operator|null);

            /** ComparisonExpression leftValue */
            leftValue?: (flyteidl.core.IOperand|null);

            /** ComparisonExpression rightValue */
            rightValue?: (flyteidl.core.IOperand|null);
        }

        /** Represents a ComparisonExpression. */
        class ComparisonExpression implements IComparisonExpression {

            /**
             * Constructs a new ComparisonExpression.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IComparisonExpression);

            /** ComparisonExpression operator. */
            public operator: flyteidl.core.ComparisonExpression.Operator;

            /** ComparisonExpression leftValue. */
            public leftValue?: (flyteidl.core.IOperand|null);

            /** ComparisonExpression rightValue. */
            public rightValue?: (flyteidl.core.IOperand|null);

            /**
             * Creates a new ComparisonExpression instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ComparisonExpression instance
             */
            public static create(properties?: flyteidl.core.IComparisonExpression): flyteidl.core.ComparisonExpression;

            /**
             * Encodes the specified ComparisonExpression message. Does not implicitly {@link flyteidl.core.ComparisonExpression.verify|verify} messages.
             * @param message ComparisonExpression message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IComparisonExpression, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ComparisonExpression message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ComparisonExpression
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ComparisonExpression;

            /**
             * Verifies a ComparisonExpression message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace ComparisonExpression {

            /** Operator enum. */
            enum Operator {
                EQ = 0,
                NEQ = 1,
                GT = 2,
                GTE = 3,
                LT = 4,
                LTE = 5
            }
        }

        /** Properties of an Operand. */
        interface IOperand {

            /** Operand primitive */
            primitive?: (flyteidl.core.IPrimitive|null);

            /** Operand var */
            "var"?: (string|null);

            /** Operand scalar */
            scalar?: (flyteidl.core.IScalar|null);
        }

        /** Represents an Operand. */
        class Operand implements IOperand {

            /**
             * Constructs a new Operand.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IOperand);

            /** Operand primitive. */
            public primitive?: (flyteidl.core.IPrimitive|null);

            /** Operand var. */
            public var: string;

            /** Operand scalar. */
            public scalar?: (flyteidl.core.IScalar|null);

            /** Operand val. */
            public val?: ("primitive"|"var"|"scalar");

            /**
             * Creates a new Operand instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Operand instance
             */
            public static create(properties?: flyteidl.core.IOperand): flyteidl.core.Operand;

            /**
             * Encodes the specified Operand message. Does not implicitly {@link flyteidl.core.Operand.verify|verify} messages.
             * @param message Operand message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IOperand, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Operand message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Operand
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Operand;

            /**
             * Verifies an Operand message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a BooleanExpression. */
        interface IBooleanExpression {

            /** BooleanExpression conjunction */
            conjunction?: (flyteidl.core.IConjunctionExpression|null);

            /** BooleanExpression comparison */
            comparison?: (flyteidl.core.IComparisonExpression|null);
        }

        /** Represents a BooleanExpression. */
        class BooleanExpression implements IBooleanExpression {

            /**
             * Constructs a new BooleanExpression.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IBooleanExpression);

            /** BooleanExpression conjunction. */
            public conjunction?: (flyteidl.core.IConjunctionExpression|null);

            /** BooleanExpression comparison. */
            public comparison?: (flyteidl.core.IComparisonExpression|null);

            /** BooleanExpression expr. */
            public expr?: ("conjunction"|"comparison");

            /**
             * Creates a new BooleanExpression instance using the specified properties.
             * @param [properties] Properties to set
             * @returns BooleanExpression instance
             */
            public static create(properties?: flyteidl.core.IBooleanExpression): flyteidl.core.BooleanExpression;

            /**
             * Encodes the specified BooleanExpression message. Does not implicitly {@link flyteidl.core.BooleanExpression.verify|verify} messages.
             * @param message BooleanExpression message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IBooleanExpression, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a BooleanExpression message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns BooleanExpression
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.BooleanExpression;

            /**
             * Verifies a BooleanExpression message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ConjunctionExpression. */
        interface IConjunctionExpression {

            /** ConjunctionExpression operator */
            operator?: (flyteidl.core.ConjunctionExpression.LogicalOperator|null);

            /** ConjunctionExpression leftExpression */
            leftExpression?: (flyteidl.core.IBooleanExpression|null);

            /** ConjunctionExpression rightExpression */
            rightExpression?: (flyteidl.core.IBooleanExpression|null);
        }

        /** Represents a ConjunctionExpression. */
        class ConjunctionExpression implements IConjunctionExpression {

            /**
             * Constructs a new ConjunctionExpression.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IConjunctionExpression);

            /** ConjunctionExpression operator. */
            public operator: flyteidl.core.ConjunctionExpression.LogicalOperator;

            /** ConjunctionExpression leftExpression. */
            public leftExpression?: (flyteidl.core.IBooleanExpression|null);

            /** ConjunctionExpression rightExpression. */
            public rightExpression?: (flyteidl.core.IBooleanExpression|null);

            /**
             * Creates a new ConjunctionExpression instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ConjunctionExpression instance
             */
            public static create(properties?: flyteidl.core.IConjunctionExpression): flyteidl.core.ConjunctionExpression;

            /**
             * Encodes the specified ConjunctionExpression message. Does not implicitly {@link flyteidl.core.ConjunctionExpression.verify|verify} messages.
             * @param message ConjunctionExpression message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IConjunctionExpression, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ConjunctionExpression message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ConjunctionExpression
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ConjunctionExpression;

            /**
             * Verifies a ConjunctionExpression message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace ConjunctionExpression {

            /** LogicalOperator enum. */
            enum LogicalOperator {
                AND = 0,
                OR = 1
            }
        }

        /** Properties of a WorkflowExecution. */
        interface IWorkflowExecution {
        }

        /** Represents a WorkflowExecution. */
        class WorkflowExecution implements IWorkflowExecution {

            /**
             * Constructs a new WorkflowExecution.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IWorkflowExecution);

            /**
             * Creates a new WorkflowExecution instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecution instance
             */
            public static create(properties?: flyteidl.core.IWorkflowExecution): flyteidl.core.WorkflowExecution;

            /**
             * Encodes the specified WorkflowExecution message. Does not implicitly {@link flyteidl.core.WorkflowExecution.verify|verify} messages.
             * @param message WorkflowExecution message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IWorkflowExecution, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecution message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecution
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.WorkflowExecution;

            /**
             * Verifies a WorkflowExecution message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace WorkflowExecution {

            /** Phase enum. */
            enum Phase {
                UNDEFINED = 0,
                QUEUED = 1,
                RUNNING = 2,
                SUCCEEDING = 3,
                SUCCEEDED = 4,
                FAILING = 5,
                FAILED = 6,
                ABORTED = 7,
                TIMED_OUT = 8,
                ABORTING = 9
            }
        }

        /** Properties of a NodeExecution. */
        interface INodeExecution {
        }

        /** Represents a NodeExecution. */
        class NodeExecution implements INodeExecution {

            /**
             * Constructs a new NodeExecution.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.INodeExecution);

            /**
             * Creates a new NodeExecution instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecution instance
             */
            public static create(properties?: flyteidl.core.INodeExecution): flyteidl.core.NodeExecution;

            /**
             * Encodes the specified NodeExecution message. Does not implicitly {@link flyteidl.core.NodeExecution.verify|verify} messages.
             * @param message NodeExecution message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.INodeExecution, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecution message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecution
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.NodeExecution;

            /**
             * Verifies a NodeExecution message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace NodeExecution {

            /** Phase enum. */
            enum Phase {
                UNDEFINED = 0,
                QUEUED = 1,
                RUNNING = 2,
                SUCCEEDED = 3,
                FAILING = 4,
                FAILED = 5,
                ABORTED = 6,
                SKIPPED = 7,
                TIMED_OUT = 8,
                DYNAMIC_RUNNING = 9,
                RECOVERED = 10
            }
        }

        /** Properties of a TaskExecution. */
        interface ITaskExecution {
        }

        /** Represents a TaskExecution. */
        class TaskExecution implements ITaskExecution {

            /**
             * Constructs a new TaskExecution.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITaskExecution);

            /**
             * Creates a new TaskExecution instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskExecution instance
             */
            public static create(properties?: flyteidl.core.ITaskExecution): flyteidl.core.TaskExecution;

            /**
             * Encodes the specified TaskExecution message. Does not implicitly {@link flyteidl.core.TaskExecution.verify|verify} messages.
             * @param message TaskExecution message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITaskExecution, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskExecution message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskExecution
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TaskExecution;

            /**
             * Verifies a TaskExecution message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace TaskExecution {

            /** Phase enum. */
            enum Phase {
                UNDEFINED = 0,
                QUEUED = 1,
                RUNNING = 2,
                SUCCEEDED = 3,
                ABORTED = 4,
                FAILED = 5,
                INITIALIZING = 6,
                WAITING_FOR_RESOURCES = 7
            }
        }

        /** Properties of an ExecutionError. */
        interface IExecutionError {

            /** ExecutionError code */
            code?: (string|null);

            /** ExecutionError message */
            message?: (string|null);

            /** ExecutionError errorUri */
            errorUri?: (string|null);

            /** ExecutionError kind */
            kind?: (flyteidl.core.ExecutionError.ErrorKind|null);
        }

        /** Represents an ExecutionError. */
        class ExecutionError implements IExecutionError {

            /**
             * Constructs a new ExecutionError.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IExecutionError);

            /** ExecutionError code. */
            public code: string;

            /** ExecutionError message. */
            public message: string;

            /** ExecutionError errorUri. */
            public errorUri: string;

            /** ExecutionError kind. */
            public kind: flyteidl.core.ExecutionError.ErrorKind;

            /**
             * Creates a new ExecutionError instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionError instance
             */
            public static create(properties?: flyteidl.core.IExecutionError): flyteidl.core.ExecutionError;

            /**
             * Encodes the specified ExecutionError message. Does not implicitly {@link flyteidl.core.ExecutionError.verify|verify} messages.
             * @param message ExecutionError message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IExecutionError, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionError message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionError
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ExecutionError;

            /**
             * Verifies an ExecutionError message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace ExecutionError {

            /** ErrorKind enum. */
            enum ErrorKind {
                UNKNOWN = 0,
                USER = 1,
                SYSTEM = 2
            }
        }

        /** Properties of a TaskLog. */
        interface ITaskLog {

            /** TaskLog uri */
            uri?: (string|null);

            /** TaskLog name */
            name?: (string|null);

            /** TaskLog messageFormat */
            messageFormat?: (flyteidl.core.TaskLog.MessageFormat|null);

            /** TaskLog ttl */
            ttl?: (google.protobuf.IDuration|null);

            /** TaskLog ShowWhilePending */
            ShowWhilePending?: (boolean|null);

            /** TaskLog HideOnceFinished */
            HideOnceFinished?: (boolean|null);
        }

        /** Represents a TaskLog. */
        class TaskLog implements ITaskLog {

            /**
             * Constructs a new TaskLog.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITaskLog);

            /** TaskLog uri. */
            public uri: string;

            /** TaskLog name. */
            public name: string;

            /** TaskLog messageFormat. */
            public messageFormat: flyteidl.core.TaskLog.MessageFormat;

            /** TaskLog ttl. */
            public ttl?: (google.protobuf.IDuration|null);

            /** TaskLog ShowWhilePending. */
            public ShowWhilePending: boolean;

            /** TaskLog HideOnceFinished. */
            public HideOnceFinished: boolean;

            /**
             * Creates a new TaskLog instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskLog instance
             */
            public static create(properties?: flyteidl.core.ITaskLog): flyteidl.core.TaskLog;

            /**
             * Encodes the specified TaskLog message. Does not implicitly {@link flyteidl.core.TaskLog.verify|verify} messages.
             * @param message TaskLog message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITaskLog, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskLog message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskLog
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TaskLog;

            /**
             * Verifies a TaskLog message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace TaskLog {

            /** MessageFormat enum. */
            enum MessageFormat {
                UNKNOWN = 0,
                CSV = 1,
                JSON = 2
            }
        }

        /** Properties of a QualityOfServiceSpec. */
        interface IQualityOfServiceSpec {

            /** QualityOfServiceSpec queueingBudget */
            queueingBudget?: (google.protobuf.IDuration|null);
        }

        /** Represents a QualityOfServiceSpec. */
        class QualityOfServiceSpec implements IQualityOfServiceSpec {

            /**
             * Constructs a new QualityOfServiceSpec.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IQualityOfServiceSpec);

            /** QualityOfServiceSpec queueingBudget. */
            public queueingBudget?: (google.protobuf.IDuration|null);

            /**
             * Creates a new QualityOfServiceSpec instance using the specified properties.
             * @param [properties] Properties to set
             * @returns QualityOfServiceSpec instance
             */
            public static create(properties?: flyteidl.core.IQualityOfServiceSpec): flyteidl.core.QualityOfServiceSpec;

            /**
             * Encodes the specified QualityOfServiceSpec message. Does not implicitly {@link flyteidl.core.QualityOfServiceSpec.verify|verify} messages.
             * @param message QualityOfServiceSpec message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IQualityOfServiceSpec, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a QualityOfServiceSpec message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns QualityOfServiceSpec
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.QualityOfServiceSpec;

            /**
             * Verifies a QualityOfServiceSpec message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a QualityOfService. */
        interface IQualityOfService {

            /** QualityOfService tier */
            tier?: (flyteidl.core.QualityOfService.Tier|null);

            /** QualityOfService spec */
            spec?: (flyteidl.core.IQualityOfServiceSpec|null);
        }

        /** Represents a QualityOfService. */
        class QualityOfService implements IQualityOfService {

            /**
             * Constructs a new QualityOfService.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IQualityOfService);

            /** QualityOfService tier. */
            public tier: flyteidl.core.QualityOfService.Tier;

            /** QualityOfService spec. */
            public spec?: (flyteidl.core.IQualityOfServiceSpec|null);

            /** QualityOfService designation. */
            public designation?: ("tier"|"spec");

            /**
             * Creates a new QualityOfService instance using the specified properties.
             * @param [properties] Properties to set
             * @returns QualityOfService instance
             */
            public static create(properties?: flyteidl.core.IQualityOfService): flyteidl.core.QualityOfService;

            /**
             * Encodes the specified QualityOfService message. Does not implicitly {@link flyteidl.core.QualityOfService.verify|verify} messages.
             * @param message QualityOfService message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IQualityOfService, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a QualityOfService message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns QualityOfService
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.QualityOfService;

            /**
             * Verifies a QualityOfService message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace QualityOfService {

            /** Tier enum. */
            enum Tier {
                UNDEFINED = 0,
                HIGH = 1,
                MEDIUM = 2,
                LOW = 3
            }
        }

        /** Properties of a Resources. */
        interface IResources {

            /** Resources requests */
            requests?: (flyteidl.core.Resources.IResourceEntry[]|null);

            /** Resources limits */
            limits?: (flyteidl.core.Resources.IResourceEntry[]|null);
        }

        /** Represents a Resources. */
        class Resources implements IResources {

            /**
             * Constructs a new Resources.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IResources);

            /** Resources requests. */
            public requests: flyteidl.core.Resources.IResourceEntry[];

            /** Resources limits. */
            public limits: flyteidl.core.Resources.IResourceEntry[];

            /**
             * Creates a new Resources instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Resources instance
             */
            public static create(properties?: flyteidl.core.IResources): flyteidl.core.Resources;

            /**
             * Encodes the specified Resources message. Does not implicitly {@link flyteidl.core.Resources.verify|verify} messages.
             * @param message Resources message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IResources, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Resources message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Resources
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Resources;

            /**
             * Verifies a Resources message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace Resources {

            /** ResourceName enum. */
            enum ResourceName {
                UNKNOWN = 0,
                CPU = 1,
                GPU = 2,
                MEMORY = 3,
                STORAGE = 4,
                EPHEMERAL_STORAGE = 5
            }

            /** Properties of a ResourceEntry. */
            interface IResourceEntry {

                /** ResourceEntry name */
                name?: (flyteidl.core.Resources.ResourceName|null);

                /** ResourceEntry value */
                value?: (string|null);
            }

            /** Represents a ResourceEntry. */
            class ResourceEntry implements IResourceEntry {

                /**
                 * Constructs a new ResourceEntry.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: flyteidl.core.Resources.IResourceEntry);

                /** ResourceEntry name. */
                public name: flyteidl.core.Resources.ResourceName;

                /** ResourceEntry value. */
                public value: string;

                /**
                 * Creates a new ResourceEntry instance using the specified properties.
                 * @param [properties] Properties to set
                 * @returns ResourceEntry instance
                 */
                public static create(properties?: flyteidl.core.Resources.IResourceEntry): flyteidl.core.Resources.ResourceEntry;

                /**
                 * Encodes the specified ResourceEntry message. Does not implicitly {@link flyteidl.core.Resources.ResourceEntry.verify|verify} messages.
                 * @param message ResourceEntry message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: flyteidl.core.Resources.IResourceEntry, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes a ResourceEntry message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns ResourceEntry
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Resources.ResourceEntry;

                /**
                 * Verifies a ResourceEntry message.
                 * @param message Plain object to verify
                 * @returns `null` if valid, otherwise the reason why it is not
                 */
                public static verify(message: { [k: string]: any }): (string|null);
            }
        }

        /** Properties of a GPUAccelerator. */
        interface IGPUAccelerator {

            /** GPUAccelerator device */
            device?: (string|null);

            /** GPUAccelerator unpartitioned */
            unpartitioned?: (boolean|null);

            /** GPUAccelerator partitionSize */
            partitionSize?: (string|null);
        }

        /** Represents a GPUAccelerator. */
        class GPUAccelerator implements IGPUAccelerator {

            /**
             * Constructs a new GPUAccelerator.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IGPUAccelerator);

            /** GPUAccelerator device. */
            public device: string;

            /** GPUAccelerator unpartitioned. */
            public unpartitioned: boolean;

            /** GPUAccelerator partitionSize. */
            public partitionSize: string;

            /** GPUAccelerator partitionSizeValue. */
            public partitionSizeValue?: ("unpartitioned"|"partitionSize");

            /**
             * Creates a new GPUAccelerator instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GPUAccelerator instance
             */
            public static create(properties?: flyteidl.core.IGPUAccelerator): flyteidl.core.GPUAccelerator;

            /**
             * Encodes the specified GPUAccelerator message. Does not implicitly {@link flyteidl.core.GPUAccelerator.verify|verify} messages.
             * @param message GPUAccelerator message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IGPUAccelerator, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GPUAccelerator message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GPUAccelerator
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.GPUAccelerator;

            /**
             * Verifies a GPUAccelerator message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExtendedResources. */
        interface IExtendedResources {

            /** ExtendedResources gpuAccelerator */
            gpuAccelerator?: (flyteidl.core.IGPUAccelerator|null);
        }

        /** Represents an ExtendedResources. */
        class ExtendedResources implements IExtendedResources {

            /**
             * Constructs a new ExtendedResources.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IExtendedResources);

            /** ExtendedResources gpuAccelerator. */
            public gpuAccelerator?: (flyteidl.core.IGPUAccelerator|null);

            /**
             * Creates a new ExtendedResources instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExtendedResources instance
             */
            public static create(properties?: flyteidl.core.IExtendedResources): flyteidl.core.ExtendedResources;

            /**
             * Encodes the specified ExtendedResources message. Does not implicitly {@link flyteidl.core.ExtendedResources.verify|verify} messages.
             * @param message ExtendedResources message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IExtendedResources, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExtendedResources message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExtendedResources
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ExtendedResources;

            /**
             * Verifies an ExtendedResources message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a RuntimeMetadata. */
        interface IRuntimeMetadata {

            /** RuntimeMetadata type */
            type?: (flyteidl.core.RuntimeMetadata.RuntimeType|null);

            /** RuntimeMetadata version */
            version?: (string|null);

            /** RuntimeMetadata flavor */
            flavor?: (string|null);
        }

        /** Represents a RuntimeMetadata. */
        class RuntimeMetadata implements IRuntimeMetadata {

            /**
             * Constructs a new RuntimeMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IRuntimeMetadata);

            /** RuntimeMetadata type. */
            public type: flyteidl.core.RuntimeMetadata.RuntimeType;

            /** RuntimeMetadata version. */
            public version: string;

            /** RuntimeMetadata flavor. */
            public flavor: string;

            /**
             * Creates a new RuntimeMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns RuntimeMetadata instance
             */
            public static create(properties?: flyteidl.core.IRuntimeMetadata): flyteidl.core.RuntimeMetadata;

            /**
             * Encodes the specified RuntimeMetadata message. Does not implicitly {@link flyteidl.core.RuntimeMetadata.verify|verify} messages.
             * @param message RuntimeMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IRuntimeMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a RuntimeMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns RuntimeMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.RuntimeMetadata;

            /**
             * Verifies a RuntimeMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace RuntimeMetadata {

            /** RuntimeType enum. */
            enum RuntimeType {
                OTHER = 0,
                FLYTE_SDK = 1
            }
        }

        /** Properties of a TaskMetadata. */
        interface ITaskMetadata {

            /** TaskMetadata discoverable */
            discoverable?: (boolean|null);

            /** TaskMetadata runtime */
            runtime?: (flyteidl.core.IRuntimeMetadata|null);

            /** TaskMetadata timeout */
            timeout?: (google.protobuf.IDuration|null);

            /** TaskMetadata retries */
            retries?: (flyteidl.core.IRetryStrategy|null);

            /** TaskMetadata discoveryVersion */
            discoveryVersion?: (string|null);

            /** TaskMetadata deprecatedErrorMessage */
            deprecatedErrorMessage?: (string|null);

            /** TaskMetadata interruptible */
            interruptible?: (boolean|null);

            /** TaskMetadata cacheSerializable */
            cacheSerializable?: (boolean|null);

            /** TaskMetadata generatesDeck */
            generatesDeck?: (boolean|null);

            /** TaskMetadata tags */
            tags?: ({ [k: string]: string }|null);

            /** TaskMetadata podTemplateName */
            podTemplateName?: (string|null);

            /** TaskMetadata cacheIgnoreInputVars */
            cacheIgnoreInputVars?: (string[]|null);
        }

        /** Represents a TaskMetadata. */
        class TaskMetadata implements ITaskMetadata {

            /**
             * Constructs a new TaskMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITaskMetadata);

            /** TaskMetadata discoverable. */
            public discoverable: boolean;

            /** TaskMetadata runtime. */
            public runtime?: (flyteidl.core.IRuntimeMetadata|null);

            /** TaskMetadata timeout. */
            public timeout?: (google.protobuf.IDuration|null);

            /** TaskMetadata retries. */
            public retries?: (flyteidl.core.IRetryStrategy|null);

            /** TaskMetadata discoveryVersion. */
            public discoveryVersion: string;

            /** TaskMetadata deprecatedErrorMessage. */
            public deprecatedErrorMessage: string;

            /** TaskMetadata interruptible. */
            public interruptible: boolean;

            /** TaskMetadata cacheSerializable. */
            public cacheSerializable: boolean;

            /** TaskMetadata generatesDeck. */
            public generatesDeck: boolean;

            /** TaskMetadata tags. */
            public tags: { [k: string]: string };

            /** TaskMetadata podTemplateName. */
            public podTemplateName: string;

            /** TaskMetadata cacheIgnoreInputVars. */
            public cacheIgnoreInputVars: string[];

            /** TaskMetadata interruptibleValue. */
            public interruptibleValue?: "interruptible";

            /**
             * Creates a new TaskMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskMetadata instance
             */
            public static create(properties?: flyteidl.core.ITaskMetadata): flyteidl.core.TaskMetadata;

            /**
             * Encodes the specified TaskMetadata message. Does not implicitly {@link flyteidl.core.TaskMetadata.verify|verify} messages.
             * @param message TaskMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITaskMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TaskMetadata;

            /**
             * Verifies a TaskMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskTemplate. */
        interface ITaskTemplate {

            /** TaskTemplate id */
            id?: (flyteidl.core.IIdentifier|null);

            /** TaskTemplate type */
            type?: (string|null);

            /** TaskTemplate metadata */
            metadata?: (flyteidl.core.ITaskMetadata|null);

            /** TaskTemplate interface */
            "interface"?: (flyteidl.core.ITypedInterface|null);

            /** TaskTemplate custom */
            custom?: (google.protobuf.IStruct|null);

            /** TaskTemplate container */
            container?: (flyteidl.core.IContainer|null);

            /** TaskTemplate k8sPod */
            k8sPod?: (flyteidl.core.IK8sPod|null);

            /** TaskTemplate sql */
            sql?: (flyteidl.core.ISql|null);

            /** TaskTemplate taskTypeVersion */
            taskTypeVersion?: (number|null);

            /** TaskTemplate securityContext */
            securityContext?: (flyteidl.core.ISecurityContext|null);

            /** TaskTemplate extendedResources */
            extendedResources?: (flyteidl.core.IExtendedResources|null);

            /** TaskTemplate config */
            config?: ({ [k: string]: string }|null);
        }

        /** Represents a TaskTemplate. */
        class TaskTemplate implements ITaskTemplate {

            /**
             * Constructs a new TaskTemplate.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ITaskTemplate);

            /** TaskTemplate id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** TaskTemplate type. */
            public type: string;

            /** TaskTemplate metadata. */
            public metadata?: (flyteidl.core.ITaskMetadata|null);

            /** TaskTemplate interface. */
            public interface?: (flyteidl.core.ITypedInterface|null);

            /** TaskTemplate custom. */
            public custom?: (google.protobuf.IStruct|null);

            /** TaskTemplate container. */
            public container?: (flyteidl.core.IContainer|null);

            /** TaskTemplate k8sPod. */
            public k8sPod?: (flyteidl.core.IK8sPod|null);

            /** TaskTemplate sql. */
            public sql?: (flyteidl.core.ISql|null);

            /** TaskTemplate taskTypeVersion. */
            public taskTypeVersion: number;

            /** TaskTemplate securityContext. */
            public securityContext?: (flyteidl.core.ISecurityContext|null);

            /** TaskTemplate extendedResources. */
            public extendedResources?: (flyteidl.core.IExtendedResources|null);

            /** TaskTemplate config. */
            public config: { [k: string]: string };

            /** TaskTemplate target. */
            public target?: ("container"|"k8sPod"|"sql");

            /**
             * Creates a new TaskTemplate instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskTemplate instance
             */
            public static create(properties?: flyteidl.core.ITaskTemplate): flyteidl.core.TaskTemplate;

            /**
             * Encodes the specified TaskTemplate message. Does not implicitly {@link flyteidl.core.TaskTemplate.verify|verify} messages.
             * @param message TaskTemplate message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ITaskTemplate, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskTemplate message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskTemplate
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.TaskTemplate;

            /**
             * Verifies a TaskTemplate message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ContainerPort. */
        interface IContainerPort {

            /** ContainerPort containerPort */
            containerPort?: (number|null);
        }

        /** Represents a ContainerPort. */
        class ContainerPort implements IContainerPort {

            /**
             * Constructs a new ContainerPort.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IContainerPort);

            /** ContainerPort containerPort. */
            public containerPort: number;

            /**
             * Creates a new ContainerPort instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ContainerPort instance
             */
            public static create(properties?: flyteidl.core.IContainerPort): flyteidl.core.ContainerPort;

            /**
             * Encodes the specified ContainerPort message. Does not implicitly {@link flyteidl.core.ContainerPort.verify|verify} messages.
             * @param message ContainerPort message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IContainerPort, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ContainerPort message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ContainerPort
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ContainerPort;

            /**
             * Verifies a ContainerPort message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Container. */
        interface IContainer {

            /** Container image */
            image?: (string|null);

            /** Container command */
            command?: (string[]|null);

            /** Container args */
            args?: (string[]|null);

            /** Container resources */
            resources?: (flyteidl.core.IResources|null);

            /** Container env */
            env?: (flyteidl.core.IKeyValuePair[]|null);

            /** Container config */
            config?: (flyteidl.core.IKeyValuePair[]|null);

            /** Container ports */
            ports?: (flyteidl.core.IContainerPort[]|null);

            /** Container dataConfig */
            dataConfig?: (flyteidl.core.IDataLoadingConfig|null);

            /** Container architecture */
            architecture?: (flyteidl.core.Container.Architecture|null);
        }

        /** Represents a Container. */
        class Container implements IContainer {

            /**
             * Constructs a new Container.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IContainer);

            /** Container image. */
            public image: string;

            /** Container command. */
            public command: string[];

            /** Container args. */
            public args: string[];

            /** Container resources. */
            public resources?: (flyteidl.core.IResources|null);

            /** Container env. */
            public env: flyteidl.core.IKeyValuePair[];

            /** Container config. */
            public config: flyteidl.core.IKeyValuePair[];

            /** Container ports. */
            public ports: flyteidl.core.IContainerPort[];

            /** Container dataConfig. */
            public dataConfig?: (flyteidl.core.IDataLoadingConfig|null);

            /** Container architecture. */
            public architecture: flyteidl.core.Container.Architecture;

            /**
             * Creates a new Container instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Container instance
             */
            public static create(properties?: flyteidl.core.IContainer): flyteidl.core.Container;

            /**
             * Encodes the specified Container message. Does not implicitly {@link flyteidl.core.Container.verify|verify} messages.
             * @param message Container message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IContainer, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Container message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Container
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Container;

            /**
             * Verifies a Container message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace Container {

            /** Architecture enum. */
            enum Architecture {
                UNKNOWN = 0,
                AMD64 = 1,
                ARM64 = 2,
                ARM_V6 = 3,
                ARM_V7 = 4
            }
        }

        /** Properties of a IOStrategy. */
        interface IIOStrategy {

            /** IOStrategy downloadMode */
            downloadMode?: (flyteidl.core.IOStrategy.DownloadMode|null);

            /** IOStrategy uploadMode */
            uploadMode?: (flyteidl.core.IOStrategy.UploadMode|null);
        }

        /** Represents a IOStrategy. */
        class IOStrategy implements IIOStrategy {

            /**
             * Constructs a new IOStrategy.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IIOStrategy);

            /** IOStrategy downloadMode. */
            public downloadMode: flyteidl.core.IOStrategy.DownloadMode;

            /** IOStrategy uploadMode. */
            public uploadMode: flyteidl.core.IOStrategy.UploadMode;

            /**
             * Creates a new IOStrategy instance using the specified properties.
             * @param [properties] Properties to set
             * @returns IOStrategy instance
             */
            public static create(properties?: flyteidl.core.IIOStrategy): flyteidl.core.IOStrategy;

            /**
             * Encodes the specified IOStrategy message. Does not implicitly {@link flyteidl.core.IOStrategy.verify|verify} messages.
             * @param message IOStrategy message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IIOStrategy, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a IOStrategy message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns IOStrategy
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.IOStrategy;

            /**
             * Verifies a IOStrategy message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace IOStrategy {

            /** DownloadMode enum. */
            enum DownloadMode {
                DOWNLOAD_EAGER = 0,
                DOWNLOAD_STREAM = 1,
                DO_NOT_DOWNLOAD = 2
            }

            /** UploadMode enum. */
            enum UploadMode {
                UPLOAD_ON_EXIT = 0,
                UPLOAD_EAGER = 1,
                DO_NOT_UPLOAD = 2
            }
        }

        /** Properties of a DataLoadingConfig. */
        interface IDataLoadingConfig {

            /** DataLoadingConfig enabled */
            enabled?: (boolean|null);

            /** DataLoadingConfig inputPath */
            inputPath?: (string|null);

            /** DataLoadingConfig outputPath */
            outputPath?: (string|null);

            /** DataLoadingConfig format */
            format?: (flyteidl.core.DataLoadingConfig.LiteralMapFormat|null);

            /** DataLoadingConfig ioStrategy */
            ioStrategy?: (flyteidl.core.IIOStrategy|null);
        }

        /** Represents a DataLoadingConfig. */
        class DataLoadingConfig implements IDataLoadingConfig {

            /**
             * Constructs a new DataLoadingConfig.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IDataLoadingConfig);

            /** DataLoadingConfig enabled. */
            public enabled: boolean;

            /** DataLoadingConfig inputPath. */
            public inputPath: string;

            /** DataLoadingConfig outputPath. */
            public outputPath: string;

            /** DataLoadingConfig format. */
            public format: flyteidl.core.DataLoadingConfig.LiteralMapFormat;

            /** DataLoadingConfig ioStrategy. */
            public ioStrategy?: (flyteidl.core.IIOStrategy|null);

            /**
             * Creates a new DataLoadingConfig instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DataLoadingConfig instance
             */
            public static create(properties?: flyteidl.core.IDataLoadingConfig): flyteidl.core.DataLoadingConfig;

            /**
             * Encodes the specified DataLoadingConfig message. Does not implicitly {@link flyteidl.core.DataLoadingConfig.verify|verify} messages.
             * @param message DataLoadingConfig message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IDataLoadingConfig, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DataLoadingConfig message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DataLoadingConfig
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.DataLoadingConfig;

            /**
             * Verifies a DataLoadingConfig message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace DataLoadingConfig {

            /** LiteralMapFormat enum. */
            enum LiteralMapFormat {
                JSON = 0,
                YAML = 1,
                PROTO = 2
            }
        }

        /** Properties of a K8sPod. */
        interface IK8sPod {

            /** K8sPod metadata */
            metadata?: (flyteidl.core.IK8sObjectMetadata|null);

            /** K8sPod podSpec */
            podSpec?: (google.protobuf.IStruct|null);

            /** K8sPod dataConfig */
            dataConfig?: (flyteidl.core.IDataLoadingConfig|null);
        }

        /** Represents a K8sPod. */
        class K8sPod implements IK8sPod {

            /**
             * Constructs a new K8sPod.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IK8sPod);

            /** K8sPod metadata. */
            public metadata?: (flyteidl.core.IK8sObjectMetadata|null);

            /** K8sPod podSpec. */
            public podSpec?: (google.protobuf.IStruct|null);

            /** K8sPod dataConfig. */
            public dataConfig?: (flyteidl.core.IDataLoadingConfig|null);

            /**
             * Creates a new K8sPod instance using the specified properties.
             * @param [properties] Properties to set
             * @returns K8sPod instance
             */
            public static create(properties?: flyteidl.core.IK8sPod): flyteidl.core.K8sPod;

            /**
             * Encodes the specified K8sPod message. Does not implicitly {@link flyteidl.core.K8sPod.verify|verify} messages.
             * @param message K8sPod message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IK8sPod, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a K8sPod message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns K8sPod
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.K8sPod;

            /**
             * Verifies a K8sPod message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a K8sObjectMetadata. */
        interface IK8sObjectMetadata {

            /** K8sObjectMetadata labels */
            labels?: ({ [k: string]: string }|null);

            /** K8sObjectMetadata annotations */
            annotations?: ({ [k: string]: string }|null);
        }

        /** Represents a K8sObjectMetadata. */
        class K8sObjectMetadata implements IK8sObjectMetadata {

            /**
             * Constructs a new K8sObjectMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IK8sObjectMetadata);

            /** K8sObjectMetadata labels. */
            public labels: { [k: string]: string };

            /** K8sObjectMetadata annotations. */
            public annotations: { [k: string]: string };

            /**
             * Creates a new K8sObjectMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns K8sObjectMetadata instance
             */
            public static create(properties?: flyteidl.core.IK8sObjectMetadata): flyteidl.core.K8sObjectMetadata;

            /**
             * Encodes the specified K8sObjectMetadata message. Does not implicitly {@link flyteidl.core.K8sObjectMetadata.verify|verify} messages.
             * @param message K8sObjectMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IK8sObjectMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a K8sObjectMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns K8sObjectMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.K8sObjectMetadata;

            /**
             * Verifies a K8sObjectMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Sql. */
        interface ISql {

            /** Sql statement */
            statement?: (string|null);

            /** Sql dialect */
            dialect?: (flyteidl.core.Sql.Dialect|null);
        }

        /** Represents a Sql. */
        class Sql implements ISql {

            /**
             * Constructs a new Sql.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISql);

            /** Sql statement. */
            public statement: string;

            /** Sql dialect. */
            public dialect: flyteidl.core.Sql.Dialect;

            /**
             * Creates a new Sql instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Sql instance
             */
            public static create(properties?: flyteidl.core.ISql): flyteidl.core.Sql;

            /**
             * Encodes the specified Sql message. Does not implicitly {@link flyteidl.core.Sql.verify|verify} messages.
             * @param message Sql message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISql, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Sql message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Sql
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Sql;

            /**
             * Verifies a Sql message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace Sql {

            /** Dialect enum. */
            enum Dialect {
                UNDEFINED = 0,
                ANSI = 1,
                HIVE = 2,
                OTHER = 3
            }
        }

        /** Properties of a Secret. */
        interface ISecret {

            /** Secret group */
            group?: (string|null);

            /** Secret groupVersion */
            groupVersion?: (string|null);

            /** Secret key */
            key?: (string|null);

            /** Secret mountRequirement */
            mountRequirement?: (flyteidl.core.Secret.MountType|null);
        }

        /** Represents a Secret. */
        class Secret implements ISecret {

            /**
             * Constructs a new Secret.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISecret);

            /** Secret group. */
            public group: string;

            /** Secret groupVersion. */
            public groupVersion: string;

            /** Secret key. */
            public key: string;

            /** Secret mountRequirement. */
            public mountRequirement: flyteidl.core.Secret.MountType;

            /**
             * Creates a new Secret instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Secret instance
             */
            public static create(properties?: flyteidl.core.ISecret): flyteidl.core.Secret;

            /**
             * Encodes the specified Secret message. Does not implicitly {@link flyteidl.core.Secret.verify|verify} messages.
             * @param message Secret message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISecret, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Secret message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Secret
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Secret;

            /**
             * Verifies a Secret message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace Secret {

            /** MountType enum. */
            enum MountType {
                ANY = 0,
                ENV_VAR = 1,
                FILE = 2
            }
        }

        /** Properties of a OAuth2Client. */
        interface IOAuth2Client {

            /** OAuth2Client clientId */
            clientId?: (string|null);

            /** OAuth2Client clientSecret */
            clientSecret?: (flyteidl.core.ISecret|null);
        }

        /** Represents a OAuth2Client. */
        class OAuth2Client implements IOAuth2Client {

            /**
             * Constructs a new OAuth2Client.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IOAuth2Client);

            /** OAuth2Client clientId. */
            public clientId: string;

            /** OAuth2Client clientSecret. */
            public clientSecret?: (flyteidl.core.ISecret|null);

            /**
             * Creates a new OAuth2Client instance using the specified properties.
             * @param [properties] Properties to set
             * @returns OAuth2Client instance
             */
            public static create(properties?: flyteidl.core.IOAuth2Client): flyteidl.core.OAuth2Client;

            /**
             * Encodes the specified OAuth2Client message. Does not implicitly {@link flyteidl.core.OAuth2Client.verify|verify} messages.
             * @param message OAuth2Client message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IOAuth2Client, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a OAuth2Client message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns OAuth2Client
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.OAuth2Client;

            /**
             * Verifies a OAuth2Client message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Identity. */
        interface IIdentity {

            /** Identity iamRole */
            iamRole?: (string|null);

            /** Identity k8sServiceAccount */
            k8sServiceAccount?: (string|null);

            /** Identity oauth2Client */
            oauth2Client?: (flyteidl.core.IOAuth2Client|null);

            /** Identity executionIdentity */
            executionIdentity?: (string|null);
        }

        /** Represents an Identity. */
        class Identity implements IIdentity {

            /**
             * Constructs a new Identity.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IIdentity);

            /** Identity iamRole. */
            public iamRole: string;

            /** Identity k8sServiceAccount. */
            public k8sServiceAccount: string;

            /** Identity oauth2Client. */
            public oauth2Client?: (flyteidl.core.IOAuth2Client|null);

            /** Identity executionIdentity. */
            public executionIdentity: string;

            /**
             * Creates a new Identity instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Identity instance
             */
            public static create(properties?: flyteidl.core.IIdentity): flyteidl.core.Identity;

            /**
             * Encodes the specified Identity message. Does not implicitly {@link flyteidl.core.Identity.verify|verify} messages.
             * @param message Identity message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IIdentity, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Identity message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Identity
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Identity;

            /**
             * Verifies an Identity message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a OAuth2TokenRequest. */
        interface IOAuth2TokenRequest {

            /** OAuth2TokenRequest name */
            name?: (string|null);

            /** OAuth2TokenRequest type */
            type?: (flyteidl.core.OAuth2TokenRequest.Type|null);

            /** OAuth2TokenRequest client */
            client?: (flyteidl.core.IOAuth2Client|null);

            /** OAuth2TokenRequest idpDiscoveryEndpoint */
            idpDiscoveryEndpoint?: (string|null);

            /** OAuth2TokenRequest tokenEndpoint */
            tokenEndpoint?: (string|null);
        }

        /** Represents a OAuth2TokenRequest. */
        class OAuth2TokenRequest implements IOAuth2TokenRequest {

            /**
             * Constructs a new OAuth2TokenRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IOAuth2TokenRequest);

            /** OAuth2TokenRequest name. */
            public name: string;

            /** OAuth2TokenRequest type. */
            public type: flyteidl.core.OAuth2TokenRequest.Type;

            /** OAuth2TokenRequest client. */
            public client?: (flyteidl.core.IOAuth2Client|null);

            /** OAuth2TokenRequest idpDiscoveryEndpoint. */
            public idpDiscoveryEndpoint: string;

            /** OAuth2TokenRequest tokenEndpoint. */
            public tokenEndpoint: string;

            /**
             * Creates a new OAuth2TokenRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns OAuth2TokenRequest instance
             */
            public static create(properties?: flyteidl.core.IOAuth2TokenRequest): flyteidl.core.OAuth2TokenRequest;

            /**
             * Encodes the specified OAuth2TokenRequest message. Does not implicitly {@link flyteidl.core.OAuth2TokenRequest.verify|verify} messages.
             * @param message OAuth2TokenRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IOAuth2TokenRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a OAuth2TokenRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns OAuth2TokenRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.OAuth2TokenRequest;

            /**
             * Verifies a OAuth2TokenRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace OAuth2TokenRequest {

            /** Type enum. */
            enum Type {
                CLIENT_CREDENTIALS = 0
            }
        }

        /** Properties of a SecurityContext. */
        interface ISecurityContext {

            /** SecurityContext runAs */
            runAs?: (flyteidl.core.IIdentity|null);

            /** SecurityContext secrets */
            secrets?: (flyteidl.core.ISecret[]|null);

            /** SecurityContext tokens */
            tokens?: (flyteidl.core.IOAuth2TokenRequest[]|null);
        }

        /** Represents a SecurityContext. */
        class SecurityContext implements ISecurityContext {

            /**
             * Constructs a new SecurityContext.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISecurityContext);

            /** SecurityContext runAs. */
            public runAs?: (flyteidl.core.IIdentity|null);

            /** SecurityContext secrets. */
            public secrets: flyteidl.core.ISecret[];

            /** SecurityContext tokens. */
            public tokens: flyteidl.core.IOAuth2TokenRequest[];

            /**
             * Creates a new SecurityContext instance using the specified properties.
             * @param [properties] Properties to set
             * @returns SecurityContext instance
             */
            public static create(properties?: flyteidl.core.ISecurityContext): flyteidl.core.SecurityContext;

            /**
             * Encodes the specified SecurityContext message. Does not implicitly {@link flyteidl.core.SecurityContext.verify|verify} messages.
             * @param message SecurityContext message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISecurityContext, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SecurityContext message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SecurityContext
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.SecurityContext;

            /**
             * Verifies a SecurityContext message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DynamicJobSpec. */
        interface IDynamicJobSpec {

            /** DynamicJobSpec nodes */
            nodes?: (flyteidl.core.INode[]|null);

            /** DynamicJobSpec minSuccesses */
            minSuccesses?: (Long|null);

            /** DynamicJobSpec outputs */
            outputs?: (flyteidl.core.IBinding[]|null);

            /** DynamicJobSpec tasks */
            tasks?: (flyteidl.core.ITaskTemplate[]|null);

            /** DynamicJobSpec subworkflows */
            subworkflows?: (flyteidl.core.IWorkflowTemplate[]|null);
        }

        /** Represents a DynamicJobSpec. */
        class DynamicJobSpec implements IDynamicJobSpec {

            /**
             * Constructs a new DynamicJobSpec.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IDynamicJobSpec);

            /** DynamicJobSpec nodes. */
            public nodes: flyteidl.core.INode[];

            /** DynamicJobSpec minSuccesses. */
            public minSuccesses: Long;

            /** DynamicJobSpec outputs. */
            public outputs: flyteidl.core.IBinding[];

            /** DynamicJobSpec tasks. */
            public tasks: flyteidl.core.ITaskTemplate[];

            /** DynamicJobSpec subworkflows. */
            public subworkflows: flyteidl.core.IWorkflowTemplate[];

            /**
             * Creates a new DynamicJobSpec instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DynamicJobSpec instance
             */
            public static create(properties?: flyteidl.core.IDynamicJobSpec): flyteidl.core.DynamicJobSpec;

            /**
             * Encodes the specified DynamicJobSpec message. Does not implicitly {@link flyteidl.core.DynamicJobSpec.verify|verify} messages.
             * @param message DynamicJobSpec message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IDynamicJobSpec, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DynamicJobSpec message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DynamicJobSpec
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.DynamicJobSpec;

            /**
             * Verifies a DynamicJobSpec message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ContainerError. */
        interface IContainerError {

            /** ContainerError code */
            code?: (string|null);

            /** ContainerError message */
            message?: (string|null);

            /** ContainerError kind */
            kind?: (flyteidl.core.ContainerError.Kind|null);

            /** ContainerError origin */
            origin?: (flyteidl.core.ExecutionError.ErrorKind|null);
        }

        /** Represents a ContainerError. */
        class ContainerError implements IContainerError {

            /**
             * Constructs a new ContainerError.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IContainerError);

            /** ContainerError code. */
            public code: string;

            /** ContainerError message. */
            public message: string;

            /** ContainerError kind. */
            public kind: flyteidl.core.ContainerError.Kind;

            /** ContainerError origin. */
            public origin: flyteidl.core.ExecutionError.ErrorKind;

            /**
             * Creates a new ContainerError instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ContainerError instance
             */
            public static create(properties?: flyteidl.core.IContainerError): flyteidl.core.ContainerError;

            /**
             * Encodes the specified ContainerError message. Does not implicitly {@link flyteidl.core.ContainerError.verify|verify} messages.
             * @param message ContainerError message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IContainerError, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ContainerError message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ContainerError
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ContainerError;

            /**
             * Verifies a ContainerError message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace ContainerError {

            /** Kind enum. */
            enum Kind {
                NON_RECOVERABLE = 0,
                RECOVERABLE = 1
            }
        }

        /** Properties of an ErrorDocument. */
        interface IErrorDocument {

            /** ErrorDocument error */
            error?: (flyteidl.core.IContainerError|null);
        }

        /** Represents an ErrorDocument. */
        class ErrorDocument implements IErrorDocument {

            /**
             * Constructs a new ErrorDocument.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IErrorDocument);

            /** ErrorDocument error. */
            public error?: (flyteidl.core.IContainerError|null);

            /**
             * Creates a new ErrorDocument instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ErrorDocument instance
             */
            public static create(properties?: flyteidl.core.IErrorDocument): flyteidl.core.ErrorDocument;

            /**
             * Encodes the specified ErrorDocument message. Does not implicitly {@link flyteidl.core.ErrorDocument.verify|verify} messages.
             * @param message ErrorDocument message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IErrorDocument, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ErrorDocument message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ErrorDocument
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ErrorDocument;

            /**
             * Verifies an ErrorDocument message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionEnvAssignment. */
        interface IExecutionEnvAssignment {

            /** ExecutionEnvAssignment nodeIds */
            nodeIds?: (string[]|null);

            /** ExecutionEnvAssignment taskType */
            taskType?: (string|null);

            /** ExecutionEnvAssignment executionEnv */
            executionEnv?: (flyteidl.core.IExecutionEnv|null);
        }

        /** Represents an ExecutionEnvAssignment. */
        class ExecutionEnvAssignment implements IExecutionEnvAssignment {

            /**
             * Constructs a new ExecutionEnvAssignment.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IExecutionEnvAssignment);

            /** ExecutionEnvAssignment nodeIds. */
            public nodeIds: string[];

            /** ExecutionEnvAssignment taskType. */
            public taskType: string;

            /** ExecutionEnvAssignment executionEnv. */
            public executionEnv?: (flyteidl.core.IExecutionEnv|null);

            /**
             * Creates a new ExecutionEnvAssignment instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionEnvAssignment instance
             */
            public static create(properties?: flyteidl.core.IExecutionEnvAssignment): flyteidl.core.ExecutionEnvAssignment;

            /**
             * Encodes the specified ExecutionEnvAssignment message. Does not implicitly {@link flyteidl.core.ExecutionEnvAssignment.verify|verify} messages.
             * @param message ExecutionEnvAssignment message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IExecutionEnvAssignment, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionEnvAssignment message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionEnvAssignment
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ExecutionEnvAssignment;

            /**
             * Verifies an ExecutionEnvAssignment message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionEnv. */
        interface IExecutionEnv {

            /** ExecutionEnv id */
            id?: (string|null);

            /** ExecutionEnv type */
            type?: (string|null);

            /** ExecutionEnv extant */
            extant?: (google.protobuf.IStruct|null);

            /** ExecutionEnv spec */
            spec?: (google.protobuf.IStruct|null);
        }

        /** Represents an ExecutionEnv. */
        class ExecutionEnv implements IExecutionEnv {

            /**
             * Constructs a new ExecutionEnv.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IExecutionEnv);

            /** ExecutionEnv id. */
            public id: string;

            /** ExecutionEnv type. */
            public type: string;

            /** ExecutionEnv extant. */
            public extant?: (google.protobuf.IStruct|null);

            /** ExecutionEnv spec. */
            public spec?: (google.protobuf.IStruct|null);

            /** ExecutionEnv environment. */
            public environment?: ("extant"|"spec");

            /**
             * Creates a new ExecutionEnv instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionEnv instance
             */
            public static create(properties?: flyteidl.core.IExecutionEnv): flyteidl.core.ExecutionEnv;

            /**
             * Encodes the specified ExecutionEnv message. Does not implicitly {@link flyteidl.core.ExecutionEnv.verify|verify} messages.
             * @param message ExecutionEnv message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IExecutionEnv, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionEnv message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionEnv
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ExecutionEnv;

            /**
             * Verifies an ExecutionEnv message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Span. */
        interface ISpan {

            /** Span startTime */
            startTime?: (google.protobuf.ITimestamp|null);

            /** Span endTime */
            endTime?: (google.protobuf.ITimestamp|null);

            /** Span workflowId */
            workflowId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** Span nodeId */
            nodeId?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** Span taskId */
            taskId?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** Span operationId */
            operationId?: (string|null);

            /** Span spans */
            spans?: (flyteidl.core.ISpan[]|null);
        }

        /** Represents a Span. */
        class Span implements ISpan {

            /**
             * Constructs a new Span.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.ISpan);

            /** Span startTime. */
            public startTime?: (google.protobuf.ITimestamp|null);

            /** Span endTime. */
            public endTime?: (google.protobuf.ITimestamp|null);

            /** Span workflowId. */
            public workflowId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** Span nodeId. */
            public nodeId?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** Span taskId. */
            public taskId?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** Span operationId. */
            public operationId: string;

            /** Span spans. */
            public spans: flyteidl.core.ISpan[];

            /** Span id. */
            public id?: ("workflowId"|"nodeId"|"taskId"|"operationId");

            /**
             * Creates a new Span instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Span instance
             */
            public static create(properties?: flyteidl.core.ISpan): flyteidl.core.Span;

            /**
             * Encodes the specified Span message. Does not implicitly {@link flyteidl.core.Span.verify|verify} messages.
             * @param message Span message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.ISpan, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Span message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Span
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.Span;

            /**
             * Verifies a Span message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionMetricResult. */
        interface IExecutionMetricResult {

            /** ExecutionMetricResult metric */
            metric?: (string|null);

            /** ExecutionMetricResult data */
            data?: (google.protobuf.IStruct|null);
        }

        /** Represents an ExecutionMetricResult. */
        class ExecutionMetricResult implements IExecutionMetricResult {

            /**
             * Constructs a new ExecutionMetricResult.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IExecutionMetricResult);

            /** ExecutionMetricResult metric. */
            public metric: string;

            /** ExecutionMetricResult data. */
            public data?: (google.protobuf.IStruct|null);

            /**
             * Creates a new ExecutionMetricResult instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionMetricResult instance
             */
            public static create(properties?: flyteidl.core.IExecutionMetricResult): flyteidl.core.ExecutionMetricResult;

            /**
             * Encodes the specified ExecutionMetricResult message. Does not implicitly {@link flyteidl.core.ExecutionMetricResult.verify|verify} messages.
             * @param message ExecutionMetricResult message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IExecutionMetricResult, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionMetricResult message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionMetricResult
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.ExecutionMetricResult;

            /**
             * Verifies an ExecutionMetricResult message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowClosure. */
        interface IWorkflowClosure {

            /** WorkflowClosure workflow */
            workflow?: (flyteidl.core.IWorkflowTemplate|null);

            /** WorkflowClosure tasks */
            tasks?: (flyteidl.core.ITaskTemplate[]|null);
        }

        /** Represents a WorkflowClosure. */
        class WorkflowClosure implements IWorkflowClosure {

            /**
             * Constructs a new WorkflowClosure.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.core.IWorkflowClosure);

            /** WorkflowClosure workflow. */
            public workflow?: (flyteidl.core.IWorkflowTemplate|null);

            /** WorkflowClosure tasks. */
            public tasks: flyteidl.core.ITaskTemplate[];

            /**
             * Creates a new WorkflowClosure instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowClosure instance
             */
            public static create(properties?: flyteidl.core.IWorkflowClosure): flyteidl.core.WorkflowClosure;

            /**
             * Encodes the specified WorkflowClosure message. Does not implicitly {@link flyteidl.core.WorkflowClosure.verify|verify} messages.
             * @param message WorkflowClosure message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.core.IWorkflowClosure, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowClosure message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowClosure
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.core.WorkflowClosure;

            /**
             * Verifies a WorkflowClosure message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }
    }

    /** Namespace event. */
    namespace event {

        /** Properties of a CloudEventWorkflowExecution. */
        interface ICloudEventWorkflowExecution {

            /** CloudEventWorkflowExecution rawEvent */
            rawEvent?: (flyteidl.event.IWorkflowExecutionEvent|null);

            /** CloudEventWorkflowExecution outputInterface */
            outputInterface?: (flyteidl.core.ITypedInterface|null);

            /** CloudEventWorkflowExecution artifactIds */
            artifactIds?: (flyteidl.core.IArtifactID[]|null);

            /** CloudEventWorkflowExecution referenceExecution */
            referenceExecution?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** CloudEventWorkflowExecution principal */
            principal?: (string|null);

            /** CloudEventWorkflowExecution launchPlanId */
            launchPlanId?: (flyteidl.core.IIdentifier|null);
        }

        /** Represents a CloudEventWorkflowExecution. */
        class CloudEventWorkflowExecution implements ICloudEventWorkflowExecution {

            /**
             * Constructs a new CloudEventWorkflowExecution.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.ICloudEventWorkflowExecution);

            /** CloudEventWorkflowExecution rawEvent. */
            public rawEvent?: (flyteidl.event.IWorkflowExecutionEvent|null);

            /** CloudEventWorkflowExecution outputInterface. */
            public outputInterface?: (flyteidl.core.ITypedInterface|null);

            /** CloudEventWorkflowExecution artifactIds. */
            public artifactIds: flyteidl.core.IArtifactID[];

            /** CloudEventWorkflowExecution referenceExecution. */
            public referenceExecution?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** CloudEventWorkflowExecution principal. */
            public principal: string;

            /** CloudEventWorkflowExecution launchPlanId. */
            public launchPlanId?: (flyteidl.core.IIdentifier|null);

            /**
             * Creates a new CloudEventWorkflowExecution instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CloudEventWorkflowExecution instance
             */
            public static create(properties?: flyteidl.event.ICloudEventWorkflowExecution): flyteidl.event.CloudEventWorkflowExecution;

            /**
             * Encodes the specified CloudEventWorkflowExecution message. Does not implicitly {@link flyteidl.event.CloudEventWorkflowExecution.verify|verify} messages.
             * @param message CloudEventWorkflowExecution message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.ICloudEventWorkflowExecution, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CloudEventWorkflowExecution message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CloudEventWorkflowExecution
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.CloudEventWorkflowExecution;

            /**
             * Verifies a CloudEventWorkflowExecution message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CloudEventNodeExecution. */
        interface ICloudEventNodeExecution {

            /** CloudEventNodeExecution rawEvent */
            rawEvent?: (flyteidl.event.INodeExecutionEvent|null);

            /** CloudEventNodeExecution taskExecId */
            taskExecId?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** CloudEventNodeExecution outputInterface */
            outputInterface?: (flyteidl.core.ITypedInterface|null);

            /** CloudEventNodeExecution artifactIds */
            artifactIds?: (flyteidl.core.IArtifactID[]|null);

            /** CloudEventNodeExecution principal */
            principal?: (string|null);

            /** CloudEventNodeExecution launchPlanId */
            launchPlanId?: (flyteidl.core.IIdentifier|null);
        }

        /** Represents a CloudEventNodeExecution. */
        class CloudEventNodeExecution implements ICloudEventNodeExecution {

            /**
             * Constructs a new CloudEventNodeExecution.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.ICloudEventNodeExecution);

            /** CloudEventNodeExecution rawEvent. */
            public rawEvent?: (flyteidl.event.INodeExecutionEvent|null);

            /** CloudEventNodeExecution taskExecId. */
            public taskExecId?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** CloudEventNodeExecution outputInterface. */
            public outputInterface?: (flyteidl.core.ITypedInterface|null);

            /** CloudEventNodeExecution artifactIds. */
            public artifactIds: flyteidl.core.IArtifactID[];

            /** CloudEventNodeExecution principal. */
            public principal: string;

            /** CloudEventNodeExecution launchPlanId. */
            public launchPlanId?: (flyteidl.core.IIdentifier|null);

            /**
             * Creates a new CloudEventNodeExecution instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CloudEventNodeExecution instance
             */
            public static create(properties?: flyteidl.event.ICloudEventNodeExecution): flyteidl.event.CloudEventNodeExecution;

            /**
             * Encodes the specified CloudEventNodeExecution message. Does not implicitly {@link flyteidl.event.CloudEventNodeExecution.verify|verify} messages.
             * @param message CloudEventNodeExecution message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.ICloudEventNodeExecution, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CloudEventNodeExecution message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CloudEventNodeExecution
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.CloudEventNodeExecution;

            /**
             * Verifies a CloudEventNodeExecution message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CloudEventTaskExecution. */
        interface ICloudEventTaskExecution {

            /** CloudEventTaskExecution rawEvent */
            rawEvent?: (flyteidl.event.ITaskExecutionEvent|null);
        }

        /** Represents a CloudEventTaskExecution. */
        class CloudEventTaskExecution implements ICloudEventTaskExecution {

            /**
             * Constructs a new CloudEventTaskExecution.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.ICloudEventTaskExecution);

            /** CloudEventTaskExecution rawEvent. */
            public rawEvent?: (flyteidl.event.ITaskExecutionEvent|null);

            /**
             * Creates a new CloudEventTaskExecution instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CloudEventTaskExecution instance
             */
            public static create(properties?: flyteidl.event.ICloudEventTaskExecution): flyteidl.event.CloudEventTaskExecution;

            /**
             * Encodes the specified CloudEventTaskExecution message. Does not implicitly {@link flyteidl.event.CloudEventTaskExecution.verify|verify} messages.
             * @param message CloudEventTaskExecution message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.ICloudEventTaskExecution, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CloudEventTaskExecution message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CloudEventTaskExecution
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.CloudEventTaskExecution;

            /**
             * Verifies a CloudEventTaskExecution message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CloudEventExecutionStart. */
        interface ICloudEventExecutionStart {

            /** CloudEventExecutionStart executionId */
            executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** CloudEventExecutionStart launchPlanId */
            launchPlanId?: (flyteidl.core.IIdentifier|null);

            /** CloudEventExecutionStart workflowId */
            workflowId?: (flyteidl.core.IIdentifier|null);

            /** CloudEventExecutionStart artifactIds */
            artifactIds?: (flyteidl.core.IArtifactID[]|null);

            /** CloudEventExecutionStart artifactTrackers */
            artifactTrackers?: (string[]|null);

            /** CloudEventExecutionStart principal */
            principal?: (string|null);
        }

        /** Represents a CloudEventExecutionStart. */
        class CloudEventExecutionStart implements ICloudEventExecutionStart {

            /**
             * Constructs a new CloudEventExecutionStart.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.ICloudEventExecutionStart);

            /** CloudEventExecutionStart executionId. */
            public executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** CloudEventExecutionStart launchPlanId. */
            public launchPlanId?: (flyteidl.core.IIdentifier|null);

            /** CloudEventExecutionStart workflowId. */
            public workflowId?: (flyteidl.core.IIdentifier|null);

            /** CloudEventExecutionStart artifactIds. */
            public artifactIds: flyteidl.core.IArtifactID[];

            /** CloudEventExecutionStart artifactTrackers. */
            public artifactTrackers: string[];

            /** CloudEventExecutionStart principal. */
            public principal: string;

            /**
             * Creates a new CloudEventExecutionStart instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CloudEventExecutionStart instance
             */
            public static create(properties?: flyteidl.event.ICloudEventExecutionStart): flyteidl.event.CloudEventExecutionStart;

            /**
             * Encodes the specified CloudEventExecutionStart message. Does not implicitly {@link flyteidl.event.CloudEventExecutionStart.verify|verify} messages.
             * @param message CloudEventExecutionStart message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.ICloudEventExecutionStart, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CloudEventExecutionStart message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CloudEventExecutionStart
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.CloudEventExecutionStart;

            /**
             * Verifies a CloudEventExecutionStart message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionEvent. */
        interface IWorkflowExecutionEvent {

            /** WorkflowExecutionEvent executionId */
            executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** WorkflowExecutionEvent producerId */
            producerId?: (string|null);

            /** WorkflowExecutionEvent phase */
            phase?: (flyteidl.core.WorkflowExecution.Phase|null);

            /** WorkflowExecutionEvent occurredAt */
            occurredAt?: (google.protobuf.ITimestamp|null);

            /** WorkflowExecutionEvent outputUri */
            outputUri?: (string|null);

            /** WorkflowExecutionEvent error */
            error?: (flyteidl.core.IExecutionError|null);

            /** WorkflowExecutionEvent outputData */
            outputData?: (flyteidl.core.ILiteralMap|null);
        }

        /** Represents a WorkflowExecutionEvent. */
        class WorkflowExecutionEvent implements IWorkflowExecutionEvent {

            /**
             * Constructs a new WorkflowExecutionEvent.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.IWorkflowExecutionEvent);

            /** WorkflowExecutionEvent executionId. */
            public executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** WorkflowExecutionEvent producerId. */
            public producerId: string;

            /** WorkflowExecutionEvent phase. */
            public phase: flyteidl.core.WorkflowExecution.Phase;

            /** WorkflowExecutionEvent occurredAt. */
            public occurredAt?: (google.protobuf.ITimestamp|null);

            /** WorkflowExecutionEvent outputUri. */
            public outputUri: string;

            /** WorkflowExecutionEvent error. */
            public error?: (flyteidl.core.IExecutionError|null);

            /** WorkflowExecutionEvent outputData. */
            public outputData?: (flyteidl.core.ILiteralMap|null);

            /** WorkflowExecutionEvent outputResult. */
            public outputResult?: ("outputUri"|"error"|"outputData");

            /**
             * Creates a new WorkflowExecutionEvent instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionEvent instance
             */
            public static create(properties?: flyteidl.event.IWorkflowExecutionEvent): flyteidl.event.WorkflowExecutionEvent;

            /**
             * Encodes the specified WorkflowExecutionEvent message. Does not implicitly {@link flyteidl.event.WorkflowExecutionEvent.verify|verify} messages.
             * @param message WorkflowExecutionEvent message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.IWorkflowExecutionEvent, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionEvent message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionEvent
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.WorkflowExecutionEvent;

            /**
             * Verifies a WorkflowExecutionEvent message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionEvent. */
        interface INodeExecutionEvent {

            /** NodeExecutionEvent id */
            id?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** NodeExecutionEvent producerId */
            producerId?: (string|null);

            /** NodeExecutionEvent phase */
            phase?: (flyteidl.core.NodeExecution.Phase|null);

            /** NodeExecutionEvent occurredAt */
            occurredAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionEvent inputUri */
            inputUri?: (string|null);

            /** NodeExecutionEvent inputData */
            inputData?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionEvent outputUri */
            outputUri?: (string|null);

            /** NodeExecutionEvent error */
            error?: (flyteidl.core.IExecutionError|null);

            /** NodeExecutionEvent outputData */
            outputData?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionEvent workflowNodeMetadata */
            workflowNodeMetadata?: (flyteidl.event.IWorkflowNodeMetadata|null);

            /** NodeExecutionEvent taskNodeMetadata */
            taskNodeMetadata?: (flyteidl.event.ITaskNodeMetadata|null);

            /** NodeExecutionEvent parentTaskMetadata */
            parentTaskMetadata?: (flyteidl.event.IParentTaskExecutionMetadata|null);

            /** NodeExecutionEvent parentNodeMetadata */
            parentNodeMetadata?: (flyteidl.event.IParentNodeExecutionMetadata|null);

            /** NodeExecutionEvent retryGroup */
            retryGroup?: (string|null);

            /** NodeExecutionEvent specNodeId */
            specNodeId?: (string|null);

            /** NodeExecutionEvent nodeName */
            nodeName?: (string|null);

            /** NodeExecutionEvent eventVersion */
            eventVersion?: (number|null);

            /** NodeExecutionEvent isParent */
            isParent?: (boolean|null);

            /** NodeExecutionEvent isDynamic */
            isDynamic?: (boolean|null);

            /** NodeExecutionEvent deckUri */
            deckUri?: (string|null);

            /** NodeExecutionEvent reportedAt */
            reportedAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionEvent isArray */
            isArray?: (boolean|null);
        }

        /** Represents a NodeExecutionEvent. */
        class NodeExecutionEvent implements INodeExecutionEvent {

            /**
             * Constructs a new NodeExecutionEvent.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.INodeExecutionEvent);

            /** NodeExecutionEvent id. */
            public id?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** NodeExecutionEvent producerId. */
            public producerId: string;

            /** NodeExecutionEvent phase. */
            public phase: flyteidl.core.NodeExecution.Phase;

            /** NodeExecutionEvent occurredAt. */
            public occurredAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionEvent inputUri. */
            public inputUri: string;

            /** NodeExecutionEvent inputData. */
            public inputData?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionEvent outputUri. */
            public outputUri: string;

            /** NodeExecutionEvent error. */
            public error?: (flyteidl.core.IExecutionError|null);

            /** NodeExecutionEvent outputData. */
            public outputData?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionEvent workflowNodeMetadata. */
            public workflowNodeMetadata?: (flyteidl.event.IWorkflowNodeMetadata|null);

            /** NodeExecutionEvent taskNodeMetadata. */
            public taskNodeMetadata?: (flyteidl.event.ITaskNodeMetadata|null);

            /** NodeExecutionEvent parentTaskMetadata. */
            public parentTaskMetadata?: (flyteidl.event.IParentTaskExecutionMetadata|null);

            /** NodeExecutionEvent parentNodeMetadata. */
            public parentNodeMetadata?: (flyteidl.event.IParentNodeExecutionMetadata|null);

            /** NodeExecutionEvent retryGroup. */
            public retryGroup: string;

            /** NodeExecutionEvent specNodeId. */
            public specNodeId: string;

            /** NodeExecutionEvent nodeName. */
            public nodeName: string;

            /** NodeExecutionEvent eventVersion. */
            public eventVersion: number;

            /** NodeExecutionEvent isParent. */
            public isParent: boolean;

            /** NodeExecutionEvent isDynamic. */
            public isDynamic: boolean;

            /** NodeExecutionEvent deckUri. */
            public deckUri: string;

            /** NodeExecutionEvent reportedAt. */
            public reportedAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionEvent isArray. */
            public isArray: boolean;

            /** NodeExecutionEvent inputValue. */
            public inputValue?: ("inputUri"|"inputData");

            /** NodeExecutionEvent outputResult. */
            public outputResult?: ("outputUri"|"error"|"outputData");

            /** NodeExecutionEvent targetMetadata. */
            public targetMetadata?: ("workflowNodeMetadata"|"taskNodeMetadata");

            /**
             * Creates a new NodeExecutionEvent instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionEvent instance
             */
            public static create(properties?: flyteidl.event.INodeExecutionEvent): flyteidl.event.NodeExecutionEvent;

            /**
             * Encodes the specified NodeExecutionEvent message. Does not implicitly {@link flyteidl.event.NodeExecutionEvent.verify|verify} messages.
             * @param message NodeExecutionEvent message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.INodeExecutionEvent, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionEvent message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionEvent
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.NodeExecutionEvent;

            /**
             * Verifies a NodeExecutionEvent message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowNodeMetadata. */
        interface IWorkflowNodeMetadata {

            /** WorkflowNodeMetadata executionId */
            executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);
        }

        /** Represents a WorkflowNodeMetadata. */
        class WorkflowNodeMetadata implements IWorkflowNodeMetadata {

            /**
             * Constructs a new WorkflowNodeMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.IWorkflowNodeMetadata);

            /** WorkflowNodeMetadata executionId. */
            public executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /**
             * Creates a new WorkflowNodeMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowNodeMetadata instance
             */
            public static create(properties?: flyteidl.event.IWorkflowNodeMetadata): flyteidl.event.WorkflowNodeMetadata;

            /**
             * Encodes the specified WorkflowNodeMetadata message. Does not implicitly {@link flyteidl.event.WorkflowNodeMetadata.verify|verify} messages.
             * @param message WorkflowNodeMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.IWorkflowNodeMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowNodeMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowNodeMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.WorkflowNodeMetadata;

            /**
             * Verifies a WorkflowNodeMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskNodeMetadata. */
        interface ITaskNodeMetadata {

            /** TaskNodeMetadata cacheStatus */
            cacheStatus?: (flyteidl.core.CatalogCacheStatus|null);

            /** TaskNodeMetadata catalogKey */
            catalogKey?: (flyteidl.core.ICatalogMetadata|null);

            /** TaskNodeMetadata reservationStatus */
            reservationStatus?: (flyteidl.core.CatalogReservation.Status|null);

            /** TaskNodeMetadata checkpointUri */
            checkpointUri?: (string|null);

            /** TaskNodeMetadata dynamicWorkflow */
            dynamicWorkflow?: (flyteidl.event.IDynamicWorkflowNodeMetadata|null);
        }

        /** Represents a TaskNodeMetadata. */
        class TaskNodeMetadata implements ITaskNodeMetadata {

            /**
             * Constructs a new TaskNodeMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.ITaskNodeMetadata);

            /** TaskNodeMetadata cacheStatus. */
            public cacheStatus: flyteidl.core.CatalogCacheStatus;

            /** TaskNodeMetadata catalogKey. */
            public catalogKey?: (flyteidl.core.ICatalogMetadata|null);

            /** TaskNodeMetadata reservationStatus. */
            public reservationStatus: flyteidl.core.CatalogReservation.Status;

            /** TaskNodeMetadata checkpointUri. */
            public checkpointUri: string;

            /** TaskNodeMetadata dynamicWorkflow. */
            public dynamicWorkflow?: (flyteidl.event.IDynamicWorkflowNodeMetadata|null);

            /**
             * Creates a new TaskNodeMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskNodeMetadata instance
             */
            public static create(properties?: flyteidl.event.ITaskNodeMetadata): flyteidl.event.TaskNodeMetadata;

            /**
             * Encodes the specified TaskNodeMetadata message. Does not implicitly {@link flyteidl.event.TaskNodeMetadata.verify|verify} messages.
             * @param message TaskNodeMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.ITaskNodeMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskNodeMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskNodeMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.TaskNodeMetadata;

            /**
             * Verifies a TaskNodeMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DynamicWorkflowNodeMetadata. */
        interface IDynamicWorkflowNodeMetadata {

            /** DynamicWorkflowNodeMetadata id */
            id?: (flyteidl.core.IIdentifier|null);

            /** DynamicWorkflowNodeMetadata compiledWorkflow */
            compiledWorkflow?: (flyteidl.core.ICompiledWorkflowClosure|null);

            /** DynamicWorkflowNodeMetadata dynamicJobSpecUri */
            dynamicJobSpecUri?: (string|null);
        }

        /** Represents a DynamicWorkflowNodeMetadata. */
        class DynamicWorkflowNodeMetadata implements IDynamicWorkflowNodeMetadata {

            /**
             * Constructs a new DynamicWorkflowNodeMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.IDynamicWorkflowNodeMetadata);

            /** DynamicWorkflowNodeMetadata id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** DynamicWorkflowNodeMetadata compiledWorkflow. */
            public compiledWorkflow?: (flyteidl.core.ICompiledWorkflowClosure|null);

            /** DynamicWorkflowNodeMetadata dynamicJobSpecUri. */
            public dynamicJobSpecUri: string;

            /**
             * Creates a new DynamicWorkflowNodeMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DynamicWorkflowNodeMetadata instance
             */
            public static create(properties?: flyteidl.event.IDynamicWorkflowNodeMetadata): flyteidl.event.DynamicWorkflowNodeMetadata;

            /**
             * Encodes the specified DynamicWorkflowNodeMetadata message. Does not implicitly {@link flyteidl.event.DynamicWorkflowNodeMetadata.verify|verify} messages.
             * @param message DynamicWorkflowNodeMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.IDynamicWorkflowNodeMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DynamicWorkflowNodeMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DynamicWorkflowNodeMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.DynamicWorkflowNodeMetadata;

            /**
             * Verifies a DynamicWorkflowNodeMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ParentTaskExecutionMetadata. */
        interface IParentTaskExecutionMetadata {

            /** ParentTaskExecutionMetadata id */
            id?: (flyteidl.core.ITaskExecutionIdentifier|null);
        }

        /** Represents a ParentTaskExecutionMetadata. */
        class ParentTaskExecutionMetadata implements IParentTaskExecutionMetadata {

            /**
             * Constructs a new ParentTaskExecutionMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.IParentTaskExecutionMetadata);

            /** ParentTaskExecutionMetadata id. */
            public id?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /**
             * Creates a new ParentTaskExecutionMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ParentTaskExecutionMetadata instance
             */
            public static create(properties?: flyteidl.event.IParentTaskExecutionMetadata): flyteidl.event.ParentTaskExecutionMetadata;

            /**
             * Encodes the specified ParentTaskExecutionMetadata message. Does not implicitly {@link flyteidl.event.ParentTaskExecutionMetadata.verify|verify} messages.
             * @param message ParentTaskExecutionMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.IParentTaskExecutionMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ParentTaskExecutionMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ParentTaskExecutionMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.ParentTaskExecutionMetadata;

            /**
             * Verifies a ParentTaskExecutionMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ParentNodeExecutionMetadata. */
        interface IParentNodeExecutionMetadata {

            /** ParentNodeExecutionMetadata nodeId */
            nodeId?: (string|null);
        }

        /** Represents a ParentNodeExecutionMetadata. */
        class ParentNodeExecutionMetadata implements IParentNodeExecutionMetadata {

            /**
             * Constructs a new ParentNodeExecutionMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.IParentNodeExecutionMetadata);

            /** ParentNodeExecutionMetadata nodeId. */
            public nodeId: string;

            /**
             * Creates a new ParentNodeExecutionMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ParentNodeExecutionMetadata instance
             */
            public static create(properties?: flyteidl.event.IParentNodeExecutionMetadata): flyteidl.event.ParentNodeExecutionMetadata;

            /**
             * Encodes the specified ParentNodeExecutionMetadata message. Does not implicitly {@link flyteidl.event.ParentNodeExecutionMetadata.verify|verify} messages.
             * @param message ParentNodeExecutionMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.IParentNodeExecutionMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ParentNodeExecutionMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ParentNodeExecutionMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.ParentNodeExecutionMetadata;

            /**
             * Verifies a ParentNodeExecutionMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an EventReason. */
        interface IEventReason {

            /** EventReason reason */
            reason?: (string|null);

            /** EventReason occurredAt */
            occurredAt?: (google.protobuf.ITimestamp|null);
        }

        /** Represents an EventReason. */
        class EventReason implements IEventReason {

            /**
             * Constructs a new EventReason.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.IEventReason);

            /** EventReason reason. */
            public reason: string;

            /** EventReason occurredAt. */
            public occurredAt?: (google.protobuf.ITimestamp|null);

            /**
             * Creates a new EventReason instance using the specified properties.
             * @param [properties] Properties to set
             * @returns EventReason instance
             */
            public static create(properties?: flyteidl.event.IEventReason): flyteidl.event.EventReason;

            /**
             * Encodes the specified EventReason message. Does not implicitly {@link flyteidl.event.EventReason.verify|verify} messages.
             * @param message EventReason message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.IEventReason, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EventReason message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EventReason
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.EventReason;

            /**
             * Verifies an EventReason message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskExecutionEvent. */
        interface ITaskExecutionEvent {

            /** TaskExecutionEvent taskId */
            taskId?: (flyteidl.core.IIdentifier|null);

            /** TaskExecutionEvent parentNodeExecutionId */
            parentNodeExecutionId?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** TaskExecutionEvent retryAttempt */
            retryAttempt?: (number|null);

            /** TaskExecutionEvent phase */
            phase?: (flyteidl.core.TaskExecution.Phase|null);

            /** TaskExecutionEvent producerId */
            producerId?: (string|null);

            /** TaskExecutionEvent logs */
            logs?: (flyteidl.core.ITaskLog[]|null);

            /** TaskExecutionEvent occurredAt */
            occurredAt?: (google.protobuf.ITimestamp|null);

            /** TaskExecutionEvent inputUri */
            inputUri?: (string|null);

            /** TaskExecutionEvent inputData */
            inputData?: (flyteidl.core.ILiteralMap|null);

            /** TaskExecutionEvent outputUri */
            outputUri?: (string|null);

            /** TaskExecutionEvent error */
            error?: (flyteidl.core.IExecutionError|null);

            /** TaskExecutionEvent outputData */
            outputData?: (flyteidl.core.ILiteralMap|null);

            /** TaskExecutionEvent customInfo */
            customInfo?: (google.protobuf.IStruct|null);

            /** TaskExecutionEvent phaseVersion */
            phaseVersion?: (number|null);

            /** TaskExecutionEvent reason */
            reason?: (string|null);

            /** TaskExecutionEvent reasons */
            reasons?: (flyteidl.event.IEventReason[]|null);

            /** TaskExecutionEvent taskType */
            taskType?: (string|null);

            /** TaskExecutionEvent metadata */
            metadata?: (flyteidl.event.ITaskExecutionMetadata|null);

            /** TaskExecutionEvent eventVersion */
            eventVersion?: (number|null);

            /** TaskExecutionEvent reportedAt */
            reportedAt?: (google.protobuf.ITimestamp|null);
        }

        /** Represents a TaskExecutionEvent. */
        class TaskExecutionEvent implements ITaskExecutionEvent {

            /**
             * Constructs a new TaskExecutionEvent.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.ITaskExecutionEvent);

            /** TaskExecutionEvent taskId. */
            public taskId?: (flyteidl.core.IIdentifier|null);

            /** TaskExecutionEvent parentNodeExecutionId. */
            public parentNodeExecutionId?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** TaskExecutionEvent retryAttempt. */
            public retryAttempt: number;

            /** TaskExecutionEvent phase. */
            public phase: flyteidl.core.TaskExecution.Phase;

            /** TaskExecutionEvent producerId. */
            public producerId: string;

            /** TaskExecutionEvent logs. */
            public logs: flyteidl.core.ITaskLog[];

            /** TaskExecutionEvent occurredAt. */
            public occurredAt?: (google.protobuf.ITimestamp|null);

            /** TaskExecutionEvent inputUri. */
            public inputUri: string;

            /** TaskExecutionEvent inputData. */
            public inputData?: (flyteidl.core.ILiteralMap|null);

            /** TaskExecutionEvent outputUri. */
            public outputUri: string;

            /** TaskExecutionEvent error. */
            public error?: (flyteidl.core.IExecutionError|null);

            /** TaskExecutionEvent outputData. */
            public outputData?: (flyteidl.core.ILiteralMap|null);

            /** TaskExecutionEvent customInfo. */
            public customInfo?: (google.protobuf.IStruct|null);

            /** TaskExecutionEvent phaseVersion. */
            public phaseVersion: number;

            /** TaskExecutionEvent reason. */
            public reason: string;

            /** TaskExecutionEvent reasons. */
            public reasons: flyteidl.event.IEventReason[];

            /** TaskExecutionEvent taskType. */
            public taskType: string;

            /** TaskExecutionEvent metadata. */
            public metadata?: (flyteidl.event.ITaskExecutionMetadata|null);

            /** TaskExecutionEvent eventVersion. */
            public eventVersion: number;

            /** TaskExecutionEvent reportedAt. */
            public reportedAt?: (google.protobuf.ITimestamp|null);

            /** TaskExecutionEvent inputValue. */
            public inputValue?: ("inputUri"|"inputData");

            /** TaskExecutionEvent outputResult. */
            public outputResult?: ("outputUri"|"error"|"outputData");

            /**
             * Creates a new TaskExecutionEvent instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskExecutionEvent instance
             */
            public static create(properties?: flyteidl.event.ITaskExecutionEvent): flyteidl.event.TaskExecutionEvent;

            /**
             * Encodes the specified TaskExecutionEvent message. Does not implicitly {@link flyteidl.event.TaskExecutionEvent.verify|verify} messages.
             * @param message TaskExecutionEvent message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.ITaskExecutionEvent, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskExecutionEvent message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskExecutionEvent
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.TaskExecutionEvent;

            /**
             * Verifies a TaskExecutionEvent message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExternalResourceInfo. */
        interface IExternalResourceInfo {

            /** ExternalResourceInfo externalId */
            externalId?: (string|null);

            /** ExternalResourceInfo index */
            index?: (number|null);

            /** ExternalResourceInfo retryAttempt */
            retryAttempt?: (number|null);

            /** ExternalResourceInfo phase */
            phase?: (flyteidl.core.TaskExecution.Phase|null);

            /** ExternalResourceInfo cacheStatus */
            cacheStatus?: (flyteidl.core.CatalogCacheStatus|null);

            /** ExternalResourceInfo logs */
            logs?: (flyteidl.core.ITaskLog[]|null);
        }

        /** Represents an ExternalResourceInfo. */
        class ExternalResourceInfo implements IExternalResourceInfo {

            /**
             * Constructs a new ExternalResourceInfo.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.IExternalResourceInfo);

            /** ExternalResourceInfo externalId. */
            public externalId: string;

            /** ExternalResourceInfo index. */
            public index: number;

            /** ExternalResourceInfo retryAttempt. */
            public retryAttempt: number;

            /** ExternalResourceInfo phase. */
            public phase: flyteidl.core.TaskExecution.Phase;

            /** ExternalResourceInfo cacheStatus. */
            public cacheStatus: flyteidl.core.CatalogCacheStatus;

            /** ExternalResourceInfo logs. */
            public logs: flyteidl.core.ITaskLog[];

            /**
             * Creates a new ExternalResourceInfo instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExternalResourceInfo instance
             */
            public static create(properties?: flyteidl.event.IExternalResourceInfo): flyteidl.event.ExternalResourceInfo;

            /**
             * Encodes the specified ExternalResourceInfo message. Does not implicitly {@link flyteidl.event.ExternalResourceInfo.verify|verify} messages.
             * @param message ExternalResourceInfo message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.IExternalResourceInfo, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExternalResourceInfo message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExternalResourceInfo
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.ExternalResourceInfo;

            /**
             * Verifies an ExternalResourceInfo message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ResourcePoolInfo. */
        interface IResourcePoolInfo {

            /** ResourcePoolInfo allocationToken */
            allocationToken?: (string|null);

            /** ResourcePoolInfo namespace */
            namespace?: (string|null);
        }

        /** Represents a ResourcePoolInfo. */
        class ResourcePoolInfo implements IResourcePoolInfo {

            /**
             * Constructs a new ResourcePoolInfo.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.IResourcePoolInfo);

            /** ResourcePoolInfo allocationToken. */
            public allocationToken: string;

            /** ResourcePoolInfo namespace. */
            public namespace: string;

            /**
             * Creates a new ResourcePoolInfo instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ResourcePoolInfo instance
             */
            public static create(properties?: flyteidl.event.IResourcePoolInfo): flyteidl.event.ResourcePoolInfo;

            /**
             * Encodes the specified ResourcePoolInfo message. Does not implicitly {@link flyteidl.event.ResourcePoolInfo.verify|verify} messages.
             * @param message ResourcePoolInfo message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.IResourcePoolInfo, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ResourcePoolInfo message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ResourcePoolInfo
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.ResourcePoolInfo;

            /**
             * Verifies a ResourcePoolInfo message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskExecutionMetadata. */
        interface ITaskExecutionMetadata {

            /** TaskExecutionMetadata generatedName */
            generatedName?: (string|null);

            /** TaskExecutionMetadata externalResources */
            externalResources?: (flyteidl.event.IExternalResourceInfo[]|null);

            /** TaskExecutionMetadata resourcePoolInfo */
            resourcePoolInfo?: (flyteidl.event.IResourcePoolInfo[]|null);

            /** TaskExecutionMetadata pluginIdentifier */
            pluginIdentifier?: (string|null);

            /** TaskExecutionMetadata instanceClass */
            instanceClass?: (flyteidl.event.TaskExecutionMetadata.InstanceClass|null);
        }

        /** Represents a TaskExecutionMetadata. */
        class TaskExecutionMetadata implements ITaskExecutionMetadata {

            /**
             * Constructs a new TaskExecutionMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.event.ITaskExecutionMetadata);

            /** TaskExecutionMetadata generatedName. */
            public generatedName: string;

            /** TaskExecutionMetadata externalResources. */
            public externalResources: flyteidl.event.IExternalResourceInfo[];

            /** TaskExecutionMetadata resourcePoolInfo. */
            public resourcePoolInfo: flyteidl.event.IResourcePoolInfo[];

            /** TaskExecutionMetadata pluginIdentifier. */
            public pluginIdentifier: string;

            /** TaskExecutionMetadata instanceClass. */
            public instanceClass: flyteidl.event.TaskExecutionMetadata.InstanceClass;

            /**
             * Creates a new TaskExecutionMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskExecutionMetadata instance
             */
            public static create(properties?: flyteidl.event.ITaskExecutionMetadata): flyteidl.event.TaskExecutionMetadata;

            /**
             * Encodes the specified TaskExecutionMetadata message. Does not implicitly {@link flyteidl.event.TaskExecutionMetadata.verify|verify} messages.
             * @param message TaskExecutionMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.event.ITaskExecutionMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskExecutionMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskExecutionMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.event.TaskExecutionMetadata;

            /**
             * Verifies a TaskExecutionMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace TaskExecutionMetadata {

            /** InstanceClass enum. */
            enum InstanceClass {
                DEFAULT = 0,
                INTERRUPTIBLE = 1
            }
        }
    }

    /** Namespace admin. */
    namespace admin {

        /** State enum. */
        enum State {
            RETRYABLE_FAILURE = 0,
            PERMANENT_FAILURE = 1,
            PENDING = 2,
            RUNNING = 3,
            SUCCEEDED = 4
        }

        /** Properties of a TaskExecutionMetadata. */
        interface ITaskExecutionMetadata {

            /** TaskExecutionMetadata taskExecutionId */
            taskExecutionId?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** TaskExecutionMetadata namespace */
            namespace?: (string|null);

            /** TaskExecutionMetadata labels */
            labels?: ({ [k: string]: string }|null);

            /** TaskExecutionMetadata annotations */
            annotations?: ({ [k: string]: string }|null);

            /** TaskExecutionMetadata k8sServiceAccount */
            k8sServiceAccount?: (string|null);

            /** TaskExecutionMetadata environmentVariables */
            environmentVariables?: ({ [k: string]: string }|null);

            /** TaskExecutionMetadata maxAttempts */
            maxAttempts?: (number|null);

            /** TaskExecutionMetadata interruptible */
            interruptible?: (boolean|null);

            /** TaskExecutionMetadata interruptibleFailureThreshold */
            interruptibleFailureThreshold?: (number|null);

            /** TaskExecutionMetadata overrides */
            overrides?: (flyteidl.core.ITaskNodeOverrides|null);

            /** TaskExecutionMetadata identity */
            identity?: (flyteidl.core.IIdentity|null);
        }

        /** Represents a TaskExecutionMetadata. */
        class TaskExecutionMetadata implements ITaskExecutionMetadata {

            /**
             * Constructs a new TaskExecutionMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ITaskExecutionMetadata);

            /** TaskExecutionMetadata taskExecutionId. */
            public taskExecutionId?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** TaskExecutionMetadata namespace. */
            public namespace: string;

            /** TaskExecutionMetadata labels. */
            public labels: { [k: string]: string };

            /** TaskExecutionMetadata annotations. */
            public annotations: { [k: string]: string };

            /** TaskExecutionMetadata k8sServiceAccount. */
            public k8sServiceAccount: string;

            /** TaskExecutionMetadata environmentVariables. */
            public environmentVariables: { [k: string]: string };

            /** TaskExecutionMetadata maxAttempts. */
            public maxAttempts: number;

            /** TaskExecutionMetadata interruptible. */
            public interruptible: boolean;

            /** TaskExecutionMetadata interruptibleFailureThreshold. */
            public interruptibleFailureThreshold: number;

            /** TaskExecutionMetadata overrides. */
            public overrides?: (flyteidl.core.ITaskNodeOverrides|null);

            /** TaskExecutionMetadata identity. */
            public identity?: (flyteidl.core.IIdentity|null);

            /**
             * Creates a new TaskExecutionMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskExecutionMetadata instance
             */
            public static create(properties?: flyteidl.admin.ITaskExecutionMetadata): flyteidl.admin.TaskExecutionMetadata;

            /**
             * Encodes the specified TaskExecutionMetadata message. Does not implicitly {@link flyteidl.admin.TaskExecutionMetadata.verify|verify} messages.
             * @param message TaskExecutionMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ITaskExecutionMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskExecutionMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskExecutionMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.TaskExecutionMetadata;

            /**
             * Verifies a TaskExecutionMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CreateTaskRequest. */
        interface ICreateTaskRequest {

            /** CreateTaskRequest inputs */
            inputs?: (flyteidl.core.ILiteralMap|null);

            /** CreateTaskRequest template */
            template?: (flyteidl.core.ITaskTemplate|null);

            /** CreateTaskRequest outputPrefix */
            outputPrefix?: (string|null);

            /** CreateTaskRequest taskExecutionMetadata */
            taskExecutionMetadata?: (flyteidl.admin.ITaskExecutionMetadata|null);
        }

        /** Represents a CreateTaskRequest. */
        class CreateTaskRequest implements ICreateTaskRequest {

            /**
             * Constructs a new CreateTaskRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ICreateTaskRequest);

            /** CreateTaskRequest inputs. */
            public inputs?: (flyteidl.core.ILiteralMap|null);

            /** CreateTaskRequest template. */
            public template?: (flyteidl.core.ITaskTemplate|null);

            /** CreateTaskRequest outputPrefix. */
            public outputPrefix: string;

            /** CreateTaskRequest taskExecutionMetadata. */
            public taskExecutionMetadata?: (flyteidl.admin.ITaskExecutionMetadata|null);

            /**
             * Creates a new CreateTaskRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CreateTaskRequest instance
             */
            public static create(properties?: flyteidl.admin.ICreateTaskRequest): flyteidl.admin.CreateTaskRequest;

            /**
             * Encodes the specified CreateTaskRequest message. Does not implicitly {@link flyteidl.admin.CreateTaskRequest.verify|verify} messages.
             * @param message CreateTaskRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ICreateTaskRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CreateTaskRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CreateTaskRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.CreateTaskRequest;

            /**
             * Verifies a CreateTaskRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CreateTaskResponse. */
        interface ICreateTaskResponse {

            /** CreateTaskResponse resourceMeta */
            resourceMeta?: (Uint8Array|null);
        }

        /** Represents a CreateTaskResponse. */
        class CreateTaskResponse implements ICreateTaskResponse {

            /**
             * Constructs a new CreateTaskResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ICreateTaskResponse);

            /** CreateTaskResponse resourceMeta. */
            public resourceMeta: Uint8Array;

            /**
             * Creates a new CreateTaskResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CreateTaskResponse instance
             */
            public static create(properties?: flyteidl.admin.ICreateTaskResponse): flyteidl.admin.CreateTaskResponse;

            /**
             * Encodes the specified CreateTaskResponse message. Does not implicitly {@link flyteidl.admin.CreateTaskResponse.verify|verify} messages.
             * @param message CreateTaskResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ICreateTaskResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CreateTaskResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CreateTaskResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.CreateTaskResponse;

            /**
             * Verifies a CreateTaskResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CreateRequestHeader. */
        interface ICreateRequestHeader {

            /** CreateRequestHeader template */
            template?: (flyteidl.core.ITaskTemplate|null);

            /** CreateRequestHeader outputPrefix */
            outputPrefix?: (string|null);

            /** CreateRequestHeader taskExecutionMetadata */
            taskExecutionMetadata?: (flyteidl.admin.ITaskExecutionMetadata|null);

            /** CreateRequestHeader maxDatasetSizeBytes */
            maxDatasetSizeBytes?: (Long|null);
        }

        /** Represents a CreateRequestHeader. */
        class CreateRequestHeader implements ICreateRequestHeader {

            /**
             * Constructs a new CreateRequestHeader.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ICreateRequestHeader);

            /** CreateRequestHeader template. */
            public template?: (flyteidl.core.ITaskTemplate|null);

            /** CreateRequestHeader outputPrefix. */
            public outputPrefix: string;

            /** CreateRequestHeader taskExecutionMetadata. */
            public taskExecutionMetadata?: (flyteidl.admin.ITaskExecutionMetadata|null);

            /** CreateRequestHeader maxDatasetSizeBytes. */
            public maxDatasetSizeBytes: Long;

            /**
             * Creates a new CreateRequestHeader instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CreateRequestHeader instance
             */
            public static create(properties?: flyteidl.admin.ICreateRequestHeader): flyteidl.admin.CreateRequestHeader;

            /**
             * Encodes the specified CreateRequestHeader message. Does not implicitly {@link flyteidl.admin.CreateRequestHeader.verify|verify} messages.
             * @param message CreateRequestHeader message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ICreateRequestHeader, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CreateRequestHeader message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CreateRequestHeader
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.CreateRequestHeader;

            /**
             * Verifies a CreateRequestHeader message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecuteTaskSyncRequest. */
        interface IExecuteTaskSyncRequest {

            /** ExecuteTaskSyncRequest header */
            header?: (flyteidl.admin.ICreateRequestHeader|null);

            /** ExecuteTaskSyncRequest inputs */
            inputs?: (flyteidl.core.ILiteralMap|null);
        }

        /** Represents an ExecuteTaskSyncRequest. */
        class ExecuteTaskSyncRequest implements IExecuteTaskSyncRequest {

            /**
             * Constructs a new ExecuteTaskSyncRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecuteTaskSyncRequest);

            /** ExecuteTaskSyncRequest header. */
            public header?: (flyteidl.admin.ICreateRequestHeader|null);

            /** ExecuteTaskSyncRequest inputs. */
            public inputs?: (flyteidl.core.ILiteralMap|null);

            /** ExecuteTaskSyncRequest part. */
            public part?: ("header"|"inputs");

            /**
             * Creates a new ExecuteTaskSyncRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecuteTaskSyncRequest instance
             */
            public static create(properties?: flyteidl.admin.IExecuteTaskSyncRequest): flyteidl.admin.ExecuteTaskSyncRequest;

            /**
             * Encodes the specified ExecuteTaskSyncRequest message. Does not implicitly {@link flyteidl.admin.ExecuteTaskSyncRequest.verify|verify} messages.
             * @param message ExecuteTaskSyncRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecuteTaskSyncRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecuteTaskSyncRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecuteTaskSyncRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecuteTaskSyncRequest;

            /**
             * Verifies an ExecuteTaskSyncRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecuteTaskSyncResponseHeader. */
        interface IExecuteTaskSyncResponseHeader {

            /** ExecuteTaskSyncResponseHeader resource */
            resource?: (flyteidl.admin.IResource|null);
        }

        /** Represents an ExecuteTaskSyncResponseHeader. */
        class ExecuteTaskSyncResponseHeader implements IExecuteTaskSyncResponseHeader {

            /**
             * Constructs a new ExecuteTaskSyncResponseHeader.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecuteTaskSyncResponseHeader);

            /** ExecuteTaskSyncResponseHeader resource. */
            public resource?: (flyteidl.admin.IResource|null);

            /**
             * Creates a new ExecuteTaskSyncResponseHeader instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecuteTaskSyncResponseHeader instance
             */
            public static create(properties?: flyteidl.admin.IExecuteTaskSyncResponseHeader): flyteidl.admin.ExecuteTaskSyncResponseHeader;

            /**
             * Encodes the specified ExecuteTaskSyncResponseHeader message. Does not implicitly {@link flyteidl.admin.ExecuteTaskSyncResponseHeader.verify|verify} messages.
             * @param message ExecuteTaskSyncResponseHeader message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecuteTaskSyncResponseHeader, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecuteTaskSyncResponseHeader message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecuteTaskSyncResponseHeader
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecuteTaskSyncResponseHeader;

            /**
             * Verifies an ExecuteTaskSyncResponseHeader message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecuteTaskSyncResponse. */
        interface IExecuteTaskSyncResponse {

            /** ExecuteTaskSyncResponse header */
            header?: (flyteidl.admin.IExecuteTaskSyncResponseHeader|null);

            /** ExecuteTaskSyncResponse outputs */
            outputs?: (flyteidl.core.ILiteralMap|null);
        }

        /** Represents an ExecuteTaskSyncResponse. */
        class ExecuteTaskSyncResponse implements IExecuteTaskSyncResponse {

            /**
             * Constructs a new ExecuteTaskSyncResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecuteTaskSyncResponse);

            /** ExecuteTaskSyncResponse header. */
            public header?: (flyteidl.admin.IExecuteTaskSyncResponseHeader|null);

            /** ExecuteTaskSyncResponse outputs. */
            public outputs?: (flyteidl.core.ILiteralMap|null);

            /** ExecuteTaskSyncResponse res. */
            public res?: ("header"|"outputs");

            /**
             * Creates a new ExecuteTaskSyncResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecuteTaskSyncResponse instance
             */
            public static create(properties?: flyteidl.admin.IExecuteTaskSyncResponse): flyteidl.admin.ExecuteTaskSyncResponse;

            /**
             * Encodes the specified ExecuteTaskSyncResponse message. Does not implicitly {@link flyteidl.admin.ExecuteTaskSyncResponse.verify|verify} messages.
             * @param message ExecuteTaskSyncResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecuteTaskSyncResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecuteTaskSyncResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecuteTaskSyncResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecuteTaskSyncResponse;

            /**
             * Verifies an ExecuteTaskSyncResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetTaskRequest. */
        interface IGetTaskRequest {

            /** GetTaskRequest taskType */
            taskType?: (string|null);

            /** GetTaskRequest resourceMeta */
            resourceMeta?: (Uint8Array|null);

            /** GetTaskRequest taskCategory */
            taskCategory?: (flyteidl.admin.ITaskCategory|null);
        }

        /** Represents a GetTaskRequest. */
        class GetTaskRequest implements IGetTaskRequest {

            /**
             * Constructs a new GetTaskRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetTaskRequest);

            /** GetTaskRequest taskType. */
            public taskType: string;

            /** GetTaskRequest resourceMeta. */
            public resourceMeta: Uint8Array;

            /** GetTaskRequest taskCategory. */
            public taskCategory?: (flyteidl.admin.ITaskCategory|null);

            /**
             * Creates a new GetTaskRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetTaskRequest instance
             */
            public static create(properties?: flyteidl.admin.IGetTaskRequest): flyteidl.admin.GetTaskRequest;

            /**
             * Encodes the specified GetTaskRequest message. Does not implicitly {@link flyteidl.admin.GetTaskRequest.verify|verify} messages.
             * @param message GetTaskRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetTaskRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetTaskRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetTaskRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetTaskRequest;

            /**
             * Verifies a GetTaskRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetTaskResponse. */
        interface IGetTaskResponse {

            /** GetTaskResponse resource */
            resource?: (flyteidl.admin.IResource|null);
        }

        /** Represents a GetTaskResponse. */
        class GetTaskResponse implements IGetTaskResponse {

            /**
             * Constructs a new GetTaskResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetTaskResponse);

            /** GetTaskResponse resource. */
            public resource?: (flyteidl.admin.IResource|null);

            /**
             * Creates a new GetTaskResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetTaskResponse instance
             */
            public static create(properties?: flyteidl.admin.IGetTaskResponse): flyteidl.admin.GetTaskResponse;

            /**
             * Encodes the specified GetTaskResponse message. Does not implicitly {@link flyteidl.admin.GetTaskResponse.verify|verify} messages.
             * @param message GetTaskResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetTaskResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetTaskResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetTaskResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetTaskResponse;

            /**
             * Verifies a GetTaskResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Resource. */
        interface IResource {

            /** Resource state */
            state?: (flyteidl.admin.State|null);

            /** Resource outputs */
            outputs?: (flyteidl.core.ILiteralMap|null);

            /** Resource message */
            message?: (string|null);

            /** Resource logLinks */
            logLinks?: (flyteidl.core.ITaskLog[]|null);

            /** Resource phase */
            phase?: (flyteidl.core.TaskExecution.Phase|null);

            /** Resource customInfo */
            customInfo?: (google.protobuf.IStruct|null);
        }

        /** Represents a Resource. */
        class Resource implements IResource {

            /**
             * Constructs a new Resource.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IResource);

            /** Resource state. */
            public state: flyteidl.admin.State;

            /** Resource outputs. */
            public outputs?: (flyteidl.core.ILiteralMap|null);

            /** Resource message. */
            public message: string;

            /** Resource logLinks. */
            public logLinks: flyteidl.core.ITaskLog[];

            /** Resource phase. */
            public phase: flyteidl.core.TaskExecution.Phase;

            /** Resource customInfo. */
            public customInfo?: (google.protobuf.IStruct|null);

            /**
             * Creates a new Resource instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Resource instance
             */
            public static create(properties?: flyteidl.admin.IResource): flyteidl.admin.Resource;

            /**
             * Encodes the specified Resource message. Does not implicitly {@link flyteidl.admin.Resource.verify|verify} messages.
             * @param message Resource message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IResource, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Resource message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Resource
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Resource;

            /**
             * Verifies a Resource message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DeleteTaskRequest. */
        interface IDeleteTaskRequest {

            /** DeleteTaskRequest taskType */
            taskType?: (string|null);

            /** DeleteTaskRequest resourceMeta */
            resourceMeta?: (Uint8Array|null);

            /** DeleteTaskRequest taskCategory */
            taskCategory?: (flyteidl.admin.ITaskCategory|null);
        }

        /** Represents a DeleteTaskRequest. */
        class DeleteTaskRequest implements IDeleteTaskRequest {

            /**
             * Constructs a new DeleteTaskRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDeleteTaskRequest);

            /** DeleteTaskRequest taskType. */
            public taskType: string;

            /** DeleteTaskRequest resourceMeta. */
            public resourceMeta: Uint8Array;

            /** DeleteTaskRequest taskCategory. */
            public taskCategory?: (flyteidl.admin.ITaskCategory|null);

            /**
             * Creates a new DeleteTaskRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DeleteTaskRequest instance
             */
            public static create(properties?: flyteidl.admin.IDeleteTaskRequest): flyteidl.admin.DeleteTaskRequest;

            /**
             * Encodes the specified DeleteTaskRequest message. Does not implicitly {@link flyteidl.admin.DeleteTaskRequest.verify|verify} messages.
             * @param message DeleteTaskRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDeleteTaskRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DeleteTaskRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DeleteTaskRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.DeleteTaskRequest;

            /**
             * Verifies a DeleteTaskRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DeleteTaskResponse. */
        interface IDeleteTaskResponse {
        }

        /** Represents a DeleteTaskResponse. */
        class DeleteTaskResponse implements IDeleteTaskResponse {

            /**
             * Constructs a new DeleteTaskResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDeleteTaskResponse);

            /**
             * Creates a new DeleteTaskResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DeleteTaskResponse instance
             */
            public static create(properties?: flyteidl.admin.IDeleteTaskResponse): flyteidl.admin.DeleteTaskResponse;

            /**
             * Encodes the specified DeleteTaskResponse message. Does not implicitly {@link flyteidl.admin.DeleteTaskResponse.verify|verify} messages.
             * @param message DeleteTaskResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDeleteTaskResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DeleteTaskResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DeleteTaskResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.DeleteTaskResponse;

            /**
             * Verifies a DeleteTaskResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Agent. */
        interface IAgent {

            /** Agent name */
            name?: (string|null);

            /** Agent supportedTaskTypes */
            supportedTaskTypes?: (string[]|null);

            /** Agent isSync */
            isSync?: (boolean|null);

            /** Agent supportedTaskCategories */
            supportedTaskCategories?: (flyteidl.admin.ITaskCategory[]|null);
        }

        /** Represents an Agent. */
        class Agent implements IAgent {

            /**
             * Constructs a new Agent.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IAgent);

            /** Agent name. */
            public name: string;

            /** Agent supportedTaskTypes. */
            public supportedTaskTypes: string[];

            /** Agent isSync. */
            public isSync: boolean;

            /** Agent supportedTaskCategories. */
            public supportedTaskCategories: flyteidl.admin.ITaskCategory[];

            /**
             * Creates a new Agent instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Agent instance
             */
            public static create(properties?: flyteidl.admin.IAgent): flyteidl.admin.Agent;

            /**
             * Encodes the specified Agent message. Does not implicitly {@link flyteidl.admin.Agent.verify|verify} messages.
             * @param message Agent message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IAgent, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Agent message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Agent
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Agent;

            /**
             * Verifies an Agent message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskCategory. */
        interface ITaskCategory {

            /** TaskCategory name */
            name?: (string|null);

            /** TaskCategory version */
            version?: (number|null);
        }

        /** Represents a TaskCategory. */
        class TaskCategory implements ITaskCategory {

            /**
             * Constructs a new TaskCategory.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ITaskCategory);

            /** TaskCategory name. */
            public name: string;

            /** TaskCategory version. */
            public version: number;

            /**
             * Creates a new TaskCategory instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskCategory instance
             */
            public static create(properties?: flyteidl.admin.ITaskCategory): flyteidl.admin.TaskCategory;

            /**
             * Encodes the specified TaskCategory message. Does not implicitly {@link flyteidl.admin.TaskCategory.verify|verify} messages.
             * @param message TaskCategory message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ITaskCategory, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskCategory message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskCategory
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.TaskCategory;

            /**
             * Verifies a TaskCategory message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetAgentRequest. */
        interface IGetAgentRequest {

            /** GetAgentRequest name */
            name?: (string|null);
        }

        /** Represents a GetAgentRequest. */
        class GetAgentRequest implements IGetAgentRequest {

            /**
             * Constructs a new GetAgentRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetAgentRequest);

            /** GetAgentRequest name. */
            public name: string;

            /**
             * Creates a new GetAgentRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetAgentRequest instance
             */
            public static create(properties?: flyteidl.admin.IGetAgentRequest): flyteidl.admin.GetAgentRequest;

            /**
             * Encodes the specified GetAgentRequest message. Does not implicitly {@link flyteidl.admin.GetAgentRequest.verify|verify} messages.
             * @param message GetAgentRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetAgentRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetAgentRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetAgentRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetAgentRequest;

            /**
             * Verifies a GetAgentRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetAgentResponse. */
        interface IGetAgentResponse {

            /** GetAgentResponse agent */
            agent?: (flyteidl.admin.IAgent|null);
        }

        /** Represents a GetAgentResponse. */
        class GetAgentResponse implements IGetAgentResponse {

            /**
             * Constructs a new GetAgentResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetAgentResponse);

            /** GetAgentResponse agent. */
            public agent?: (flyteidl.admin.IAgent|null);

            /**
             * Creates a new GetAgentResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetAgentResponse instance
             */
            public static create(properties?: flyteidl.admin.IGetAgentResponse): flyteidl.admin.GetAgentResponse;

            /**
             * Encodes the specified GetAgentResponse message. Does not implicitly {@link flyteidl.admin.GetAgentResponse.verify|verify} messages.
             * @param message GetAgentResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetAgentResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetAgentResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetAgentResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetAgentResponse;

            /**
             * Verifies a GetAgentResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ListAgentsRequest. */
        interface IListAgentsRequest {
        }

        /** Represents a ListAgentsRequest. */
        class ListAgentsRequest implements IListAgentsRequest {

            /**
             * Constructs a new ListAgentsRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IListAgentsRequest);

            /**
             * Creates a new ListAgentsRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ListAgentsRequest instance
             */
            public static create(properties?: flyteidl.admin.IListAgentsRequest): flyteidl.admin.ListAgentsRequest;

            /**
             * Encodes the specified ListAgentsRequest message. Does not implicitly {@link flyteidl.admin.ListAgentsRequest.verify|verify} messages.
             * @param message ListAgentsRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IListAgentsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ListAgentsRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ListAgentsRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ListAgentsRequest;

            /**
             * Verifies a ListAgentsRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ListAgentsResponse. */
        interface IListAgentsResponse {

            /** ListAgentsResponse agents */
            agents?: (flyteidl.admin.IAgent[]|null);
        }

        /** Represents a ListAgentsResponse. */
        class ListAgentsResponse implements IListAgentsResponse {

            /**
             * Constructs a new ListAgentsResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IListAgentsResponse);

            /** ListAgentsResponse agents. */
            public agents: flyteidl.admin.IAgent[];

            /**
             * Creates a new ListAgentsResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ListAgentsResponse instance
             */
            public static create(properties?: flyteidl.admin.IListAgentsResponse): flyteidl.admin.ListAgentsResponse;

            /**
             * Encodes the specified ListAgentsResponse message. Does not implicitly {@link flyteidl.admin.ListAgentsResponse.verify|verify} messages.
             * @param message ListAgentsResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IListAgentsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ListAgentsResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ListAgentsResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ListAgentsResponse;

            /**
             * Verifies a ListAgentsResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetTaskMetricsRequest. */
        interface IGetTaskMetricsRequest {

            /** GetTaskMetricsRequest taskType */
            taskType?: (string|null);

            /** GetTaskMetricsRequest resourceMeta */
            resourceMeta?: (Uint8Array|null);

            /** GetTaskMetricsRequest queries */
            queries?: (string[]|null);

            /** GetTaskMetricsRequest startTime */
            startTime?: (google.protobuf.ITimestamp|null);

            /** GetTaskMetricsRequest endTime */
            endTime?: (google.protobuf.ITimestamp|null);

            /** GetTaskMetricsRequest step */
            step?: (google.protobuf.IDuration|null);

            /** GetTaskMetricsRequest taskCategory */
            taskCategory?: (flyteidl.admin.ITaskCategory|null);
        }

        /** Represents a GetTaskMetricsRequest. */
        class GetTaskMetricsRequest implements IGetTaskMetricsRequest {

            /**
             * Constructs a new GetTaskMetricsRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetTaskMetricsRequest);

            /** GetTaskMetricsRequest taskType. */
            public taskType: string;

            /** GetTaskMetricsRequest resourceMeta. */
            public resourceMeta: Uint8Array;

            /** GetTaskMetricsRequest queries. */
            public queries: string[];

            /** GetTaskMetricsRequest startTime. */
            public startTime?: (google.protobuf.ITimestamp|null);

            /** GetTaskMetricsRequest endTime. */
            public endTime?: (google.protobuf.ITimestamp|null);

            /** GetTaskMetricsRequest step. */
            public step?: (google.protobuf.IDuration|null);

            /** GetTaskMetricsRequest taskCategory. */
            public taskCategory?: (flyteidl.admin.ITaskCategory|null);

            /**
             * Creates a new GetTaskMetricsRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetTaskMetricsRequest instance
             */
            public static create(properties?: flyteidl.admin.IGetTaskMetricsRequest): flyteidl.admin.GetTaskMetricsRequest;

            /**
             * Encodes the specified GetTaskMetricsRequest message. Does not implicitly {@link flyteidl.admin.GetTaskMetricsRequest.verify|verify} messages.
             * @param message GetTaskMetricsRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetTaskMetricsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetTaskMetricsRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetTaskMetricsRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetTaskMetricsRequest;

            /**
             * Verifies a GetTaskMetricsRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetTaskMetricsResponse. */
        interface IGetTaskMetricsResponse {

            /** GetTaskMetricsResponse results */
            results?: (flyteidl.core.IExecutionMetricResult[]|null);
        }

        /** Represents a GetTaskMetricsResponse. */
        class GetTaskMetricsResponse implements IGetTaskMetricsResponse {

            /**
             * Constructs a new GetTaskMetricsResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetTaskMetricsResponse);

            /** GetTaskMetricsResponse results. */
            public results: flyteidl.core.IExecutionMetricResult[];

            /**
             * Creates a new GetTaskMetricsResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetTaskMetricsResponse instance
             */
            public static create(properties?: flyteidl.admin.IGetTaskMetricsResponse): flyteidl.admin.GetTaskMetricsResponse;

            /**
             * Encodes the specified GetTaskMetricsResponse message. Does not implicitly {@link flyteidl.admin.GetTaskMetricsResponse.verify|verify} messages.
             * @param message GetTaskMetricsResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetTaskMetricsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetTaskMetricsResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetTaskMetricsResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetTaskMetricsResponse;

            /**
             * Verifies a GetTaskMetricsResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetTaskLogsRequest. */
        interface IGetTaskLogsRequest {

            /** GetTaskLogsRequest taskType */
            taskType?: (string|null);

            /** GetTaskLogsRequest resourceMeta */
            resourceMeta?: (Uint8Array|null);

            /** GetTaskLogsRequest lines */
            lines?: (Long|null);

            /** GetTaskLogsRequest token */
            token?: (string|null);

            /** GetTaskLogsRequest taskCategory */
            taskCategory?: (flyteidl.admin.ITaskCategory|null);
        }

        /** Represents a GetTaskLogsRequest. */
        class GetTaskLogsRequest implements IGetTaskLogsRequest {

            /**
             * Constructs a new GetTaskLogsRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetTaskLogsRequest);

            /** GetTaskLogsRequest taskType. */
            public taskType: string;

            /** GetTaskLogsRequest resourceMeta. */
            public resourceMeta: Uint8Array;

            /** GetTaskLogsRequest lines. */
            public lines: Long;

            /** GetTaskLogsRequest token. */
            public token: string;

            /** GetTaskLogsRequest taskCategory. */
            public taskCategory?: (flyteidl.admin.ITaskCategory|null);

            /**
             * Creates a new GetTaskLogsRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetTaskLogsRequest instance
             */
            public static create(properties?: flyteidl.admin.IGetTaskLogsRequest): flyteidl.admin.GetTaskLogsRequest;

            /**
             * Encodes the specified GetTaskLogsRequest message. Does not implicitly {@link flyteidl.admin.GetTaskLogsRequest.verify|verify} messages.
             * @param message GetTaskLogsRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetTaskLogsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetTaskLogsRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetTaskLogsRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetTaskLogsRequest;

            /**
             * Verifies a GetTaskLogsRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetTaskLogsResponseHeader. */
        interface IGetTaskLogsResponseHeader {

            /** GetTaskLogsResponseHeader token */
            token?: (string|null);
        }

        /** Represents a GetTaskLogsResponseHeader. */
        class GetTaskLogsResponseHeader implements IGetTaskLogsResponseHeader {

            /**
             * Constructs a new GetTaskLogsResponseHeader.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetTaskLogsResponseHeader);

            /** GetTaskLogsResponseHeader token. */
            public token: string;

            /**
             * Creates a new GetTaskLogsResponseHeader instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetTaskLogsResponseHeader instance
             */
            public static create(properties?: flyteidl.admin.IGetTaskLogsResponseHeader): flyteidl.admin.GetTaskLogsResponseHeader;

            /**
             * Encodes the specified GetTaskLogsResponseHeader message. Does not implicitly {@link flyteidl.admin.GetTaskLogsResponseHeader.verify|verify} messages.
             * @param message GetTaskLogsResponseHeader message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetTaskLogsResponseHeader, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetTaskLogsResponseHeader message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetTaskLogsResponseHeader
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetTaskLogsResponseHeader;

            /**
             * Verifies a GetTaskLogsResponseHeader message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetTaskLogsResponseBody. */
        interface IGetTaskLogsResponseBody {

            /** GetTaskLogsResponseBody results */
            results?: (string[]|null);
        }

        /** Represents a GetTaskLogsResponseBody. */
        class GetTaskLogsResponseBody implements IGetTaskLogsResponseBody {

            /**
             * Constructs a new GetTaskLogsResponseBody.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetTaskLogsResponseBody);

            /** GetTaskLogsResponseBody results. */
            public results: string[];

            /**
             * Creates a new GetTaskLogsResponseBody instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetTaskLogsResponseBody instance
             */
            public static create(properties?: flyteidl.admin.IGetTaskLogsResponseBody): flyteidl.admin.GetTaskLogsResponseBody;

            /**
             * Encodes the specified GetTaskLogsResponseBody message. Does not implicitly {@link flyteidl.admin.GetTaskLogsResponseBody.verify|verify} messages.
             * @param message GetTaskLogsResponseBody message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetTaskLogsResponseBody, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetTaskLogsResponseBody message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetTaskLogsResponseBody
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetTaskLogsResponseBody;

            /**
             * Verifies a GetTaskLogsResponseBody message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetTaskLogsResponse. */
        interface IGetTaskLogsResponse {

            /** GetTaskLogsResponse header */
            header?: (flyteidl.admin.IGetTaskLogsResponseHeader|null);

            /** GetTaskLogsResponse body */
            body?: (flyteidl.admin.IGetTaskLogsResponseBody|null);
        }

        /** Represents a GetTaskLogsResponse. */
        class GetTaskLogsResponse implements IGetTaskLogsResponse {

            /**
             * Constructs a new GetTaskLogsResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetTaskLogsResponse);

            /** GetTaskLogsResponse header. */
            public header?: (flyteidl.admin.IGetTaskLogsResponseHeader|null);

            /** GetTaskLogsResponse body. */
            public body?: (flyteidl.admin.IGetTaskLogsResponseBody|null);

            /** GetTaskLogsResponse part. */
            public part?: ("header"|"body");

            /**
             * Creates a new GetTaskLogsResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetTaskLogsResponse instance
             */
            public static create(properties?: flyteidl.admin.IGetTaskLogsResponse): flyteidl.admin.GetTaskLogsResponse;

            /**
             * Encodes the specified GetTaskLogsResponse message. Does not implicitly {@link flyteidl.admin.GetTaskLogsResponse.verify|verify} messages.
             * @param message GetTaskLogsResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetTaskLogsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetTaskLogsResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetTaskLogsResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetTaskLogsResponse;

            /**
             * Verifies a GetTaskLogsResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ClusterAssignment. */
        interface IClusterAssignment {

            /** ClusterAssignment clusterPoolName */
            clusterPoolName?: (string|null);
        }

        /** Represents a ClusterAssignment. */
        class ClusterAssignment implements IClusterAssignment {

            /**
             * Constructs a new ClusterAssignment.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IClusterAssignment);

            /** ClusterAssignment clusterPoolName. */
            public clusterPoolName: string;

            /**
             * Creates a new ClusterAssignment instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ClusterAssignment instance
             */
            public static create(properties?: flyteidl.admin.IClusterAssignment): flyteidl.admin.ClusterAssignment;

            /**
             * Encodes the specified ClusterAssignment message. Does not implicitly {@link flyteidl.admin.ClusterAssignment.verify|verify} messages.
             * @param message ClusterAssignment message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IClusterAssignment, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ClusterAssignment message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ClusterAssignment
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ClusterAssignment;

            /**
             * Verifies a ClusterAssignment message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NamedEntityIdentifier. */
        interface INamedEntityIdentifier {

            /** NamedEntityIdentifier project */
            project?: (string|null);

            /** NamedEntityIdentifier domain */
            domain?: (string|null);

            /** NamedEntityIdentifier name */
            name?: (string|null);

            /** NamedEntityIdentifier org */
            org?: (string|null);
        }

        /** Represents a NamedEntityIdentifier. */
        class NamedEntityIdentifier implements INamedEntityIdentifier {

            /**
             * Constructs a new NamedEntityIdentifier.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityIdentifier);

            /** NamedEntityIdentifier project. */
            public project: string;

            /** NamedEntityIdentifier domain. */
            public domain: string;

            /** NamedEntityIdentifier name. */
            public name: string;

            /** NamedEntityIdentifier org. */
            public org: string;

            /**
             * Creates a new NamedEntityIdentifier instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityIdentifier instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityIdentifier): flyteidl.admin.NamedEntityIdentifier;

            /**
             * Encodes the specified NamedEntityIdentifier message. Does not implicitly {@link flyteidl.admin.NamedEntityIdentifier.verify|verify} messages.
             * @param message NamedEntityIdentifier message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityIdentifier, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityIdentifier message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityIdentifier
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityIdentifier;

            /**
             * Verifies a NamedEntityIdentifier message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** NamedEntityState enum. */
        enum NamedEntityState {
            NAMED_ENTITY_ACTIVE = 0,
            NAMED_ENTITY_ARCHIVED = 1,
            SYSTEM_GENERATED = 2
        }

        /** Properties of a NamedEntityMetadata. */
        interface INamedEntityMetadata {

            /** NamedEntityMetadata description */
            description?: (string|null);

            /** NamedEntityMetadata state */
            state?: (flyteidl.admin.NamedEntityState|null);
        }

        /** Represents a NamedEntityMetadata. */
        class NamedEntityMetadata implements INamedEntityMetadata {

            /**
             * Constructs a new NamedEntityMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityMetadata);

            /** NamedEntityMetadata description. */
            public description: string;

            /** NamedEntityMetadata state. */
            public state: flyteidl.admin.NamedEntityState;

            /**
             * Creates a new NamedEntityMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityMetadata instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityMetadata): flyteidl.admin.NamedEntityMetadata;

            /**
             * Encodes the specified NamedEntityMetadata message. Does not implicitly {@link flyteidl.admin.NamedEntityMetadata.verify|verify} messages.
             * @param message NamedEntityMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityMetadata;

            /**
             * Verifies a NamedEntityMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NamedEntity. */
        interface INamedEntity {

            /** NamedEntity resourceType */
            resourceType?: (flyteidl.core.ResourceType|null);

            /** NamedEntity id */
            id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /** NamedEntity metadata */
            metadata?: (flyteidl.admin.INamedEntityMetadata|null);
        }

        /** Represents a NamedEntity. */
        class NamedEntity implements INamedEntity {

            /**
             * Constructs a new NamedEntity.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntity);

            /** NamedEntity resourceType. */
            public resourceType: flyteidl.core.ResourceType;

            /** NamedEntity id. */
            public id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /** NamedEntity metadata. */
            public metadata?: (flyteidl.admin.INamedEntityMetadata|null);

            /**
             * Creates a new NamedEntity instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntity instance
             */
            public static create(properties?: flyteidl.admin.INamedEntity): flyteidl.admin.NamedEntity;

            /**
             * Encodes the specified NamedEntity message. Does not implicitly {@link flyteidl.admin.NamedEntity.verify|verify} messages.
             * @param message NamedEntity message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntity, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntity message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntity
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntity;

            /**
             * Verifies a NamedEntity message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Sort. */
        interface ISort {

            /** Sort key */
            key?: (string|null);

            /** Sort direction */
            direction?: (flyteidl.admin.Sort.Direction|null);
        }

        /** Represents a Sort. */
        class Sort implements ISort {

            /**
             * Constructs a new Sort.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ISort);

            /** Sort key. */
            public key: string;

            /** Sort direction. */
            public direction: flyteidl.admin.Sort.Direction;

            /**
             * Creates a new Sort instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Sort instance
             */
            public static create(properties?: flyteidl.admin.ISort): flyteidl.admin.Sort;

            /**
             * Encodes the specified Sort message. Does not implicitly {@link flyteidl.admin.Sort.verify|verify} messages.
             * @param message Sort message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ISort, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Sort message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Sort
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Sort;

            /**
             * Verifies a Sort message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace Sort {

            /** Direction enum. */
            enum Direction {
                DESCENDING = 0,
                ASCENDING = 1
            }
        }

        /** Properties of a NamedEntityIdentifierListRequest. */
        interface INamedEntityIdentifierListRequest {

            /** NamedEntityIdentifierListRequest project */
            project?: (string|null);

            /** NamedEntityIdentifierListRequest domain */
            domain?: (string|null);

            /** NamedEntityIdentifierListRequest limit */
            limit?: (number|null);

            /** NamedEntityIdentifierListRequest token */
            token?: (string|null);

            /** NamedEntityIdentifierListRequest sortBy */
            sortBy?: (flyteidl.admin.ISort|null);

            /** NamedEntityIdentifierListRequest filters */
            filters?: (string|null);

            /** NamedEntityIdentifierListRequest org */
            org?: (string|null);
        }

        /** Represents a NamedEntityIdentifierListRequest. */
        class NamedEntityIdentifierListRequest implements INamedEntityIdentifierListRequest {

            /**
             * Constructs a new NamedEntityIdentifierListRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityIdentifierListRequest);

            /** NamedEntityIdentifierListRequest project. */
            public project: string;

            /** NamedEntityIdentifierListRequest domain. */
            public domain: string;

            /** NamedEntityIdentifierListRequest limit. */
            public limit: number;

            /** NamedEntityIdentifierListRequest token. */
            public token: string;

            /** NamedEntityIdentifierListRequest sortBy. */
            public sortBy?: (flyteidl.admin.ISort|null);

            /** NamedEntityIdentifierListRequest filters. */
            public filters: string;

            /** NamedEntityIdentifierListRequest org. */
            public org: string;

            /**
             * Creates a new NamedEntityIdentifierListRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityIdentifierListRequest instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityIdentifierListRequest): flyteidl.admin.NamedEntityIdentifierListRequest;

            /**
             * Encodes the specified NamedEntityIdentifierListRequest message. Does not implicitly {@link flyteidl.admin.NamedEntityIdentifierListRequest.verify|verify} messages.
             * @param message NamedEntityIdentifierListRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityIdentifierListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityIdentifierListRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityIdentifierListRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityIdentifierListRequest;

            /**
             * Verifies a NamedEntityIdentifierListRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NamedEntityListRequest. */
        interface INamedEntityListRequest {

            /** NamedEntityListRequest resourceType */
            resourceType?: (flyteidl.core.ResourceType|null);

            /** NamedEntityListRequest project */
            project?: (string|null);

            /** NamedEntityListRequest domain */
            domain?: (string|null);

            /** NamedEntityListRequest limit */
            limit?: (number|null);

            /** NamedEntityListRequest token */
            token?: (string|null);

            /** NamedEntityListRequest sortBy */
            sortBy?: (flyteidl.admin.ISort|null);

            /** NamedEntityListRequest filters */
            filters?: (string|null);

            /** NamedEntityListRequest org */
            org?: (string|null);
        }

        /** Represents a NamedEntityListRequest. */
        class NamedEntityListRequest implements INamedEntityListRequest {

            /**
             * Constructs a new NamedEntityListRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityListRequest);

            /** NamedEntityListRequest resourceType. */
            public resourceType: flyteidl.core.ResourceType;

            /** NamedEntityListRequest project. */
            public project: string;

            /** NamedEntityListRequest domain. */
            public domain: string;

            /** NamedEntityListRequest limit. */
            public limit: number;

            /** NamedEntityListRequest token. */
            public token: string;

            /** NamedEntityListRequest sortBy. */
            public sortBy?: (flyteidl.admin.ISort|null);

            /** NamedEntityListRequest filters. */
            public filters: string;

            /** NamedEntityListRequest org. */
            public org: string;

            /**
             * Creates a new NamedEntityListRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityListRequest instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityListRequest): flyteidl.admin.NamedEntityListRequest;

            /**
             * Encodes the specified NamedEntityListRequest message. Does not implicitly {@link flyteidl.admin.NamedEntityListRequest.verify|verify} messages.
             * @param message NamedEntityListRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityListRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityListRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityListRequest;

            /**
             * Verifies a NamedEntityListRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NamedEntityIdentifierList. */
        interface INamedEntityIdentifierList {

            /** NamedEntityIdentifierList entities */
            entities?: (flyteidl.admin.INamedEntityIdentifier[]|null);

            /** NamedEntityIdentifierList token */
            token?: (string|null);
        }

        /** Represents a NamedEntityIdentifierList. */
        class NamedEntityIdentifierList implements INamedEntityIdentifierList {

            /**
             * Constructs a new NamedEntityIdentifierList.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityIdentifierList);

            /** NamedEntityIdentifierList entities. */
            public entities: flyteidl.admin.INamedEntityIdentifier[];

            /** NamedEntityIdentifierList token. */
            public token: string;

            /**
             * Creates a new NamedEntityIdentifierList instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityIdentifierList instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityIdentifierList): flyteidl.admin.NamedEntityIdentifierList;

            /**
             * Encodes the specified NamedEntityIdentifierList message. Does not implicitly {@link flyteidl.admin.NamedEntityIdentifierList.verify|verify} messages.
             * @param message NamedEntityIdentifierList message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityIdentifierList, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityIdentifierList message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityIdentifierList
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityIdentifierList;

            /**
             * Verifies a NamedEntityIdentifierList message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NamedEntityList. */
        interface INamedEntityList {

            /** NamedEntityList entities */
            entities?: (flyteidl.admin.INamedEntity[]|null);

            /** NamedEntityList token */
            token?: (string|null);
        }

        /** Represents a NamedEntityList. */
        class NamedEntityList implements INamedEntityList {

            /**
             * Constructs a new NamedEntityList.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityList);

            /** NamedEntityList entities. */
            public entities: flyteidl.admin.INamedEntity[];

            /** NamedEntityList token. */
            public token: string;

            /**
             * Creates a new NamedEntityList instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityList instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityList): flyteidl.admin.NamedEntityList;

            /**
             * Encodes the specified NamedEntityList message. Does not implicitly {@link flyteidl.admin.NamedEntityList.verify|verify} messages.
             * @param message NamedEntityList message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityList, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityList message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityList
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityList;

            /**
             * Verifies a NamedEntityList message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NamedEntityGetRequest. */
        interface INamedEntityGetRequest {

            /** NamedEntityGetRequest resourceType */
            resourceType?: (flyteidl.core.ResourceType|null);

            /** NamedEntityGetRequest id */
            id?: (flyteidl.admin.INamedEntityIdentifier|null);
        }

        /** Represents a NamedEntityGetRequest. */
        class NamedEntityGetRequest implements INamedEntityGetRequest {

            /**
             * Constructs a new NamedEntityGetRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityGetRequest);

            /** NamedEntityGetRequest resourceType. */
            public resourceType: flyteidl.core.ResourceType;

            /** NamedEntityGetRequest id. */
            public id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /**
             * Creates a new NamedEntityGetRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityGetRequest instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityGetRequest): flyteidl.admin.NamedEntityGetRequest;

            /**
             * Encodes the specified NamedEntityGetRequest message. Does not implicitly {@link flyteidl.admin.NamedEntityGetRequest.verify|verify} messages.
             * @param message NamedEntityGetRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityGetRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityGetRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityGetRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityGetRequest;

            /**
             * Verifies a NamedEntityGetRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NamedEntityUpdateRequest. */
        interface INamedEntityUpdateRequest {

            /** NamedEntityUpdateRequest resourceType */
            resourceType?: (flyteidl.core.ResourceType|null);

            /** NamedEntityUpdateRequest id */
            id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /** NamedEntityUpdateRequest metadata */
            metadata?: (flyteidl.admin.INamedEntityMetadata|null);
        }

        /** Represents a NamedEntityUpdateRequest. */
        class NamedEntityUpdateRequest implements INamedEntityUpdateRequest {

            /**
             * Constructs a new NamedEntityUpdateRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityUpdateRequest);

            /** NamedEntityUpdateRequest resourceType. */
            public resourceType: flyteidl.core.ResourceType;

            /** NamedEntityUpdateRequest id. */
            public id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /** NamedEntityUpdateRequest metadata. */
            public metadata?: (flyteidl.admin.INamedEntityMetadata|null);

            /**
             * Creates a new NamedEntityUpdateRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityUpdateRequest instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityUpdateRequest): flyteidl.admin.NamedEntityUpdateRequest;

            /**
             * Encodes the specified NamedEntityUpdateRequest message. Does not implicitly {@link flyteidl.admin.NamedEntityUpdateRequest.verify|verify} messages.
             * @param message NamedEntityUpdateRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityUpdateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityUpdateRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityUpdateRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityUpdateRequest;

            /**
             * Verifies a NamedEntityUpdateRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NamedEntityUpdateResponse. */
        interface INamedEntityUpdateResponse {
        }

        /** Represents a NamedEntityUpdateResponse. */
        class NamedEntityUpdateResponse implements INamedEntityUpdateResponse {

            /**
             * Constructs a new NamedEntityUpdateResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INamedEntityUpdateResponse);

            /**
             * Creates a new NamedEntityUpdateResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NamedEntityUpdateResponse instance
             */
            public static create(properties?: flyteidl.admin.INamedEntityUpdateResponse): flyteidl.admin.NamedEntityUpdateResponse;

            /**
             * Encodes the specified NamedEntityUpdateResponse message. Does not implicitly {@link flyteidl.admin.NamedEntityUpdateResponse.verify|verify} messages.
             * @param message NamedEntityUpdateResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INamedEntityUpdateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NamedEntityUpdateResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NamedEntityUpdateResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NamedEntityUpdateResponse;

            /**
             * Verifies a NamedEntityUpdateResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ObjectGetRequest. */
        interface IObjectGetRequest {

            /** ObjectGetRequest id */
            id?: (flyteidl.core.IIdentifier|null);
        }

        /** Represents an ObjectGetRequest. */
        class ObjectGetRequest implements IObjectGetRequest {

            /**
             * Constructs a new ObjectGetRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IObjectGetRequest);

            /** ObjectGetRequest id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /**
             * Creates a new ObjectGetRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ObjectGetRequest instance
             */
            public static create(properties?: flyteidl.admin.IObjectGetRequest): flyteidl.admin.ObjectGetRequest;

            /**
             * Encodes the specified ObjectGetRequest message. Does not implicitly {@link flyteidl.admin.ObjectGetRequest.verify|verify} messages.
             * @param message ObjectGetRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IObjectGetRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ObjectGetRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ObjectGetRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ObjectGetRequest;

            /**
             * Verifies an ObjectGetRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ResourceListRequest. */
        interface IResourceListRequest {

            /** ResourceListRequest id */
            id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /** ResourceListRequest limit */
            limit?: (number|null);

            /** ResourceListRequest token */
            token?: (string|null);

            /** ResourceListRequest filters */
            filters?: (string|null);

            /** ResourceListRequest sortBy */
            sortBy?: (flyteidl.admin.ISort|null);
        }

        /** Represents a ResourceListRequest. */
        class ResourceListRequest implements IResourceListRequest {

            /**
             * Constructs a new ResourceListRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IResourceListRequest);

            /** ResourceListRequest id. */
            public id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /** ResourceListRequest limit. */
            public limit: number;

            /** ResourceListRequest token. */
            public token: string;

            /** ResourceListRequest filters. */
            public filters: string;

            /** ResourceListRequest sortBy. */
            public sortBy?: (flyteidl.admin.ISort|null);

            /**
             * Creates a new ResourceListRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ResourceListRequest instance
             */
            public static create(properties?: flyteidl.admin.IResourceListRequest): flyteidl.admin.ResourceListRequest;

            /**
             * Encodes the specified ResourceListRequest message. Does not implicitly {@link flyteidl.admin.ResourceListRequest.verify|verify} messages.
             * @param message ResourceListRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IResourceListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ResourceListRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ResourceListRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ResourceListRequest;

            /**
             * Verifies a ResourceListRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an EmailNotification. */
        interface IEmailNotification {

            /** EmailNotification recipientsEmail */
            recipientsEmail?: (string[]|null);
        }

        /** Represents an EmailNotification. */
        class EmailNotification implements IEmailNotification {

            /**
             * Constructs a new EmailNotification.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IEmailNotification);

            /** EmailNotification recipientsEmail. */
            public recipientsEmail: string[];

            /**
             * Creates a new EmailNotification instance using the specified properties.
             * @param [properties] Properties to set
             * @returns EmailNotification instance
             */
            public static create(properties?: flyteidl.admin.IEmailNotification): flyteidl.admin.EmailNotification;

            /**
             * Encodes the specified EmailNotification message. Does not implicitly {@link flyteidl.admin.EmailNotification.verify|verify} messages.
             * @param message EmailNotification message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IEmailNotification, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EmailNotification message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EmailNotification
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.EmailNotification;

            /**
             * Verifies an EmailNotification message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a PagerDutyNotification. */
        interface IPagerDutyNotification {

            /** PagerDutyNotification recipientsEmail */
            recipientsEmail?: (string[]|null);
        }

        /** Represents a PagerDutyNotification. */
        class PagerDutyNotification implements IPagerDutyNotification {

            /**
             * Constructs a new PagerDutyNotification.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IPagerDutyNotification);

            /** PagerDutyNotification recipientsEmail. */
            public recipientsEmail: string[];

            /**
             * Creates a new PagerDutyNotification instance using the specified properties.
             * @param [properties] Properties to set
             * @returns PagerDutyNotification instance
             */
            public static create(properties?: flyteidl.admin.IPagerDutyNotification): flyteidl.admin.PagerDutyNotification;

            /**
             * Encodes the specified PagerDutyNotification message. Does not implicitly {@link flyteidl.admin.PagerDutyNotification.verify|verify} messages.
             * @param message PagerDutyNotification message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IPagerDutyNotification, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a PagerDutyNotification message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns PagerDutyNotification
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.PagerDutyNotification;

            /**
             * Verifies a PagerDutyNotification message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a SlackNotification. */
        interface ISlackNotification {

            /** SlackNotification recipientsEmail */
            recipientsEmail?: (string[]|null);
        }

        /** Represents a SlackNotification. */
        class SlackNotification implements ISlackNotification {

            /**
             * Constructs a new SlackNotification.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ISlackNotification);

            /** SlackNotification recipientsEmail. */
            public recipientsEmail: string[];

            /**
             * Creates a new SlackNotification instance using the specified properties.
             * @param [properties] Properties to set
             * @returns SlackNotification instance
             */
            public static create(properties?: flyteidl.admin.ISlackNotification): flyteidl.admin.SlackNotification;

            /**
             * Encodes the specified SlackNotification message. Does not implicitly {@link flyteidl.admin.SlackNotification.verify|verify} messages.
             * @param message SlackNotification message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ISlackNotification, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SlackNotification message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SlackNotification
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.SlackNotification;

            /**
             * Verifies a SlackNotification message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Notification. */
        interface INotification {

            /** Notification phases */
            phases?: (flyteidl.core.WorkflowExecution.Phase[]|null);

            /** Notification email */
            email?: (flyteidl.admin.IEmailNotification|null);

            /** Notification pagerDuty */
            pagerDuty?: (flyteidl.admin.IPagerDutyNotification|null);

            /** Notification slack */
            slack?: (flyteidl.admin.ISlackNotification|null);
        }

        /** Represents a Notification. */
        class Notification implements INotification {

            /**
             * Constructs a new Notification.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INotification);

            /** Notification phases. */
            public phases: flyteidl.core.WorkflowExecution.Phase[];

            /** Notification email. */
            public email?: (flyteidl.admin.IEmailNotification|null);

            /** Notification pagerDuty. */
            public pagerDuty?: (flyteidl.admin.IPagerDutyNotification|null);

            /** Notification slack. */
            public slack?: (flyteidl.admin.ISlackNotification|null);

            /** Notification type. */
            public type?: ("email"|"pagerDuty"|"slack");

            /**
             * Creates a new Notification instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Notification instance
             */
            public static create(properties?: flyteidl.admin.INotification): flyteidl.admin.Notification;

            /**
             * Encodes the specified Notification message. Does not implicitly {@link flyteidl.admin.Notification.verify|verify} messages.
             * @param message Notification message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INotification, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Notification message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Notification
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Notification;

            /**
             * Verifies a Notification message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an UrlBlob. */
        interface IUrlBlob {

            /** UrlBlob url */
            url?: (string|null);

            /** UrlBlob bytes */
            bytes?: (Long|null);
        }

        /** Represents an UrlBlob. */
        class UrlBlob implements IUrlBlob {

            /**
             * Constructs a new UrlBlob.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IUrlBlob);

            /** UrlBlob url. */
            public url: string;

            /** UrlBlob bytes. */
            public bytes: Long;

            /**
             * Creates a new UrlBlob instance using the specified properties.
             * @param [properties] Properties to set
             * @returns UrlBlob instance
             */
            public static create(properties?: flyteidl.admin.IUrlBlob): flyteidl.admin.UrlBlob;

            /**
             * Encodes the specified UrlBlob message. Does not implicitly {@link flyteidl.admin.UrlBlob.verify|verify} messages.
             * @param message UrlBlob message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IUrlBlob, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an UrlBlob message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns UrlBlob
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.UrlBlob;

            /**
             * Verifies an UrlBlob message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Labels. */
        interface ILabels {

            /** Labels values */
            values?: ({ [k: string]: string }|null);
        }

        /** Represents a Labels. */
        class Labels implements ILabels {

            /**
             * Constructs a new Labels.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILabels);

            /** Labels values. */
            public values: { [k: string]: string };

            /**
             * Creates a new Labels instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Labels instance
             */
            public static create(properties?: flyteidl.admin.ILabels): flyteidl.admin.Labels;

            /**
             * Encodes the specified Labels message. Does not implicitly {@link flyteidl.admin.Labels.verify|verify} messages.
             * @param message Labels message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILabels, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Labels message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Labels
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Labels;

            /**
             * Verifies a Labels message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Annotations. */
        interface IAnnotations {

            /** Annotations values */
            values?: ({ [k: string]: string }|null);
        }

        /** Represents an Annotations. */
        class Annotations implements IAnnotations {

            /**
             * Constructs a new Annotations.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IAnnotations);

            /** Annotations values. */
            public values: { [k: string]: string };

            /**
             * Creates a new Annotations instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Annotations instance
             */
            public static create(properties?: flyteidl.admin.IAnnotations): flyteidl.admin.Annotations;

            /**
             * Encodes the specified Annotations message. Does not implicitly {@link flyteidl.admin.Annotations.verify|verify} messages.
             * @param message Annotations message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IAnnotations, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Annotations message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Annotations
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Annotations;

            /**
             * Verifies an Annotations message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Envs. */
        interface IEnvs {

            /** Envs values */
            values?: (flyteidl.core.IKeyValuePair[]|null);
        }

        /** Represents an Envs. */
        class Envs implements IEnvs {

            /**
             * Constructs a new Envs.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IEnvs);

            /** Envs values. */
            public values: flyteidl.core.IKeyValuePair[];

            /**
             * Creates a new Envs instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Envs instance
             */
            public static create(properties?: flyteidl.admin.IEnvs): flyteidl.admin.Envs;

            /**
             * Encodes the specified Envs message. Does not implicitly {@link flyteidl.admin.Envs.verify|verify} messages.
             * @param message Envs message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IEnvs, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Envs message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Envs
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Envs;

            /**
             * Verifies an Envs message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an AuthRole. */
        interface IAuthRole {

            /** AuthRole assumableIamRole */
            assumableIamRole?: (string|null);

            /** AuthRole kubernetesServiceAccount */
            kubernetesServiceAccount?: (string|null);
        }

        /** Represents an AuthRole. */
        class AuthRole implements IAuthRole {

            /**
             * Constructs a new AuthRole.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IAuthRole);

            /** AuthRole assumableIamRole. */
            public assumableIamRole: string;

            /** AuthRole kubernetesServiceAccount. */
            public kubernetesServiceAccount: string;

            /**
             * Creates a new AuthRole instance using the specified properties.
             * @param [properties] Properties to set
             * @returns AuthRole instance
             */
            public static create(properties?: flyteidl.admin.IAuthRole): flyteidl.admin.AuthRole;

            /**
             * Encodes the specified AuthRole message. Does not implicitly {@link flyteidl.admin.AuthRole.verify|verify} messages.
             * @param message AuthRole message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IAuthRole, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an AuthRole message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns AuthRole
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.AuthRole;

            /**
             * Verifies an AuthRole message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a RawOutputDataConfig. */
        interface IRawOutputDataConfig {

            /** RawOutputDataConfig outputLocationPrefix */
            outputLocationPrefix?: (string|null);
        }

        /** Represents a RawOutputDataConfig. */
        class RawOutputDataConfig implements IRawOutputDataConfig {

            /**
             * Constructs a new RawOutputDataConfig.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IRawOutputDataConfig);

            /** RawOutputDataConfig outputLocationPrefix. */
            public outputLocationPrefix: string;

            /**
             * Creates a new RawOutputDataConfig instance using the specified properties.
             * @param [properties] Properties to set
             * @returns RawOutputDataConfig instance
             */
            public static create(properties?: flyteidl.admin.IRawOutputDataConfig): flyteidl.admin.RawOutputDataConfig;

            /**
             * Encodes the specified RawOutputDataConfig message. Does not implicitly {@link flyteidl.admin.RawOutputDataConfig.verify|verify} messages.
             * @param message RawOutputDataConfig message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IRawOutputDataConfig, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a RawOutputDataConfig message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns RawOutputDataConfig
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.RawOutputDataConfig;

            /**
             * Verifies a RawOutputDataConfig message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a FlyteURLs. */
        interface IFlyteURLs {

            /** FlyteURLs inputs */
            inputs?: (string|null);

            /** FlyteURLs outputs */
            outputs?: (string|null);

            /** FlyteURLs deck */
            deck?: (string|null);
        }

        /** Represents a FlyteURLs. */
        class FlyteURLs implements IFlyteURLs {

            /**
             * Constructs a new FlyteURLs.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IFlyteURLs);

            /** FlyteURLs inputs. */
            public inputs: string;

            /** FlyteURLs outputs. */
            public outputs: string;

            /** FlyteURLs deck. */
            public deck: string;

            /**
             * Creates a new FlyteURLs instance using the specified properties.
             * @param [properties] Properties to set
             * @returns FlyteURLs instance
             */
            public static create(properties?: flyteidl.admin.IFlyteURLs): flyteidl.admin.FlyteURLs;

            /**
             * Encodes the specified FlyteURLs message. Does not implicitly {@link flyteidl.admin.FlyteURLs.verify|verify} messages.
             * @param message FlyteURLs message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IFlyteURLs, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a FlyteURLs message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns FlyteURLs
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.FlyteURLs;

            /**
             * Verifies a FlyteURLs message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DescriptionEntity. */
        interface IDescriptionEntity {

            /** DescriptionEntity id */
            id?: (flyteidl.core.IIdentifier|null);

            /** DescriptionEntity shortDescription */
            shortDescription?: (string|null);

            /** DescriptionEntity longDescription */
            longDescription?: (flyteidl.admin.IDescription|null);

            /** DescriptionEntity sourceCode */
            sourceCode?: (flyteidl.admin.ISourceCode|null);

            /** DescriptionEntity tags */
            tags?: (string[]|null);
        }

        /** Represents a DescriptionEntity. */
        class DescriptionEntity implements IDescriptionEntity {

            /**
             * Constructs a new DescriptionEntity.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDescriptionEntity);

            /** DescriptionEntity id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** DescriptionEntity shortDescription. */
            public shortDescription: string;

            /** DescriptionEntity longDescription. */
            public longDescription?: (flyteidl.admin.IDescription|null);

            /** DescriptionEntity sourceCode. */
            public sourceCode?: (flyteidl.admin.ISourceCode|null);

            /** DescriptionEntity tags. */
            public tags: string[];

            /**
             * Creates a new DescriptionEntity instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DescriptionEntity instance
             */
            public static create(properties?: flyteidl.admin.IDescriptionEntity): flyteidl.admin.DescriptionEntity;

            /**
             * Encodes the specified DescriptionEntity message. Does not implicitly {@link flyteidl.admin.DescriptionEntity.verify|verify} messages.
             * @param message DescriptionEntity message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDescriptionEntity, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DescriptionEntity message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DescriptionEntity
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.DescriptionEntity;

            /**
             * Verifies a DescriptionEntity message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** DescriptionFormat enum. */
        enum DescriptionFormat {
            DESCRIPTION_FORMAT_UNKNOWN = 0,
            DESCRIPTION_FORMAT_MARKDOWN = 1,
            DESCRIPTION_FORMAT_HTML = 2,
            DESCRIPTION_FORMAT_RST = 3
        }

        /** Properties of a Description. */
        interface IDescription {

            /** Description value */
            value?: (string|null);

            /** Description uri */
            uri?: (string|null);

            /** Description format */
            format?: (flyteidl.admin.DescriptionFormat|null);

            /** Description iconLink */
            iconLink?: (string|null);
        }

        /** Represents a Description. */
        class Description implements IDescription {

            /**
             * Constructs a new Description.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDescription);

            /** Description value. */
            public value: string;

            /** Description uri. */
            public uri: string;

            /** Description format. */
            public format: flyteidl.admin.DescriptionFormat;

            /** Description iconLink. */
            public iconLink: string;

            /** Description content. */
            public content?: ("value"|"uri");

            /**
             * Creates a new Description instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Description instance
             */
            public static create(properties?: flyteidl.admin.IDescription): flyteidl.admin.Description;

            /**
             * Encodes the specified Description message. Does not implicitly {@link flyteidl.admin.Description.verify|verify} messages.
             * @param message Description message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDescription, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Description message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Description
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Description;

            /**
             * Verifies a Description message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a SourceCode. */
        interface ISourceCode {

            /** SourceCode link */
            link?: (string|null);
        }

        /** Represents a SourceCode. */
        class SourceCode implements ISourceCode {

            /**
             * Constructs a new SourceCode.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ISourceCode);

            /** SourceCode link. */
            public link: string;

            /**
             * Creates a new SourceCode instance using the specified properties.
             * @param [properties] Properties to set
             * @returns SourceCode instance
             */
            public static create(properties?: flyteidl.admin.ISourceCode): flyteidl.admin.SourceCode;

            /**
             * Encodes the specified SourceCode message. Does not implicitly {@link flyteidl.admin.SourceCode.verify|verify} messages.
             * @param message SourceCode message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ISourceCode, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SourceCode message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SourceCode
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.SourceCode;

            /**
             * Verifies a SourceCode message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DescriptionEntityList. */
        interface IDescriptionEntityList {

            /** DescriptionEntityList descriptionEntities */
            descriptionEntities?: (flyteidl.admin.IDescriptionEntity[]|null);

            /** DescriptionEntityList token */
            token?: (string|null);
        }

        /** Represents a DescriptionEntityList. */
        class DescriptionEntityList implements IDescriptionEntityList {

            /**
             * Constructs a new DescriptionEntityList.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDescriptionEntityList);

            /** DescriptionEntityList descriptionEntities. */
            public descriptionEntities: flyteidl.admin.IDescriptionEntity[];

            /** DescriptionEntityList token. */
            public token: string;

            /**
             * Creates a new DescriptionEntityList instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DescriptionEntityList instance
             */
            public static create(properties?: flyteidl.admin.IDescriptionEntityList): flyteidl.admin.DescriptionEntityList;

            /**
             * Encodes the specified DescriptionEntityList message. Does not implicitly {@link flyteidl.admin.DescriptionEntityList.verify|verify} messages.
             * @param message DescriptionEntityList message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDescriptionEntityList, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DescriptionEntityList message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DescriptionEntityList
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.DescriptionEntityList;

            /**
             * Verifies a DescriptionEntityList message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DescriptionEntityListRequest. */
        interface IDescriptionEntityListRequest {

            /** DescriptionEntityListRequest resourceType */
            resourceType?: (flyteidl.core.ResourceType|null);

            /** DescriptionEntityListRequest id */
            id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /** DescriptionEntityListRequest limit */
            limit?: (number|null);

            /** DescriptionEntityListRequest token */
            token?: (string|null);

            /** DescriptionEntityListRequest filters */
            filters?: (string|null);

            /** DescriptionEntityListRequest sortBy */
            sortBy?: (flyteidl.admin.ISort|null);
        }

        /** Represents a DescriptionEntityListRequest. */
        class DescriptionEntityListRequest implements IDescriptionEntityListRequest {

            /**
             * Constructs a new DescriptionEntityListRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDescriptionEntityListRequest);

            /** DescriptionEntityListRequest resourceType. */
            public resourceType: flyteidl.core.ResourceType;

            /** DescriptionEntityListRequest id. */
            public id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /** DescriptionEntityListRequest limit. */
            public limit: number;

            /** DescriptionEntityListRequest token. */
            public token: string;

            /** DescriptionEntityListRequest filters. */
            public filters: string;

            /** DescriptionEntityListRequest sortBy. */
            public sortBy?: (flyteidl.admin.ISort|null);

            /**
             * Creates a new DescriptionEntityListRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DescriptionEntityListRequest instance
             */
            public static create(properties?: flyteidl.admin.IDescriptionEntityListRequest): flyteidl.admin.DescriptionEntityListRequest;

            /**
             * Encodes the specified DescriptionEntityListRequest message. Does not implicitly {@link flyteidl.admin.DescriptionEntityListRequest.verify|verify} messages.
             * @param message DescriptionEntityListRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDescriptionEntityListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DescriptionEntityListRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DescriptionEntityListRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.DescriptionEntityListRequest;

            /**
             * Verifies a DescriptionEntityListRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an EventErrorAlreadyInTerminalState. */
        interface IEventErrorAlreadyInTerminalState {

            /** EventErrorAlreadyInTerminalState currentPhase */
            currentPhase?: (string|null);
        }

        /** Represents an EventErrorAlreadyInTerminalState. */
        class EventErrorAlreadyInTerminalState implements IEventErrorAlreadyInTerminalState {

            /**
             * Constructs a new EventErrorAlreadyInTerminalState.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IEventErrorAlreadyInTerminalState);

            /** EventErrorAlreadyInTerminalState currentPhase. */
            public currentPhase: string;

            /**
             * Creates a new EventErrorAlreadyInTerminalState instance using the specified properties.
             * @param [properties] Properties to set
             * @returns EventErrorAlreadyInTerminalState instance
             */
            public static create(properties?: flyteidl.admin.IEventErrorAlreadyInTerminalState): flyteidl.admin.EventErrorAlreadyInTerminalState;

            /**
             * Encodes the specified EventErrorAlreadyInTerminalState message. Does not implicitly {@link flyteidl.admin.EventErrorAlreadyInTerminalState.verify|verify} messages.
             * @param message EventErrorAlreadyInTerminalState message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IEventErrorAlreadyInTerminalState, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EventErrorAlreadyInTerminalState message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EventErrorAlreadyInTerminalState
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.EventErrorAlreadyInTerminalState;

            /**
             * Verifies an EventErrorAlreadyInTerminalState message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an EventErrorIncompatibleCluster. */
        interface IEventErrorIncompatibleCluster {

            /** EventErrorIncompatibleCluster cluster */
            cluster?: (string|null);
        }

        /** Represents an EventErrorIncompatibleCluster. */
        class EventErrorIncompatibleCluster implements IEventErrorIncompatibleCluster {

            /**
             * Constructs a new EventErrorIncompatibleCluster.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IEventErrorIncompatibleCluster);

            /** EventErrorIncompatibleCluster cluster. */
            public cluster: string;

            /**
             * Creates a new EventErrorIncompatibleCluster instance using the specified properties.
             * @param [properties] Properties to set
             * @returns EventErrorIncompatibleCluster instance
             */
            public static create(properties?: flyteidl.admin.IEventErrorIncompatibleCluster): flyteidl.admin.EventErrorIncompatibleCluster;

            /**
             * Encodes the specified EventErrorIncompatibleCluster message. Does not implicitly {@link flyteidl.admin.EventErrorIncompatibleCluster.verify|verify} messages.
             * @param message EventErrorIncompatibleCluster message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IEventErrorIncompatibleCluster, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EventErrorIncompatibleCluster message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EventErrorIncompatibleCluster
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.EventErrorIncompatibleCluster;

            /**
             * Verifies an EventErrorIncompatibleCluster message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an EventFailureReason. */
        interface IEventFailureReason {

            /** EventFailureReason alreadyInTerminalState */
            alreadyInTerminalState?: (flyteidl.admin.IEventErrorAlreadyInTerminalState|null);

            /** EventFailureReason incompatibleCluster */
            incompatibleCluster?: (flyteidl.admin.IEventErrorIncompatibleCluster|null);
        }

        /** Represents an EventFailureReason. */
        class EventFailureReason implements IEventFailureReason {

            /**
             * Constructs a new EventFailureReason.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IEventFailureReason);

            /** EventFailureReason alreadyInTerminalState. */
            public alreadyInTerminalState?: (flyteidl.admin.IEventErrorAlreadyInTerminalState|null);

            /** EventFailureReason incompatibleCluster. */
            public incompatibleCluster?: (flyteidl.admin.IEventErrorIncompatibleCluster|null);

            /** EventFailureReason reason. */
            public reason?: ("alreadyInTerminalState"|"incompatibleCluster");

            /**
             * Creates a new EventFailureReason instance using the specified properties.
             * @param [properties] Properties to set
             * @returns EventFailureReason instance
             */
            public static create(properties?: flyteidl.admin.IEventFailureReason): flyteidl.admin.EventFailureReason;

            /**
             * Encodes the specified EventFailureReason message. Does not implicitly {@link flyteidl.admin.EventFailureReason.verify|verify} messages.
             * @param message EventFailureReason message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IEventFailureReason, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EventFailureReason message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EventFailureReason
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.EventFailureReason;

            /**
             * Verifies an EventFailureReason message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionEventRequest. */
        interface IWorkflowExecutionEventRequest {

            /** WorkflowExecutionEventRequest requestId */
            requestId?: (string|null);

            /** WorkflowExecutionEventRequest event */
            event?: (flyteidl.event.IWorkflowExecutionEvent|null);
        }

        /** Represents a WorkflowExecutionEventRequest. */
        class WorkflowExecutionEventRequest implements IWorkflowExecutionEventRequest {

            /**
             * Constructs a new WorkflowExecutionEventRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowExecutionEventRequest);

            /** WorkflowExecutionEventRequest requestId. */
            public requestId: string;

            /** WorkflowExecutionEventRequest event. */
            public event?: (flyteidl.event.IWorkflowExecutionEvent|null);

            /**
             * Creates a new WorkflowExecutionEventRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionEventRequest instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowExecutionEventRequest): flyteidl.admin.WorkflowExecutionEventRequest;

            /**
             * Encodes the specified WorkflowExecutionEventRequest message. Does not implicitly {@link flyteidl.admin.WorkflowExecutionEventRequest.verify|verify} messages.
             * @param message WorkflowExecutionEventRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowExecutionEventRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionEventRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionEventRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowExecutionEventRequest;

            /**
             * Verifies a WorkflowExecutionEventRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionEventResponse. */
        interface IWorkflowExecutionEventResponse {
        }

        /** Represents a WorkflowExecutionEventResponse. */
        class WorkflowExecutionEventResponse implements IWorkflowExecutionEventResponse {

            /**
             * Constructs a new WorkflowExecutionEventResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowExecutionEventResponse);

            /**
             * Creates a new WorkflowExecutionEventResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionEventResponse instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowExecutionEventResponse): flyteidl.admin.WorkflowExecutionEventResponse;

            /**
             * Encodes the specified WorkflowExecutionEventResponse message. Does not implicitly {@link flyteidl.admin.WorkflowExecutionEventResponse.verify|verify} messages.
             * @param message WorkflowExecutionEventResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowExecutionEventResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionEventResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionEventResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowExecutionEventResponse;

            /**
             * Verifies a WorkflowExecutionEventResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionEventRequest. */
        interface INodeExecutionEventRequest {

            /** NodeExecutionEventRequest requestId */
            requestId?: (string|null);

            /** NodeExecutionEventRequest event */
            event?: (flyteidl.event.INodeExecutionEvent|null);
        }

        /** Represents a NodeExecutionEventRequest. */
        class NodeExecutionEventRequest implements INodeExecutionEventRequest {

            /**
             * Constructs a new NodeExecutionEventRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionEventRequest);

            /** NodeExecutionEventRequest requestId. */
            public requestId: string;

            /** NodeExecutionEventRequest event. */
            public event?: (flyteidl.event.INodeExecutionEvent|null);

            /**
             * Creates a new NodeExecutionEventRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionEventRequest instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionEventRequest): flyteidl.admin.NodeExecutionEventRequest;

            /**
             * Encodes the specified NodeExecutionEventRequest message. Does not implicitly {@link flyteidl.admin.NodeExecutionEventRequest.verify|verify} messages.
             * @param message NodeExecutionEventRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionEventRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionEventRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionEventRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionEventRequest;

            /**
             * Verifies a NodeExecutionEventRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionEventResponse. */
        interface INodeExecutionEventResponse {
        }

        /** Represents a NodeExecutionEventResponse. */
        class NodeExecutionEventResponse implements INodeExecutionEventResponse {

            /**
             * Constructs a new NodeExecutionEventResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionEventResponse);

            /**
             * Creates a new NodeExecutionEventResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionEventResponse instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionEventResponse): flyteidl.admin.NodeExecutionEventResponse;

            /**
             * Encodes the specified NodeExecutionEventResponse message. Does not implicitly {@link flyteidl.admin.NodeExecutionEventResponse.verify|verify} messages.
             * @param message NodeExecutionEventResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionEventResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionEventResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionEventResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionEventResponse;

            /**
             * Verifies a NodeExecutionEventResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskExecutionEventRequest. */
        interface ITaskExecutionEventRequest {

            /** TaskExecutionEventRequest requestId */
            requestId?: (string|null);

            /** TaskExecutionEventRequest event */
            event?: (flyteidl.event.ITaskExecutionEvent|null);
        }

        /** Represents a TaskExecutionEventRequest. */
        class TaskExecutionEventRequest implements ITaskExecutionEventRequest {

            /**
             * Constructs a new TaskExecutionEventRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ITaskExecutionEventRequest);

            /** TaskExecutionEventRequest requestId. */
            public requestId: string;

            /** TaskExecutionEventRequest event. */
            public event?: (flyteidl.event.ITaskExecutionEvent|null);

            /**
             * Creates a new TaskExecutionEventRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskExecutionEventRequest instance
             */
            public static create(properties?: flyteidl.admin.ITaskExecutionEventRequest): flyteidl.admin.TaskExecutionEventRequest;

            /**
             * Encodes the specified TaskExecutionEventRequest message. Does not implicitly {@link flyteidl.admin.TaskExecutionEventRequest.verify|verify} messages.
             * @param message TaskExecutionEventRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ITaskExecutionEventRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskExecutionEventRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskExecutionEventRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.TaskExecutionEventRequest;

            /**
             * Verifies a TaskExecutionEventRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskExecutionEventResponse. */
        interface ITaskExecutionEventResponse {
        }

        /** Represents a TaskExecutionEventResponse. */
        class TaskExecutionEventResponse implements ITaskExecutionEventResponse {

            /**
             * Constructs a new TaskExecutionEventResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ITaskExecutionEventResponse);

            /**
             * Creates a new TaskExecutionEventResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskExecutionEventResponse instance
             */
            public static create(properties?: flyteidl.admin.ITaskExecutionEventResponse): flyteidl.admin.TaskExecutionEventResponse;

            /**
             * Encodes the specified TaskExecutionEventResponse message. Does not implicitly {@link flyteidl.admin.TaskExecutionEventResponse.verify|verify} messages.
             * @param message TaskExecutionEventResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ITaskExecutionEventResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskExecutionEventResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskExecutionEventResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.TaskExecutionEventResponse;

            /**
             * Verifies a TaskExecutionEventResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionCreateRequest. */
        interface IExecutionCreateRequest {

            /** ExecutionCreateRequest project */
            project?: (string|null);

            /** ExecutionCreateRequest domain */
            domain?: (string|null);

            /** ExecutionCreateRequest name */
            name?: (string|null);

            /** ExecutionCreateRequest spec */
            spec?: (flyteidl.admin.IExecutionSpec|null);

            /** ExecutionCreateRequest inputs */
            inputs?: (flyteidl.core.ILiteralMap|null);

            /** ExecutionCreateRequest org */
            org?: (string|null);
        }

        /** Represents an ExecutionCreateRequest. */
        class ExecutionCreateRequest implements IExecutionCreateRequest {

            /**
             * Constructs a new ExecutionCreateRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionCreateRequest);

            /** ExecutionCreateRequest project. */
            public project: string;

            /** ExecutionCreateRequest domain. */
            public domain: string;

            /** ExecutionCreateRequest name. */
            public name: string;

            /** ExecutionCreateRequest spec. */
            public spec?: (flyteidl.admin.IExecutionSpec|null);

            /** ExecutionCreateRequest inputs. */
            public inputs?: (flyteidl.core.ILiteralMap|null);

            /** ExecutionCreateRequest org. */
            public org: string;

            /**
             * Creates a new ExecutionCreateRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionCreateRequest instance
             */
            public static create(properties?: flyteidl.admin.IExecutionCreateRequest): flyteidl.admin.ExecutionCreateRequest;

            /**
             * Encodes the specified ExecutionCreateRequest message. Does not implicitly {@link flyteidl.admin.ExecutionCreateRequest.verify|verify} messages.
             * @param message ExecutionCreateRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionCreateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionCreateRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionCreateRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionCreateRequest;

            /**
             * Verifies an ExecutionCreateRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionRelaunchRequest. */
        interface IExecutionRelaunchRequest {

            /** ExecutionRelaunchRequest id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionRelaunchRequest name */
            name?: (string|null);

            /** ExecutionRelaunchRequest overwriteCache */
            overwriteCache?: (boolean|null);
        }

        /** Represents an ExecutionRelaunchRequest. */
        class ExecutionRelaunchRequest implements IExecutionRelaunchRequest {

            /**
             * Constructs a new ExecutionRelaunchRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionRelaunchRequest);

            /** ExecutionRelaunchRequest id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionRelaunchRequest name. */
            public name: string;

            /** ExecutionRelaunchRequest overwriteCache. */
            public overwriteCache: boolean;

            /**
             * Creates a new ExecutionRelaunchRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionRelaunchRequest instance
             */
            public static create(properties?: flyteidl.admin.IExecutionRelaunchRequest): flyteidl.admin.ExecutionRelaunchRequest;

            /**
             * Encodes the specified ExecutionRelaunchRequest message. Does not implicitly {@link flyteidl.admin.ExecutionRelaunchRequest.verify|verify} messages.
             * @param message ExecutionRelaunchRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionRelaunchRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionRelaunchRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionRelaunchRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionRelaunchRequest;

            /**
             * Verifies an ExecutionRelaunchRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionRecoverRequest. */
        interface IExecutionRecoverRequest {

            /** ExecutionRecoverRequest id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionRecoverRequest name */
            name?: (string|null);

            /** ExecutionRecoverRequest metadata */
            metadata?: (flyteidl.admin.IExecutionMetadata|null);
        }

        /** Represents an ExecutionRecoverRequest. */
        class ExecutionRecoverRequest implements IExecutionRecoverRequest {

            /**
             * Constructs a new ExecutionRecoverRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionRecoverRequest);

            /** ExecutionRecoverRequest id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionRecoverRequest name. */
            public name: string;

            /** ExecutionRecoverRequest metadata. */
            public metadata?: (flyteidl.admin.IExecutionMetadata|null);

            /**
             * Creates a new ExecutionRecoverRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionRecoverRequest instance
             */
            public static create(properties?: flyteidl.admin.IExecutionRecoverRequest): flyteidl.admin.ExecutionRecoverRequest;

            /**
             * Encodes the specified ExecutionRecoverRequest message. Does not implicitly {@link flyteidl.admin.ExecutionRecoverRequest.verify|verify} messages.
             * @param message ExecutionRecoverRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionRecoverRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionRecoverRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionRecoverRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionRecoverRequest;

            /**
             * Verifies an ExecutionRecoverRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionCreateResponse. */
        interface IExecutionCreateResponse {

            /** ExecutionCreateResponse id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);
        }

        /** Represents an ExecutionCreateResponse. */
        class ExecutionCreateResponse implements IExecutionCreateResponse {

            /**
             * Constructs a new ExecutionCreateResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionCreateResponse);

            /** ExecutionCreateResponse id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /**
             * Creates a new ExecutionCreateResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionCreateResponse instance
             */
            public static create(properties?: flyteidl.admin.IExecutionCreateResponse): flyteidl.admin.ExecutionCreateResponse;

            /**
             * Encodes the specified ExecutionCreateResponse message. Does not implicitly {@link flyteidl.admin.ExecutionCreateResponse.verify|verify} messages.
             * @param message ExecutionCreateResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionCreateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionCreateResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionCreateResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionCreateResponse;

            /**
             * Verifies an ExecutionCreateResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionGetRequest. */
        interface IWorkflowExecutionGetRequest {

            /** WorkflowExecutionGetRequest id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);
        }

        /** Represents a WorkflowExecutionGetRequest. */
        class WorkflowExecutionGetRequest implements IWorkflowExecutionGetRequest {

            /**
             * Constructs a new WorkflowExecutionGetRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowExecutionGetRequest);

            /** WorkflowExecutionGetRequest id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /**
             * Creates a new WorkflowExecutionGetRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionGetRequest instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowExecutionGetRequest): flyteidl.admin.WorkflowExecutionGetRequest;

            /**
             * Encodes the specified WorkflowExecutionGetRequest message. Does not implicitly {@link flyteidl.admin.WorkflowExecutionGetRequest.verify|verify} messages.
             * @param message WorkflowExecutionGetRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowExecutionGetRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionGetRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionGetRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowExecutionGetRequest;

            /**
             * Verifies a WorkflowExecutionGetRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Execution. */
        interface IExecution {

            /** Execution id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** Execution spec */
            spec?: (flyteidl.admin.IExecutionSpec|null);

            /** Execution closure */
            closure?: (flyteidl.admin.IExecutionClosure|null);
        }

        /** Represents an Execution. */
        class Execution implements IExecution {

            /**
             * Constructs a new Execution.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecution);

            /** Execution id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** Execution spec. */
            public spec?: (flyteidl.admin.IExecutionSpec|null);

            /** Execution closure. */
            public closure?: (flyteidl.admin.IExecutionClosure|null);

            /**
             * Creates a new Execution instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Execution instance
             */
            public static create(properties?: flyteidl.admin.IExecution): flyteidl.admin.Execution;

            /**
             * Encodes the specified Execution message. Does not implicitly {@link flyteidl.admin.Execution.verify|verify} messages.
             * @param message Execution message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecution, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Execution message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Execution
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Execution;

            /**
             * Verifies an Execution message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionList. */
        interface IExecutionList {

            /** ExecutionList executions */
            executions?: (flyteidl.admin.IExecution[]|null);

            /** ExecutionList token */
            token?: (string|null);
        }

        /** Represents an ExecutionList. */
        class ExecutionList implements IExecutionList {

            /**
             * Constructs a new ExecutionList.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionList);

            /** ExecutionList executions. */
            public executions: flyteidl.admin.IExecution[];

            /** ExecutionList token. */
            public token: string;

            /**
             * Creates a new ExecutionList instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionList instance
             */
            public static create(properties?: flyteidl.admin.IExecutionList): flyteidl.admin.ExecutionList;

            /**
             * Encodes the specified ExecutionList message. Does not implicitly {@link flyteidl.admin.ExecutionList.verify|verify} messages.
             * @param message ExecutionList message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionList, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionList message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionList
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionList;

            /**
             * Verifies an ExecutionList message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LiteralMapBlob. */
        interface ILiteralMapBlob {

            /** LiteralMapBlob values */
            values?: (flyteidl.core.ILiteralMap|null);

            /** LiteralMapBlob uri */
            uri?: (string|null);
        }

        /** Represents a LiteralMapBlob. */
        class LiteralMapBlob implements ILiteralMapBlob {

            /**
             * Constructs a new LiteralMapBlob.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILiteralMapBlob);

            /** LiteralMapBlob values. */
            public values?: (flyteidl.core.ILiteralMap|null);

            /** LiteralMapBlob uri. */
            public uri: string;

            /** LiteralMapBlob data. */
            public data?: ("values"|"uri");

            /**
             * Creates a new LiteralMapBlob instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LiteralMapBlob instance
             */
            public static create(properties?: flyteidl.admin.ILiteralMapBlob): flyteidl.admin.LiteralMapBlob;

            /**
             * Encodes the specified LiteralMapBlob message. Does not implicitly {@link flyteidl.admin.LiteralMapBlob.verify|verify} messages.
             * @param message LiteralMapBlob message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILiteralMapBlob, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LiteralMapBlob message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LiteralMapBlob
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LiteralMapBlob;

            /**
             * Verifies a LiteralMapBlob message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an AbortMetadata. */
        interface IAbortMetadata {

            /** AbortMetadata cause */
            cause?: (string|null);

            /** AbortMetadata principal */
            principal?: (string|null);
        }

        /** Represents an AbortMetadata. */
        class AbortMetadata implements IAbortMetadata {

            /**
             * Constructs a new AbortMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IAbortMetadata);

            /** AbortMetadata cause. */
            public cause: string;

            /** AbortMetadata principal. */
            public principal: string;

            /**
             * Creates a new AbortMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns AbortMetadata instance
             */
            public static create(properties?: flyteidl.admin.IAbortMetadata): flyteidl.admin.AbortMetadata;

            /**
             * Encodes the specified AbortMetadata message. Does not implicitly {@link flyteidl.admin.AbortMetadata.verify|verify} messages.
             * @param message AbortMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IAbortMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an AbortMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns AbortMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.AbortMetadata;

            /**
             * Verifies an AbortMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionClosure. */
        interface IExecutionClosure {

            /** ExecutionClosure outputs */
            outputs?: (flyteidl.admin.ILiteralMapBlob|null);

            /** ExecutionClosure error */
            error?: (flyteidl.core.IExecutionError|null);

            /** ExecutionClosure abortCause */
            abortCause?: (string|null);

            /** ExecutionClosure abortMetadata */
            abortMetadata?: (flyteidl.admin.IAbortMetadata|null);

            /** ExecutionClosure outputData */
            outputData?: (flyteidl.core.ILiteralMap|null);

            /** ExecutionClosure computedInputs */
            computedInputs?: (flyteidl.core.ILiteralMap|null);

            /** ExecutionClosure phase */
            phase?: (flyteidl.core.WorkflowExecution.Phase|null);

            /** ExecutionClosure startedAt */
            startedAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionClosure duration */
            duration?: (google.protobuf.IDuration|null);

            /** ExecutionClosure createdAt */
            createdAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionClosure updatedAt */
            updatedAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionClosure notifications */
            notifications?: (flyteidl.admin.INotification[]|null);

            /** ExecutionClosure workflowId */
            workflowId?: (flyteidl.core.IIdentifier|null);

            /** ExecutionClosure stateChangeDetails */
            stateChangeDetails?: (flyteidl.admin.IExecutionStateChangeDetails|null);
        }

        /** Represents an ExecutionClosure. */
        class ExecutionClosure implements IExecutionClosure {

            /**
             * Constructs a new ExecutionClosure.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionClosure);

            /** ExecutionClosure outputs. */
            public outputs?: (flyteidl.admin.ILiteralMapBlob|null);

            /** ExecutionClosure error. */
            public error?: (flyteidl.core.IExecutionError|null);

            /** ExecutionClosure abortCause. */
            public abortCause: string;

            /** ExecutionClosure abortMetadata. */
            public abortMetadata?: (flyteidl.admin.IAbortMetadata|null);

            /** ExecutionClosure outputData. */
            public outputData?: (flyteidl.core.ILiteralMap|null);

            /** ExecutionClosure computedInputs. */
            public computedInputs?: (flyteidl.core.ILiteralMap|null);

            /** ExecutionClosure phase. */
            public phase: flyteidl.core.WorkflowExecution.Phase;

            /** ExecutionClosure startedAt. */
            public startedAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionClosure duration. */
            public duration?: (google.protobuf.IDuration|null);

            /** ExecutionClosure createdAt. */
            public createdAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionClosure updatedAt. */
            public updatedAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionClosure notifications. */
            public notifications: flyteidl.admin.INotification[];

            /** ExecutionClosure workflowId. */
            public workflowId?: (flyteidl.core.IIdentifier|null);

            /** ExecutionClosure stateChangeDetails. */
            public stateChangeDetails?: (flyteidl.admin.IExecutionStateChangeDetails|null);

            /** ExecutionClosure outputResult. */
            public outputResult?: ("outputs"|"error"|"abortCause"|"abortMetadata"|"outputData");

            /**
             * Creates a new ExecutionClosure instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionClosure instance
             */
            public static create(properties?: flyteidl.admin.IExecutionClosure): flyteidl.admin.ExecutionClosure;

            /**
             * Encodes the specified ExecutionClosure message. Does not implicitly {@link flyteidl.admin.ExecutionClosure.verify|verify} messages.
             * @param message ExecutionClosure message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionClosure, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionClosure message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionClosure
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionClosure;

            /**
             * Verifies an ExecutionClosure message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a SystemMetadata. */
        interface ISystemMetadata {

            /** SystemMetadata executionCluster */
            executionCluster?: (string|null);

            /** SystemMetadata namespace */
            namespace?: (string|null);
        }

        /** Represents a SystemMetadata. */
        class SystemMetadata implements ISystemMetadata {

            /**
             * Constructs a new SystemMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ISystemMetadata);

            /** SystemMetadata executionCluster. */
            public executionCluster: string;

            /** SystemMetadata namespace. */
            public namespace: string;

            /**
             * Creates a new SystemMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns SystemMetadata instance
             */
            public static create(properties?: flyteidl.admin.ISystemMetadata): flyteidl.admin.SystemMetadata;

            /**
             * Encodes the specified SystemMetadata message. Does not implicitly {@link flyteidl.admin.SystemMetadata.verify|verify} messages.
             * @param message SystemMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ISystemMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SystemMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SystemMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.SystemMetadata;

            /**
             * Verifies a SystemMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionMetadata. */
        interface IExecutionMetadata {

            /** ExecutionMetadata mode */
            mode?: (flyteidl.admin.ExecutionMetadata.ExecutionMode|null);

            /** ExecutionMetadata principal */
            principal?: (string|null);

            /** ExecutionMetadata nesting */
            nesting?: (number|null);

            /** ExecutionMetadata scheduledAt */
            scheduledAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionMetadata parentNodeExecution */
            parentNodeExecution?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** ExecutionMetadata referenceExecution */
            referenceExecution?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionMetadata systemMetadata */
            systemMetadata?: (flyteidl.admin.ISystemMetadata|null);

            /** ExecutionMetadata artifactIds */
            artifactIds?: (flyteidl.core.IArtifactID[]|null);
        }

        /** Represents an ExecutionMetadata. */
        class ExecutionMetadata implements IExecutionMetadata {

            /**
             * Constructs a new ExecutionMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionMetadata);

            /** ExecutionMetadata mode. */
            public mode: flyteidl.admin.ExecutionMetadata.ExecutionMode;

            /** ExecutionMetadata principal. */
            public principal: string;

            /** ExecutionMetadata nesting. */
            public nesting: number;

            /** ExecutionMetadata scheduledAt. */
            public scheduledAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionMetadata parentNodeExecution. */
            public parentNodeExecution?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** ExecutionMetadata referenceExecution. */
            public referenceExecution?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionMetadata systemMetadata. */
            public systemMetadata?: (flyteidl.admin.ISystemMetadata|null);

            /** ExecutionMetadata artifactIds. */
            public artifactIds: flyteidl.core.IArtifactID[];

            /**
             * Creates a new ExecutionMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionMetadata instance
             */
            public static create(properties?: flyteidl.admin.IExecutionMetadata): flyteidl.admin.ExecutionMetadata;

            /**
             * Encodes the specified ExecutionMetadata message. Does not implicitly {@link flyteidl.admin.ExecutionMetadata.verify|verify} messages.
             * @param message ExecutionMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionMetadata;

            /**
             * Verifies an ExecutionMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace ExecutionMetadata {

            /** ExecutionMode enum. */
            enum ExecutionMode {
                MANUAL = 0,
                SCHEDULED = 1,
                SYSTEM = 2,
                RELAUNCH = 3,
                CHILD_WORKFLOW = 4,
                RECOVERED = 5,
                TRIGGER = 6
            }
        }

        /** Properties of a NotificationList. */
        interface INotificationList {

            /** NotificationList notifications */
            notifications?: (flyteidl.admin.INotification[]|null);
        }

        /** Represents a NotificationList. */
        class NotificationList implements INotificationList {

            /**
             * Constructs a new NotificationList.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INotificationList);

            /** NotificationList notifications. */
            public notifications: flyteidl.admin.INotification[];

            /**
             * Creates a new NotificationList instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NotificationList instance
             */
            public static create(properties?: flyteidl.admin.INotificationList): flyteidl.admin.NotificationList;

            /**
             * Encodes the specified NotificationList message. Does not implicitly {@link flyteidl.admin.NotificationList.verify|verify} messages.
             * @param message NotificationList message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INotificationList, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NotificationList message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NotificationList
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NotificationList;

            /**
             * Verifies a NotificationList message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionSpec. */
        interface IExecutionSpec {

            /** ExecutionSpec launchPlan */
            launchPlan?: (flyteidl.core.IIdentifier|null);

            /** ExecutionSpec inputs */
            inputs?: (flyteidl.core.ILiteralMap|null);

            /** ExecutionSpec metadata */
            metadata?: (flyteidl.admin.IExecutionMetadata|null);

            /** ExecutionSpec notifications */
            notifications?: (flyteidl.admin.INotificationList|null);

            /** ExecutionSpec disableAll */
            disableAll?: (boolean|null);

            /** ExecutionSpec labels */
            labels?: (flyteidl.admin.ILabels|null);

            /** ExecutionSpec annotations */
            annotations?: (flyteidl.admin.IAnnotations|null);

            /** ExecutionSpec securityContext */
            securityContext?: (flyteidl.core.ISecurityContext|null);

            /** ExecutionSpec authRole */
            authRole?: (flyteidl.admin.IAuthRole|null);

            /** ExecutionSpec qualityOfService */
            qualityOfService?: (flyteidl.core.IQualityOfService|null);

            /** ExecutionSpec maxParallelism */
            maxParallelism?: (number|null);

            /** ExecutionSpec rawOutputDataConfig */
            rawOutputDataConfig?: (flyteidl.admin.IRawOutputDataConfig|null);

            /** ExecutionSpec clusterAssignment */
            clusterAssignment?: (flyteidl.admin.IClusterAssignment|null);

            /** ExecutionSpec interruptible */
            interruptible?: (google.protobuf.IBoolValue|null);

            /** ExecutionSpec overwriteCache */
            overwriteCache?: (boolean|null);

            /** ExecutionSpec envs */
            envs?: (flyteidl.admin.IEnvs|null);

            /** ExecutionSpec tags */
            tags?: (string[]|null);

            /** ExecutionSpec executionClusterLabel */
            executionClusterLabel?: (flyteidl.admin.IExecutionClusterLabel|null);

            /** ExecutionSpec executionEnvAssignments */
            executionEnvAssignments?: (flyteidl.core.IExecutionEnvAssignment[]|null);
        }

        /** Represents an ExecutionSpec. */
        class ExecutionSpec implements IExecutionSpec {

            /**
             * Constructs a new ExecutionSpec.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionSpec);

            /** ExecutionSpec launchPlan. */
            public launchPlan?: (flyteidl.core.IIdentifier|null);

            /** ExecutionSpec inputs. */
            public inputs?: (flyteidl.core.ILiteralMap|null);

            /** ExecutionSpec metadata. */
            public metadata?: (flyteidl.admin.IExecutionMetadata|null);

            /** ExecutionSpec notifications. */
            public notifications?: (flyteidl.admin.INotificationList|null);

            /** ExecutionSpec disableAll. */
            public disableAll: boolean;

            /** ExecutionSpec labels. */
            public labels?: (flyteidl.admin.ILabels|null);

            /** ExecutionSpec annotations. */
            public annotations?: (flyteidl.admin.IAnnotations|null);

            /** ExecutionSpec securityContext. */
            public securityContext?: (flyteidl.core.ISecurityContext|null);

            /** ExecutionSpec authRole. */
            public authRole?: (flyteidl.admin.IAuthRole|null);

            /** ExecutionSpec qualityOfService. */
            public qualityOfService?: (flyteidl.core.IQualityOfService|null);

            /** ExecutionSpec maxParallelism. */
            public maxParallelism: number;

            /** ExecutionSpec rawOutputDataConfig. */
            public rawOutputDataConfig?: (flyteidl.admin.IRawOutputDataConfig|null);

            /** ExecutionSpec clusterAssignment. */
            public clusterAssignment?: (flyteidl.admin.IClusterAssignment|null);

            /** ExecutionSpec interruptible. */
            public interruptible?: (google.protobuf.IBoolValue|null);

            /** ExecutionSpec overwriteCache. */
            public overwriteCache: boolean;

            /** ExecutionSpec envs. */
            public envs?: (flyteidl.admin.IEnvs|null);

            /** ExecutionSpec tags. */
            public tags: string[];

            /** ExecutionSpec executionClusterLabel. */
            public executionClusterLabel?: (flyteidl.admin.IExecutionClusterLabel|null);

            /** ExecutionSpec executionEnvAssignments. */
            public executionEnvAssignments: flyteidl.core.IExecutionEnvAssignment[];

            /** ExecutionSpec notificationOverrides. */
            public notificationOverrides?: ("notifications"|"disableAll");

            /**
             * Creates a new ExecutionSpec instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionSpec instance
             */
            public static create(properties?: flyteidl.admin.IExecutionSpec): flyteidl.admin.ExecutionSpec;

            /**
             * Encodes the specified ExecutionSpec message. Does not implicitly {@link flyteidl.admin.ExecutionSpec.verify|verify} messages.
             * @param message ExecutionSpec message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionSpec, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionSpec message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionSpec
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionSpec;

            /**
             * Verifies an ExecutionSpec message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionTerminateRequest. */
        interface IExecutionTerminateRequest {

            /** ExecutionTerminateRequest id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionTerminateRequest cause */
            cause?: (string|null);
        }

        /** Represents an ExecutionTerminateRequest. */
        class ExecutionTerminateRequest implements IExecutionTerminateRequest {

            /**
             * Constructs a new ExecutionTerminateRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionTerminateRequest);

            /** ExecutionTerminateRequest id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionTerminateRequest cause. */
            public cause: string;

            /**
             * Creates a new ExecutionTerminateRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionTerminateRequest instance
             */
            public static create(properties?: flyteidl.admin.IExecutionTerminateRequest): flyteidl.admin.ExecutionTerminateRequest;

            /**
             * Encodes the specified ExecutionTerminateRequest message. Does not implicitly {@link flyteidl.admin.ExecutionTerminateRequest.verify|verify} messages.
             * @param message ExecutionTerminateRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionTerminateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionTerminateRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionTerminateRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionTerminateRequest;

            /**
             * Verifies an ExecutionTerminateRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionTerminateResponse. */
        interface IExecutionTerminateResponse {
        }

        /** Represents an ExecutionTerminateResponse. */
        class ExecutionTerminateResponse implements IExecutionTerminateResponse {

            /**
             * Constructs a new ExecutionTerminateResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionTerminateResponse);

            /**
             * Creates a new ExecutionTerminateResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionTerminateResponse instance
             */
            public static create(properties?: flyteidl.admin.IExecutionTerminateResponse): flyteidl.admin.ExecutionTerminateResponse;

            /**
             * Encodes the specified ExecutionTerminateResponse message. Does not implicitly {@link flyteidl.admin.ExecutionTerminateResponse.verify|verify} messages.
             * @param message ExecutionTerminateResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionTerminateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionTerminateResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionTerminateResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionTerminateResponse;

            /**
             * Verifies an ExecutionTerminateResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionGetDataRequest. */
        interface IWorkflowExecutionGetDataRequest {

            /** WorkflowExecutionGetDataRequest id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);
        }

        /** Represents a WorkflowExecutionGetDataRequest. */
        class WorkflowExecutionGetDataRequest implements IWorkflowExecutionGetDataRequest {

            /**
             * Constructs a new WorkflowExecutionGetDataRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowExecutionGetDataRequest);

            /** WorkflowExecutionGetDataRequest id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /**
             * Creates a new WorkflowExecutionGetDataRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionGetDataRequest instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowExecutionGetDataRequest): flyteidl.admin.WorkflowExecutionGetDataRequest;

            /**
             * Encodes the specified WorkflowExecutionGetDataRequest message. Does not implicitly {@link flyteidl.admin.WorkflowExecutionGetDataRequest.verify|verify} messages.
             * @param message WorkflowExecutionGetDataRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowExecutionGetDataRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionGetDataRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionGetDataRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowExecutionGetDataRequest;

            /**
             * Verifies a WorkflowExecutionGetDataRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionGetDataResponse. */
        interface IWorkflowExecutionGetDataResponse {

            /** WorkflowExecutionGetDataResponse outputs */
            outputs?: (flyteidl.admin.IUrlBlob|null);

            /** WorkflowExecutionGetDataResponse inputs */
            inputs?: (flyteidl.admin.IUrlBlob|null);

            /** WorkflowExecutionGetDataResponse fullInputs */
            fullInputs?: (flyteidl.core.ILiteralMap|null);

            /** WorkflowExecutionGetDataResponse fullOutputs */
            fullOutputs?: (flyteidl.core.ILiteralMap|null);
        }

        /** Represents a WorkflowExecutionGetDataResponse. */
        class WorkflowExecutionGetDataResponse implements IWorkflowExecutionGetDataResponse {

            /**
             * Constructs a new WorkflowExecutionGetDataResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowExecutionGetDataResponse);

            /** WorkflowExecutionGetDataResponse outputs. */
            public outputs?: (flyteidl.admin.IUrlBlob|null);

            /** WorkflowExecutionGetDataResponse inputs. */
            public inputs?: (flyteidl.admin.IUrlBlob|null);

            /** WorkflowExecutionGetDataResponse fullInputs. */
            public fullInputs?: (flyteidl.core.ILiteralMap|null);

            /** WorkflowExecutionGetDataResponse fullOutputs. */
            public fullOutputs?: (flyteidl.core.ILiteralMap|null);

            /**
             * Creates a new WorkflowExecutionGetDataResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionGetDataResponse instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowExecutionGetDataResponse): flyteidl.admin.WorkflowExecutionGetDataResponse;

            /**
             * Encodes the specified WorkflowExecutionGetDataResponse message. Does not implicitly {@link flyteidl.admin.WorkflowExecutionGetDataResponse.verify|verify} messages.
             * @param message WorkflowExecutionGetDataResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowExecutionGetDataResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionGetDataResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionGetDataResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowExecutionGetDataResponse;

            /**
             * Verifies a WorkflowExecutionGetDataResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** ExecutionState enum. */
        enum ExecutionState {
            EXECUTION_ACTIVE = 0,
            EXECUTION_ARCHIVED = 1
        }

        /** Properties of an ExecutionUpdateRequest. */
        interface IExecutionUpdateRequest {

            /** ExecutionUpdateRequest id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionUpdateRequest state */
            state?: (flyteidl.admin.ExecutionState|null);
        }

        /** Represents an ExecutionUpdateRequest. */
        class ExecutionUpdateRequest implements IExecutionUpdateRequest {

            /**
             * Constructs a new ExecutionUpdateRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionUpdateRequest);

            /** ExecutionUpdateRequest id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** ExecutionUpdateRequest state. */
            public state: flyteidl.admin.ExecutionState;

            /**
             * Creates a new ExecutionUpdateRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionUpdateRequest instance
             */
            public static create(properties?: flyteidl.admin.IExecutionUpdateRequest): flyteidl.admin.ExecutionUpdateRequest;

            /**
             * Encodes the specified ExecutionUpdateRequest message. Does not implicitly {@link flyteidl.admin.ExecutionUpdateRequest.verify|verify} messages.
             * @param message ExecutionUpdateRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionUpdateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionUpdateRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionUpdateRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionUpdateRequest;

            /**
             * Verifies an ExecutionUpdateRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionStateChangeDetails. */
        interface IExecutionStateChangeDetails {

            /** ExecutionStateChangeDetails state */
            state?: (flyteidl.admin.ExecutionState|null);

            /** ExecutionStateChangeDetails occurredAt */
            occurredAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionStateChangeDetails principal */
            principal?: (string|null);
        }

        /** Represents an ExecutionStateChangeDetails. */
        class ExecutionStateChangeDetails implements IExecutionStateChangeDetails {

            /**
             * Constructs a new ExecutionStateChangeDetails.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionStateChangeDetails);

            /** ExecutionStateChangeDetails state. */
            public state: flyteidl.admin.ExecutionState;

            /** ExecutionStateChangeDetails occurredAt. */
            public occurredAt?: (google.protobuf.ITimestamp|null);

            /** ExecutionStateChangeDetails principal. */
            public principal: string;

            /**
             * Creates a new ExecutionStateChangeDetails instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionStateChangeDetails instance
             */
            public static create(properties?: flyteidl.admin.IExecutionStateChangeDetails): flyteidl.admin.ExecutionStateChangeDetails;

            /**
             * Encodes the specified ExecutionStateChangeDetails message. Does not implicitly {@link flyteidl.admin.ExecutionStateChangeDetails.verify|verify} messages.
             * @param message ExecutionStateChangeDetails message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionStateChangeDetails, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionStateChangeDetails message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionStateChangeDetails
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionStateChangeDetails;

            /**
             * Verifies an ExecutionStateChangeDetails message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionUpdateResponse. */
        interface IExecutionUpdateResponse {
        }

        /** Represents an ExecutionUpdateResponse. */
        class ExecutionUpdateResponse implements IExecutionUpdateResponse {

            /**
             * Constructs a new ExecutionUpdateResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionUpdateResponse);

            /**
             * Creates a new ExecutionUpdateResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionUpdateResponse instance
             */
            public static create(properties?: flyteidl.admin.IExecutionUpdateResponse): flyteidl.admin.ExecutionUpdateResponse;

            /**
             * Encodes the specified ExecutionUpdateResponse message. Does not implicitly {@link flyteidl.admin.ExecutionUpdateResponse.verify|verify} messages.
             * @param message ExecutionUpdateResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionUpdateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionUpdateResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionUpdateResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionUpdateResponse;

            /**
             * Verifies an ExecutionUpdateResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionGetMetricsRequest. */
        interface IWorkflowExecutionGetMetricsRequest {

            /** WorkflowExecutionGetMetricsRequest id */
            id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** WorkflowExecutionGetMetricsRequest depth */
            depth?: (number|null);
        }

        /** Represents a WorkflowExecutionGetMetricsRequest. */
        class WorkflowExecutionGetMetricsRequest implements IWorkflowExecutionGetMetricsRequest {

            /**
             * Constructs a new WorkflowExecutionGetMetricsRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowExecutionGetMetricsRequest);

            /** WorkflowExecutionGetMetricsRequest id. */
            public id?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** WorkflowExecutionGetMetricsRequest depth. */
            public depth: number;

            /**
             * Creates a new WorkflowExecutionGetMetricsRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionGetMetricsRequest instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowExecutionGetMetricsRequest): flyteidl.admin.WorkflowExecutionGetMetricsRequest;

            /**
             * Encodes the specified WorkflowExecutionGetMetricsRequest message. Does not implicitly {@link flyteidl.admin.WorkflowExecutionGetMetricsRequest.verify|verify} messages.
             * @param message WorkflowExecutionGetMetricsRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowExecutionGetMetricsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionGetMetricsRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionGetMetricsRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowExecutionGetMetricsRequest;

            /**
             * Verifies a WorkflowExecutionGetMetricsRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionGetMetricsResponse. */
        interface IWorkflowExecutionGetMetricsResponse {

            /** WorkflowExecutionGetMetricsResponse span */
            span?: (flyteidl.core.ISpan|null);
        }

        /** Represents a WorkflowExecutionGetMetricsResponse. */
        class WorkflowExecutionGetMetricsResponse implements IWorkflowExecutionGetMetricsResponse {

            /**
             * Constructs a new WorkflowExecutionGetMetricsResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowExecutionGetMetricsResponse);

            /** WorkflowExecutionGetMetricsResponse span. */
            public span?: (flyteidl.core.ISpan|null);

            /**
             * Creates a new WorkflowExecutionGetMetricsResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionGetMetricsResponse instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowExecutionGetMetricsResponse): flyteidl.admin.WorkflowExecutionGetMetricsResponse;

            /**
             * Encodes the specified WorkflowExecutionGetMetricsResponse message. Does not implicitly {@link flyteidl.admin.WorkflowExecutionGetMetricsResponse.verify|verify} messages.
             * @param message WorkflowExecutionGetMetricsResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowExecutionGetMetricsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionGetMetricsResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionGetMetricsResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowExecutionGetMetricsResponse;

            /**
             * Verifies a WorkflowExecutionGetMetricsResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** MatchableResource enum. */
        enum MatchableResource {
            TASK_RESOURCE = 0,
            CLUSTER_RESOURCE = 1,
            EXECUTION_QUEUE = 2,
            EXECUTION_CLUSTER_LABEL = 3,
            QUALITY_OF_SERVICE_SPECIFICATION = 4,
            PLUGIN_OVERRIDE = 5,
            WORKFLOW_EXECUTION_CONFIG = 6,
            CLUSTER_ASSIGNMENT = 7
        }

        /** Properties of a TaskResourceSpec. */
        interface ITaskResourceSpec {

            /** TaskResourceSpec cpu */
            cpu?: (string|null);

            /** TaskResourceSpec gpu */
            gpu?: (string|null);

            /** TaskResourceSpec memory */
            memory?: (string|null);

            /** TaskResourceSpec storage */
            storage?: (string|null);

            /** TaskResourceSpec ephemeralStorage */
            ephemeralStorage?: (string|null);
        }

        /** Represents a TaskResourceSpec. */
        class TaskResourceSpec implements ITaskResourceSpec {

            /**
             * Constructs a new TaskResourceSpec.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ITaskResourceSpec);

            /** TaskResourceSpec cpu. */
            public cpu: string;

            /** TaskResourceSpec gpu. */
            public gpu: string;

            /** TaskResourceSpec memory. */
            public memory: string;

            /** TaskResourceSpec storage. */
            public storage: string;

            /** TaskResourceSpec ephemeralStorage. */
            public ephemeralStorage: string;

            /**
             * Creates a new TaskResourceSpec instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskResourceSpec instance
             */
            public static create(properties?: flyteidl.admin.ITaskResourceSpec): flyteidl.admin.TaskResourceSpec;

            /**
             * Encodes the specified TaskResourceSpec message. Does not implicitly {@link flyteidl.admin.TaskResourceSpec.verify|verify} messages.
             * @param message TaskResourceSpec message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ITaskResourceSpec, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskResourceSpec message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskResourceSpec
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.TaskResourceSpec;

            /**
             * Verifies a TaskResourceSpec message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskResourceAttributes. */
        interface ITaskResourceAttributes {

            /** TaskResourceAttributes defaults */
            defaults?: (flyteidl.admin.ITaskResourceSpec|null);

            /** TaskResourceAttributes limits */
            limits?: (flyteidl.admin.ITaskResourceSpec|null);
        }

        /** Represents a TaskResourceAttributes. */
        class TaskResourceAttributes implements ITaskResourceAttributes {

            /**
             * Constructs a new TaskResourceAttributes.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ITaskResourceAttributes);

            /** TaskResourceAttributes defaults. */
            public defaults?: (flyteidl.admin.ITaskResourceSpec|null);

            /** TaskResourceAttributes limits. */
            public limits?: (flyteidl.admin.ITaskResourceSpec|null);

            /**
             * Creates a new TaskResourceAttributes instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskResourceAttributes instance
             */
            public static create(properties?: flyteidl.admin.ITaskResourceAttributes): flyteidl.admin.TaskResourceAttributes;

            /**
             * Encodes the specified TaskResourceAttributes message. Does not implicitly {@link flyteidl.admin.TaskResourceAttributes.verify|verify} messages.
             * @param message TaskResourceAttributes message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ITaskResourceAttributes, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskResourceAttributes message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskResourceAttributes
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.TaskResourceAttributes;

            /**
             * Verifies a TaskResourceAttributes message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ClusterResourceAttributes. */
        interface IClusterResourceAttributes {

            /** ClusterResourceAttributes attributes */
            attributes?: ({ [k: string]: string }|null);
        }

        /** Represents a ClusterResourceAttributes. */
        class ClusterResourceAttributes implements IClusterResourceAttributes {

            /**
             * Constructs a new ClusterResourceAttributes.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IClusterResourceAttributes);

            /** ClusterResourceAttributes attributes. */
            public attributes: { [k: string]: string };

            /**
             * Creates a new ClusterResourceAttributes instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ClusterResourceAttributes instance
             */
            public static create(properties?: flyteidl.admin.IClusterResourceAttributes): flyteidl.admin.ClusterResourceAttributes;

            /**
             * Encodes the specified ClusterResourceAttributes message. Does not implicitly {@link flyteidl.admin.ClusterResourceAttributes.verify|verify} messages.
             * @param message ClusterResourceAttributes message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IClusterResourceAttributes, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ClusterResourceAttributes message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ClusterResourceAttributes
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ClusterResourceAttributes;

            /**
             * Verifies a ClusterResourceAttributes message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionQueueAttributes. */
        interface IExecutionQueueAttributes {

            /** ExecutionQueueAttributes tags */
            tags?: (string[]|null);
        }

        /** Represents an ExecutionQueueAttributes. */
        class ExecutionQueueAttributes implements IExecutionQueueAttributes {

            /**
             * Constructs a new ExecutionQueueAttributes.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionQueueAttributes);

            /** ExecutionQueueAttributes tags. */
            public tags: string[];

            /**
             * Creates a new ExecutionQueueAttributes instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionQueueAttributes instance
             */
            public static create(properties?: flyteidl.admin.IExecutionQueueAttributes): flyteidl.admin.ExecutionQueueAttributes;

            /**
             * Encodes the specified ExecutionQueueAttributes message. Does not implicitly {@link flyteidl.admin.ExecutionQueueAttributes.verify|verify} messages.
             * @param message ExecutionQueueAttributes message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionQueueAttributes, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionQueueAttributes message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionQueueAttributes
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionQueueAttributes;

            /**
             * Verifies an ExecutionQueueAttributes message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ExecutionClusterLabel. */
        interface IExecutionClusterLabel {

            /** ExecutionClusterLabel value */
            value?: (string|null);
        }

        /** Represents an ExecutionClusterLabel. */
        class ExecutionClusterLabel implements IExecutionClusterLabel {

            /**
             * Constructs a new ExecutionClusterLabel.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IExecutionClusterLabel);

            /** ExecutionClusterLabel value. */
            public value: string;

            /**
             * Creates a new ExecutionClusterLabel instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ExecutionClusterLabel instance
             */
            public static create(properties?: flyteidl.admin.IExecutionClusterLabel): flyteidl.admin.ExecutionClusterLabel;

            /**
             * Encodes the specified ExecutionClusterLabel message. Does not implicitly {@link flyteidl.admin.ExecutionClusterLabel.verify|verify} messages.
             * @param message ExecutionClusterLabel message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IExecutionClusterLabel, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ExecutionClusterLabel message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ExecutionClusterLabel
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ExecutionClusterLabel;

            /**
             * Verifies an ExecutionClusterLabel message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a PluginOverride. */
        interface IPluginOverride {

            /** PluginOverride taskType */
            taskType?: (string|null);

            /** PluginOverride pluginId */
            pluginId?: (string[]|null);

            /** PluginOverride missingPluginBehavior */
            missingPluginBehavior?: (flyteidl.admin.PluginOverride.MissingPluginBehavior|null);
        }

        /** Represents a PluginOverride. */
        class PluginOverride implements IPluginOverride {

            /**
             * Constructs a new PluginOverride.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IPluginOverride);

            /** PluginOverride taskType. */
            public taskType: string;

            /** PluginOverride pluginId. */
            public pluginId: string[];

            /** PluginOverride missingPluginBehavior. */
            public missingPluginBehavior: flyteidl.admin.PluginOverride.MissingPluginBehavior;

            /**
             * Creates a new PluginOverride instance using the specified properties.
             * @param [properties] Properties to set
             * @returns PluginOverride instance
             */
            public static create(properties?: flyteidl.admin.IPluginOverride): flyteidl.admin.PluginOverride;

            /**
             * Encodes the specified PluginOverride message. Does not implicitly {@link flyteidl.admin.PluginOverride.verify|verify} messages.
             * @param message PluginOverride message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IPluginOverride, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a PluginOverride message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns PluginOverride
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.PluginOverride;

            /**
             * Verifies a PluginOverride message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace PluginOverride {

            /** MissingPluginBehavior enum. */
            enum MissingPluginBehavior {
                FAIL = 0,
                USE_DEFAULT = 1
            }
        }

        /** Properties of a PluginOverrides. */
        interface IPluginOverrides {

            /** PluginOverrides overrides */
            overrides?: (flyteidl.admin.IPluginOverride[]|null);
        }

        /** Represents a PluginOverrides. */
        class PluginOverrides implements IPluginOverrides {

            /**
             * Constructs a new PluginOverrides.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IPluginOverrides);

            /** PluginOverrides overrides. */
            public overrides: flyteidl.admin.IPluginOverride[];

            /**
             * Creates a new PluginOverrides instance using the specified properties.
             * @param [properties] Properties to set
             * @returns PluginOverrides instance
             */
            public static create(properties?: flyteidl.admin.IPluginOverrides): flyteidl.admin.PluginOverrides;

            /**
             * Encodes the specified PluginOverrides message. Does not implicitly {@link flyteidl.admin.PluginOverrides.verify|verify} messages.
             * @param message PluginOverrides message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IPluginOverrides, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a PluginOverrides message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns PluginOverrides
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.PluginOverrides;

            /**
             * Verifies a PluginOverrides message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowExecutionConfig. */
        interface IWorkflowExecutionConfig {

            /** WorkflowExecutionConfig maxParallelism */
            maxParallelism?: (number|null);

            /** WorkflowExecutionConfig securityContext */
            securityContext?: (flyteidl.core.ISecurityContext|null);

            /** WorkflowExecutionConfig rawOutputDataConfig */
            rawOutputDataConfig?: (flyteidl.admin.IRawOutputDataConfig|null);

            /** WorkflowExecutionConfig labels */
            labels?: (flyteidl.admin.ILabels|null);

            /** WorkflowExecutionConfig annotations */
            annotations?: (flyteidl.admin.IAnnotations|null);

            /** WorkflowExecutionConfig interruptible */
            interruptible?: (google.protobuf.IBoolValue|null);

            /** WorkflowExecutionConfig overwriteCache */
            overwriteCache?: (boolean|null);

            /** WorkflowExecutionConfig envs */
            envs?: (flyteidl.admin.IEnvs|null);

            /** WorkflowExecutionConfig executionEnvAssignments */
            executionEnvAssignments?: (flyteidl.core.IExecutionEnvAssignment[]|null);
        }

        /** Represents a WorkflowExecutionConfig. */
        class WorkflowExecutionConfig implements IWorkflowExecutionConfig {

            /**
             * Constructs a new WorkflowExecutionConfig.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowExecutionConfig);

            /** WorkflowExecutionConfig maxParallelism. */
            public maxParallelism: number;

            /** WorkflowExecutionConfig securityContext. */
            public securityContext?: (flyteidl.core.ISecurityContext|null);

            /** WorkflowExecutionConfig rawOutputDataConfig. */
            public rawOutputDataConfig?: (flyteidl.admin.IRawOutputDataConfig|null);

            /** WorkflowExecutionConfig labels. */
            public labels?: (flyteidl.admin.ILabels|null);

            /** WorkflowExecutionConfig annotations. */
            public annotations?: (flyteidl.admin.IAnnotations|null);

            /** WorkflowExecutionConfig interruptible. */
            public interruptible?: (google.protobuf.IBoolValue|null);

            /** WorkflowExecutionConfig overwriteCache. */
            public overwriteCache: boolean;

            /** WorkflowExecutionConfig envs. */
            public envs?: (flyteidl.admin.IEnvs|null);

            /** WorkflowExecutionConfig executionEnvAssignments. */
            public executionEnvAssignments: flyteidl.core.IExecutionEnvAssignment[];

            /**
             * Creates a new WorkflowExecutionConfig instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowExecutionConfig instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowExecutionConfig): flyteidl.admin.WorkflowExecutionConfig;

            /**
             * Encodes the specified WorkflowExecutionConfig message. Does not implicitly {@link flyteidl.admin.WorkflowExecutionConfig.verify|verify} messages.
             * @param message WorkflowExecutionConfig message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowExecutionConfig, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowExecutionConfig message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowExecutionConfig
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowExecutionConfig;

            /**
             * Verifies a WorkflowExecutionConfig message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a MatchingAttributes. */
        interface IMatchingAttributes {

            /** MatchingAttributes taskResourceAttributes */
            taskResourceAttributes?: (flyteidl.admin.ITaskResourceAttributes|null);

            /** MatchingAttributes clusterResourceAttributes */
            clusterResourceAttributes?: (flyteidl.admin.IClusterResourceAttributes|null);

            /** MatchingAttributes executionQueueAttributes */
            executionQueueAttributes?: (flyteidl.admin.IExecutionQueueAttributes|null);

            /** MatchingAttributes executionClusterLabel */
            executionClusterLabel?: (flyteidl.admin.IExecutionClusterLabel|null);

            /** MatchingAttributes qualityOfService */
            qualityOfService?: (flyteidl.core.IQualityOfService|null);

            /** MatchingAttributes pluginOverrides */
            pluginOverrides?: (flyteidl.admin.IPluginOverrides|null);

            /** MatchingAttributes workflowExecutionConfig */
            workflowExecutionConfig?: (flyteidl.admin.IWorkflowExecutionConfig|null);

            /** MatchingAttributes clusterAssignment */
            clusterAssignment?: (flyteidl.admin.IClusterAssignment|null);
        }

        /** Represents a MatchingAttributes. */
        class MatchingAttributes implements IMatchingAttributes {

            /**
             * Constructs a new MatchingAttributes.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IMatchingAttributes);

            /** MatchingAttributes taskResourceAttributes. */
            public taskResourceAttributes?: (flyteidl.admin.ITaskResourceAttributes|null);

            /** MatchingAttributes clusterResourceAttributes. */
            public clusterResourceAttributes?: (flyteidl.admin.IClusterResourceAttributes|null);

            /** MatchingAttributes executionQueueAttributes. */
            public executionQueueAttributes?: (flyteidl.admin.IExecutionQueueAttributes|null);

            /** MatchingAttributes executionClusterLabel. */
            public executionClusterLabel?: (flyteidl.admin.IExecutionClusterLabel|null);

            /** MatchingAttributes qualityOfService. */
            public qualityOfService?: (flyteidl.core.IQualityOfService|null);

            /** MatchingAttributes pluginOverrides. */
            public pluginOverrides?: (flyteidl.admin.IPluginOverrides|null);

            /** MatchingAttributes workflowExecutionConfig. */
            public workflowExecutionConfig?: (flyteidl.admin.IWorkflowExecutionConfig|null);

            /** MatchingAttributes clusterAssignment. */
            public clusterAssignment?: (flyteidl.admin.IClusterAssignment|null);

            /** MatchingAttributes target. */
            public target?: ("taskResourceAttributes"|"clusterResourceAttributes"|"executionQueueAttributes"|"executionClusterLabel"|"qualityOfService"|"pluginOverrides"|"workflowExecutionConfig"|"clusterAssignment");

            /**
             * Creates a new MatchingAttributes instance using the specified properties.
             * @param [properties] Properties to set
             * @returns MatchingAttributes instance
             */
            public static create(properties?: flyteidl.admin.IMatchingAttributes): flyteidl.admin.MatchingAttributes;

            /**
             * Encodes the specified MatchingAttributes message. Does not implicitly {@link flyteidl.admin.MatchingAttributes.verify|verify} messages.
             * @param message MatchingAttributes message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IMatchingAttributes, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a MatchingAttributes message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns MatchingAttributes
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.MatchingAttributes;

            /**
             * Verifies a MatchingAttributes message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a MatchableAttributesConfiguration. */
        interface IMatchableAttributesConfiguration {

            /** MatchableAttributesConfiguration attributes */
            attributes?: (flyteidl.admin.IMatchingAttributes|null);

            /** MatchableAttributesConfiguration domain */
            domain?: (string|null);

            /** MatchableAttributesConfiguration project */
            project?: (string|null);

            /** MatchableAttributesConfiguration workflow */
            workflow?: (string|null);

            /** MatchableAttributesConfiguration launchPlan */
            launchPlan?: (string|null);

            /** MatchableAttributesConfiguration org */
            org?: (string|null);
        }

        /** Represents a MatchableAttributesConfiguration. */
        class MatchableAttributesConfiguration implements IMatchableAttributesConfiguration {

            /**
             * Constructs a new MatchableAttributesConfiguration.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IMatchableAttributesConfiguration);

            /** MatchableAttributesConfiguration attributes. */
            public attributes?: (flyteidl.admin.IMatchingAttributes|null);

            /** MatchableAttributesConfiguration domain. */
            public domain: string;

            /** MatchableAttributesConfiguration project. */
            public project: string;

            /** MatchableAttributesConfiguration workflow. */
            public workflow: string;

            /** MatchableAttributesConfiguration launchPlan. */
            public launchPlan: string;

            /** MatchableAttributesConfiguration org. */
            public org: string;

            /**
             * Creates a new MatchableAttributesConfiguration instance using the specified properties.
             * @param [properties] Properties to set
             * @returns MatchableAttributesConfiguration instance
             */
            public static create(properties?: flyteidl.admin.IMatchableAttributesConfiguration): flyteidl.admin.MatchableAttributesConfiguration;

            /**
             * Encodes the specified MatchableAttributesConfiguration message. Does not implicitly {@link flyteidl.admin.MatchableAttributesConfiguration.verify|verify} messages.
             * @param message MatchableAttributesConfiguration message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IMatchableAttributesConfiguration, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a MatchableAttributesConfiguration message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns MatchableAttributesConfiguration
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.MatchableAttributesConfiguration;

            /**
             * Verifies a MatchableAttributesConfiguration message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ListMatchableAttributesRequest. */
        interface IListMatchableAttributesRequest {

            /** ListMatchableAttributesRequest resourceType */
            resourceType?: (flyteidl.admin.MatchableResource|null);

            /** ListMatchableAttributesRequest org */
            org?: (string|null);
        }

        /** Represents a ListMatchableAttributesRequest. */
        class ListMatchableAttributesRequest implements IListMatchableAttributesRequest {

            /**
             * Constructs a new ListMatchableAttributesRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IListMatchableAttributesRequest);

            /** ListMatchableAttributesRequest resourceType. */
            public resourceType: flyteidl.admin.MatchableResource;

            /** ListMatchableAttributesRequest org. */
            public org: string;

            /**
             * Creates a new ListMatchableAttributesRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ListMatchableAttributesRequest instance
             */
            public static create(properties?: flyteidl.admin.IListMatchableAttributesRequest): flyteidl.admin.ListMatchableAttributesRequest;

            /**
             * Encodes the specified ListMatchableAttributesRequest message. Does not implicitly {@link flyteidl.admin.ListMatchableAttributesRequest.verify|verify} messages.
             * @param message ListMatchableAttributesRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IListMatchableAttributesRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ListMatchableAttributesRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ListMatchableAttributesRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ListMatchableAttributesRequest;

            /**
             * Verifies a ListMatchableAttributesRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ListMatchableAttributesResponse. */
        interface IListMatchableAttributesResponse {

            /** ListMatchableAttributesResponse configurations */
            configurations?: (flyteidl.admin.IMatchableAttributesConfiguration[]|null);
        }

        /** Represents a ListMatchableAttributesResponse. */
        class ListMatchableAttributesResponse implements IListMatchableAttributesResponse {

            /**
             * Constructs a new ListMatchableAttributesResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IListMatchableAttributesResponse);

            /** ListMatchableAttributesResponse configurations. */
            public configurations: flyteidl.admin.IMatchableAttributesConfiguration[];

            /**
             * Creates a new ListMatchableAttributesResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ListMatchableAttributesResponse instance
             */
            public static create(properties?: flyteidl.admin.IListMatchableAttributesResponse): flyteidl.admin.ListMatchableAttributesResponse;

            /**
             * Encodes the specified ListMatchableAttributesResponse message. Does not implicitly {@link flyteidl.admin.ListMatchableAttributesResponse.verify|verify} messages.
             * @param message ListMatchableAttributesResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IListMatchableAttributesResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ListMatchableAttributesResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ListMatchableAttributesResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ListMatchableAttributesResponse;

            /**
             * Verifies a ListMatchableAttributesResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanCreateRequest. */
        interface ILaunchPlanCreateRequest {

            /** LaunchPlanCreateRequest id */
            id?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlanCreateRequest spec */
            spec?: (flyteidl.admin.ILaunchPlanSpec|null);
        }

        /** Represents a LaunchPlanCreateRequest. */
        class LaunchPlanCreateRequest implements ILaunchPlanCreateRequest {

            /**
             * Constructs a new LaunchPlanCreateRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlanCreateRequest);

            /** LaunchPlanCreateRequest id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlanCreateRequest spec. */
            public spec?: (flyteidl.admin.ILaunchPlanSpec|null);

            /**
             * Creates a new LaunchPlanCreateRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanCreateRequest instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlanCreateRequest): flyteidl.admin.LaunchPlanCreateRequest;

            /**
             * Encodes the specified LaunchPlanCreateRequest message. Does not implicitly {@link flyteidl.admin.LaunchPlanCreateRequest.verify|verify} messages.
             * @param message LaunchPlanCreateRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlanCreateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanCreateRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanCreateRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlanCreateRequest;

            /**
             * Verifies a LaunchPlanCreateRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanCreateResponse. */
        interface ILaunchPlanCreateResponse {
        }

        /** Represents a LaunchPlanCreateResponse. */
        class LaunchPlanCreateResponse implements ILaunchPlanCreateResponse {

            /**
             * Constructs a new LaunchPlanCreateResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlanCreateResponse);

            /**
             * Creates a new LaunchPlanCreateResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanCreateResponse instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlanCreateResponse): flyteidl.admin.LaunchPlanCreateResponse;

            /**
             * Encodes the specified LaunchPlanCreateResponse message. Does not implicitly {@link flyteidl.admin.LaunchPlanCreateResponse.verify|verify} messages.
             * @param message LaunchPlanCreateResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlanCreateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanCreateResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanCreateResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlanCreateResponse;

            /**
             * Verifies a LaunchPlanCreateResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** LaunchPlanState enum. */
        enum LaunchPlanState {
            INACTIVE = 0,
            ACTIVE = 1
        }

        /** Properties of a LaunchPlan. */
        interface ILaunchPlan {

            /** LaunchPlan id */
            id?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlan spec */
            spec?: (flyteidl.admin.ILaunchPlanSpec|null);

            /** LaunchPlan closure */
            closure?: (flyteidl.admin.ILaunchPlanClosure|null);
        }

        /** Represents a LaunchPlan. */
        class LaunchPlan implements ILaunchPlan {

            /**
             * Constructs a new LaunchPlan.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlan);

            /** LaunchPlan id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlan spec. */
            public spec?: (flyteidl.admin.ILaunchPlanSpec|null);

            /** LaunchPlan closure. */
            public closure?: (flyteidl.admin.ILaunchPlanClosure|null);

            /**
             * Creates a new LaunchPlan instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlan instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlan): flyteidl.admin.LaunchPlan;

            /**
             * Encodes the specified LaunchPlan message. Does not implicitly {@link flyteidl.admin.LaunchPlan.verify|verify} messages.
             * @param message LaunchPlan message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlan, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlan message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlan
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlan;

            /**
             * Verifies a LaunchPlan message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanList. */
        interface ILaunchPlanList {

            /** LaunchPlanList launchPlans */
            launchPlans?: (flyteidl.admin.ILaunchPlan[]|null);

            /** LaunchPlanList token */
            token?: (string|null);
        }

        /** Represents a LaunchPlanList. */
        class LaunchPlanList implements ILaunchPlanList {

            /**
             * Constructs a new LaunchPlanList.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlanList);

            /** LaunchPlanList launchPlans. */
            public launchPlans: flyteidl.admin.ILaunchPlan[];

            /** LaunchPlanList token. */
            public token: string;

            /**
             * Creates a new LaunchPlanList instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanList instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlanList): flyteidl.admin.LaunchPlanList;

            /**
             * Encodes the specified LaunchPlanList message. Does not implicitly {@link flyteidl.admin.LaunchPlanList.verify|verify} messages.
             * @param message LaunchPlanList message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlanList, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanList message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanList
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlanList;

            /**
             * Verifies a LaunchPlanList message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an Auth. */
        interface IAuth {

            /** Auth assumableIamRole */
            assumableIamRole?: (string|null);

            /** Auth kubernetesServiceAccount */
            kubernetesServiceAccount?: (string|null);
        }

        /** Represents an Auth. */
        class Auth implements IAuth {

            /**
             * Constructs a new Auth.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IAuth);

            /** Auth assumableIamRole. */
            public assumableIamRole: string;

            /** Auth kubernetesServiceAccount. */
            public kubernetesServiceAccount: string;

            /**
             * Creates a new Auth instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Auth instance
             */
            public static create(properties?: flyteidl.admin.IAuth): flyteidl.admin.Auth;

            /**
             * Encodes the specified Auth message. Does not implicitly {@link flyteidl.admin.Auth.verify|verify} messages.
             * @param message Auth message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IAuth, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Auth message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Auth
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Auth;

            /**
             * Verifies an Auth message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanSpec. */
        interface ILaunchPlanSpec {

            /** LaunchPlanSpec workflowId */
            workflowId?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlanSpec entityMetadata */
            entityMetadata?: (flyteidl.admin.ILaunchPlanMetadata|null);

            /** LaunchPlanSpec defaultInputs */
            defaultInputs?: (flyteidl.core.IParameterMap|null);

            /** LaunchPlanSpec fixedInputs */
            fixedInputs?: (flyteidl.core.ILiteralMap|null);

            /** LaunchPlanSpec role */
            role?: (string|null);

            /** LaunchPlanSpec labels */
            labels?: (flyteidl.admin.ILabels|null);

            /** LaunchPlanSpec annotations */
            annotations?: (flyteidl.admin.IAnnotations|null);

            /** LaunchPlanSpec auth */
            auth?: (flyteidl.admin.IAuth|null);

            /** LaunchPlanSpec authRole */
            authRole?: (flyteidl.admin.IAuthRole|null);

            /** LaunchPlanSpec securityContext */
            securityContext?: (flyteidl.core.ISecurityContext|null);

            /** LaunchPlanSpec qualityOfService */
            qualityOfService?: (flyteidl.core.IQualityOfService|null);

            /** LaunchPlanSpec rawOutputDataConfig */
            rawOutputDataConfig?: (flyteidl.admin.IRawOutputDataConfig|null);

            /** LaunchPlanSpec maxParallelism */
            maxParallelism?: (number|null);

            /** LaunchPlanSpec interruptible */
            interruptible?: (google.protobuf.IBoolValue|null);

            /** LaunchPlanSpec overwriteCache */
            overwriteCache?: (boolean|null);

            /** LaunchPlanSpec envs */
            envs?: (flyteidl.admin.IEnvs|null);

            /** LaunchPlanSpec executionEnvAssignments */
            executionEnvAssignments?: (flyteidl.core.IExecutionEnvAssignment[]|null);
        }

        /** Represents a LaunchPlanSpec. */
        class LaunchPlanSpec implements ILaunchPlanSpec {

            /**
             * Constructs a new LaunchPlanSpec.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlanSpec);

            /** LaunchPlanSpec workflowId. */
            public workflowId?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlanSpec entityMetadata. */
            public entityMetadata?: (flyteidl.admin.ILaunchPlanMetadata|null);

            /** LaunchPlanSpec defaultInputs. */
            public defaultInputs?: (flyteidl.core.IParameterMap|null);

            /** LaunchPlanSpec fixedInputs. */
            public fixedInputs?: (flyteidl.core.ILiteralMap|null);

            /** LaunchPlanSpec role. */
            public role: string;

            /** LaunchPlanSpec labels. */
            public labels?: (flyteidl.admin.ILabels|null);

            /** LaunchPlanSpec annotations. */
            public annotations?: (flyteidl.admin.IAnnotations|null);

            /** LaunchPlanSpec auth. */
            public auth?: (flyteidl.admin.IAuth|null);

            /** LaunchPlanSpec authRole. */
            public authRole?: (flyteidl.admin.IAuthRole|null);

            /** LaunchPlanSpec securityContext. */
            public securityContext?: (flyteidl.core.ISecurityContext|null);

            /** LaunchPlanSpec qualityOfService. */
            public qualityOfService?: (flyteidl.core.IQualityOfService|null);

            /** LaunchPlanSpec rawOutputDataConfig. */
            public rawOutputDataConfig?: (flyteidl.admin.IRawOutputDataConfig|null);

            /** LaunchPlanSpec maxParallelism. */
            public maxParallelism: number;

            /** LaunchPlanSpec interruptible. */
            public interruptible?: (google.protobuf.IBoolValue|null);

            /** LaunchPlanSpec overwriteCache. */
            public overwriteCache: boolean;

            /** LaunchPlanSpec envs. */
            public envs?: (flyteidl.admin.IEnvs|null);

            /** LaunchPlanSpec executionEnvAssignments. */
            public executionEnvAssignments: flyteidl.core.IExecutionEnvAssignment[];

            /**
             * Creates a new LaunchPlanSpec instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanSpec instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlanSpec): flyteidl.admin.LaunchPlanSpec;

            /**
             * Encodes the specified LaunchPlanSpec message. Does not implicitly {@link flyteidl.admin.LaunchPlanSpec.verify|verify} messages.
             * @param message LaunchPlanSpec message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlanSpec, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanSpec message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanSpec
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlanSpec;

            /**
             * Verifies a LaunchPlanSpec message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanClosure. */
        interface ILaunchPlanClosure {

            /** LaunchPlanClosure state */
            state?: (flyteidl.admin.LaunchPlanState|null);

            /** LaunchPlanClosure expectedInputs */
            expectedInputs?: (flyteidl.core.IParameterMap|null);

            /** LaunchPlanClosure expectedOutputs */
            expectedOutputs?: (flyteidl.core.IVariableMap|null);

            /** LaunchPlanClosure createdAt */
            createdAt?: (google.protobuf.ITimestamp|null);

            /** LaunchPlanClosure updatedAt */
            updatedAt?: (google.protobuf.ITimestamp|null);
        }

        /** Represents a LaunchPlanClosure. */
        class LaunchPlanClosure implements ILaunchPlanClosure {

            /**
             * Constructs a new LaunchPlanClosure.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlanClosure);

            /** LaunchPlanClosure state. */
            public state: flyteidl.admin.LaunchPlanState;

            /** LaunchPlanClosure expectedInputs. */
            public expectedInputs?: (flyteidl.core.IParameterMap|null);

            /** LaunchPlanClosure expectedOutputs. */
            public expectedOutputs?: (flyteidl.core.IVariableMap|null);

            /** LaunchPlanClosure createdAt. */
            public createdAt?: (google.protobuf.ITimestamp|null);

            /** LaunchPlanClosure updatedAt. */
            public updatedAt?: (google.protobuf.ITimestamp|null);

            /**
             * Creates a new LaunchPlanClosure instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanClosure instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlanClosure): flyteidl.admin.LaunchPlanClosure;

            /**
             * Encodes the specified LaunchPlanClosure message. Does not implicitly {@link flyteidl.admin.LaunchPlanClosure.verify|verify} messages.
             * @param message LaunchPlanClosure message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlanClosure, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanClosure message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanClosure
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlanClosure;

            /**
             * Verifies a LaunchPlanClosure message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanMetadata. */
        interface ILaunchPlanMetadata {

            /** LaunchPlanMetadata schedule */
            schedule?: (flyteidl.admin.ISchedule|null);

            /** LaunchPlanMetadata notifications */
            notifications?: (flyteidl.admin.INotification[]|null);

            /** LaunchPlanMetadata launchConditions */
            launchConditions?: (google.protobuf.IAny|null);
        }

        /** Represents a LaunchPlanMetadata. */
        class LaunchPlanMetadata implements ILaunchPlanMetadata {

            /**
             * Constructs a new LaunchPlanMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlanMetadata);

            /** LaunchPlanMetadata schedule. */
            public schedule?: (flyteidl.admin.ISchedule|null);

            /** LaunchPlanMetadata notifications. */
            public notifications: flyteidl.admin.INotification[];

            /** LaunchPlanMetadata launchConditions. */
            public launchConditions?: (google.protobuf.IAny|null);

            /**
             * Creates a new LaunchPlanMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanMetadata instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlanMetadata): flyteidl.admin.LaunchPlanMetadata;

            /**
             * Encodes the specified LaunchPlanMetadata message. Does not implicitly {@link flyteidl.admin.LaunchPlanMetadata.verify|verify} messages.
             * @param message LaunchPlanMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlanMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlanMetadata;

            /**
             * Verifies a LaunchPlanMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanUpdateRequest. */
        interface ILaunchPlanUpdateRequest {

            /** LaunchPlanUpdateRequest id */
            id?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlanUpdateRequest state */
            state?: (flyteidl.admin.LaunchPlanState|null);
        }

        /** Represents a LaunchPlanUpdateRequest. */
        class LaunchPlanUpdateRequest implements ILaunchPlanUpdateRequest {

            /**
             * Constructs a new LaunchPlanUpdateRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlanUpdateRequest);

            /** LaunchPlanUpdateRequest id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** LaunchPlanUpdateRequest state. */
            public state: flyteidl.admin.LaunchPlanState;

            /**
             * Creates a new LaunchPlanUpdateRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanUpdateRequest instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlanUpdateRequest): flyteidl.admin.LaunchPlanUpdateRequest;

            /**
             * Encodes the specified LaunchPlanUpdateRequest message. Does not implicitly {@link flyteidl.admin.LaunchPlanUpdateRequest.verify|verify} messages.
             * @param message LaunchPlanUpdateRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlanUpdateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanUpdateRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanUpdateRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlanUpdateRequest;

            /**
             * Verifies a LaunchPlanUpdateRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a LaunchPlanUpdateResponse. */
        interface ILaunchPlanUpdateResponse {
        }

        /** Represents a LaunchPlanUpdateResponse. */
        class LaunchPlanUpdateResponse implements ILaunchPlanUpdateResponse {

            /**
             * Constructs a new LaunchPlanUpdateResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ILaunchPlanUpdateResponse);

            /**
             * Creates a new LaunchPlanUpdateResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns LaunchPlanUpdateResponse instance
             */
            public static create(properties?: flyteidl.admin.ILaunchPlanUpdateResponse): flyteidl.admin.LaunchPlanUpdateResponse;

            /**
             * Encodes the specified LaunchPlanUpdateResponse message. Does not implicitly {@link flyteidl.admin.LaunchPlanUpdateResponse.verify|verify} messages.
             * @param message LaunchPlanUpdateResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ILaunchPlanUpdateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a LaunchPlanUpdateResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns LaunchPlanUpdateResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.LaunchPlanUpdateResponse;

            /**
             * Verifies a LaunchPlanUpdateResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ActiveLaunchPlanRequest. */
        interface IActiveLaunchPlanRequest {

            /** ActiveLaunchPlanRequest id */
            id?: (flyteidl.admin.INamedEntityIdentifier|null);
        }

        /** Represents an ActiveLaunchPlanRequest. */
        class ActiveLaunchPlanRequest implements IActiveLaunchPlanRequest {

            /**
             * Constructs a new ActiveLaunchPlanRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IActiveLaunchPlanRequest);

            /** ActiveLaunchPlanRequest id. */
            public id?: (flyteidl.admin.INamedEntityIdentifier|null);

            /**
             * Creates a new ActiveLaunchPlanRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ActiveLaunchPlanRequest instance
             */
            public static create(properties?: flyteidl.admin.IActiveLaunchPlanRequest): flyteidl.admin.ActiveLaunchPlanRequest;

            /**
             * Encodes the specified ActiveLaunchPlanRequest message. Does not implicitly {@link flyteidl.admin.ActiveLaunchPlanRequest.verify|verify} messages.
             * @param message ActiveLaunchPlanRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IActiveLaunchPlanRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ActiveLaunchPlanRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ActiveLaunchPlanRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ActiveLaunchPlanRequest;

            /**
             * Verifies an ActiveLaunchPlanRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an ActiveLaunchPlanListRequest. */
        interface IActiveLaunchPlanListRequest {

            /** ActiveLaunchPlanListRequest project */
            project?: (string|null);

            /** ActiveLaunchPlanListRequest domain */
            domain?: (string|null);

            /** ActiveLaunchPlanListRequest limit */
            limit?: (number|null);

            /** ActiveLaunchPlanListRequest token */
            token?: (string|null);

            /** ActiveLaunchPlanListRequest sortBy */
            sortBy?: (flyteidl.admin.ISort|null);

            /** ActiveLaunchPlanListRequest org */
            org?: (string|null);
        }

        /** Represents an ActiveLaunchPlanListRequest. */
        class ActiveLaunchPlanListRequest implements IActiveLaunchPlanListRequest {

            /**
             * Constructs a new ActiveLaunchPlanListRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IActiveLaunchPlanListRequest);

            /** ActiveLaunchPlanListRequest project. */
            public project: string;

            /** ActiveLaunchPlanListRequest domain. */
            public domain: string;

            /** ActiveLaunchPlanListRequest limit. */
            public limit: number;

            /** ActiveLaunchPlanListRequest token. */
            public token: string;

            /** ActiveLaunchPlanListRequest sortBy. */
            public sortBy?: (flyteidl.admin.ISort|null);

            /** ActiveLaunchPlanListRequest org. */
            public org: string;

            /**
             * Creates a new ActiveLaunchPlanListRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns ActiveLaunchPlanListRequest instance
             */
            public static create(properties?: flyteidl.admin.IActiveLaunchPlanListRequest): flyteidl.admin.ActiveLaunchPlanListRequest;

            /**
             * Encodes the specified ActiveLaunchPlanListRequest message. Does not implicitly {@link flyteidl.admin.ActiveLaunchPlanListRequest.verify|verify} messages.
             * @param message ActiveLaunchPlanListRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IActiveLaunchPlanListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an ActiveLaunchPlanListRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ActiveLaunchPlanListRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.ActiveLaunchPlanListRequest;

            /**
             * Verifies an ActiveLaunchPlanListRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** FixedRateUnit enum. */
        enum FixedRateUnit {
            MINUTE = 0,
            HOUR = 1,
            DAY = 2
        }

        /** Properties of a FixedRate. */
        interface IFixedRate {

            /** FixedRate value */
            value?: (number|null);

            /** FixedRate unit */
            unit?: (flyteidl.admin.FixedRateUnit|null);
        }

        /** Represents a FixedRate. */
        class FixedRate implements IFixedRate {

            /**
             * Constructs a new FixedRate.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IFixedRate);

            /** FixedRate value. */
            public value: number;

            /** FixedRate unit. */
            public unit: flyteidl.admin.FixedRateUnit;

            /**
             * Creates a new FixedRate instance using the specified properties.
             * @param [properties] Properties to set
             * @returns FixedRate instance
             */
            public static create(properties?: flyteidl.admin.IFixedRate): flyteidl.admin.FixedRate;

            /**
             * Encodes the specified FixedRate message. Does not implicitly {@link flyteidl.admin.FixedRate.verify|verify} messages.
             * @param message FixedRate message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IFixedRate, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a FixedRate message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns FixedRate
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.FixedRate;

            /**
             * Verifies a FixedRate message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a CronSchedule. */
        interface ICronSchedule {

            /** CronSchedule schedule */
            schedule?: (string|null);

            /** CronSchedule offset */
            offset?: (string|null);
        }

        /** Represents a CronSchedule. */
        class CronSchedule implements ICronSchedule {

            /**
             * Constructs a new CronSchedule.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ICronSchedule);

            /** CronSchedule schedule. */
            public schedule: string;

            /** CronSchedule offset. */
            public offset: string;

            /**
             * Creates a new CronSchedule instance using the specified properties.
             * @param [properties] Properties to set
             * @returns CronSchedule instance
             */
            public static create(properties?: flyteidl.admin.ICronSchedule): flyteidl.admin.CronSchedule;

            /**
             * Encodes the specified CronSchedule message. Does not implicitly {@link flyteidl.admin.CronSchedule.verify|verify} messages.
             * @param message CronSchedule message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ICronSchedule, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CronSchedule message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CronSchedule
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.CronSchedule;

            /**
             * Verifies a CronSchedule message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Schedule. */
        interface ISchedule {

            /** Schedule cronExpression */
            cronExpression?: (string|null);

            /** Schedule rate */
            rate?: (flyteidl.admin.IFixedRate|null);

            /** Schedule cronSchedule */
            cronSchedule?: (flyteidl.admin.ICronSchedule|null);

            /** Schedule kickoffTimeInputArg */
            kickoffTimeInputArg?: (string|null);
        }

        /** Represents a Schedule. */
        class Schedule implements ISchedule {

            /**
             * Constructs a new Schedule.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ISchedule);

            /** Schedule cronExpression. */
            public cronExpression: string;

            /** Schedule rate. */
            public rate?: (flyteidl.admin.IFixedRate|null);

            /** Schedule cronSchedule. */
            public cronSchedule?: (flyteidl.admin.ICronSchedule|null);

            /** Schedule kickoffTimeInputArg. */
            public kickoffTimeInputArg: string;

            /** Schedule ScheduleExpression. */
            public ScheduleExpression?: ("cronExpression"|"rate"|"cronSchedule");

            /**
             * Creates a new Schedule instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Schedule instance
             */
            public static create(properties?: flyteidl.admin.ISchedule): flyteidl.admin.Schedule;

            /**
             * Encodes the specified Schedule message. Does not implicitly {@link flyteidl.admin.Schedule.verify|verify} messages.
             * @param message Schedule message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ISchedule, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Schedule message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Schedule
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Schedule;

            /**
             * Verifies a Schedule message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionGetRequest. */
        interface INodeExecutionGetRequest {

            /** NodeExecutionGetRequest id */
            id?: (flyteidl.core.INodeExecutionIdentifier|null);
        }

        /** Represents a NodeExecutionGetRequest. */
        class NodeExecutionGetRequest implements INodeExecutionGetRequest {

            /**
             * Constructs a new NodeExecutionGetRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionGetRequest);

            /** NodeExecutionGetRequest id. */
            public id?: (flyteidl.core.INodeExecutionIdentifier|null);

            /**
             * Creates a new NodeExecutionGetRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionGetRequest instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionGetRequest): flyteidl.admin.NodeExecutionGetRequest;

            /**
             * Encodes the specified NodeExecutionGetRequest message. Does not implicitly {@link flyteidl.admin.NodeExecutionGetRequest.verify|verify} messages.
             * @param message NodeExecutionGetRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionGetRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionGetRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionGetRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionGetRequest;

            /**
             * Verifies a NodeExecutionGetRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionListRequest. */
        interface INodeExecutionListRequest {

            /** NodeExecutionListRequest workflowExecutionId */
            workflowExecutionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** NodeExecutionListRequest limit */
            limit?: (number|null);

            /** NodeExecutionListRequest token */
            token?: (string|null);

            /** NodeExecutionListRequest filters */
            filters?: (string|null);

            /** NodeExecutionListRequest sortBy */
            sortBy?: (flyteidl.admin.ISort|null);

            /** NodeExecutionListRequest uniqueParentId */
            uniqueParentId?: (string|null);
        }

        /** Represents a NodeExecutionListRequest. */
        class NodeExecutionListRequest implements INodeExecutionListRequest {

            /**
             * Constructs a new NodeExecutionListRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionListRequest);

            /** NodeExecutionListRequest workflowExecutionId. */
            public workflowExecutionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /** NodeExecutionListRequest limit. */
            public limit: number;

            /** NodeExecutionListRequest token. */
            public token: string;

            /** NodeExecutionListRequest filters. */
            public filters: string;

            /** NodeExecutionListRequest sortBy. */
            public sortBy?: (flyteidl.admin.ISort|null);

            /** NodeExecutionListRequest uniqueParentId. */
            public uniqueParentId: string;

            /**
             * Creates a new NodeExecutionListRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionListRequest instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionListRequest): flyteidl.admin.NodeExecutionListRequest;

            /**
             * Encodes the specified NodeExecutionListRequest message. Does not implicitly {@link flyteidl.admin.NodeExecutionListRequest.verify|verify} messages.
             * @param message NodeExecutionListRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionListRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionListRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionListRequest;

            /**
             * Verifies a NodeExecutionListRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionForTaskListRequest. */
        interface INodeExecutionForTaskListRequest {

            /** NodeExecutionForTaskListRequest taskExecutionId */
            taskExecutionId?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** NodeExecutionForTaskListRequest limit */
            limit?: (number|null);

            /** NodeExecutionForTaskListRequest token */
            token?: (string|null);

            /** NodeExecutionForTaskListRequest filters */
            filters?: (string|null);

            /** NodeExecutionForTaskListRequest sortBy */
            sortBy?: (flyteidl.admin.ISort|null);
        }

        /** Represents a NodeExecutionForTaskListRequest. */
        class NodeExecutionForTaskListRequest implements INodeExecutionForTaskListRequest {

            /**
             * Constructs a new NodeExecutionForTaskListRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionForTaskListRequest);

            /** NodeExecutionForTaskListRequest taskExecutionId. */
            public taskExecutionId?: (flyteidl.core.ITaskExecutionIdentifier|null);

            /** NodeExecutionForTaskListRequest limit. */
            public limit: number;

            /** NodeExecutionForTaskListRequest token. */
            public token: string;

            /** NodeExecutionForTaskListRequest filters. */
            public filters: string;

            /** NodeExecutionForTaskListRequest sortBy. */
            public sortBy?: (flyteidl.admin.ISort|null);

            /**
             * Creates a new NodeExecutionForTaskListRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionForTaskListRequest instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionForTaskListRequest): flyteidl.admin.NodeExecutionForTaskListRequest;

            /**
             * Encodes the specified NodeExecutionForTaskListRequest message. Does not implicitly {@link flyteidl.admin.NodeExecutionForTaskListRequest.verify|verify} messages.
             * @param message NodeExecutionForTaskListRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionForTaskListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionForTaskListRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionForTaskListRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionForTaskListRequest;

            /**
             * Verifies a NodeExecutionForTaskListRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecution. */
        interface INodeExecution {

            /** NodeExecution id */
            id?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** NodeExecution inputUri */
            inputUri?: (string|null);

            /** NodeExecution closure */
            closure?: (flyteidl.admin.INodeExecutionClosure|null);

            /** NodeExecution metadata */
            metadata?: (flyteidl.admin.INodeExecutionMetaData|null);
        }

        /** Represents a NodeExecution. */
        class NodeExecution implements INodeExecution {

            /**
             * Constructs a new NodeExecution.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecution);

            /** NodeExecution id. */
            public id?: (flyteidl.core.INodeExecutionIdentifier|null);

            /** NodeExecution inputUri. */
            public inputUri: string;

            /** NodeExecution closure. */
            public closure?: (flyteidl.admin.INodeExecutionClosure|null);

            /** NodeExecution metadata. */
            public metadata?: (flyteidl.admin.INodeExecutionMetaData|null);

            /**
             * Creates a new NodeExecution instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecution instance
             */
            public static create(properties?: flyteidl.admin.INodeExecution): flyteidl.admin.NodeExecution;

            /**
             * Encodes the specified NodeExecution message. Does not implicitly {@link flyteidl.admin.NodeExecution.verify|verify} messages.
             * @param message NodeExecution message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecution, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecution message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecution
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecution;

            /**
             * Verifies a NodeExecution message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionMetaData. */
        interface INodeExecutionMetaData {

            /** NodeExecutionMetaData retryGroup */
            retryGroup?: (string|null);

            /** NodeExecutionMetaData isParentNode */
            isParentNode?: (boolean|null);

            /** NodeExecutionMetaData specNodeId */
            specNodeId?: (string|null);

            /** NodeExecutionMetaData isDynamic */
            isDynamic?: (boolean|null);

            /** NodeExecutionMetaData isArray */
            isArray?: (boolean|null);
        }

        /** Represents a NodeExecutionMetaData. */
        class NodeExecutionMetaData implements INodeExecutionMetaData {

            /**
             * Constructs a new NodeExecutionMetaData.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionMetaData);

            /** NodeExecutionMetaData retryGroup. */
            public retryGroup: string;

            /** NodeExecutionMetaData isParentNode. */
            public isParentNode: boolean;

            /** NodeExecutionMetaData specNodeId. */
            public specNodeId: string;

            /** NodeExecutionMetaData isDynamic. */
            public isDynamic: boolean;

            /** NodeExecutionMetaData isArray. */
            public isArray: boolean;

            /**
             * Creates a new NodeExecutionMetaData instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionMetaData instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionMetaData): flyteidl.admin.NodeExecutionMetaData;

            /**
             * Encodes the specified NodeExecutionMetaData message. Does not implicitly {@link flyteidl.admin.NodeExecutionMetaData.verify|verify} messages.
             * @param message NodeExecutionMetaData message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionMetaData, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionMetaData message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionMetaData
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionMetaData;

            /**
             * Verifies a NodeExecutionMetaData message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionList. */
        interface INodeExecutionList {

            /** NodeExecutionList nodeExecutions */
            nodeExecutions?: (flyteidl.admin.INodeExecution[]|null);

            /** NodeExecutionList token */
            token?: (string|null);
        }

        /** Represents a NodeExecutionList. */
        class NodeExecutionList implements INodeExecutionList {

            /**
             * Constructs a new NodeExecutionList.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionList);

            /** NodeExecutionList nodeExecutions. */
            public nodeExecutions: flyteidl.admin.INodeExecution[];

            /** NodeExecutionList token. */
            public token: string;

            /**
             * Creates a new NodeExecutionList instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionList instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionList): flyteidl.admin.NodeExecutionList;

            /**
             * Encodes the specified NodeExecutionList message. Does not implicitly {@link flyteidl.admin.NodeExecutionList.verify|verify} messages.
             * @param message NodeExecutionList message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionList, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionList message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionList
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionList;

            /**
             * Verifies a NodeExecutionList message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionClosure. */
        interface INodeExecutionClosure {

            /** NodeExecutionClosure outputUri */
            outputUri?: (string|null);

            /** NodeExecutionClosure error */
            error?: (flyteidl.core.IExecutionError|null);

            /** NodeExecutionClosure outputData */
            outputData?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionClosure phase */
            phase?: (flyteidl.core.NodeExecution.Phase|null);

            /** NodeExecutionClosure startedAt */
            startedAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionClosure duration */
            duration?: (google.protobuf.IDuration|null);

            /** NodeExecutionClosure createdAt */
            createdAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionClosure updatedAt */
            updatedAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionClosure workflowNodeMetadata */
            workflowNodeMetadata?: (flyteidl.admin.IWorkflowNodeMetadata|null);

            /** NodeExecutionClosure taskNodeMetadata */
            taskNodeMetadata?: (flyteidl.admin.ITaskNodeMetadata|null);

            /** NodeExecutionClosure deckUri */
            deckUri?: (string|null);

            /** NodeExecutionClosure dynamicJobSpecUri */
            dynamicJobSpecUri?: (string|null);
        }

        /** Represents a NodeExecutionClosure. */
        class NodeExecutionClosure implements INodeExecutionClosure {

            /**
             * Constructs a new NodeExecutionClosure.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionClosure);

            /** NodeExecutionClosure outputUri. */
            public outputUri: string;

            /** NodeExecutionClosure error. */
            public error?: (flyteidl.core.IExecutionError|null);

            /** NodeExecutionClosure outputData. */
            public outputData?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionClosure phase. */
            public phase: flyteidl.core.NodeExecution.Phase;

            /** NodeExecutionClosure startedAt. */
            public startedAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionClosure duration. */
            public duration?: (google.protobuf.IDuration|null);

            /** NodeExecutionClosure createdAt. */
            public createdAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionClosure updatedAt. */
            public updatedAt?: (google.protobuf.ITimestamp|null);

            /** NodeExecutionClosure workflowNodeMetadata. */
            public workflowNodeMetadata?: (flyteidl.admin.IWorkflowNodeMetadata|null);

            /** NodeExecutionClosure taskNodeMetadata. */
            public taskNodeMetadata?: (flyteidl.admin.ITaskNodeMetadata|null);

            /** NodeExecutionClosure deckUri. */
            public deckUri: string;

            /** NodeExecutionClosure dynamicJobSpecUri. */
            public dynamicJobSpecUri: string;

            /** NodeExecutionClosure outputResult. */
            public outputResult?: ("outputUri"|"error"|"outputData");

            /** NodeExecutionClosure targetMetadata. */
            public targetMetadata?: ("workflowNodeMetadata"|"taskNodeMetadata");

            /**
             * Creates a new NodeExecutionClosure instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionClosure instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionClosure): flyteidl.admin.NodeExecutionClosure;

            /**
             * Encodes the specified NodeExecutionClosure message. Does not implicitly {@link flyteidl.admin.NodeExecutionClosure.verify|verify} messages.
             * @param message NodeExecutionClosure message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionClosure, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionClosure message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionClosure
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionClosure;

            /**
             * Verifies a NodeExecutionClosure message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a WorkflowNodeMetadata. */
        interface IWorkflowNodeMetadata {

            /** WorkflowNodeMetadata executionId */
            executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);
        }

        /** Represents a WorkflowNodeMetadata. */
        class WorkflowNodeMetadata implements IWorkflowNodeMetadata {

            /**
             * Constructs a new WorkflowNodeMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IWorkflowNodeMetadata);

            /** WorkflowNodeMetadata executionId. */
            public executionId?: (flyteidl.core.IWorkflowExecutionIdentifier|null);

            /**
             * Creates a new WorkflowNodeMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns WorkflowNodeMetadata instance
             */
            public static create(properties?: flyteidl.admin.IWorkflowNodeMetadata): flyteidl.admin.WorkflowNodeMetadata;

            /**
             * Encodes the specified WorkflowNodeMetadata message. Does not implicitly {@link flyteidl.admin.WorkflowNodeMetadata.verify|verify} messages.
             * @param message WorkflowNodeMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IWorkflowNodeMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a WorkflowNodeMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns WorkflowNodeMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.WorkflowNodeMetadata;

            /**
             * Verifies a WorkflowNodeMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a TaskNodeMetadata. */
        interface ITaskNodeMetadata {

            /** TaskNodeMetadata cacheStatus */
            cacheStatus?: (flyteidl.core.CatalogCacheStatus|null);

            /** TaskNodeMetadata catalogKey */
            catalogKey?: (flyteidl.core.ICatalogMetadata|null);

            /** TaskNodeMetadata checkpointUri */
            checkpointUri?: (string|null);
        }

        /** Represents a TaskNodeMetadata. */
        class TaskNodeMetadata implements ITaskNodeMetadata {

            /**
             * Constructs a new TaskNodeMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.ITaskNodeMetadata);

            /** TaskNodeMetadata cacheStatus. */
            public cacheStatus: flyteidl.core.CatalogCacheStatus;

            /** TaskNodeMetadata catalogKey. */
            public catalogKey?: (flyteidl.core.ICatalogMetadata|null);

            /** TaskNodeMetadata checkpointUri. */
            public checkpointUri: string;

            /**
             * Creates a new TaskNodeMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns TaskNodeMetadata instance
             */
            public static create(properties?: flyteidl.admin.ITaskNodeMetadata): flyteidl.admin.TaskNodeMetadata;

            /**
             * Encodes the specified TaskNodeMetadata message. Does not implicitly {@link flyteidl.admin.TaskNodeMetadata.verify|verify} messages.
             * @param message TaskNodeMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.ITaskNodeMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a TaskNodeMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns TaskNodeMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.TaskNodeMetadata;

            /**
             * Verifies a TaskNodeMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DynamicWorkflowNodeMetadata. */
        interface IDynamicWorkflowNodeMetadata {

            /** DynamicWorkflowNodeMetadata id */
            id?: (flyteidl.core.IIdentifier|null);

            /** DynamicWorkflowNodeMetadata compiledWorkflow */
            compiledWorkflow?: (flyteidl.core.ICompiledWorkflowClosure|null);

            /** DynamicWorkflowNodeMetadata dynamicJobSpecUri */
            dynamicJobSpecUri?: (string|null);
        }

        /** Represents a DynamicWorkflowNodeMetadata. */
        class DynamicWorkflowNodeMetadata implements IDynamicWorkflowNodeMetadata {

            /**
             * Constructs a new DynamicWorkflowNodeMetadata.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDynamicWorkflowNodeMetadata);

            /** DynamicWorkflowNodeMetadata id. */
            public id?: (flyteidl.core.IIdentifier|null);

            /** DynamicWorkflowNodeMetadata compiledWorkflow. */
            public compiledWorkflow?: (flyteidl.core.ICompiledWorkflowClosure|null);

            /** DynamicWorkflowNodeMetadata dynamicJobSpecUri. */
            public dynamicJobSpecUri: string;

            /**
             * Creates a new DynamicWorkflowNodeMetadata instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DynamicWorkflowNodeMetadata instance
             */
            public static create(properties?: flyteidl.admin.IDynamicWorkflowNodeMetadata): flyteidl.admin.DynamicWorkflowNodeMetadata;

            /**
             * Encodes the specified DynamicWorkflowNodeMetadata message. Does not implicitly {@link flyteidl.admin.DynamicWorkflowNodeMetadata.verify|verify} messages.
             * @param message DynamicWorkflowNodeMetadata message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDynamicWorkflowNodeMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DynamicWorkflowNodeMetadata message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DynamicWorkflowNodeMetadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.DynamicWorkflowNodeMetadata;

            /**
             * Verifies a DynamicWorkflowNodeMetadata message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionGetDataRequest. */
        interface INodeExecutionGetDataRequest {

            /** NodeExecutionGetDataRequest id */
            id?: (flyteidl.core.INodeExecutionIdentifier|null);
        }

        /** Represents a NodeExecutionGetDataRequest. */
        class NodeExecutionGetDataRequest implements INodeExecutionGetDataRequest {

            /**
             * Constructs a new NodeExecutionGetDataRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionGetDataRequest);

            /** NodeExecutionGetDataRequest id. */
            public id?: (flyteidl.core.INodeExecutionIdentifier|null);

            /**
             * Creates a new NodeExecutionGetDataRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionGetDataRequest instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionGetDataRequest): flyteidl.admin.NodeExecutionGetDataRequest;

            /**
             * Encodes the specified NodeExecutionGetDataRequest message. Does not implicitly {@link flyteidl.admin.NodeExecutionGetDataRequest.verify|verify} messages.
             * @param message NodeExecutionGetDataRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionGetDataRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionGetDataRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionGetDataRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionGetDataRequest;

            /**
             * Verifies a NodeExecutionGetDataRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a NodeExecutionGetDataResponse. */
        interface INodeExecutionGetDataResponse {

            /** NodeExecutionGetDataResponse inputs */
            inputs?: (flyteidl.admin.IUrlBlob|null);

            /** NodeExecutionGetDataResponse outputs */
            outputs?: (flyteidl.admin.IUrlBlob|null);

            /** NodeExecutionGetDataResponse fullInputs */
            fullInputs?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionGetDataResponse fullOutputs */
            fullOutputs?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionGetDataResponse dynamicWorkflow */
            dynamicWorkflow?: (flyteidl.admin.IDynamicWorkflowNodeMetadata|null);

            /** NodeExecutionGetDataResponse flyteUrls */
            flyteUrls?: (flyteidl.admin.IFlyteURLs|null);
        }

        /** Represents a NodeExecutionGetDataResponse. */
        class NodeExecutionGetDataResponse implements INodeExecutionGetDataResponse {

            /**
             * Constructs a new NodeExecutionGetDataResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.INodeExecutionGetDataResponse);

            /** NodeExecutionGetDataResponse inputs. */
            public inputs?: (flyteidl.admin.IUrlBlob|null);

            /** NodeExecutionGetDataResponse outputs. */
            public outputs?: (flyteidl.admin.IUrlBlob|null);

            /** NodeExecutionGetDataResponse fullInputs. */
            public fullInputs?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionGetDataResponse fullOutputs. */
            public fullOutputs?: (flyteidl.core.ILiteralMap|null);

            /** NodeExecutionGetDataResponse dynamicWorkflow. */
            public dynamicWorkflow?: (flyteidl.admin.IDynamicWorkflowNodeMetadata|null);

            /** NodeExecutionGetDataResponse flyteUrls. */
            public flyteUrls?: (flyteidl.admin.IFlyteURLs|null);

            /**
             * Creates a new NodeExecutionGetDataResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns NodeExecutionGetDataResponse instance
             */
            public static create(properties?: flyteidl.admin.INodeExecutionGetDataResponse): flyteidl.admin.NodeExecutionGetDataResponse;

            /**
             * Encodes the specified NodeExecutionGetDataResponse message. Does not implicitly {@link flyteidl.admin.NodeExecutionGetDataResponse.verify|verify} messages.
             * @param message NodeExecutionGetDataResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.INodeExecutionGetDataResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a NodeExecutionGetDataResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns NodeExecutionGetDataResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.NodeExecutionGetDataResponse;

            /**
             * Verifies a NodeExecutionGetDataResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetDynamicNodeWorkflowRequest. */
        interface IGetDynamicNodeWorkflowRequest {

            /** GetDynamicNodeWorkflowRequest id */
            id?: (flyteidl.core.INodeExecutionIdentifier|null);
        }

        /** Represents a GetDynamicNodeWorkflowRequest. */
        class GetDynamicNodeWorkflowRequest implements IGetDynamicNodeWorkflowRequest {

            /**
             * Constructs a new GetDynamicNodeWorkflowRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetDynamicNodeWorkflowRequest);

            /** GetDynamicNodeWorkflowRequest id. */
            public id?: (flyteidl.core.INodeExecutionIdentifier|null);

            /**
             * Creates a new GetDynamicNodeWorkflowRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetDynamicNodeWorkflowRequest instance
             */
            public static create(properties?: flyteidl.admin.IGetDynamicNodeWorkflowRequest): flyteidl.admin.GetDynamicNodeWorkflowRequest;

            /**
             * Encodes the specified GetDynamicNodeWorkflowRequest message. Does not implicitly {@link flyteidl.admin.GetDynamicNodeWorkflowRequest.verify|verify} messages.
             * @param message GetDynamicNodeWorkflowRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetDynamicNodeWorkflowRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetDynamicNodeWorkflowRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetDynamicNodeWorkflowRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetDynamicNodeWorkflowRequest;

            /**
             * Verifies a GetDynamicNodeWorkflowRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a DynamicNodeWorkflowResponse. */
        interface IDynamicNodeWorkflowResponse {

            /** DynamicNodeWorkflowResponse compiledWorkflow */
            compiledWorkflow?: (flyteidl.core.ICompiledWorkflowClosure|null);
        }

        /** Represents a DynamicNodeWorkflowResponse. */
        class DynamicNodeWorkflowResponse implements IDynamicNodeWorkflowResponse {

            /**
             * Constructs a new DynamicNodeWorkflowResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDynamicNodeWorkflowResponse);

            /** DynamicNodeWorkflowResponse compiledWorkflow. */
            public compiledWorkflow?: (flyteidl.core.ICompiledWorkflowClosure|null);

            /**
             * Creates a new DynamicNodeWorkflowResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns DynamicNodeWorkflowResponse instance
             */
            public static create(properties?: flyteidl.admin.IDynamicNodeWorkflowResponse): flyteidl.admin.DynamicNodeWorkflowResponse;

            /**
             * Encodes the specified DynamicNodeWorkflowResponse message. Does not implicitly {@link flyteidl.admin.DynamicNodeWorkflowResponse.verify|verify} messages.
             * @param message DynamicNodeWorkflowResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDynamicNodeWorkflowResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DynamicNodeWorkflowResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DynamicNodeWorkflowResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.DynamicNodeWorkflowResponse;

            /**
             * Verifies a DynamicNodeWorkflowResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of an EmailMessage. */
        interface IEmailMessage {

            /** EmailMessage recipientsEmail */
            recipientsEmail?: (string[]|null);

            /** EmailMessage senderEmail */
            senderEmail?: (string|null);

            /** EmailMessage subjectLine */
            subjectLine?: (string|null);

            /** EmailMessage body */
            body?: (string|null);
        }

        /** Represents an EmailMessage. */
        class EmailMessage implements IEmailMessage {

            /**
             * Constructs a new EmailMessage.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IEmailMessage);

            /** EmailMessage recipientsEmail. */
            public recipientsEmail: string[];

            /** EmailMessage senderEmail. */
            public senderEmail: string;

            /** EmailMessage subjectLine. */
            public subjectLine: string;

            /** EmailMessage body. */
            public body: string;

            /**
             * Creates a new EmailMessage instance using the specified properties.
             * @param [properties] Properties to set
             * @returns EmailMessage instance
             */
            public static create(properties?: flyteidl.admin.IEmailMessage): flyteidl.admin.EmailMessage;

            /**
             * Encodes the specified EmailMessage message. Does not implicitly {@link flyteidl.admin.EmailMessage.verify|verify} messages.
             * @param message EmailMessage message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IEmailMessage, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EmailMessage message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EmailMessage
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.EmailMessage;

            /**
             * Verifies an EmailMessage message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetDomainRequest. */
        interface IGetDomainRequest {
        }

        /** Represents a GetDomainRequest. */
        class GetDomainRequest implements IGetDomainRequest {

            /**
             * Constructs a new GetDomainRequest.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetDomainRequest);

            /**
             * Creates a new GetDomainRequest instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetDomainRequest instance
             */
            public static create(properties?: flyteidl.admin.IGetDomainRequest): flyteidl.admin.GetDomainRequest;

            /**
             * Encodes the specified GetDomainRequest message. Does not implicitly {@link flyteidl.admin.GetDomainRequest.verify|verify} messages.
             * @param message GetDomainRequest message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetDomainRequest, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetDomainRequest message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetDomainRequest
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetDomainRequest;

            /**
             * Verifies a GetDomainRequest message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Domain. */
        interface IDomain {

            /** Domain id */
            id?: (string|null);

            /** Domain name */
            name?: (string|null);
        }

        /** Represents a Domain. */
        class Domain implements IDomain {

            /**
             * Constructs a new Domain.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IDomain);

            /** Domain id. */
            public id: string;

            /** Domain name. */
            public name: string;

            /**
             * Creates a new Domain instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Domain instance
             */
            public static create(properties?: flyteidl.admin.IDomain): flyteidl.admin.Domain;

            /**
             * Encodes the specified Domain message. Does not implicitly {@link flyteidl.admin.Domain.verify|verify} messages.
             * @param message Domain message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IDomain, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Domain message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Domain
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Domain;

            /**
             * Verifies a Domain message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a GetDomainsResponse. */
        interface IGetDomainsResponse {

            /** GetDomainsResponse domains */
            domains?: (flyteidl.admin.IDomain[]|null);
        }

        /** Represents a GetDomainsResponse. */
        class GetDomainsResponse implements IGetDomainsResponse {

            /**
             * Constructs a new GetDomainsResponse.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IGetDomainsResponse);

            /** GetDomainsResponse domains. */
            public domains: flyteidl.admin.IDomain[];

            /**
             * Creates a new GetDomainsResponse instance using the specified properties.
             * @param [properties] Properties to set
             * @returns GetDomainsResponse instance
             */
            public static create(properties?: flyteidl.admin.IGetDomainsResponse): flyteidl.admin.GetDomainsResponse;

            /**
             * Encodes the specified GetDomainsResponse message. Does not implicitly {@link flyteidl.admin.GetDomainsResponse.verify|verify} messages.
             * @param message GetDomainsResponse message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IGetDomainsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GetDomainsResponse message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GetDomainsResponse
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.GetDomainsResponse;

            /**
             * Verifies a GetDomainsResponse message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a Project. */
        interface IProject {

            /** Project id */
            id?: (string|null);

            /** Project name */
            name?: (string|null);

            /** Project domains */
            domains?: (flyteidl.admin.IDomain[]|null);

            /** Project description */
            description?: (string|null);

            /** Project labels */
            labels?: (flyteidl.admin.ILabels|null);

            /** Project state */
            state?: (flyteidl.admin.Project.ProjectState|null);

            /** Project org */
            org?: (string|null);
        }

        /** Represents a Project. */
        class Project implements IProject {

            /**
             * Constructs a new Project.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IProject);

            /** Project id. */
            public id: string;

            /** Project name. */
            public name: string;

            /** Project domains. */
            public domains: flyteidl.admin.IDomain[];

            /** Project description. */
            public description: string;

            /** Project labels. */
            public labels?: (flyteidl.admin.ILabels|null);

            /** Project state. */
            public state: flyteidl.admin.Project.ProjectState;

            /** Project org. */
            public org: string;

            /**
             * Creates a new Project instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Project instance
             */
            public static create(properties?: flyteidl.admin.IProject): flyteidl.admin.Project;

            /**
             * Encodes the specified Project message. Does not implicitly {@link flyteidl.admin.Project.verify|verify} messages.
             * @param message Project message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IProject, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Project message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Project
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Project;

            /**
             * Verifies a Project message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        namespace Project {

            /** ProjectState enum. */
            enum ProjectState {
                ACTIVE = 0,
                ARCHIVED = 1,
                SYSTEM_GENERATED = 2
            }
        }

        /** Properties of a Projects. */
        interface IProjects {

            /** Projects projects */
            projects?: (flyteidl.admin.IProject[]|null);

            /** Projects token */
            token?: (string|null);
        }

        /** Represents a Projects. */
        class Projects implements IProjects {

            /**
             * Constructs a new Projects.
             * @param [properties] Properties to set
             */
            constructor(properties?: flyteidl.admin.IProjects);

            /** Projects projects. */
            public projects: flyteidl.admin.IProject[];

            /** Projects token. */
            public token: string;

            /**
             * Creates a new Projects instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Projects instance
             */
            public static create(properties?: flyteidl.admin.IProjects): flyteidl.admin.Projects;

            /**
             * Encodes the specified Projects message. Does not implicitly {@link flyteidl.admin.Projects.verify|verify} messages.
             * @param message Projects message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: flyteidl.admin.IProjects, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Projects message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Projects
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): flyteidl.admin.Projects;

            /**
             * Verifies a Projects message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);
        }

        /** Properties of a ProjectListRequest. */
        interface IProjectListRequest {

            /** ProjectListRequest limit */
            limit?: (number|null);

            /** ProjectListRequest token */
            token?: (string|null);

            /** ProjectListRequest filters */
            filters?: (string|null);

            /** ProjectListRequest sortBy */
            sortBy?: (flyteidl.admin.ISort|null);

            /** ProjectListRequest org */
            org?: (string|null);
        }

        /** Represents a ProjectListRequest. */
        class ProjectListRequest implements IProjectListRequest {
