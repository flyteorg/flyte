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
            public annotations?