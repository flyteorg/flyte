import Typography from '@material-ui/core/Typography';
import * as React from 'react';

export interface SectionSubheaderProps {
    value: string;
}
export const SectionSubheader: React.FC<SectionSubheaderProps> = ({
    value
}) => {
    return (
        <div>
            <Typography variant="body2">{value}</Typography>
        </div>
    );
};
