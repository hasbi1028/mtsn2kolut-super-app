import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/device_fingerprint.dart';
import 'package:mobile/src/exam_session_store.dart';

void main() {
  group('DeviceFingerprintStore', () {
    test('persists generated install-scoped id securely', () async {
      final secureStore = MemorySnapshotValueStore();
      final store = DeviceFingerprintStore(
        secureStore: secureStore,
        operatingSystem: 'android',
        hostName: 'student phone',
      );

      final first = await store.loadFingerprint();
      final second = await DeviceFingerprintStore(
        secureStore: secureStore,
        operatingSystem: 'android',
        hostName: 'student phone',
      ).loadFingerprint();

      expect(first, second);
      expect(first, startsWith('android:student-phone:install-'));
      expect(secureStore.values['exam_device_install_id'], isNotNull);
      expect(secureStore.values['exam_device_install_id'], hasLength(32));
    });

    test('keeps install id while reflecting platform metadata', () async {
      final secureStore = MemorySnapshotValueStore();
      await secureStore.write('exam_device_install_id', 'abc123');

      final fingerprint = await DeviceFingerprintStore(
        secureStore: secureStore,
        operatingSystem: 'android',
        hostName: 'Lab Tablet',
      ).loadFingerprint();

      expect(fingerprint, 'android:Lab-Tablet:install-abc123');
    });
  });
}

class MemorySnapshotValueStore implements SnapshotValueStore {
  final Map<String, String> values = <String, String>{};

  @override
  Future<void> delete(String key) async {
    values.remove(key);
  }

  @override
  Future<String?> read(String key) async => values[key];

  @override
  Future<void> write(String key, String value) async {
    values[key] = value;
  }
}
