import { Button, IconButton, TextField, Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import * as React from 'react';
import RemoveIcon from '@material-ui/icons/Remove';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import { requiredInputSuffix } from './constants';
import { InputProps, InputType, InputTypeDefinition, InputValue } from './types';
import { formatType, toMappedTypeValue } from './utils';

const useStyles = makeStyles((theme: Theme) => ({
  formControl: {
    width: '100%',
    marginTop: theme.spacing(1),
  },
  controls: {
    margin: theme.spacing(1),
    width: '100%',
    display: 'flex',
    alignItems: 'flex-start',
    flexDirection: 'row',
  },
  keyControl: {
    marginRight: theme.spacing(1),
  },
  valueControl: {
    flexGrow: 1,
  },
  addButton: {
    display: 'flex',
    justifyContent: 'flex-end',
    marginTop: theme.spacing(1),
  },
  error: {
    border: '1px solid #f44336',
  },
}));

interface MapInputItemProps {
  data: MapInputItem;
  subtype?: InputTypeDefinition;
  setKey: (key: string) => void;
  setValue: (value: string) => void;
  isValid: (value: string) => boolean;
  onDeleteItem: () => void;
}

const MapSingleInputItem = (props: MapInputItemProps) => {
  const classes = useStyles();
  const { data, subtype, setKey, setValue, isValid, onDeleteItem } = props;
  const [error, setError] = React.useState(false);

  const isOneLineType = subtype?.type === InputType.String || subtype?.type === InputType.Integer;

  return (
    <div className={classes.controls}>
      <TextField
        label={`string${requiredInputSuffix}`}
        onChange={({ target: { value } }: React.ChangeEvent<HTMLInputElement>) => {
          setKey(value);
          setError(!!value && !isValid(value));
        }}
        value={data.key}
        error={error}
        placeholder="key"
        variant="outlined"
        helperText={error ? 'This key already defined' : ''}
        className={classes.keyControl}
      />
      <TextField
        label={subtype ? `${formatType(subtype)}${requiredInputSuffix}` : ''}
        onChange={({ target: { value } }: React.ChangeEvent<HTMLInputElement>) => {
          setValue(value);
        }}
        value={data.value}
        variant="outlined"
        className={classes.valueControl}
        multiline={!isOneLineType}
        type={subtype?.type === InputType.Integer ? 'number' : 'text'}
      />
      <IconButton onClick={onDeleteItem}>
        <RemoveIcon />
      </IconButton>
    </div>
  );
};

type MapInputItem = {
  id: number | null;
  key: string;
  value: string;
};

const getNewMapItem = (id, key = '', value = ''): MapInputItem => {
  return { id, key, value };
};

function parseMappedTypeValue(value?: InputValue): MapInputItem[] {
  const fallback = [getNewMapItem(0)];
  if (!value) {
    return fallback;
  }
  try {
    const mapObj = JSON.parse(value.toString());
    if (typeof mapObj === 'object') {
      return Object.keys(mapObj).map((key, index) => getNewMapItem(index, key, mapObj[key]));
    }
  } catch (e) {
    // do nothing
  }

  return fallback;
}

export const MapInput = (props: InputProps) => {
  const {
    value,
    label,
    onChange,
    typeDefinition: { subtype },
  } = props;
  const classes = useStyles();

  const [data, setData] = React.useState<MapInputItem[]>(parseMappedTypeValue(value));

  const onAddItem = () => {
    setData((data) => [...data, getNewMapItem(data.length)]);
  };

  const updateUpperStream = (newData: MapInputItem[]) => {
    const newPairs = newData
      .filter((item) => {
        // we filter out delted values and items with errors or empty keys/values
        return item.id !== null && !!item.key && !!item.value;
      })
      .map((item) => {
        return {
          key: item.key,
          value: item.value,
        };
      });
    const newValue = toMappedTypeValue(newPairs);
    onChange(newValue);
  };

  const onSetKey = (id: number | null, key: string) => {
    if (id === null) return;
    const newData = [...data];
    newData[id].key = key;
    setData([...newData]);
    updateUpperStream([...newData]);
  };

  const onSetValue = (id: number | null, value: string) => {
    if (id === null) return;
    const newData = [...data];
    newData[id].value = value;
    setData([...newData]);
    updateUpperStream([...newData]);
  };

  const onDeleteItem = (id: number | null) => {
    if (id === null) return;
    const newData = [...data];
    const dataIndex = newData.findIndex((item) => item.id === id);
    if (dataIndex >= 0 && dataIndex < newData.length) {
      newData[dataIndex].id = null;
    }
    setData([...newData]);
    updateUpperStream([...newData]);
  };

  const isValid = (id: number | null, value: string) => {
    if (id === null) return true;
    // findIndex returns -1 if value is not found, which means we can use that key
    return (
      data
        .filter((item) => item.id !== null && item.id !== id)
        .findIndex((item) => item.key === value) === -1
    );
  };

  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="body1" component="label">
          {label}
        </Typography>
        {data
          .filter((item) => item.id !== null)
          .map((item) => {
            return (
              <MapSingleInputItem
                key={item.id}
                data={item}
                subtype={subtype}
                setKey={(key) => onSetKey(item.id, key)}
                setValue={(value) => onSetValue(item.id, value)}
                isValid={(value) => isValid(item.id, value)}
                onDeleteItem={() => onDeleteItem(item.id)}
              />
            );
          })}
        <div className={classes.addButton}>
          <Button onClick={onAddItem}>+ ADD ITEM</Button>
        </div>
      </CardContent>
    </Card>
  );
};
