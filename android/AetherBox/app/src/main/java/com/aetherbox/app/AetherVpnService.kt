package com.aetherbox.app

import android.net.VpnService
import android.os.ParcelFileDescriptor
import android.util.Log
import java.io.FileInputStream
import java.io.FileOutputStream

/**
 * AetherVpnService is the core VPN service for AetherBox.
 * It captures IP packets from selected apps via VpnService TUN
 * and forwards them through the Go tunnel engine (gomobile binding).
 */
class AetherVpnService : VpnService() {

    companion object {
        private const val TAG = "AetherVpnService"
        private const val MTU = 1500
        private const val TUN_ADDRESS = "10.8.0.2"
        private const val TUN_PREFIX = 24
    }

    private var tunFd: ParcelFileDescriptor? = null
    private var tunnelEngine: Any? = null // mobile.TunnelEngine from gomobile
    private var isRunning = false

    override fun onStartCommand(intent: android.content.Intent?, flags: Int, startId: Int): Int {
        Log.i(TAG, "VPN service starting")

        val serverAddr = intent?.getStringExtra("server_addr") ?: return START_NOT_STICKY
        val psk = intent.getStringExtra("psk") ?: return START_NOT_STICKY
        val sni = intent.getStringExtra("sni") ?: ""
        val allowedApps = intent.getStringArrayListExtra("allowed_apps") ?: arrayListOf()

        // Build VpnService with split tunneling
        val builder = Builder()
            .setMtu(MTU)
            .addAddress(TUN_ADDRESS, TUN_PREFIX)
            .addRoute("0.0.0.0", 0) // Route all traffic through tunnel
            .setSession("AetherBox")
            .setBlocking(true)

        // App-Based Split Tunneling: only allowed apps go through tunnel
        if (allowedApps.isNotEmpty()) {
            allowedApps.forEach { pkg ->
                try {
                    builder.addAllowedApplication(pkg)
                } catch (e: Exception) {
                    Log.w(TAG, "Failed to add allowed app: $pkg", e)
                }
            }
            Log.i(TAG, "Split tunneling: ${allowedApps.size} apps allowed")
        }

        try {
            tunFd = builder.establish()
        } catch (e: Exception) {
            Log.e(TAG, "Failed to establish VPN", e)
            return START_NOT_STICKY
        }

        if (tunFd == null) {
            Log.e(TAG, "Failed to get TUN file descriptor")
            return START_NOT_STICKY
        }

        // Initialize Go tunnel engine via gomobile binding
        // tunnelEngine = Mobile.newTunnelEngine(
        //     Mobile.TunnelConfig(serverAddr, psk, sni, false)
        // )
        // tunnelEngine.connect()

        isRunning = true

        // Start packet forwarding threads
        Thread({ tunToTunnel() }, "TUN->Tunnel").start()
        Thread({ tunnelToTun() }, "Tunnel->TUN").start()

        return START_STICKY
    }

    /**
     * Reads packets from TUN fd and writes them into the tunnel.
     */
    private fun tunToTunnel() {
        val input = FileInputStream(tunFd!!.fileDescriptor)
        val buffer = ByteArray(MTU)

        try {
            while (isRunning) {
                val n = input.read(buffer)
                if (n <= 0) continue

                val packet = buffer.copyOf(n)
                // tunnelEngine.writePacket(packet)
                Log.d(TAG, "TUN->Tunnel: ${n} bytes")
            }
        } catch (e: Exception) {
            if (isRunning) Log.e(TAG, "TUN read error", e)
        }
    }

    /**
     * Reads packets from the tunnel and writes them to TUN fd.
     */
    private fun tunnelToTun() {
        val output = FileOutputStream(tunFd!!.fileDescriptor)

        try {
            while (isRunning) {
                // val packet = tunnelEngine.readPacket()
                // if (packet != null) {
                //     output.write(packet)
                // }
                Thread.sleep(1) // Placeholder until gomobile binding is integrated
            }
        } catch (e: Exception) {
            if (isRunning) Log.e(TAG, "TUN write error", e)
        }
    }

    override fun onRevoke() {
        Log.i(TAG, "VPN revoked")
        stopTunnel()
    }

    override fun onDestroy() {
        Log.i(TAG, "VPN service destroyed")
        stopTunnel()
        super.onDestroy()
    }

    private fun stopTunnel() {
        isRunning = false
        // tunnelEngine?.disconnect()
        tunFd?.close()
        tunFd = null
    }
}