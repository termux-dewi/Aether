const { withAndroidManifest, withDangerousMod } = require('@expo/config-plugins');
const fs = require('fs');
const path = require('path');

/**
 * Expo config plugin that integrates the Aether gomobile AAR binding
 * into the generated Android project during `expo prebuild`.
 */

// Step 1: Copy AAR + patch build.gradle
const withAetherAarGradle = (config) => {
  return withDangerousMod(config, [
    'android',
    async (cfg) => {
      const appDir = path.join(cfg.modRequest.platformProjectRoot, 'app');
      const libsDir = path.join(appDir, 'libs');

      // Copy AAR into android/app/libs/
      const sourceAar = path.join(cfg.modRequest.projectRoot, 'libs', 'aether.aar');
      const destAar = path.join(libsDir, 'aether.aar');
      if (!fs.existsSync(libsDir)) fs.mkdirSync(libsDir, { recursive: true });
      if (fs.existsSync(sourceAar)) fs.copyFileSync(sourceAar, destAar);

      // Patch build.gradle to add AAR dependency
      const buildGradlePath = path.join(appDir, 'build.gradle');
      let gradle = fs.readFileSync(buildGradlePath, 'utf8');
      if (!gradle.includes("implementation files('libs/aether.aar')")) {
        gradle = gradle.replace(
          /dependencies\s*\{/,
          "dependencies {\n    implementation files('libs/aether.aar')"
        );
        fs.writeFileSync(buildGradlePath, gradle);
      }

      // Patch gradle.properties to ARM-only architectures (AAR is ARM-only)
      const propsPath = path.join(cfg.modRequest.platformProjectRoot, 'gradle.properties');
      let props = fs.readFileSync(propsPath, 'utf8');
      props = props.replace(
        /reactNativeArchitectures=.*/,
        'reactNativeArchitectures=armeabi-v7a,arm64-v8a'
      );
      fs.writeFileSync(propsPath, props);

      return cfg;
    },
  ]);
};

// Step 2: Add VpnService to AndroidManifest.xml
const withAetherVpnManifest = (config) => {
  return withAndroidManifest(config, async (cfg) => {
    const app = cfg.modResults.manifest.application[0];

    // Ensure service array exists
    if (!app.service) app.service = [];

    const hasVpn = app.service.some(
      (s) => s.$['android:name'] === '.AetherVpnService'
    );
    if (!hasVpn) {
      app.service.push({
        $: {
          'android:name': '.AetherVpnService',
          'android:permission': 'android.permission.BIND_VPN_SERVICE',
          'android:exported': 'true',
          'android:foregroundServiceType': 'dataSync',
        },
        'intent-filter': [
          { action: [{ $: { 'android:name': 'android.net.VpnService' } }] },
        ],
      });
    }

    // Add FOREGROUND_SERVICE_DATA_SYNC permission
    const perms = cfg.modResults.manifest['uses-permission'] || [];
    const hasFgDataSync = perms.some(
      (p) => p.$['android:name'] === 'android.permission.FOREGROUND_SERVICE_DATA_SYNC'
    );
    if (!hasFgDataSync) {
      perms.push({ $: { 'android:name': 'android.permission.FOREGROUND_SERVICE_DATA_SYNC' } });
    }
    cfg.modResults.manifest['uses-permission'] = perms;

    return cfg;
  });
};

// Step 3: Copy AetherVpnService.kt into the android source
const withAetherVpnServiceKt = (config) => {
  return withDangerousMod(config, [
    'android',
    async (cfg) => {
      const srcKt = path.join(cfg.modRequest.projectRoot, 'native', 'AetherVpnService.kt');
      const destDir = path.join(
        cfg.modRequest.platformProjectRoot,
        'app', 'src', 'main', 'java', 'com', 'aetherbox', 'app'
      );
      const destKt = path.join(destDir, 'AetherVpnService.kt');
      if (!fs.existsSync(destDir)) fs.mkdirSync(destDir, { recursive: true });
      if (fs.existsSync(srcKt)) fs.copyFileSync(srcKt, destKt);
      return cfg;
    },
  ]);
};

module.exports = function withAetherAar(config) {
  config = withAetherAarGradle(config);
  config = withAetherVpnManifest(config);
  config = withAetherVpnServiceKt(config);
  return config;
};