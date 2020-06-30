import { Spacing } from '@material-ui/core/styles/createSpacing';
import { measureText } from 'components/common/utils';
import { TaskLog } from 'models';
import { nameColumnLeftMarginGridWidth } from './styles';

interface MeasuredTaskLog extends TaskLog {
    width: number;
}

/** For a list of pre-measured TaskLog items, will split the list into
 * two smaller lists. The first list contains the items which can be rendered
 * into the given width, the second list contains the remaining items.
 */
export function splitLogLinksAtWidth(
    logs: MeasuredTaskLog[],
    width: number
): [TaskLog[], TaskLog[]] {
    const { taken, left } = logs.reduce(
        (out, log) => {
            const { width } = log;
            // Accounting for icon decoration and spacing
            const nameWidth = width + 16;
            if (nameWidth > out.pixelsRemaining) {
                out.pixelsRemaining = 0;
                out.left.push(log);
            } else {
                out.pixelsRemaining = out.pixelsRemaining - nameWidth;
                out.taken.push(log);
            }
            return out;
        },
        {
            pixelsRemaining: width,
            taken: <TaskLog[]>[],
            left: <TaskLog[]>[]
        }
    );
    return [taken, left];
}

export function calculateNodeExecutionRowLeftSpacing(
    level: number,
    spacing: Spacing
) {
    return spacing(nameColumnLeftMarginGridWidth + 3 * level);
}
