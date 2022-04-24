import * as React from 'react';
import { IconProps } from './interface';

export const InfoIcon: React.FC<IconProps> = ({ size = 14, className, onClick }) => {
  return (
    <svg
      className={className}
      width={size}
      height={size}
      viewBox="0 0 14 14"
      fill="none"
      onClick={onClick}
    >
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M7 1.00024C10.3139 1.00024 13 3.68695 13 7.00024C13 10.3135 10.3139 13.0002 7 13.0002C3.6867 13.0002 1 10.3135 1 7.00024C1 3.68695 3.6867 1.00024 7 1.00024Z"
        stroke="#2F49C6"
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M6.99856 4.04517V7.48482"
        stroke="#2F49C6"
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M6.99858 9.95486H7.00636"
        stroke="#2F49C6"
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
