import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import * as React from 'react';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import { InputProps, InputType, InputTypeDefinition, UnionValue, InputValue } from './types';
import { formatType } from './utils';
import { getComponentForInput } from './LaunchFormInputs';
import { getHelperForInput } from './inputHelpers/getHelperForInput';
import { SearchableSelector, SearchableSelectorOption } from './SearchableSelector';
import t from '../../common/strings';

const useStyles = makeStyles((theme: Theme) => ({
  inlineTitle: {
    display: 'flex',
    gap: theme.spacing(1),
    alignItems: 'center',
    paddingBottom: theme.spacing(3),
  },
}));

const generateInputTypeToValueMap = (
  listOfSubTypes: InputTypeDefinition[] | undefined,
  initialInputValue: UnionValue | undefined,
  initialType: InputTypeDefinition,
): Record<InputType, UnionValue> | {} => {
  if (!listOfSubTypes?.length) {
    return {};
  }

  return listOfSubTypes.reduce(function (map, subType) {
    if (initialInputValue && subType === initialType) {
      map[subType.type] = initialInputValue;
    } else {
      map[subType.type] = { value: '', typeDefinition: subType };
    }
    return map;
  }, {});
};

const generateSearchableSelectorOption = (
  inputTypeDefinition: InputTypeDefinition,
): SearchableSelectorOption<InputType> => {
  return {
    id: inputTypeDefinition.type,
    data: inputTypeDefinition.type,
    name: formatType(inputTypeDefinition),
  } as SearchableSelectorOption<InputType>;
};

const generateListOfSearchableSelectorOptions = (
  listOfInputTypeDefinition: InputTypeDefinition[],
): SearchableSelectorOption<InputType>[] => {
  return listOfInputTypeDefinition.map((inputTypeDefinition) =>
    generateSearchableSelectorOption(inputTypeDefinition),
  );
};

export const UnionInput = (props: InputProps) => {
  const {
    initialValue,
    required,
    label,
    onChange,
    typeDefinition,
    error,
    description,
    setIsError,
  } = props;

  const classes = useStyles();

  const listOfSubTypes = typeDefinition?.listOfSubTypes;

  if (!listOfSubTypes?.length) {
    return <></>;
  }

  const inputTypeToInputTypeDefinition = listOfSubTypes.reduce(
    (previous, current) => ({ ...previous, [current.type]: current }),
    {},
  );

  const initialInputValue =
    initialValue &&
    (getHelperForInput(typeDefinition.type).fromLiteral(
      initialValue,
      typeDefinition,
    ) as UnionValue);

  const initialInputTypeDefinition = initialInputValue?.typeDefinition ?? listOfSubTypes[0];

  if (!initialInputTypeDefinition) {
    return <></>;
  }

  const [inputTypeToValueMap, setInputTypeToValueMap] = React.useState<
    Record<InputType, UnionValue> | {}
  >(generateInputTypeToValueMap(listOfSubTypes, initialInputValue, initialInputTypeDefinition));

  const [selectedInputType, setSelectedInputType] = React.useState<InputType>(
    initialInputTypeDefinition.type,
  );

  const selectedInputTypeDefintion = inputTypeToInputTypeDefinition[
    selectedInputType
  ] as InputTypeDefinition;

  // change the selected union input value when change the selected union input type
  React.useEffect(() => {
    if (inputTypeToValueMap[selectedInputType]) {
      handleSubTypeOnChange(inputTypeToValueMap[selectedInputType].value);
    }
  }, [selectedInputTypeDefintion]);

  const handleTypeOnSelectionChanged = (value: SearchableSelectorOption<InputType>) => {
    setSelectedInputType(value.data);
  };

  const handleSubTypeOnChange = (input: InputValue) => {
    onChange({
      value: input,
      typeDefinition: selectedInputTypeDefintion,
    } as UnionValue);
    setInputTypeToValueMap({
      ...inputTypeToValueMap,
      [selectedInputType]: {
        value: input,
        typeDefinition: selectedInputTypeDefintion,
      } as UnionValue,
    });
  };

  return (
    <Card
      variant="outlined"
      style={{
        overflow: 'visible',
      }}
    >
      <CardContent>
        <div className={classes.inlineTitle}>
          <Typography variant="body1" component="label">
            {label}
          </Typography>

          <SearchableSelector
            label={t('type')}
            options={generateListOfSearchableSelectorOptions(listOfSubTypes)}
            selectedItem={generateSearchableSelectorOption(selectedInputTypeDefintion)}
            onSelectionChanged={handleTypeOnSelectionChanged}
          />
        </div>

        <div>
          {getComponentForInput(
            {
              description: description,
              name: `${formatType(selectedInputTypeDefintion)}`,
              label: '',
              required: required,
              typeDefinition: selectedInputTypeDefintion,
              onChange: handleSubTypeOnChange,
              value: inputTypeToValueMap[selectedInputType]?.value,
              error: error,
            } as InputProps,
            true,
            setIsError,
          )}
        </div>
      </CardContent>
    </Card>
  );
};
