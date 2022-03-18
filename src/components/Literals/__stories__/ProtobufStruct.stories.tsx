import { storiesOf } from '@storybook/react';
import { ProtobufListValue, ProtobufStruct } from 'models/Common/types';
import * as React from 'react';
import { LiteralValue } from '../LiteralValue';
import { CardDecorator } from './CardDecorator';
import { protobufValues } from './protobufValues';

const stories = storiesOf('Literals/ProtobufStruct', module);
stories.addDecorator(CardDecorator);

function renderStruct(label: string, struct: ProtobufStruct) {
  return (
    <LiteralValue
      label={label}
      literal={{
        scalar: { value: 'generic', generic: struct },
        value: 'scalar',
        hash: '',
      }}
    />
  );
}

stories.add('basic', () => renderStruct('basic_struct', { fields: protobufValues }));

stories.add('list', () =>
  renderStruct('struct_with_list', {
    fields: {
      list_value: {
        kind: 'listValue',
        listValue: {
          values: [
            ...Object.values(protobufValues),
            {
              kind: 'structValue',
              structValue: { fields: protobufValues },
            },
          ],
        } as ProtobufListValue,
      },
    },
  }),
);

stories.add('nested', () =>
  renderStruct('struct_with_nested', {
    fields: {
      struct_value: {
        kind: 'structValue',
        structValue: { fields: protobufValues },
      },
    },
  }),
);
