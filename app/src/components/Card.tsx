import React from 'react';
import { View, StyleSheet, StyleProp, ViewStyle } from 'react-native';
import { Colors, Radius } from '../theme/colors';

interface Props {
  children: React.ReactNode;
  style?: StyleProp<ViewStyle>;
  level?: 1 | 2 | 3;
  active?: boolean;
}

// Card implements the tonal surface stepping from DESIGN.md.
// level 2 is the default node/card container.
export default function Card({ children, style, level = 2, active = false }: Props) {
  const bg =
    level === 1 ? Colors.level1 : level === 3 ? Colors.level3 : Colors.level2;

  return (
    <View
      style={[
        styles.card,
        { backgroundColor: bg },
        active && styles.active,
        style,
      ]}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: Radius.md,
    borderWidth: 1,
    borderColor: Colors.borderCard,
    padding: 16,
  },
  active: {
    borderColor: Colors.primaryContainer,
    shadowColor: Colors.activeGlow,
    shadowOpacity: 1,
    shadowRadius: 4,
    shadowOffset: { width: 0, height: 0 },
    elevation: 4,
  },
});