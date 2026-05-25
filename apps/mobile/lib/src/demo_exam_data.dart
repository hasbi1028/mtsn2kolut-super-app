import 'dart:async';

import 'exam_api.dart';
import 'exam_events.dart';
import 'models.dart';

const bool kCbtDemoModeEnabled = bool.fromEnvironment(
  'CBT_DEMO_MODE',
  defaultValue: false,
);

const String kDemoExamToken = 'demo-token-local';
const String kDemoRoomToken = 'DEMO-RUANG';
const String kDemoDeviceFingerprint = 'demo-device-local';

ExamLoginPayload buildDemoExamPayload() {
  final now = DateTime.now();
  return ExamLoginPayload(
    participantId: 'demo-participant-local',
    student: const ExamStudent(nis: 'DEMO-001', nama: 'Siswa Demo'),
    session: ExamSession(
      id: 'demo-session-local',
      title: 'Mode DEMO CBT — Data Contoh',
      scheduledStart: now.subtract(const Duration(minutes: 10)),
      scheduledEnd: now.add(const Duration(minutes: 90)),
      durationMinutes: 90,
    ),
    room: const ExamRoom(roomName: 'Ruang Demo'),
    questions: const [
      ExamQuestion(
        id: 'demo-q-1',
        questionType: 'multiple_choice',
        questionText:
            'Perhatikan pernyataan berikut. Manakah pilihan yang paling tepat untuk menggambarkan fungsi utama aplikasi CBT Flutter?',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [
          ExamOption(
            label: 'A',
            text: 'Mengerjakan ujian dengan token dan pemantauan perangkat',
          ),
          ExamOption(label: 'B', text: 'Mengubah jadwal ujian madrasah'),
          ExamOption(label: 'C', text: 'Mengelola data pegawai'),
          ExamOption(label: 'D', text: 'Mencetak kartu perpustakaan'),
        ],
      ),
      ExamQuestion(
        id: 'demo-q-2',
        questionType: 'multiple_answer',
        questionText:
            'Pilih semua perilaku yang dapat dipakai untuk menguji mode anti-cheat secara manual.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [
          ExamOption(label: 'A', text: 'Membuka recent apps lalu kembali'),
          ExamOption(label: 'B', text: 'Mencoba split-screen/floating window'),
          ExamOption(label: 'C', text: 'Menjawab soal sampai selesai'),
          ExamOption(label: 'D', text: 'Mematikan koneksi internet sebentar'),
        ],
      ),
      ExamQuestion(
        id: 'demo-q-3',
        questionType: 'true_false',
        questionText:
            'Mode DEMO memakai data contoh lokal dan tidak mengirim jawaban ke server produksi.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [],
      ),
      ExamQuestion(
        id: 'demo-q-4',
        questionType: 'short_answer',
        questionText:
            'Tuliskan satu hal yang perlu diperiksa panitia saat siswa memakai browser darurat.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [],
      ),
      ExamQuestion(
        id: 'demo-q-5',
        questionType: 'essay',
        questionText:
            'Jelaskan secara singkat perbedaan jalur utama Flutter APK dan jalur Browser Darurat untuk pelaksanaan CBT.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [],
      ),
    ],
    answeredCount: 0,
    totalQuestions: 5,
    timeRemainingSeconds: 90 * 60,
  );
}

class DemoExamApiClient extends ExamApiClient {
  DemoExamApiClient()
    : super(
        baseUrl: 'https://demo.invalid',
        deviceFingerprint: kDemoDeviceFingerprint,
      );

  final Map<String, String> answers = <String, String>{};
  bool submitted = false;

  @override
  Future<ExamLoginPayload> login({
    required String token,
    required String deviceFingerprint,
    String roomToken = '',
  }) async {
    return buildDemoExamPayload();
  }

  @override
  Future<ExamStatusPayload> getStatus(String token) async {
    return ExamStatusPayload(
      answeredCount: answers.values
          .where((answer) => answer.trim().isNotEmpty)
          .length,
      totalQuestions: buildDemoExamPayload().totalQuestions,
      timeRemainingSeconds: buildDemoExamPayload().timeRemainingSeconds,
      isSubmitted: submitted,
    );
  }

  @override
  Future<void> sendHeartbeat(String token) async {}

  @override
  Future<void> sendEvent({
    required String token,
    required String eventType,
    Map<String, Object?> data = const <String, Object?>{},
  }) async {}

  @override
  Future<void> sendExamEvent({
    required String token,
    required ExamClientEvent event,
  }) async {}

  @override
  Future<void> saveAnswer({
    required String token,
    required String questionId,
    required String answer,
  }) async {
    if (!ExamApiClient.isAnswerBodyWithinLimit(
      questionId: questionId,
      answer: answer,
    )) {
      throw const ExamApiException(
        'Jawaban demo terlalu panjang.',
        statusCode: 413,
      );
    }
    answers[questionId] = answer;
  }

  @override
  Future<void> submit(String token) async {
    submitted = true;
  }
}
