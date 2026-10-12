import React, { useState } from 'react';
import { View, Text, StyleSheet, Pressable, FlatList, Alert } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import ScreenContainer from '../components/ScreenContainer';
import Card from '../components/Card';
import ProtocolChip from '../components/ProtocolChip';
import PingPill from '../components/PingPill';
import { Colors, Spacing, Radius, Typography } from '../theme/colors';

interface Profile {
  id: string;
  name: string;
  endpoint: string;
  protocol: string;
  group: string;
  ping?: number;
}

const PROFILES: Profile[] = [
  { id: '1', name: 'Aether-SG-01', endpoint: 'sg.aether.example:443', protocol: 'L3 gRPC', group: 'Auto-Fallback', ping: 42 },
  { id: '2', name: 'Aether-JP-02', endpoint: 'jp.aether.example:443', protocol: 'L3 QUIC', group: 'Auto-Fallback', ping: 88 },
  { id: '3', name: 'Aether-DE-01', endpoint: 'de.aether.example:443', protocol: 'L3 gRPC', group: 'Manual', ping: 214 },
  { id: '4', name: 'Aether-US-03', endpoint: 'us.aether.example:443', protocol: 'L3 QUIC', group: 'Manual' },
];

export default function KonfigurasiScreen() {
  const [selected, setSelected] = useState('1');
  const [testing, setTesting] = useState(false);

  const testAll = () => {
    setTesting(true);
    setTimeout(() => setTesting(false), 1200);
  };

  const renderProfile = ({ item }: { item: Profile }) => {
    const isActive = item.id === selected;
    return (
      <Pressable onPress={() => setSelected(item.id)}>
        <Card active={isActive} style={[styles.profileCard, isActive && styles.activeCard]}>
          {isActive && <View style={styles.activeBar} />}
          <View style={styles.profileTop}>
            <ProtocolChip
              label={item.protocol}
              variant={item.protocol.includes('QUIC') ? 'cyan' : 'emerald'}
            />
            <View style={styles.profileTopRight}>
              <PingPill ms={item.ping} />
              <Ionicons
                name={isActive ? 'radio-button-on' : 'radio-button-off'}
                size={20}
                color={isActive ? Colors.primary : Colors.muted}
              />
            </View>
          </View>
          <Text style={styles.profileName}>{item.name}</Text>
          <View style={styles.profileMeta}>
            <Text style={styles.profileEndpoint}>{item.endpoint}</Text>
            <View style={styles.groupBadge}>
              <Text style={styles.groupText}>{item.group}</Text>
            </View>
          </View>
        </Card>
      </Pressable>
    );
  };

  return (
    <ScreenContainer scroll={false} style={styles.container}>
      <View style={styles.headerRow}>
        <Text style={styles.title}>Daftar Konfigurasi</Text>
        <Pressable
          onPress={testAll}
          style={({ pressed }) => [styles.testBtn, pressed && { opacity: 0.85 }]}
        >
          {testing ? (
            <Ionicons name="refresh" size={16} color={Colors.primary} />
          ) : (
            <Ionicons name="speedometer" size={16} color={Colors.primary} />
          )}
          <Text style={styles.testBtnText}>{testing ? 'Testing...' : 'Test Semua'}</Text>
        </Pressable>
      </View>

      <FlatList
        data={PROFILES}
        keyExtractor={(i) => i.id}
        renderItem={renderProfile}
        contentContainerStyle={styles.list}
        showsVerticalScrollIndicator={false}
      />

      {/* FAB */}
      <Pressable
        style={({ pressed }) => [styles.fab, pressed && { transform: [{ scale: 0.9 }] }]}
        onPress={() => Alert.alert('Tambah Profil', 'Form tambah konfigurasi server L3.')}
      >
        <Ionicons name="add" size={28} color={Colors.onPrimary} />
      </Pressable>
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  container: { paddingBottom: 0, flex: 1 },
  headerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: Spacing.md,
  },
  title: { ...Typography.headlineSm, color: Colors.onSurface },
  testBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: Colors.level3,
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: Radius.sm,
  },
  testBtnText: { ...Typography.telemetry, color: Colors.primary },
  list: { paddingBottom: 90, gap: Spacing.sm },
  profileCard: { overflow: 'hidden' },
  activeCard: { borderColor: Colors.primaryContainer },
  activeBar: {
    position: 'absolute',
    left: 0,
    top: 0,
    bottom: 0,
    width: 2,
    backgroundColor: Colors.primaryContainer,
  },
  profileTop: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: Spacing.sm,
  },
  profileTopRight: { flexDirection: 'row', alignItems: 'center', gap: 10 },
  profileName: { ...Typography.titleMd, fontSize: 15, color: Colors.onSurface, marginBottom: 4 },
  profileMeta: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  profileEndpoint: { ...Typography.bodySm, color: Colors.onSurfaceVariant },
  groupBadge: {
    backgroundColor: Colors.surfaceContainerHighest,
    borderRadius: Radius.sm,
    paddingHorizontal: 8,
    paddingVertical: 2,
  },
  groupText: { ...Typography.labelSm, color: Colors.onSurfaceVariant },
  fab: {
    position: 'absolute',
    right: Spacing.marginMobile,
    bottom: 24,
    width: 56,
    height: 56,
    borderRadius: Radius.full,
    backgroundColor: Colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: Colors.activeGlow,
    shadowOpacity: 1,
    shadowRadius: 8,
    shadowOffset: { width: 0, height: 4 },
    elevation: 8,
  },
});