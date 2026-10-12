import React, { useMemo, useState } from 'react';
import { View, Text, StyleSheet, Pressable, FlatList, TextInput, Switch } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import ScreenContainer from '../components/ScreenContainer';
import Card from '../components/Card';
import { Colors, Spacing, Radius, Typography } from '../theme/colors';

interface AppItem {
  id: string;
  name: string;
  pkg: string;
  icon: keyof typeof Ionicons.glyphMap;
  color: string;
}

const APPS: AppItem[] = [
  { id: '1', name: 'Telegram', pkg: 'org.telegram.messenger', icon: 'paper-plane', color: '#2AABEE' },
  { id: '2', name: 'Spotify', pkg: 'com.spotify.music', icon: 'musical-notes', color: '#1DB954' },
  { id: '3', name: 'Chrome', pkg: 'com.android.chrome', icon: 'globe', color: '#4285F4' },
  { id: '4', name: 'YouTube', pkg: 'com.google.android.youtube', icon: 'logo-youtube', color: '#FF0000' },
  { id: '5', name: 'WhatsApp', pkg: 'com.whatsapp', icon: 'chatbubbles', color: '#25D366' },
  { id: '6', name: 'Twitter / X', pkg: 'com.twitter.android', icon: 'logo-twitter', color: '#1DA1F2' },
  { id: '7', name: 'Play Store', pkg: 'com.android.vending', icon: 'play', color: '#689F38' },
];

export default function SplitTunnelScreen() {
  const [enabled, setEnabled] = useState(true);
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<Set<string>>(new Set(['1', '2', '5']));

  const filtered = useMemo(() => {
    const q = query.toLowerCase().trim();
    if (!q) return APPS;
    return APPS.filter((a) => a.name.toLowerCase().includes(q) || a.pkg.toLowerCase().includes(q));
  }, [query]);

  const toggleApp = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const toggleAll = () => {
    setSelected((prev) => (prev.size === APPS.length ? new Set() : new Set(APPS.map((a) => a.id))));
  };

  return (
    <ScreenContainer scroll={false} style={styles.container}>
      <Text style={styles.title}>Split Tunneling</Text>
      <Text style={styles.subtitle}>
        Hanya aplikasi terpilih yang dirutekan melalui terowongan L3.
      </Text>

      <Card style={styles.masterCard}>
        <View style={styles.masterRow}>
          <View style={styles.masterLeft}>
            <Ionicons name="git-branch" size={20} color={Colors.primary} />
            <Text style={styles.masterLabel}>App-Based Split Tunnel</Text>
          </View>
          <Switch
            value={enabled}
            onValueChange={setEnabled}
            trackColor={{ false: Colors.level3, true: Colors.primaryContainer }}
            thumbColor={Colors.onSurface}
          />
        </View>
      </Card>

      <View style={styles.searchRow}>
        <View style={styles.searchBox}>
          <Ionicons name="search" size={16} color={Colors.muted} />
          <TextInput
            style={styles.searchInput}
            placeholder="Cari aplikasi atau package..."
            placeholderTextColor={Colors.muted}
            value={query}
            onChangeText={setQuery}
          />
          {query.length > 0 && (
            <Pressable onPress={() => setQuery('')}>
              <Ionicons name="close-circle" size={16} color={Colors.muted} />
            </Pressable>
          )}
        </View>
        <Pressable onPress={toggleAll} style={styles.toggleAllBtn}>
          <Text style={styles.toggleAllText}>
            {selected.size === APPS.length ? 'Batalkan' : 'Semua'}
          </Text>
        </Pressable>
      </View>

      <Text style={styles.countLabel}>
        {selected.size} dari {APPS.length} aplikasi dipilih
      </Text>

      <FlatList
        data={filtered}
        keyExtractor={(i) => i.id}
        style={styles.list}
        contentContainerStyle={styles.listContent}
        showsVerticalScrollIndicator={false}
        renderItem={({ item }) => {
          const isSelected = selected.has(item.id);
          return (
            <Pressable onPress={() => enabled && toggleApp(item.id)}>
              <View style={[styles.appRow, isSelected && styles.appRowSelected]}>
                <View style={[styles.appIcon, { backgroundColor: item.color + '22' }]}>
                  <Ionicons name={item.icon} size={20} color={item.color} />
                </View>
                <View style={styles.appInfo}>
                  <Text style={styles.appName}>{item.name}</Text>
                  <Text style={styles.appPkg}>{item.pkg}</Text>
                </View>
                <Ionicons
                  name={isSelected ? 'checkbox' : 'square-outline'}
                  size={22}
                  color={isSelected ? Colors.primary : Colors.muted}
                />
              </View>
            </Pressable>
          );
        }}
      />

      <Pressable
        style={({ pressed }) => [styles.saveBtn, pressed && { opacity: 0.9 }]}
        onPress={() => {}}
      >
        <Ionicons name="save" size={20} color={Colors.onPrimary} />
        <Text style={styles.saveText}>Simpan Aturan Split</Text>
      </Pressable>
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, paddingBottom: 0 },
  title: { ...Typography.headlineSm, color: Colors.onSurface },
  subtitle: { ...Typography.bodySm, color: Colors.onSurfaceVariant, marginTop: 4, marginBottom: Spacing.md },
  masterCard: { marginBottom: Spacing.md, paddingVertical: 12 },
  masterRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  masterLeft: { flexDirection: 'row', alignItems: 'center', gap: 10 },
  masterLabel: { ...Typography.titleMd, color: Colors.onSurface },
  searchRow: { flexDirection: 'row', gap: Spacing.sm, marginBottom: Spacing.sm },
  searchBox: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: Colors.level1,
    borderWidth: 1,
    borderColor: Colors.borderInput,
    borderRadius: Radius.sm,
    paddingHorizontal: 10,
    height: 42,
  },
  searchInput: { flex: 1, ...Typography.bodyMd, color: Colors.onSurface },
  toggleAllBtn: {
    justifyContent: 'center',
    paddingHorizontal: 12,
    backgroundColor: Colors.level3,
    borderRadius: Radius.sm,
  },
  toggleAllText: { ...Typography.telemetry, color: Colors.onSurface },
  countLabel: { ...Typography.labelMd, color: Colors.muted, marginBottom: Spacing.sm },
  list: { flex: 1 },
  listContent: { gap: Spacing.sm, paddingBottom: Spacing.md },
  appRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    backgroundColor: Colors.level2,
    borderRadius: Radius.md,
    borderWidth: 1,
    borderColor: Colors.borderCard,
    padding: 12,
  },
  appRowSelected: { borderColor: Colors.primaryContainer },
  appIcon: {
    width: 40,
    height: 40,
    borderRadius: Radius.md,
    alignItems: 'center',
    justifyContent: 'center',
  },
  appInfo: { flex: 1 },
  appName: { ...Typography.titleMd, color: Colors.onSurface },
  appPkg: { ...Typography.bodySm, color: Colors.muted },
  saveBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
    backgroundColor: Colors.primary,
    borderRadius: Radius.lg,
    paddingVertical: 14,
    marginBottom: Spacing.sm,
  },
  saveText: { ...Typography.titleMd, fontWeight: '600', color: Colors.onPrimary },
});