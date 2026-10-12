import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { Colors, Radius, Typography } from '../theme/colors';

interface Props {
  ms?: number; // undefined => timeout/dead
}

// Latency pill with color semantics from DESIGN.md.
export default function PingPill({ ms }: Props) {
  let color: string;
  let label: string;
  let strike = false;

  if (ms === undefined) {
    color = Colors.coral;
    label = 'TIMEOUT';
    strike = true;
  } else if (ms < 100) {
    color = Colors.emerald;
    label = `${ms}ms`;
  } else if (ms <= 250) {
    color = Colors.cyan;
    label = `${ms}ms`;
  } else {
    color = Colors.amber;
    label = `${ms}ms`;
  }

  return (
    <View style={[styles.pill, { borderColor: color }]}>
      <Text
        style={[
          styles.text,
          { color },
          strike && styles.strike,
        ]}
      >
        {label}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  pill: {
    borderRadius: Radius.sm,
    borderWidth: 1,
    paddingHorizontal: 8,
    paddingVertical: 2,
    alignSelf: 'flex-start',
  },
  text: {
    ...Typography.labelSm,
    fontVariant: ['tabular-nums'],
  },
  strike: {
    textDecorationLine: 'line-through',
  },
});