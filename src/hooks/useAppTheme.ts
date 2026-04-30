import { useColorScheme } from 'react-native';
import { useAppSelector } from '../store/hooks';
import { darkColors, lightColors, accentPalettes, ThemeColors } from '../constants/theme';
import { AccentColor, ColorMode } from '../types';

export interface AppTheme {
  colors: ThemeColors & {
    primary: string;
    primaryDark: string;
    primaryLight: string;
    primaryBg: string;
  };
  isDark: boolean;
  accent: AccentColor;
  mode: ColorMode;
}

export function useAppTheme(): AppTheme {
  const { mode, accent } = useAppSelector((state) => state.theme);
  const systemScheme = useColorScheme();

  const isDark =
    mode === 'system' ? systemScheme === 'dark' : mode === 'dark';

  const base = isDark ? darkColors : lightColors;
  const palette = accentPalettes[accent];

  return {
    colors: {
      ...base,
      primary: palette.primary,
      primaryDark: palette.primaryDark,
      primaryLight: palette.primaryLight,
      primaryBg: palette.primaryBg,
      borderFocus: palette.primary,
    },
    isDark,
    accent,
    mode,
  };
}
