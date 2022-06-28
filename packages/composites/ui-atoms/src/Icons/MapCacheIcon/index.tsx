import * as React from 'react';
import SvgIcon, { SvgIconProps } from '@material-ui/core/SvgIcon';

export const MapCacheIcon = React.forwardRef<SVGSVGElement, SvgIconProps>((props, ref) => {
  return (
    <SvgIcon {...props} ref={ref} viewBox="0 0 17 17">
      <g clipPath="url(#clip0_6712_89419)">
        <path
          d="M5.68615 2.99228L6.91566 7.15285L4.89438 6.0537C4.31136 7.12586 4.17812 8.3857 4.52399 9.55609C4.86986 10.7265 5.45103 11.4847 6.52318 12.0677C7.19695 12.4341 8.24054 12.5388 8.96464 12.5397L9.41341 14.0583C8.25879 14.1378 6.70623 14.0476 5.68964 13.4944C4.2601 12.717 3.51417 11.5513 3.05301 9.99079C2.59185 8.43027 2.76949 6.75048 3.54686 5.32094L1.52558 4.22179L5.68615 2.99228Z"
          fill="#666666"
        />
        <path
          d="M9.90614 2.76737C11.152 3.02647 12.2756 3.69438 13.0985 4.66504C13.9214 5.6357 14.3964 6.85346 14.4481 8.12494C14.4998 9.39642 14.1252 10.6487 13.3839 11.683C12.6425 12.7173 11.5768 13.4742 10.3561 13.8336L8.74374 8.35689L9.90614 2.76737Z"
          fill="#DDDDE5"
        />
      </g>
      <defs>
        <clipPath id="clip0_6712_89419">
          <rect width="16" height="16" fill="white" transform="translate(0.406982 0.356445)" />
        </clipPath>
      </defs>
    </SvgIcon>
  );
});
