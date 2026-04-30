import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useAppTheme } from '../../hooks/useAppTheme';

interface Props {
  onRetry?: () => void;
  fullScreen?: boolean;
}

export function ServerDownState({ onRetry, fullScreen = false }: Props) {
  const { colors } = useAppTheme();
  return (
    <View
      style={[
        styles.container,
        fullScreen && styles.fullScreen,
        { backgroundColor: colors.background },
      ]}
    >
      <View style={[styles.iconWrap, { backgroundColor: colors.warningBg }]}>
        <Ionicons name="cloud-offline-outline" size={44} color={colors.warning} />
      </View>
      <Text style={[styles.title, { color: colors.text }]}>Server unavailable</Text>
      <Text style={[styles.message, { color: colors.textSecondary }]}>
        We can't reach our servers right now. Check your connection and try again.
      </Text>
      {onRetry && (
        <TouchableOpacity
          style={[styles.btn, { backgroundColor: colors.primary }]}
          onPress={onRetry}
          activeOpacity={0.85}
        >
          <Ionicons name="refresh-outline" size={16} color="#FFFFFF" />
          <Text style={styles.btnText}>Retry</Text>
        </TouchableOpacity>
      )}
      <Text style={[styles.hint, { color: colors.textMuted }]}>
        If this keeps happening, contact support.
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    justifyContent: 'center',
    padding: 40,
    gap: 12,
  },
  fullScreen: {
    flex: 1,
  },
  iconWrap: {
    width: 88,
    height: 88,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 4,
  },
  title: {
    fontSize: 18,
    fontWeight: '600',
    textAlign: 'center',
  },
  message: {
    fontSize: 14,
    textAlign: 'center',
    lineHeight: 22,
    maxWidth: 280,
  },
  btn: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    marginTop: 8,
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 10,
  },
  btnText: {
    fontSize: 15,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  hint: {
    fontSize: 12,
    textAlign: 'center',
    marginTop: 4,
  },
});
