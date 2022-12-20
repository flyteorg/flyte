import { storiesOf } from '@storybook/react';
import * as React from 'react';
import { MapInput } from '../MapInput';
import { InputValue } from '../types';
import { inputTypes } from '../inputHelpers/test/testCases';

const stories = storiesOf('Launch/MapInput', module);

stories.addDecorator((story) => {
  return <div style={{ width: 600, height: '95vh' }}>{story()}</div>;
});

stories.add('Base', () => {
  return (
    <MapInput
      description="Something"
      name="Variable A"
      label="Variable_A"
      required={false}
      typeDefinition={inputTypes.map}
      onChange={(newValue: InputValue) => {
        console.log('onChange', newValue);
      }}
    />
  );
});
