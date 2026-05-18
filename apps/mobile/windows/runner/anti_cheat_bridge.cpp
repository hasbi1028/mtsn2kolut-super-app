#include "anti_cheat_bridge.h"

#include <flutter/encodable_value.h>
#include <flutter/event_channel.h>
#include <flutter/event_sink.h>
#include <flutter/event_stream_handler_functions.h>
#include <flutter/method_channel.h>
#include <flutter/method_result_functions.h>
#include <flutter/standard_method_codec.h>
#include <windows.h>

#include <algorithm>
#include <cstdlib>
#include <memory>
#include <string>

namespace {
constexpr char kMethodChannel[] = "id.sch.mtsn2kolutara.mobile/anti_cheat";
constexpr char kEventChannel[] = "id.sch.mtsn2kolutara.mobile/anti_cheat_events";

bool IsWindowFullscreen(HWND window) {
  if (!window) {
    return false;
  }

  RECT window_rect{};
  if (!::GetWindowRect(window, &window_rect)) {
    return false;
  }

  HMONITOR monitor = ::MonitorFromWindow(window, MONITOR_DEFAULTTONEAREST);
  MONITORINFO monitor_info{};
  monitor_info.cbSize = sizeof(monitor_info);
  if (!::GetMonitorInfo(monitor, &monitor_info)) {
    return false;
  }

  const RECT& monitor_rect = monitor_info.rcMonitor;
  constexpr int kTolerancePixels = 2;
  return std::abs(window_rect.left - monitor_rect.left) <= kTolerancePixels &&
         std::abs(window_rect.top - monitor_rect.top) <= kTolerancePixels &&
         std::abs(window_rect.right - monitor_rect.right) <= kTolerancePixels &&
         std::abs(window_rect.bottom - monitor_rect.bottom) <= kTolerancePixels;
}

bool ApplyDisplayAffinity(HWND window) {
  if (!window) {
    return false;
  }

#ifndef WDA_EXCLUDEFROMCAPTURE
#define WDA_EXCLUDEFROMCAPTURE 0x00000011
#endif
  if (::SetWindowDisplayAffinity(window, WDA_EXCLUDEFROMCAPTURE)) {
    return true;
  }
  // Older supported Windows builds may not support EXCLUDEFROMCAPTURE. Fall
  // back to monitor-only affinity instead of failing the app start.
  return ::SetWindowDisplayAffinity(window, WDA_MONITOR) != FALSE;
}

void EnterFullscreen(HWND window) {
  if (!window) {
    return;
  }

  HMONITOR monitor = ::MonitorFromWindow(window, MONITOR_DEFAULTTONEAREST);
  MONITORINFO monitor_info{};
  monitor_info.cbSize = sizeof(monitor_info);
  if (!::GetMonitorInfo(monitor, &monitor_info)) {
    return;
  }

  const LONG_PTR current_style = ::GetWindowLongPtr(window, GWL_STYLE);
  const LONG_PTR fullscreen_style = current_style & ~WS_OVERLAPPEDWINDOW;
  ::SetWindowLongPtr(window, GWL_STYLE, fullscreen_style | WS_POPUP | WS_VISIBLE);

  const RECT& rect = monitor_info.rcMonitor;
  ::SetWindowPos(window, HWND_TOPMOST, rect.left, rect.top,
                 rect.right - rect.left, rect.bottom - rect.top,
                 SWP_FRAMECHANGED | SWP_SHOWWINDOW);
}
}  // namespace

class AntiCheatBridge::Impl {
 public:
  Impl(flutter::BinaryMessenger* messenger, HWND window)
      : messenger_(messenger), window_(window) {}

  void Register() {
    method_channel_ = std::make_unique<flutter::MethodChannel<flutter::EncodableValue>>(
        messenger_, kMethodChannel,
        &flutter::StandardMethodCodec::GetInstance());

    method_channel_->SetMethodCallHandler(
        [this](const flutter::MethodCall<flutter::EncodableValue>& call,
               std::unique_ptr<flutter::MethodResult<flutter::EncodableValue>> result) {
          if (call.method_name() == "enableSecureFlag") {
            ApplyExamWindowPolicy();
            result->Success(flutter::EncodableValue());
            return;
          }
          if (call.method_name() == "getWindowState") {
            result->Success(BuildState());
            return;
          }
          result->NotImplemented();
        });

    event_channel_ = std::make_unique<flutter::EventChannel<flutter::EncodableValue>>(
        messenger_, kEventChannel,
        &flutter::StandardMethodCodec::GetInstance());

    auto handler = std::make_unique<
        flutter::StreamHandlerFunctions<flutter::EncodableValue>>(
        [this](const flutter::EncodableValue* arguments,
               std::unique_ptr<flutter::EventSink<flutter::EncodableValue>>&& events)
            -> std::unique_ptr<flutter::StreamHandlerError<flutter::EncodableValue>> {
          event_sink_ = std::move(events);
          EmitState();
          return nullptr;
        },
        [this](const flutter::EncodableValue* arguments)
            -> std::unique_ptr<flutter::StreamHandlerError<flutter::EncodableValue>> {
          event_sink_.reset();
          return nullptr;
        });
    event_channel_->SetStreamHandler(std::move(handler));
  }

  void ApplyExamWindowPolicy() {
    EnterFullscreen(window_);
    secure_flag_enabled_ = ApplyDisplayAffinity(window_);
    EmitState();
  }

  void HandleWindowMessage(UINT message, WPARAM wparam, LPARAM lparam) {
    switch (message) {
      case WM_ACTIVATE:
      case WM_ACTIVATEAPP:
      case WM_SETFOCUS:
      case WM_KILLFOCUS:
      case WM_SIZE:
      case WM_WINDOWPOSCHANGED:
      case WM_DISPLAYCHANGE:
        // Re-apply fullscreen after restore/resize attempts. This is
        // best-effort prevention; telemetry still reports the observed state.
        if (message == WM_SIZE && wparam == SIZE_RESTORED) {
          EnterFullscreen(window_);
        }
        EmitState();
        break;
      default:
        break;
    }
  }

 private:
  flutter::EncodableValue BuildState() const {
    const HWND foreground = ::GetForegroundWindow();
    const bool has_focus = foreground == window_ || ::IsChild(window_, foreground);
    const bool minimized = ::IsIconic(window_) != FALSE;
    const bool fullscreen = IsWindowFullscreen(window_);

    flutter::EncodableMap state;
    state[flutter::EncodableValue("platform")] = flutter::EncodableValue("windows");
    state[flutter::EncodableValue("isMultiWindow")] = flutter::EncodableValue(false);
    state[flutter::EncodableValue("isPictureInPicture")] = flutter::EncodableValue(false);
    state[flutter::EncodableValue("hasWindowFocus")] = flutter::EncodableValue(has_focus);
    state[flutter::EncodableValue("secureFlagEnabled")] =
        flutter::EncodableValue(secure_flag_enabled_);
    state[flutter::EncodableValue("isFullscreen")] = flutter::EncodableValue(fullscreen);
    state[flutter::EncodableValue("isMinimized")] = flutter::EncodableValue(minimized);
    return flutter::EncodableValue(state);
  }

  void EmitState() {
    if (event_sink_) {
      event_sink_->Success(BuildState());
    }
  }

  flutter::BinaryMessenger* messenger_ = nullptr;
  HWND window_ = nullptr;
  bool secure_flag_enabled_ = false;
  std::unique_ptr<flutter::MethodChannel<flutter::EncodableValue>> method_channel_;
  std::unique_ptr<flutter::EventChannel<flutter::EncodableValue>> event_channel_;
  std::unique_ptr<flutter::EventSink<flutter::EncodableValue>> event_sink_;
};

AntiCheatBridge::AntiCheatBridge(flutter::BinaryMessenger* messenger, HWND window)
    : impl_(std::make_unique<Impl>(messenger, window)) {}

AntiCheatBridge::~AntiCheatBridge() = default;

void AntiCheatBridge::Register() {
  impl_->Register();
}

void AntiCheatBridge::ApplyExamWindowPolicy() {
  impl_->ApplyExamWindowPolicy();
}

void AntiCheatBridge::HandleWindowMessage(UINT message, WPARAM wparam, LPARAM lparam) {
  impl_->HandleWindowMessage(message, wparam, lparam);
}
