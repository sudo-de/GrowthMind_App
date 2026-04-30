import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { AccentColor, ColorMode } from '../../types';

interface ThemeState {
  mode: ColorMode;
  accent: AccentColor;
}

const initialState: ThemeState = {
  mode: 'system',
  accent: 'indigo',
};

const themeSlice = createSlice({
  name: 'theme',
  initialState,
  reducers: {
    setMode(state, action: PayloadAction<ColorMode>) {
      state.mode = action.payload;
    },
    setAccent(state, action: PayloadAction<AccentColor>) {
      state.accent = action.payload;
    },
  },
});

export const { setMode, setAccent } = themeSlice.actions;
export default themeSlice.reducer;
