import React, { useState } from 'react';
import { View, Text, StyleSheet, Pressable, ScrollView } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import ScreenContainer from '../components/ScreenContainer';
import { Colors, Spacing, Radius, Typography } from '../theme/colors';

type Level = 'INFO' | 'WARN' | 'ERROR' | 'DEBUG';

interface LogEntry {
  id: number;
  time: string;
  level: Level;
  msg: string;
}

const LEVEL_COLORS: Record<Level, string> = {
  INFO: Colors.emerald,
  WARN: Colors.amber,
  ERROR: Colors.coral,
  DEBUG: Colors.cyan,
};

const INITIAL_LOGS: LogEntry[] = [
  { id: 1, time: '09:41:02', level: 'INFO', msg: 'Tunnel handshake OK (TLS 1.3, PSK auth)' },
  { id: 2, time: '09:41:02', level: 'INFO', msg: 'TUN aether0 up, addr 10.8.0.2/24 mtu 1500' },
  { id: 3, time: '09:41:03', level: 'DEBUG', msg: 'gRPC bidi stream established (HTTP/2)' },
  { id: 4, time: '09:41:05', level: 'INFO', msg: 'ICMP echo 8.8.8.8 -> reply 42ms' },
  { id: 5, time: '09:41:08', level: 'DEBUG', msg: 'UDP session bind 5353 -> forwarded' },
  { id: 6, time: '09:41:12', level: 'WARN', msg: 'Keep-alive jitter detected, resync in 1.2s' },
  { id: 7, time: '09:41:15', level: 'INFO', msg: 'Split-tunnel rule applied: 3 packages' },
  { id: 8, time: '09:41:19', level: 'DEBUG', msg: 'TCP SYN 142.250.190.46:443 encapsulated' },
];

export default function LogsScreen() {
  const [logs, setLogs] = useState<LogEntry[]>(INITIAL_LOGS);
  const [filter, setFilter] = useState<Level | 'ALL'>('ALL');

  const visible = filter === 'ALL' ? logs : logs.filter((l) => l.level === filter);

  const clear = () => setLogs([]);

  const FILTERS: Array<Level | 'ALL'> = ['ALL', 'INFO', 'WARN', 'ERROR', 'DEBUG'];

  return (
    <ScreenContainer scroll={false} style={styles.container}>
      <View style={styles.headerRow}>
        <Text style={styles.title}>Logs & Debugging</Text>
        <Pressable onPress={clear} style={styles.clearBtn}>
          <Ionicons name="trash-outline" size={16} color={Colors.coral} />
          <Text style={styles.clearText}>Bersihkan</Text>
        </Pressable>
      </View>

      {/* Filter chips */}
      <View style={styles.filterRow}>
        {FILTERS.map((f) => {
          const active = filter === f;
          return (
            <Pressable
              key={f}
              onPress={() => setFilter(f)}
              style={[
                styles.chip,
                active && { borderColor: Colors.primary, backgroundColor: Colors.activeTintBg },
              ]}
            >
              <Text style={[styles.chipText, active && { color: Colors.primary }]}>{f}</Text>
            </Pressable>
          );
        })}
      </View>

      {/* Terminal output */}
      <View style={styles.terminal}>
        <ScrollView showsVerticalScrollIndicator={false}>
          {visible.length === 0 ? (
            <Text style={styles.empty}>— tidak ada log —</Text>
          ) : (
            visible.map((entry) => (
              <View key={entry.id} style={styles.logLine}>
                <Text style={styles.logTime}>{entry.time}</Text>
                <Text style={[styles.logLevel, { color: LEVEL_COLORS[entry.level] }]}>
                  {entry.level.padEnd(5)}
                </Text>
                <Text style={styles.logMsg}>{entry.msg}</Text>
              </View>
            ))
          )}
        </ScrollView>
      </View>
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, paddingBottom: 0 },
  headerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: Spacing.md,
  },
  title: { ...Typography.headlineSm, color: Colors.onSurface },
  clearBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: Colors.level3,
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: Radius.sm,
  },
  clearText: { ...Typography.telemetry, color: Colors.coral },
  filterRow: { flexDirection: 'row', gap: Spacing.sm, marginBottom: Spacing.md },
  chip: {
    paddingHorizontal: 10,
    paddingVertical: 5,
    borderRadius: Radius.sm,
    borderWidth: 1,
    borderColor: Colors.borderCard,
    backgroundColor: Colors.level2,
  },
  chipText: { ...Typography.labelSm, color: Colors.onSurfaceVariant },
  terminal: {
    flex: 1,
    backgroundColor: Colors.surfaceContainerLowest,
    borderRadius: Radius.md,
    borderWidth: 1,
    borderColor: Colors.borderHairline,
    padding: 12,
  },
  logLine: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 4,
    alignItems: 'flex-start',
  },
  logTime: { ...Typography.bodySm, color: Colors.muted, fontVariant: ['tabular-nums'] },
  logLevel: { ...Typography.bodySm, fontWeight: '600', fontVariant: ['tabular-nums'] },
  logMsg: { ...Typography.bodySm, color: Colors.onSurface, flex: 1 },
  empty: { ...Typography.bodySm, color: Colors.muted, textAlign: 'center', marginTop: 40 },
});