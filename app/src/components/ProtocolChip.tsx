import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { Colors, Radius, Typography } from '../theme/colors';

interface Props {
  label: string;
  variant?: 'cyan' | 'emerald' | 'amber';
}

// Protocol identifier micro-chip (VLESS / VMess / WireGuard / L3).
export default function ProtocolChip({ label, variant = 'cyan' }: Props) {
  const color =
    variant === 'emerald' ? Colors.emerald : variant === 'amber' ? Colors.amber : Colors.cyan;
  const bg =
    variant === 'emerald'
      ? 'rgba(16, 185, 129, 0.1)'
      : variant === 'amber'
      ? 'rgba(245, 158, 11, 0.1)'
      : Colors.cyanTintBg;

  return (
    <View style={[styles.chip, { backgroundColor: bg, borderColor: Colors.cyanBorder }]}>
      <Text style={[styles.text, { color }]}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  chip: {
    borderRadius: Radius.sm,
    borderWidth: 1,
    paddingHorizontal: 8,
    paddingVertical: 3,
    alignSelf: 'flex-start',
  },
  text: {
    ...Typography.labelSm,
    textTransform: 'uppercase',
    fontVariant: ['tabular-nums'],
  },
});