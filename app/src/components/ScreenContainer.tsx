import React from 'react';
import { View, StyleSheet, ScrollView, ViewStyle } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Colors, Spacing } from '../theme/colors';

interface Props {
  children: React.ReactNode;
  scroll?: boolean;
  style?: ViewStyle;
}

export default function ScreenContainer({ children, scroll = true, style }: Props) {
  return (
    <SafeAreaView style={styles.safe} edges={['top']}>
      {scroll ? (
        <ScrollView
          style={styles.scroll}
          contentContainerStyle={[styles.content, style]}
          showsVerticalScrollIndicator={false}
        >
          {children}
        </ScrollView>
      ) : (
        <View style={[styles.content, style]}>{children}</View>
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  scroll: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  content: {
    padding: Spacing.marginMobile,
    paddingBottom: Spacing.xxl,
  },
});