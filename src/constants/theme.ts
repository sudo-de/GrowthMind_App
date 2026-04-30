import { AccentColor, ColorMode } from '../types';

export interface ThemeColors {
  // Backgrounds
  background: string;
  surface: string;
  card: string;
  overlay: string;
  // Text
  text: string;
  textSecondary: string;
  textMuted: string;
  textInverse: string;
  // Borders & inputs
  border: string;
  borderFocus: string;
  inputBackground: string;
  // Feedback
  success: string;
  successBg: string;
  warning: string;
  warningBg: string;
  error: string;
  errorBg: string;
  info: string;
  infoBg: string;
  // Nav
  tabBar: string;
  headerBackground: string;
}

export const lightColors: ThemeColors = {
  background: '#FFFFFF',
  surface: '#F9FAFB',
  card: '#FFFFFF',
  overlay: 'rgba(0,0,0,0.45)',
  text: '#111827',
  textSecondary: '#6B7280',
  textMuted: '#9CA3AF',
  textInverse: '#FFFFFF',
  border: '#E5E7EB',
  borderFocus: '#4F46E5',
  inputBackground: '#F9FAFB',
  success: '#10B981',
  successBg: '#ECFDF5',
  warning: '#F59E0B',
  warningBg: '#FFFBEB',
  error: '#EF4444',
  errorBg: '#FEF2F2',
  info: '#3B82F6',
  infoBg: '#EFF6FF',
  tabBar: '#FFFFFF',
  headerBackground: '#FFFFFF',
};

// True black for OLED screens
export const darkColors: ThemeColors = {
  background: '#000000',
  surface: '#111111',
  card: '#1C1C1C',
  overlay: 'rgba(0,0,0,0.65)',
  text: '#F9FAFB',
  textSecondary: '#A1A1AA',
  textMuted: '#71717A',
  textInverse: '#111111',
  border: '#27272A',
  borderFocus: '#6366F1',
  inputBackground: '#111111',
  success: '#34D399',
  successBg: '#064E3B',
  warning: '#FCD34D',
  warningBg: '#451A03',
  error: '#F87171',
  errorBg: '#450A0A',
  info: '#60A5FA',
  infoBg: '#1E3A5F',
  tabBar: '#111111',
  headerBackground: '#000000',
};

export const accentPalettes: Record<
  AccentColor,
  { primary: string; primaryDark: string; primaryLight: string; primaryBg: string }
> = {
  indigo: {
    primary: '#4F46E5',
    primaryDark: '#4338CA',
    primaryLight: '#818CF8',
    primaryBg: '#EEF2FF',
  },
  red: {
    primary: '#EF4444',
    primaryDark: '#DC2626',
    primaryLight: '#FCA5A5',
    primaryBg: '#FEF2F2',
  },
  blue: {
    primary: '#3B82F6',
    primaryDark: '#2563EB',
    primaryLight: '#93C5FD',
    primaryBg: '#EFF6FF',
  },
  green: {
    primary: '#10B981',
    primaryDark: '#059669',
    primaryLight: '#6EE7B7',
    primaryBg: '#ECFDF5',
  },
  purple: {
    primary: '#8B5CF6',
    primaryDark: '#7C3AED',
    primaryLight: '#C4B5FD',
    primaryBg: '#F5F3FF',
  },
  orange: {
    primary: '#F97316',
    primaryDark: '#EA580C',
    primaryLight: '#FDba74',
    primaryBg: '#FFF7ED',
  },
};

export type { ColorMode, AccentColor };
