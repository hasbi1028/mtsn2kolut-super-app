#ifndef RUNNER_ANTI_CHEAT_BRIDGE_H_
#define RUNNER_ANTI_CHEAT_BRIDGE_H_

#include <flutter/binary_messenger.h>
#include <windows.h>

#include <memory>

// Registers the CBT anti-cheat MethodChannel/EventChannel used by the Dart app.
// The Android implementation uses the same channel names; this Windows runner
// keeps the Dart anti-cheat guard platform-agnostic.
class AntiCheatBridge {
 public:
  AntiCheatBridge(flutter::BinaryMessenger* messenger, HWND window);
  ~AntiCheatBridge();

  AntiCheatBridge(const AntiCheatBridge&) = delete;
  AntiCheatBridge& operator=(const AntiCheatBridge&) = delete;

  void Register();
  void ApplyExamWindowPolicy();
  void HandleWindowMessage(UINT message, WPARAM wparam, LPARAM lparam);

 private:
  class Impl;
  std::unique_ptr<Impl> impl_;
};

#endif  // RUNNER_ANTI_CHEAT_BRIDGE_H_
