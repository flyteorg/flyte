import { ProtobufValue } from 'models/Common/types';

export const protobufValues: Dictionary<ProtobufValue> = {
  booleanFalse: {
    boolValue: false,
    kind: 'boolValue',
  },
  booleanTrue: {
    boolValue: true,
    kind: 'boolValue',
  },
  simpleNumber: {
    numberValue: 12345,
    kind: 'numberValue',
  },
  floatNumber: {
    numberValue: 12345.6789,
    kind: 'numberValue',
  },
  simpleString: {
    stringValue: 'I am a string value',
    kind: 'stringValue',
  },
  null: {
    kind: 'nullValue',
  },
};
