import React, { useMemo, useState } from 'react';
import { SafeAreaView, ScrollView, Switch, Text, TextInput, View } from 'react-native';

export default function App() {
  const [strictMode, setStrictMode] = useState(true);
  const [city, setCity] = useState('Toulouse');
  const stats = useMemo(() => ({ blocked: 12, blurred: 38, syncedAt: '2026-04-22 22:00 UTC' }), []);

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: '#08120d' }}>
      <ScrollView contentContainerStyle={{ padding: 20, gap: 16 }}>
        <Text style={{ color: '#fff', fontSize: 28, fontWeight: '700' }}>HalalBrowse AI</Text>
        <Text style={{ color: '#b7d9c4' }}>Local VPN filter status: ready</Text>

        <View style={{ backgroundColor: '#0f2017', padding: 16, borderRadius: 12 }}>
          <Text style={{ color: '#fff', fontSize: 18, marginBottom: 8 }}>Salah-Time Strict Mode</Text>
          <Switch value={strictMode} onValueChange={setStrictMode} />
          <Text style={{ color: '#b7d9c4', marginTop: 8 }}>
            Threshold lowers automatically around prayer windows.
          </Text>
        </View>

        <View style={{ backgroundColor: '#0f2017', padding: 16, borderRadius: 12 }}>
          <Text style={{ color: '#fff', fontSize: 18, marginBottom: 8 }}>Prayer location</Text>
          <TextInput
            value={city}
            onChangeText={setCity}
            placeholder="City"
            placeholderTextColor="#7aa58b"
            style={{ backgroundColor: '#163224', color: '#fff', padding: 12, borderRadius: 8 }}
          />
        </View>

        <View style={{ backgroundColor: '#0f2017', padding: 16, borderRadius: 12 }}>
          <Text style={{ color: '#fff', fontSize: 18, marginBottom: 8 }}>Today</Text>
          <Text style={{ color: '#b7d9c4' }}>Blocked pages: {stats.blocked}</Text>
          <Text style={{ color: '#b7d9c4' }}>Blurred nodes: {stats.blurred}</Text>
          <Text style={{ color: '#b7d9c4' }}>Blocklist sync: {stats.syncedAt}</Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
