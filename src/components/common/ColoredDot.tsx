import * as React from 'react';

export const ColoredDot: React.FC<{
    color: string;
    size?: number;
}> = ({ color, size = 10 }) => (
    <svg
        viewBox={`0 0 ${size} ${size}`}
        xmlns="http://www.w3.org/2000/svg"
        width={size}
        height={size}
    >
        <circle
            cx={size / 2}
            cy={size / 2}
            r={size / 2}
            style={{ fill: color }}
        />
    </svg>
);
