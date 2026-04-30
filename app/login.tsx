import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { Link } from 'expo-router';
import { Ionicons, AntDesign } from '@expo/vector-icons';
import * as WebBrowser from 'expo-web-browser';
import * as Google from 'expo-auth-session/providers/google';
import { makeRedirectUri } from 'expo-auth-session';
import * as AppleAuthentication from 'expo-apple-authentication';
import { useAppTheme } from '../src/hooks/useAppTheme';
import { useAppDispatch, useAppSelector } from '../src/store/hooks';
import { authRequest, authSuccess, authFailure, clearAuthError } from '../src/store/slices/authSlice';
import { GOOGLE_AUTH } from '../src/constants/googleAuth';

WebBrowser.maybeCompleteAuthSession();

export default function LoginScreen() {
  const { colors, isDark } = useAppTheme();
  const dispatch = useAppDispatch();
  const { status, error } = useAppSelector((s) => s.auth);
  const loading = status === 'loading';

  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [focused, setFocused] = useState<string | null>(null);

  const redirectUri = makeRedirectUri({
    scheme: Platform.select({
      ios: 'com.googleusercontent.apps.526828440529-i9apirka4salnalevojkfnot7697ii3a',
      android: 'com.googleusercontent.apps.526828440529-9b5t450dhd48oh13dhjpl4i92lsdrk5g',
    }),
    path: Platform.OS === 'android' ? 'oauth2redirect/google' : undefined,
  });

  const [request, response, promptAsync] = Google.useAuthRequest({
    ...GOOGLE_AUTH,
    redirectUri,
  });

  useEffect(() => {
    if (__DEV__ && request?.redirectUri) {
      console.log('[Google OAuth] redirectUri in use:', request.redirectUri);
    }
  }, [request]);

  useEffect(() => {
    if (response?.type === 'success' && response.authentication?.accessToken) {
      signInWithGoogle(response.authentication.accessToken);
    }
  }, [response]);

  const signInWithGoogle = async (accessToken: string) => {
    dispatch(authRequest());
    try {
      const res = await fetch('https://www.googleapis.com/userinfo/v2/me', {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      const profile = await res.json();
      dispatch(
        authSuccess({
          user: {
            id: profile.id,
            fullName: profile.name ?? 'Google User',
            username: (profile.email as string).split('@')[0],
            email: profile.email ?? '',
            avatarUrl: profile.picture,
          },
          token: accessToken,
        }),
      );
    } catch {
      dispatch(authFailure('Google Sign In failed. Please try again.'));
    }
  };

  useEffect(() => {
    if (error) {
      Alert.alert('Sign in failed', error, [
        { text: 'OK', onPress: () => dispatch(clearAuthError()) },
      ]);
    }
  }, [error]);

  const handleLogin = async () => {
    if (!identifier.trim() || !password) {
      Alert.alert('Missing fields', 'Please enter your email/username and password.');
      return;
    }
    dispatch(authRequest());
    try {
      // TODO: replace with real API call
      await new Promise((r) => setTimeout(r, 1000));
      dispatch(
        authSuccess({
          user: { id: '1', fullName: 'Demo User', username: identifier, email: identifier },
          token: 'demo_token',
        }),
      );
    } catch {
      dispatch(authFailure('Invalid email or password.'));
    }
  };

  const handleAppleSignIn = async () => {
    try {
      const credential = await AppleAuthentication.signInAsync({
        requestedScopes: [
          AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
          AppleAuthentication.AppleAuthenticationScope.EMAIL,
        ],
      });
      // TODO: send credential to your backend
      dispatch(
        authSuccess({
          user: {
            id: credential.user,
            fullName:
              `${credential.fullName?.givenName ?? ''} ${credential.fullName?.familyName ?? ''}`.trim() ||
              'Apple User',
            username: 'apple_user',
            email: credential.email ?? '',
          },
          token: credential.identityToken ?? '',
        }),
      );
    } catch (e: unknown) {
      const err = e as { code?: string };
      if (err.code !== 'ERR_REQUEST_CANCELED') {
        Alert.alert('Error', 'Apple Sign In failed. Please try again.');
      }
    }
  };

  const iw = (field: string) => [
    styles.inputWrapper,
    {
      borderColor: focused === field ? colors.borderFocus : colors.border,
      backgroundColor: colors.inputBackground,
    },
  ];
  const ic = (field: string) => (focused === field ? colors.primary : colors.textMuted);

  return (
    <KeyboardAvoidingView
      style={[styles.container, { backgroundColor: colors.background }]}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView
        contentContainerStyle={styles.scroll}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        {/* Brand */}
        <View style={styles.brand}>
          <View style={[styles.logoBox, { shadowColor: colors.primary }]}>
            <Text style={styles.logoText}>GM</Text>
          </View>
          <Text style={[styles.brandName, { color: colors.text }]}>GrownMind</Text>
        </View>

        <Text style={[styles.heading, { color: colors.text }]}>Welcome back</Text>
        <Text style={[styles.subheading, { color: colors.textSecondary }]}>
          Sign in to continue your growth journey
        </Text>

        {/* Form */}
        <View style={styles.form}>
          <View style={iw('id')}>
            <Ionicons name="person-outline" size={18} color={ic('id')} style={styles.icon} />
            <TextInput
              style={[styles.input, { color: colors.text }]}
              placeholder="Email or username"
              placeholderTextColor={colors.textMuted}
              value={identifier}
              onChangeText={setIdentifier}
              onFocus={() => setFocused('id')}
              onBlur={() => setFocused(null)}
              autoCapitalize="none"
              autoCorrect={false}
              keyboardType="email-address"
            />
          </View>

          <View style={iw('pw')}>
            <Ionicons name="lock-closed-outline" size={18} color={ic('pw')} style={styles.icon} />
            <TextInput
              style={[styles.input, { color: colors.text }]}
              placeholder="Password"
              placeholderTextColor={colors.textMuted}
              value={password}
              onChangeText={setPassword}
              onFocus={() => setFocused('pw')}
              onBlur={() => setFocused(null)}
              secureTextEntry={!showPassword}
            />
            <TouchableOpacity onPress={() => setShowPassword((p) => !p)} style={styles.eye}>
              <Ionicons
                name={showPassword ? 'eye-outline' : 'eye-off-outline'}
                size={18}
                color={colors.textMuted}
              />
            </TouchableOpacity>
          </View>

          <TouchableOpacity style={styles.forgotRow}>
            <Text style={[styles.forgotText, { color: colors.primary }]}>Forgot Password?</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.primaryBtn, { backgroundColor: colors.primary, shadowColor: colors.primary }, loading && styles.btnDisabled]}
            onPress={handleLogin}
            disabled={loading}
            activeOpacity={0.85}
          >
            {loading ? (
              <ActivityIndicator color="#fff" size="small" />
            ) : (
              <Text style={styles.primaryBtnText}>Sign In</Text>
            )}
          </TouchableOpacity>
        </View>

        {/* Divider */}
        <View style={styles.divider}>
          <View style={[styles.dividerLine, { backgroundColor: colors.border }]} />
          <Text style={[styles.dividerLabel, { color: colors.textMuted }]}>or continue with</Text>
          <View style={[styles.dividerLine, { backgroundColor: colors.border }]} />
        </View>

        {/* Social */}
        <View style={styles.social}>
          <TouchableOpacity
            style={[styles.googleBtn, { borderColor: colors.border, backgroundColor: colors.card }]}
            onPress={() => promptAsync()}
            disabled={!request}
            activeOpacity={0.8}
          >
            <AntDesign name="google" size={20} color="#DB4437" />
            <Text style={[styles.googleText, { color: colors.text }]}>Continue with Google</Text>
          </TouchableOpacity>

          {Platform.OS === 'ios' && (
            <AppleAuthentication.AppleAuthenticationButton
              buttonType={AppleAuthentication.AppleAuthenticationButtonType.SIGN_IN}
              buttonStyle={
                isDark
                  ? AppleAuthentication.AppleAuthenticationButtonStyle.WHITE
                  : AppleAuthentication.AppleAuthenticationButtonStyle.BLACK
              }
              cornerRadius={12}
              style={styles.appleBtn}
              onPress={handleAppleSignIn}
            />
          )}
        </View>

        {/* Footer */}
        <View style={styles.footer}>
          <Text style={[styles.footerText, { color: colors.textSecondary }]}>
            New to GrownMind?
          </Text>
          <Link href="/register" asChild>
            <TouchableOpacity>
              <Text style={[styles.footerLink, { color: colors.primary }]}> Create account</Text>
            </TouchableOpacity>
          </Link>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  scroll: {
    flexGrow: 1,
    paddingHorizontal: 24,
    paddingTop: 64,
    paddingBottom: 40,
  },
  brand: { alignItems: 'center', marginBottom: 36 },
  logoBox: {
    width: 72,
    height: 72,
    borderRadius: 20,
    backgroundColor: '#4F46E5',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 10,
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.35,
    shadowRadius: 16,
    elevation: 8,
  },
  logoText: { fontSize: 28, fontWeight: '800', color: '#FFFFFF', letterSpacing: 1 },
  brandName: { fontSize: 20, fontWeight: '700', letterSpacing: 0.5 },
  heading: { fontSize: 28, fontWeight: '700', marginBottom: 6 },
  subheading: { fontSize: 15, marginBottom: 28, lineHeight: 22 },
  form: { gap: 12 },
  inputWrapper: {
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1.5,
    borderRadius: 12,
    paddingHorizontal: 14,
    height: 54,
  },
  icon: { marginRight: 10 },
  input: { flex: 1, fontSize: 15 },
  eye: { padding: 4 },
  forgotRow: { alignSelf: 'flex-end', marginTop: 2 },
  forgotText: { fontSize: 14, fontWeight: '500' },
  primaryBtn: {
    borderRadius: 12,
    height: 54,
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 4,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 5,
  },
  btnDisabled: { opacity: 0.7 },
  primaryBtnText: { fontSize: 16, fontWeight: '600', color: '#FFFFFF', letterSpacing: 0.4 },
  divider: { flexDirection: 'row', alignItems: 'center', marginVertical: 24, gap: 12 },
  dividerLine: { flex: 1, height: 1 },
  dividerLabel: { fontSize: 13 },
  social: { gap: 12 },
  googleBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1.5,
    borderRadius: 12,
    height: 54,
    gap: 10,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.06,
    shadowRadius: 4,
    elevation: 2,
  },
  googleText: { fontSize: 15, fontWeight: '600' },
  appleBtn: { height: 54, width: '100%' },
  footer: { flexDirection: 'row', justifyContent: 'center', alignItems: 'center', marginTop: 32 },
  footerText: { fontSize: 14 },
  footerLink: { fontSize: 14, fontWeight: '600' },
});
