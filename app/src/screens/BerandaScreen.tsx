import React, { useEffect, useState } from 'react';
import { View, Text, StyleSheet, Pressable, ActivityIndicator } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import ScreenContainer from '../components/ScreenContainer';
import Card from '../components/Card';
import ProtocolChip from '../components/ProtocolChip';
import PingPill from '../components/PingPill';
import { Colors, Spacing, Radius, Typography } from '../theme/colors';

export default function BerandaScreen() {
  const [connected, setConnected] = useState(true);
  const [busy, setBusy] = useState(false);
  const [seconds, setSeconds] = useState(3724);
  const [rx, setRx] = useState(12.4);
  const [tx, setTx] = useState(2.1);

  useEffect(() => {
    if (!connected) return;
    const id = setInterval(() => {
      setSeconds((s) => s + 1);
      setRx(12.0 + Math.random() * 1.2);
      setTx(1.9 + Math.random() * 0.5);
    }, 1000);
    return () => clearInterval(id);
  }, [connected]);

  const toggle = () => {
    setBusy(true);
    setTimeout(() => {
      setConnected((c) => !c);
      setBusy(false);
      if (connected) setSeconds(0);
    }, 900);
  };

  const fmtTime = (s: number) => {
    const h = String(Math.floor(s / 3600)).padStart(2, '0');
    const m = String(Math.floor((s % 3600) / 60)).padStart(2, '0');
    const sec = String(s % 60).padStart(2, '0');
    return `${h}:${m}:${sec}`;
  };

  return (
    <ScreenContainer>
      <Card active={connected} style={styles.hero}>
        <View style={styles.heroTop}>
          <View style={styles.statusRow}>
            <View style={[styles.statusDot, { backgroundColor: connected ? Colors.emerald : Colors.muted }]} />
            <Text style={styles.statusLabel}>{connected ? 'TERHUBUNG' : 'TERPUTUS'}</Text>
          </View>
          <ProtocolChip label="L3 gRPC" variant="emerald" />
        </View>
        <View style={styles.serverBlock}>
          <Text style={styles.serverName}>Aether-SG-01</Text>
          <Text style={styles.serverEndpoint}>sg.aether.example:443 · TUN 10.8.0.2/24</Text>
        </View>
        <View style={styles.metaRow}>
          <View style={styles.metaItem}>
            <Text style={styles.metaLabel}>UPTIME</Text>
            <Text style={styles.metaValue}>{connected ? fmtTime(seconds) : '--:--:--'}</Text>
          </View>
          <View style={styles.metaItem}>
            <Text style={styles.metaLabel}>PING</Text>
            {connected ? <PingPill ms={42} /> : <PingPill />}
          </View>
        </View>
      </Card>

      <Card style={styles.trafficCard}>
        <Text style={styles.sectionTitle}>Telemetri Trafik</Text>
        <View style={styles.trafficRow}>
          <View style={styles.trafficItem}>
            <View style={styles.trafficHeader}>
              <Ionicons name="arrow-down" size={14} color={Colors.cyan} />
              <Text style={[styles.trafficLabel, { color: Colors.cyan }]}>RX / UNDUH</Text>
            </View>
            <Text style={styles.trafficValue}>
              {connected ? rx.toFixed(1) : '0.0'}
              <Text style={styles.trafficUnit}> MB/s</Text>
            </Text>
          </View>
          <View style={styles.trafficDivider} />
          <View style={styles.trafficItem}>
            <View style={styles.trafficHeader}>
              <Ionicons name="arrow-up" size={14} color={Colors.emerald} />
              <Text style={[styles.trafficLabel, { color: Colors.emerald }]}>TX / UNGGAH</Text>
            </View>
            <Text style={styles.trafficValue}>
              {connected ? tx.toFixed(1) : '0.0'}
              <Text style={styles.trafficUnit}> MB/s</Text>
            </Text>
          </View>
        </View>
      </Card>

      <Pressable
        onPress={toggle}
        disabled={busy}
        style={({ pressed }) => [
          styles.mainBtn,
          connected ? styles.mainBtnConnected : styles.mainBtnIdle,
          pressed && { opacity: 0.9 },
        ]}
      >
        {busy ? (
          <>
            <ActivityIndicator color={connected ? Colors.onPrimary : Colors.primary} size="small" />
            <Text style={[styles.mainBtnText, connected && styles.mainBtnTextConnected]}>Memutuskan...</Text>
          </>
        ) : (
          <>
            <Ionicons name={connected ? 'power' : 'flash'} size={20} color={connected ? Colors.onPrimary : Colors.primary} />
            <Text style={[styles.mainBtnText, connected && styles.mainBtnTextConnected]}>
              {connected ? 'Putuskan Sambungan' : 'Sambungkan Kembali'}
            </Text>
          </>
        )}
      </Pressable>

      <View style={styles.quickRow}>
        <Card style={styles.quickCard}>
          <Ionicons name="swap-vertical" size={18} color={Colors.secondary} />
          <Text style={styles.quickValue}>1.2 GB</Text>
          <Text style={styles.quickLabel}>Total Data</Text>
        </Card>
        <Card style={styles.quickCard}>
          <Ionicons name="pulse" size={18} color={Colors.tertiary} />
          <Text style={styles.quickValue}>42 ms</Text>
          <Text style={styles.quickLabel}>Latensi ICMP</Text>
        </Card>
        <Card style={styles.quickCard}>
          <Ionicons name="layers" size={18} color={Colors.primary} />
          <Text style={styles.quickValue}>14</Text>
          <Text style={styles.quickLabel}>App Tunnel</Text>
        </Card>
      </View>
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  hero: { marginBottom: Spacing.md },
  heroTop: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: Spacing.md },
  statusRow: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  statusDot: { width: 10, height: 10, borderRadius: Radius.full },
  statusLabel: { ...Typography.labelMd, color: Colors.onSurfaceVariant },
  serverBlock: { marginBottom: Spacing.md },
  serverName: { ...Typography.headlineMd, color: Colors.onSurface },
  serverEndpoint: { ...Typography.bodySm, color: Colors.onSurfaceVariant, marginTop: 2 },
  metaRow: { flexDirection: 'row', gap: Spacing.xl },
  metaItem: { gap: 4 },
  metaLabel: { ...Typography.labelSm, color: Colors.muted },
  metaValue: { ...Typography.telemetry, color: Colors.onSurface },
  trafficCard: { marginBottom: Spacing.md },
  sectionTitle: { ...Typography.titleMd, color: Colors.onSurface, marginBottom: Spacing.md },
  trafficRow: { flexDirection: 'row', alignItems: 'center' },
  trafficItem: { flex: 1, gap: 6 },
  trafficHeader: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  trafficLabel: { ...Typography.labelSm },
  trafficValue: { ...Typography.headlineMd, color: Colors.onSurface },
  trafficUnit: { ...Typography.bodySm, color: Colors.muted },
  trafficDivider: { width: 1, height: 48, backgroundColor: Colors.borderCard },
  mainBtn: { flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 10, paddingVertical: 15, borderRadius: Radius.lg, marginBottom: Spacing.md },
  mainBtnConnected: { backgroundColor: Colors.primary },
  mainBtnIdle: { backgroundColor: Colors.level1, borderWidth: 1.5, borderColor: Colors.primaryContainer },
  mainBtnText: { ...Typography.titleMd, fontWeight: '600', color: Colors.primary },
  mainBtnTextConnected: { color: Colors.onPrimary },
  quickRow: { flexDirection: 'row', gap: Spacing.sm },
  quickCard: { flex: 1, alignItems: 'center', gap: 6, paddingVertical: 14 },
  quickValue: { ...Typography.titleMd, color: Colors.onSurface },
  quickLabel: { ...Typography.labelSm, color: Colors.muted },
});