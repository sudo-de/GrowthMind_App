import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { AppStatus } from '../../types';

interface AppState {
  status: AppStatus;
  error: string | null;
  message: string | null;
}

const initialState: AppState = {
  status: 'idle',
  error: null,
  message: null,
};

const appSlice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    setLoading(state) {
      state.status = 'loading';
      state.error = null;
      state.message = null;
    },
    setSuccess(state, action: PayloadAction<string | undefined>) {
      state.status = 'success';
      state.message = action.payload ?? null;
      state.error = null;
    },
    setError(state, action: PayloadAction<string>) {
      state.status = 'error';
      state.error = action.payload;
      state.message = null;
    },
    setServerDown(state) {
      state.status = 'server_down';
      state.error = 'The server is currently unavailable. Please try again later.';
      state.message = null;
    },
    setEmpty(state) {
      state.status = 'empty';
      state.error = null;
      state.message = null;
    },
    resetApp(state) {
      state.status = 'idle';
      state.error = null;
      state.message = null;
    },
  },
});

export const { setLoading, setSuccess, setError, setServerDown, setEmpty, resetApp } =
  appSlice.actions;
export default appSlice.reducer;
