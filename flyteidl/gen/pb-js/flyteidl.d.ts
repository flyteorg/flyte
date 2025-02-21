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
            MINUTE = 1