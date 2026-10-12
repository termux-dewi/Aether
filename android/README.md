# AetherBox — Android Client (React Native)

Android VPN client for AetherL3-Tunnel with App-Based Split Tunneling.
The UI is built with **React Native (Expo SDK 50)** and the project source
lives in [`../app/`](../app/). The native tunneling engine remains written in
Go and is exposed through a gomobile `.aar` binding consumed by a native
Android module.

> ⚠️ **ARM only** — this project targets `armeabi-v7a` and `arm64-v8a` exclusively.
> There is **no x86 / x86_64 support** by design (configured via the
> `expo-build-properties` plugin in `app/app.json` → `abiFilters`).

## Architecture

```
┌─────────────────────────────────────────────┐
│  React Native UI (Expo, app/)               │
│  Beranda · Konfigurasi · Split Tunnel · Logs│
├─────────────────────────────────────────────┤
│  Native Module bridge (Kotlin)              │
│  - exposes Go engine to JS                  │
├─────────────────────────────────────────────┤
│  VpnService (Android Framework)             │
│  - addAllowedApplication() for split tunnel │
│  - ParcelFileDescriptor (TUN fd)            │
├─────────────────────────────────────────────┤
│  Go Engine via gomobile (.aar binding)      │
│  - internal/client engine (mobile/mobile.go)│
│  - L3 framing + gRPC/TLS 1.3 transport      │
└─────────────────────────────────────────────┘
```

## Prerequisites

1. Node.js 18+ and npm
2. Go 1.22+ with gomobile:
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ```
3. Android SDK 34 + NDK (for gomobile bind)
4. JDK 17 (for local EAS builds)
5. Expo account (only for cloud builds: `npx eas login`)

## Build Steps

```bash
# 1. Install JS dependencies
cd app && npm install

# 2. Generate the Go AAR binding (ARM only)
make gomobile-bind

# 3. Typecheck the TS sources
npm run typecheck

# 4. Build the APK with EAS CLI
npx eas build --platform android --profile preview --local   # local build
# or on EAS cloud:
npx eas build --platform android --profile preview
```

Equivalent Make targets from the repo root: `make app-install`,
`make app-typecheck`, `make app-build-apk`, `make app-build-apk-cloud`.

## Build Profiles (`app/eas.json`)

| Profile      | Output | Purpose                          |
|--------------|--------|----------------------------------|
| development  | APK    | Dev client, debuggable           |
| preview      | APK    | Internal distribution (testing)  |
| production   | APK    | Release, auto-increment version  |

## Split Tunneling

The native `VpnService` uses `Builder.addAllowedApplication(packageName)` to
route only the apps selected on the *Split Tunnel* tab through the tunnel:

```kotlin
val builder = VpnService.Builder()
selectedApps.forEach { pkg ->
    builder.addAllowedApplication(pkg)
}
```

## Project Structure

```
app/                          # React Native (Expo) project
├── App.tsx                   # Root component + navigation theme
├── app.json                  # Expo config (ARM-only abiFilters)
├── eas.json                  # EAS build profiles (APK output)
├── src/
│   ├── navigation/RootNavigator.tsx
│   ├── screens/
│   │   ├── BerandaScreen.tsx       # Dashboard + connect/disconnect
│   │   ├── KonfigurasiScreen.tsx   # Server profile list + FAB
│   │   ├── SplitTunnelScreen.tsx   # Per-app split tunneling
│   │   └── LogsScreen.tsx          # AetherL3 debug logs
│   ├── components/                 # Card, ProtocolChip, PingPill, …
│   └── theme/colors.ts             # Design tokens from Stitch reference
└── assets/                   # icon / splash / adaptive-icon

android/                      # Legacy Kotlin reference (being replaced by RN)
└── AetherBox/…
```