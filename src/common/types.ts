export interface DecodeProtobufTypedDictionaryResult {
    error?: Error;
    // tslint:disable-next-line:no-any
    value?: { [k: string]: any };
}

export interface ResourceDefinition {
    domain: string;
    name: string;
    project: string;
    version: string;
    // TODO: This should probably be an enum
    type: string;
}
