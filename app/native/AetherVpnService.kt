package com.aetherbox.app

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Intent
import android.net.VpnService
import android.os.Build
import android.os.ParcelFileDescriptor
import android.util.Log
import java.io.FileInputStream
import java.io.FileOutputStream

/**
 * AetherVpnService – core VPN service for AetherBox.
 * Captures IP packets from selected apps via VpnService TUN
 * and forwards them through the Go tunnel engine (gomobile AAR).
 */
class AetherVpnService : VpnService() {

    companion object {
        private const val TAG = "AetherVpnService"
        private const val MTU = 1500
        private const val TUN_ADDRESS = "10.8.0.2"
        private const val TUN_PREFIX = 24
        private const val CHANNEL_ID = "aether_vpn"
        private const val NOTIFICATION_ID = 1
    }

    private var tunFd: ParcelFileDescriptor? = null
    @Volatile
    private var isRunning = false

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.i(TAG, "VPN service starting")

        val serverAddr = intent?.getStringExtra("server_addr") ?: return START_NOT_STICKY
        val psk = intent.getStringExtra("psk") ?: return START_NOT_STICKY
        val sni = intent.getStringExtra("sni") ?: ""
        val allowedApps = intent.getStringArrayListExtra("allowed_apps") ?: arrayListOf()

        // Start foreground
        startForegroundNotification()

        // Build VpnService with optional split tunneling
        val builder = Builder()
            .setMtu(MTU)
            .addAddress(TUN_ADDRESS, TUN_PREFIX)
            .addRoute("0.0.0.0", 0)
            .setSession("AetherBox")
            .setBlocking(true)

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

        isRunning = true

        // TODO: Initialize Go tunnel engine via gomobile binding
        // val engine = Mobile.newTunnelEngine(serverAddr, psk, sni, false)
        // engine.connect()

        // Packet forwarding threads
        Thread({ tunToTunnel() }, "TUN->Tunnel").start()
        Thread({ tunnelToTun() }, "Tunnel->TUN").start()

        return START_STICKY
    }

    private fun tunToTunnel() {
        val input = FileInputStream(tunFd!!.fileDescriptor)
        val buffer = ByteArray(MTU)
        try {
            while (isRunning) {
                val n = input.read(buffer)
                if (n <= 0) continue
                val packet = buffer.copyOf(n)
                // TODO: engine.writePacket(packet)
                Log.d(TAG, "TUN->Tunnel: $n bytes")
            }
        } catch (e: Exception) {
            if (isRunning) Log.e(TAG, "TUN read error", e)
        }
    }

    private fun tunnelToTun() {
        val output = FileOutputStream(tunFd!!.fileDescriptor)
        try {
            while (isRunning) {
                // TODO: val packet = engine.readPacket()
                // if (packet != null) output.write(packet)
                Thread.sleep(1)
            }
        } catch (e: Exception) {
            if (isRunning) Log.e(TAG, "TUN write error", e)
        }
    }

    private fun startForegroundNotification() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID, "AetherBox VPN",
                NotificationManager.IMPORTANCE_LOW
            )
            val nm = getSystemService(NotificationManager::class.java)
            nm.createNotificationChannel(channel)
        }
        val notification = Notification.Builder(this, CHANNEL_ID)
            .setContentTitle("AetherBox")
            .setContentText("VPN tunnel active")
            .setSmallIcon(android.R.drawable.ic_lock_lock)
            .build()
        startForeground(NOTIFICATION_ID, notification)
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
        // TODO: engine?.disconnect()
        tunFd?.close()
        tunFd = null
        stopForeground(true)
        stopSelf()
    }
}