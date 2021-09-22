import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { smallFontSize } from 'components/Theme/constants';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        display: 'flex',
        flexDirection: 'column',
        marginBottom: theme.spacing(2.5)
    },
    header: {
        display: 'flex',
        justifyContent: 'space-between',
        marginBottom: theme.spacing(0.75),
        fontSize: smallFontSize,
        color: COLOR_SPECTRUM.gray40.color
    },
    body: {
        display: 'flex',
        alignItems: 'stretch',
        borderLeft: '1.04174px dashed #C1C1C1',
        borderRight: '1.04174px dashed #C1C1C1',
        minHeight: theme.spacing(10.5)
    },
    item: {
        flex: 1,
        display: 'flex',
        alignItems: 'flex-end',
        '&:last-child': {
            marginRight: 0
        }
    },
    itemBar: {
        borderRadius: 2,
        flex: 1,
        marginRight: theme.spacing(0.25),
        minHeight: theme.spacing(0.75),
        cursor: 'pointer'
    }
}));

interface BarChartData {
    value: number;
    color: string;
}

interface BarChartProps {
    data: BarChartData[];
    startDate?: string;
}

/**
 * Display individual chart item for the BarChart component
 * @param value
 * @param color
 * @constructor
 */
const BarChartItem: React.FC<BarChartData> = ({ value, color }) => {
    const styles = useStyles();

    return (
        <div className={styles.item}>
            <div
                className={styles.itemBar}
                style={{ backgroundColor: color, height: `${value}%` }}
            ></div>
        </div>
    );
};

/**
 * Display information as bar chart with value and color
 * @param data
 * @param startDate
 * @constructor
 */
export const BarChart: React.FC<BarChartProps> = ({ data, startDate }) => {
    const styles = useStyles();

    const maxHeight = React.useMemo(() => {
        return Math.max(...data.map(x => Math.log2(x.value)));
    }, [data]);

    return (
        <div className={styles.container}>
            <div className={styles.header}>
                <span>{startDate}</span>
                <span>Most Recent</span>
            </div>
            <div className={styles.body}>
                {data.map((item, index) => (
                    <BarChartItem
                        value={(Math.log2(item.value) / maxHeight) * 100}
                        color={item.color}
                        key={`bar-chart-item-${index}`}
                    />
                ))}
            </div>
        </div>
    );
};
