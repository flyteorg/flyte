import { mapValues } from 'lodash';
import {
    ParameterMap,
    SimpleType,
    TypedInterface,
    Variable
} from 'models/Common';

function simpleType(primitiveType: SimpleType, description?: string): Variable {
    return {
        description,
        type: {
            simple: primitiveType
        }
    };
}

export const mockVariables: Record<string, Variable> = {
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

export const mockWorkflowInputsInterface: TypedInterface = {
    inputs: {
        variables: { ...mockVariables }
    }
};

export const mockParameterMap: ParameterMap = {
    parameters: mapValues(mockVariables, v => ({ var: v }))
};
