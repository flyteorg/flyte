import { Spacing } from '@material-ui/core/styles/createSpacing';
import { nameColumnLeftMarginGridWidth } from './styles';

export function calculateNodeExecutionRowLeftSpacing(level: number, spacing: Spacing) {
  return spacing(nameColumnLeftMarginGridWidth + 3 * level);
}
