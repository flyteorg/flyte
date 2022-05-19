import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ReactJsonView, { ReactJsonViewProps } from 'react-json-view';
import * as copyToClipboard from 'copy-to-clipboard';
import { primaryTextColor } from 'components/Theme/constants';

const useStyles = makeStyles((theme: Theme) => ({
  jsonViewer: {
    marginLeft: '-10px',
    width: '100%',
    '& span': {
      fontWeight: 'normal !important',
    },
    '& .object-container': {
      overflowWrap: 'anywhere !important',
    },
    '& .copy-to-clipboard-container': {
      position: 'absolute',
    },
    '& .copy-icon svg': {
      color: `${theme.palette.primary.main} !important`,
    },
    '& .variable-value': {
      paddingLeft: '5px',
    },
    '& .variable-value >*': {
      color: `${primaryTextColor} !important`,
    },
    '& .object-key-val': {
      padding: '0 0 0 5px !important',
    },
    '& .object-key': {
      color: `${theme.palette.grey[500]} !important`,
    },
    '& .node-ellipsis': {
      color: `${theme.palette.grey[500]} !important`,
    },
  },
}));

/**
 *
 * Replacer functionality to pass to the JSON.stringify function that
 * does proper serialization of arrays that contain non-numeric indexes
 * @param _ parent element key
 * @param value the element being serialized
 * @returns Transformed version of input
 */
const replacer = (_, value) => {
  // Check if associative array
  if (value instanceof Array && Object.keys(value).some((v) => isNaN(+v))) {
    // Serialize associative array
    return Object.keys(value).reduce((acc, arrKey) => {
      // if:
      //     string key is encountered insert {[key]: value} into transformed array
      // else:
      //     insert original value
      acc.push(isNaN(+arrKey) ? { [arrKey]: value[arrKey] } : value[arrKey]);

      return acc;
    }, [] as any[]);
  }

  // Non-associative array. return original value to allow default JSON.stringify behavior
  return value;
};

/**
 * Custom implementation for JSON.stringify to allow
 * proper serialization of arrays that contain non-numeric indexes
 *
 * @param json Object to serialize
 * @returns A string version of the input json
 */
const customStringify = (json) => {
  return JSON.stringify(json, replacer);
};

export const ReactJsonViewWrapper: React.FC<ReactJsonViewProps> = (props) => {
  const styles = useStyles();

  return (
    <div className={styles.jsonViewer}>
      <ReactJsonView
        enableClipboard={(options) => {
          const objToCopy = options.src;
          const text = typeof objToCopy === 'object' ? customStringify(objToCopy) : objToCopy;
          copyToClipboard(text);
        }}
        displayDataTypes={false}
        quotesOnKeys={false}
        iconStyle="triangle"
        displayObjectSize={true}
        name={null}
        indentWidth={4}
        collapseStringsAfterLength={80}
        sortKeys={true}
        {...props}
      />
    </div>
  );
};
