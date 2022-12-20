import { Theme as MaterialTheme, useTheme as useMaterialTheme } from '@material-ui/core/styles';
import { Variant } from '@material-ui/core/styles/createTypography';
import { measureText } from 'components/common/utils';

export interface Theme extends MaterialTheme {
  measureTextWidth(variant: Variant, text: string): number;
}

/** Provides an enhanced version of the Material Theme */
export function useTheme(): Theme {
  const theme = useMaterialTheme<MaterialTheme>();
  const measureTextWidth = (variant: Variant, text: string) => {
    const { fontFamily, fontSize, fontWeight } = theme.typography[variant];
    const fontDefinition = `${fontWeight} ${fontSize} ${fontFamily}`;
    return measureText(fontDefinition, text).width;
  };

  return { ...theme, measureTextWidth };
}
