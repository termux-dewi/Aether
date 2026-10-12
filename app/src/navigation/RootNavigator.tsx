import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Ionicons } from '@expo/vector-icons';
import { Colors } from '../theme/colors';
import BerandaScreen from '../screens/BerandaScreen';
import KonfigurasiScreen from '../screens/KonfigurasiScreen';
import SplitTunnelScreen from '../screens/SplitTunnelScreen';
import LogsScreen from '../screens/LogsScreen';

export type RootTabParamList = {
  Beranda: undefined;
  Konfigurasi: undefined;
  SplitTunnel: undefined;
  Logs: undefined;
};

const Tab = createBottomTabNavigator<RootTabParamList>();

const TAB_ICONS: Record<keyof RootTabParamList, { active: keyof typeof Ionicons.glyphMap; inactive: keyof typeof Ionicons.glyphMap }> = {
  Beranda: { active: 'shield', inactive: 'shield-outline' },
  Konfigurasi: { active: 'server', inactive: 'server-outline' },
  SplitTunnel: { active: 'git-branch', inactive: 'git-branch-outline' },
  Logs: { active: 'terminal', inactive: 'terminal-outline' },
};

export default function RootNavigator() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarActiveTintColor: Colors.primary,
        tabBarInactiveTintColor: Colors.onSurfaceVariant,
        tabBarStyle: {
          backgroundColor: 'rgba(27, 27, 31, 0.9)',
          borderTopColor: Colors.borderHairline,
          height: 64,
          paddingTop: 4,
        },
        tabBarLabelStyle: {
          fontSize: 11,
          fontWeight: '600',
          letterSpacing: 0.5,
        },
        tabBarIcon: ({ focused, color, size }) => {
          const icons = TAB_ICONS[route.name];
          return (
            <Ionicons
              name={focused ? icons.active : icons.inactive}
              size={size}
              color={color}
            />
          );
        },
      })}
    >
      <Tab.Screen name="Beranda" component={BerandaScreen} options={{ title: 'Beranda' }} />
      <Tab.Screen name="Konfigurasi" component={KonfigurasiScreen} options={{ title: 'Konfigurasi' }} />
      <Tab.Screen name="SplitTunnel" component={SplitTunnelScreen} options={{ title: 'Split Tunnel' }} />
      <Tab.Screen name="Logs" component={LogsScreen} options={{ title: 'Logs' }} />
    </Tab.Navigator>
  );
}