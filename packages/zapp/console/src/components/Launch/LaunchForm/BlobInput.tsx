import {
  FormControl,
  FormHelperText,
  InputLabel,
  MenuItem,
  Select,
  TextField,
  Typography,
} from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { BlobDimensionality } from 'models/Common/types';
import * as React from 'react';
import { blobFormatHelperText, blobUriHelperText, defaultBlobValue } from './constants';
import { InputProps } from './types';
import { getLaunchInputId, isBlobValue } from './utils';

const useStyles = makeStyles((theme: Theme) => ({
  dimensionalityInput: {
    flex: '1 1 auto',
    marginLeft: theme.spacing(2),
  },
  formatInput: {
    flex: '1 1 auto',
  },
  inputContainer: {
    borderLeft: `1px solid ${theme.palette.divider}`,
    marginTop: theme.spacing(1),
    paddingLeft: theme.spacing(1),
  },
  metadataContainer: {
    display: 'flex',
    marginTop: theme.spacing(1),
    width: '100%',
  },
}));

/** A micro form for entering the values related to a Blob Literal */
export const BlobInput: React.FC<InputProps> = (props) => {
  const styles = useStyles();
  const { error, label, name, onChange, value: propValue, typeDefinition } = props;
  const dimensionality = typeDefinition?.literalType?.blob?.dimensionality;
  const blobValue = isBlobValue(propValue)
    ? propValue
    : {
        uri: '',
        dimensionality: dimensionality ?? BlobDimensionality.SINGLE,
      };
  const hasError = !!error;
  const helperText = hasError ? error : props.helperText;

  const handleUriChange = ({ target: { value: uri } }: React.ChangeEvent<HTMLInputElement>) => {
    onChange({
      ...blobValue,
      uri,
    });
  };

  const handleFormatChange = ({
    target: { value: format },
  }: React.ChangeEvent<HTMLInputElement>) => {
    onChange({
      ...blobValue,
      format,
    });
  };

  const handleDimensionalityChange = ({
    target: { value },
  }: React.ChangeEvent<{ value: unknown }>) => {
    onChange({
      ...blobValue,
      dimensionality: value as BlobDimensionality,
    });
  };

  const selectId = getLaunchInputId(`${name}-select`);

  return (
    <div>
      <Typography variant="body1" component="label">
        {label}
      </Typography>
      <FormHelperText error={hasError}>{helperText}</FormHelperText>
      <div className={styles.inputContainer}>
        <TextField
          id={getLaunchInputId(`${name}-uri`)}
          helperText={blobUriHelperText}
          fullWidth={true}
          label="uri"
          onChange={handleUriChange}
          value={blobValue.uri}
          variant="outlined"
        />
        <div className={styles.metadataContainer}>
          <TextField
            className={styles.formatInput}
            id={getLaunchInputId(`${name}-format`)}
            helperText={blobFormatHelperText}
            label="format"
            onChange={handleFormatChange}
            value={blobValue.format}
            variant="outlined"
          />
          <FormControl className={styles.dimensionalityInput}>
            <InputLabel id={`${selectId}-label`}>Dimensionality</InputLabel>
            <Select
              labelId={`${selectId}-label`}
              id={selectId}
              value={blobValue.dimensionality}
              onChange={handleDimensionalityChange}
              disabled
            >
              <MenuItem value={BlobDimensionality.SINGLE}>Single (File)</MenuItem>
              <MenuItem value={BlobDimensionality.MULTIPART}>Multipart (Directory)</MenuItem>
            </Select>
          </FormControl>
        </div>
      </div>
    </div>
  );
};
