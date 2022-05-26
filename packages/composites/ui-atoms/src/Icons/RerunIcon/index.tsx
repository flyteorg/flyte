import * as React from 'react';

interface IconProps {
  size?: number;
  className?: string;
  onClick?: () => void;
}

export const RerunIcon = (props: IconProps): JSX.Element => {
  const { size = 18, className, onClick } = props;
  return (
    <svg
      className={className}
      width={size}
      height={size}
      viewBox="0 0 13 11"
      fill="none"
      onClick={onClick}
    >
      <path
        d="M4.8 1.4H8.8V0L12.6 2.2L8.8 4.4V3H4.8C3.04 3 1.6 4.44 1.6 6.2C1.6 7.96 3.04 9.4 4.8 9.4H8.8V11H4.8C2.16 11 0 8.84 0 6.2C0 3.56 2.16 1.4 4.8 1.4Z"
        fill="#666666"
      />
    </svg>
  );
};
