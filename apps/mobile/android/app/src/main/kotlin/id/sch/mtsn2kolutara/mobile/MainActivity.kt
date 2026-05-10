package id.sch.mtsn2kolutara.mobile

import android.os.Bundle
import android.view.WindowManager
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.EventChannel
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {
    private var antiCheatEventSink: EventChannel.EventSink? = null
    private var hasWindowFocusNow: Boolean = true

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableSecureFlag()
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(
            flutterEngine.dartExecutor.binaryMessenger,
            ANTI_CHEAT_METHOD_CHANNEL,
        ).setMethodCallHandler { call, result ->
            when (call.method) {
                "enableSecureFlag" -> {
                    enableSecureFlag()
                    result.success(null)
                    emitAntiCheatState()
                }
                "getWindowState" -> result.success(antiCheatState())
                else -> result.notImplemented()
            }
        }
        EventChannel(
            flutterEngine.dartExecutor.binaryMessenger,
            ANTI_CHEAT_EVENT_CHANNEL,
        ).setStreamHandler(object : EventChannel.StreamHandler {
            override fun onListen(arguments: Any?, events: EventChannel.EventSink?) {
                antiCheatEventSink = events
                emitAntiCheatState()
            }

            override fun onCancel(arguments: Any?) {
                antiCheatEventSink = null
            }
        })
    }

    override fun onResume() {
        super.onResume()
        enableSecureFlag()
        emitAntiCheatState()
    }

    override fun onWindowFocusChanged(hasFocus: Boolean) {
        super.onWindowFocusChanged(hasFocus)
        hasWindowFocusNow = hasFocus
        emitAntiCheatState()
    }

    override fun onMultiWindowModeChanged(isInMultiWindowMode: Boolean) {
        super.onMultiWindowModeChanged(isInMultiWindowMode)
        emitAntiCheatState()
    }

    override fun onPictureInPictureModeChanged(isInPictureInPictureMode: Boolean) {
        super.onPictureInPictureModeChanged(isInPictureInPictureMode)
        emitAntiCheatState()
    }

    private fun enableSecureFlag() {
        window.setFlags(
            WindowManager.LayoutParams.FLAG_SECURE,
            WindowManager.LayoutParams.FLAG_SECURE,
        )
    }

    private fun secureFlagEnabled(): Boolean {
        return (window.attributes.flags and WindowManager.LayoutParams.FLAG_SECURE) != 0
    }

    private fun antiCheatState(): Map<String, Any> {
        return mapOf(
            "isMultiWindow" to isInMultiWindowMode,
            "isPictureInPicture" to isInPictureInPictureMode,
            "hasWindowFocus" to hasWindowFocusNow,
            "secureFlagEnabled" to secureFlagEnabled(),
        )
    }

    private fun emitAntiCheatState() {
        antiCheatEventSink?.success(antiCheatState())
    }

    companion object {
        private const val ANTI_CHEAT_METHOD_CHANNEL =
            "id.sch.mtsn2kolutara.mobile/anti_cheat"
        private const val ANTI_CHEAT_EVENT_CHANNEL =
            "id.sch.mtsn2kolutara.mobile/anti_cheat_events"
    }
}
