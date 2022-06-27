import { makeStyles } from '@material-ui/core/styles';
import { SvgIconProps, Theme } from '@material-ui/core';
import classnames from 'classnames';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
  svg: {
    marginTop: 0,
    marginRight: theme.spacing(2),
    display: 'inline-block',
    fontSize: '1.5rem',
    transition: 'fill 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms',
    flexShrink: 0,
    userSelect: 'none',
    color: '#666666',
  },
}));

export const MuiLaunchPlanIcon = (props: SvgIconProps): JSX.Element => {
  const { fill, className, width = '1em', height = '1em', fontSize } = props;
  const styles = useStyles();
  return (
    <svg
      className={classnames(styles.svg, className)}
      width={width}
      height={height}
      viewBox="0 0 16 16"
      xmlns="http://www.w3.org/2000/svg"
      fontSize={fontSize}
    >
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M2 15V2C2 1.44772 2.44772 1 3 1H12.7391C13.2914 1 13.7391 1.44772 13.7391 2V11.4421H9.82593H9.32593V11.9421V16H3C2.44772 16 2 15.5523 2 15ZM10.3259 12.4421H13.384L10.3259 15.5002V12.4421ZM5.1307 5.93466H11.0003V4.93466H5.1307V5.93466ZM11.0004 8.83351H5.13079V7.83351H11.0004V8.83351ZM5.13079 11.732H8.02934V10.732H5.13079V11.732Z"
        fill={fill || 'currentColor'}
      />
    </svg>
  );
};
