import { mapValues } from 'lodash';
import { SimpleType, TypedInterface, Variable } from 'models/Common';

function simpleType(primitiveType: SimpleType, description?: string): Variable {
    return {
        description,
        type: {
            simple: primitiveType
        }
    };
}

export const mockSimpleVariables: Record<string, Variable> = {
    simpleString: simpleType(SimpleType.STRING, 'a simple string value'),
    stringNoLabel: simpleType(SimpleType.STRING),
    simpleInteger: simpleType(SimpleType.INTEGER, 'a simple integer value'),
    simpleFloat: simpleType(SimpleType.FLOAT, 'a simple floating point value'),
    simpleBoolean: simpleType(SimpleType.BOOLEAN, 'a simple boolean value'),
    simpleDuration: simpleType(SimpleType.DURATION, 'a simple duration value'),
    simpleDatetime: simpleType(SimpleType.DATETIME, 'a simple datetime value'),
    simpleBinary: simpleType(SimpleType.BINARY, 'a simple binary value'),
    simpleError: simpleType(SimpleType.ERROR, 'a simple error value'),
    simpleStruct: simpleType(SimpleType.STRUCT, 'a simple struct value')
    // schema: {},
    // collection: {},
    // mapValue: {},
    // blob: {}
};

export const mockCollectionVariables: Record<string, Variable> = mapValues(
    mockSimpleVariables,
    v => ({
        description: `A collection of: ${v.description}`,
        type: { collectionType: v.type }
    })
);

export const mockNestedCollectionVariables: Record<
    string,
    Variable
> = mapValues(mockCollectionVariables, v => ({
    description: `${v.description} (nested)`,
    type: { collectionType: v.type }
}));

export function createMockWorkflowInputsInterface(
    variables: Record<string, Variable>
): TypedInterface {
    return {
        inputs: {
            variables: { ...variables }
        }
    };
}
