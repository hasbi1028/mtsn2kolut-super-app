import 'dart:io';
import 'dart:math';

import 'exam_session_store.dart';

class DeviceFingerprintStore {
  static const _installIdKey = 'exam_device_install_id';

  DeviceFingerprintStore({
    SnapshotValueStore? secureStore,
    String? operatingSystem,
    String? hostName,
    Random? random,
  }) : _secureStore = secureStore ?? const SecureSnapshotValueStore(),
       _operatingSystem = operatingSystem,
       _hostName = hostName,
       _random = random;

  final SnapshotValueStore _secureStore;
  final String? _operatingSystem;
  final String? _hostName;
  final Random? _random;

  Future<String> loadFingerprint() async {
    final installId = await _loadOrCreateInstallId();
    final os = _cleanPart(_operatingSystem ?? Platform.operatingSystem);
    final host = _cleanPart(_hostName ?? Platform.localHostname);
    // BYOD note: this is only a lightweight telemetry/resume hint, not a
    // strong device identity proof.
    return '$os:$host:install-$installId';
  }

  Future<String> _loadOrCreateInstallId() async {
    final existing = await _secureStore.read(_installIdKey);
    if (existing != null && existing.trim().isNotEmpty) {
      return existing.trim();
    }

    final generated = _generateInstallId(_random ?? Random.secure());
    await _secureStore.write(_installIdKey, generated);
    return generated;
  }
}

String _generateInstallId(Random random) {
  final buffer = StringBuffer();
  for (var i = 0; i < 16; i += 1) {
    buffer.write(random.nextInt(256).toRadixString(16).padLeft(2, '0'));
  }
  return buffer.toString();
}

String _cleanPart(String value) {
  final cleaned = value.trim().replaceAll(RegExp(r'\s+'), '-');
  return cleaned.isEmpty ? 'unknown' : cleaned;
}
