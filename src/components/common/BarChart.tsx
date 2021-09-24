import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { smallFontSize } from 'components/Theme/constants';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';
import { Tooltip, Zoom } from '@material-ui/core';

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
    metadata?: any;
    tooltip?: string;
}

interface BarChartItemProps extends BarChartData {
    onClick?: () => void;
}

interface BarChartProps {
    data: BarChartData[];
    startDate?: string;
    onClickItem?: (item: any) => void;
}

/**
 * Display individual chart item for the BarChart component
 * @param value
 * @param color
 * @constructor
 */
const BarChartItem: React.FC<BarChartItemProps> = ({
    value,
    color,
    tooltip,
    onClick
}) => {
    const styles = useStyles();
    const [position, setPosition] = React.useState({ x: 0, y: 0 });

    const content = (
        <div
            className={styles.itemBar}
            style={{ backgroundColor: color, height: `${value}%` }}
            onClick={onClick}
        />
    );

    return (
        <div className={styles.item}>
            {tooltip ? (
                <Tooltip
                    title={tooltip}
                    TransitionComponent={Zoom}
                    onMouseMove={e => setPosition({ x: e.pageX, y: e.pageY })}
                    PopperProps={{
                        anchorEl: {
                            clientHeight: 0,
                            clientWidth: 0,
                            getBoundingClientRect: () => ({
                                top: position.y,
                                left: position.x,
                                right: position.x,
                                bottom: position.y,
                                width: 0,
                                height: 0
                            })
                        }
                    }}
                >
                    {content}
                </Tooltip>
            ) : (
                content
            )}
        </div>
    );
};

/**
 * Display information as bar chart with value and color
 * @param data
 * @param startDate
 * @constructor
 */
export const BarChart: React.FC<BarChartProps> = ({
    data,
    startDate,
    onClickItem
}) => {
    const styles = useStyles();

    const maxHeight = React.useMemo(() => {
        return Math.max(...data.map(x => Math.log2(x.value)));
    }, [data]);

    const handleClickItem = React.useCallback(
        item => () => {
            if (onClickItem) {
                onClickItem(item);
            }
        },
        [onClickItem]
    );

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
                        tooltip={item.tooltip}
                        onClick={handleClickItem(item)}
                        key={`bar-chart-item-${index}`}
                    />
                ))}
            </div>
        </div>
    );
};
