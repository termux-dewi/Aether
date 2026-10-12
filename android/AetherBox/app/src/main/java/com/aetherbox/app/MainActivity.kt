package com.aetherbox.app

import android.content.Intent
import android.net.VpnService
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

/**
 * MainActivity is the entry point for AetherBox Android app.
 * Provides a NekoBox-style UI for profile management and tunnel control.
 */
class MainActivity : ComponentActivity() {

    companion object {
        private const val VPN_REQUEST_CODE = 1001
    }

    private var isConnected by mutableStateOf(false)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        setContent {
            MaterialTheme {
                AetherBoxApp(
                    isConnected = isConnected,
                    onConnect = { startVpn() },
                    onDisconnect = { stopVpn() }
                )
            }
        }
    }

    private fun startVpn() {
        val intent = VpnService.prepare(this)
        if (intent != null) {
            startActivityForResult(intent, VPN_REQUEST_CODE)
        } else {
            onActivityResult(VPN_REQUEST_CODE, RESULT_OK, null)
        }
    }

    private fun stopVpn() {
        stopService(Intent(this, AetherVpnService::class.java))
        isConnected = false
    }

    @Deprecated("Use Activity Result API")
    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)
        if (requestCode == VPN_REQUEST_CODE && resultCode == RESULT_OK) {
            val vpnIntent = Intent(this, AetherVpnService::class.java).apply {
                putExtra("server_addr", "your-vps:443")
                putExtra("psk", "your-psk-hex")
                putExtra("sni", "your-domain.com")
                putStringArrayListExtra("allowed_apps", arrayListOf())
            }
            startForegroundService(vpnIntent)
            isConnected = true
        }
    }
}

@Composable
fun AetherBoxApp(
    isConnected: Boolean,
    onConnect: () -> Unit,
    onDisconnect: () -> Unit
) {
    Surface(
        modifier = Modifier.fillMaxSize(),
        color = MaterialTheme.colorScheme.background
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = "AetherBox",
                style = MaterialTheme.typography.headlineLarge
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = if (isConnected) "Connected" else "Disconnected",
                style = MaterialTheme.typography.bodyLarge,
                color = if (isConnected)
                    MaterialTheme.colorScheme.primary
                else
                    MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(32.dp))

            Button(
                onClick = if (isConnected) onDisconnect else onConnect,
                modifier = Modifier
                    .fillMaxWidth(0.6f)
                    .height(56.dp)
            ) {
                Text(if (isConnected) "Disconnect" else "Connect")
            }
        }
    }
}