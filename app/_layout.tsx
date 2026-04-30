import React from 'react';
import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { Provider } from 'react-redux';
import { store } from '../src/store';
import { useAppTheme } from '../src/hooks/useAppTheme';

// Inner component so it runs inside the Provider and can read theme
function ThemedStack() {
  const { isDark } = useAppTheme();
  return (
    <>
      <Stack screenOptions={{ headerShown: false }} />
      <StatusBar style={isDark ? 'light' : 'dark'} />
    </>
  );
}

export default function RootLayout() {
  return (
    <Provider store={store}>
      <ThemedStack />
    </Provider>
  );
}
