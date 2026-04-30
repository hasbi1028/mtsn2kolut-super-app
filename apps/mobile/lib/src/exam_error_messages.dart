import 'exam_api.dart';

enum ExamGuidanceTone { info, warning, danger }

class ExamGuidanceNotice {
  const ExamGuidanceNotice({
    required this.title,
    required this.message,
    required this.tone,
  });

  final String title;
  final String message;
  final ExamGuidanceTone tone;
}

String loginFailureMessage(ExamApiException error) {
  switch (error.statusCode) {
    case 404:
      return 'Token ujian tidak ditemukan. Periksa kembali token dari pengawas atau kartu ujian.';
    case 403:
      return 'Sesi ujian belum aktif atau sudah berakhir. Hubungi pengawas untuk memastikan jadwal sesi.';
    case 409:
      return 'Token ini sudah terhubung dengan perangkat lain. Gunakan perangkat yang sama atau minta bantuan pengawas.';
    default:
      return error.message;
  }
}

String restoreFailureMessage(ExamApiException error) {
  switch (error.statusCode) {
    case 404:
      return 'Token sesi lama sudah tidak ditemukan lagi di server. Login ulang dengan token aktif dari pengawas jika sesi masih berlangsung.';
    case 403:
      return 'Sesi lama tidak bisa dipulihkan karena ujian belum aktif lagi atau sudah ditutup. Periksa status sesi dengan pengawas.';
    case 409:
      return 'Sesi lama terikat ke perangkat lain. Gunakan perangkat yang sama seperti sebelumnya atau minta bantuan pengawas.';
    default:
      return 'Sesi lama tidak bisa dipulihkan. ${error.message}';
  }
}

String answerFailureMessage(ExamApiException error) {
  switch (error.statusCode) {
    case 403:
      return 'Waktu ujian sudah berakhir. Jawaban tetap disimpan di perangkat ini, tetapi pengawas perlu memastikan apakah sesi masih bisa dipulihkan.';
    case 409:
      return 'Ujian ini sudah dinyatakan selesai di server. Jawaban baru tidak bisa dikirim lagi.';
    default:
      return '${error.message} Jawaban tetap disimpan di perangkat dan akan dicoba sinkron ulang.';
  }
}

ExamGuidanceNotice? answerFailureNotice(ExamApiException error) {
  switch (error.statusCode) {
    case 403:
      return const ExamGuidanceNotice(
        title: 'Waktu ujian sudah berakhir',
        message:
            'Jawaban lokal masih aman di perangkat ini, tetapi pengawas perlu memastikan apakah sesi masih bisa dipulihkan atau harus diakhiri.',
        tone: ExamGuidanceTone.warning,
      );
    case 409:
      return const ExamGuidanceNotice(
        title: 'Ujian sudah selesai di server',
        message:
            'Perangkat ini tidak dapat mengirim jawaban baru lagi. Pengawas sebaiknya mengecek apakah submit sebelumnya sudah final.',
        tone: ExamGuidanceTone.danger,
      );
    default:
      return null;
  }
}

String submitFailureMessage(
  ExamApiException error, {
  required bool autoSubmit,
}) {
  switch (error.statusCode) {
    case 403:
      return autoSubmit
          ? 'Waktu ujian sudah habis, tetapi server belum menerima submit otomatis. Segera minta pengawas memeriksa koneksi dan status sesi.'
          : 'Waktu ujian sudah berakhir menurut server. Hubungi pengawas untuk memastikan status kirim ujian.';
    case 409:
      return 'Ujian ini sudah tercatat selesai di server. Tidak perlu menekan kirim lagi.';
    default:
      return error.message;
  }
}

ExamGuidanceNotice? submitFailureNotice(
  ExamApiException error, {
  required bool autoSubmit,
}) {
  switch (error.statusCode) {
    case 403:
      return ExamGuidanceNotice(
        title: autoSubmit
            ? 'Submit otomatis belum diterima server'
            : 'Waktu ujian sudah berakhir',
        message: autoSubmit
            ? 'Pengawas perlu segera memeriksa koneksi perangkat dan memastikan status sesi di server sebelum peserta meninggalkan ujian.'
            : 'Server menilai waktu sesi sudah selesai. Pengawas perlu memastikan apakah ujian perlu ditutup manual atau cukup diverifikasi.',
        tone: ExamGuidanceTone.warning,
      );
    case 409:
      return const ExamGuidanceNotice(
        title: 'Submit sudah tercatat',
        message:
            'Server sudah menganggap ujian ini selesai. Pengawas cukup memverifikasi status akhir, tidak perlu mengirim ulang.',
        tone: ExamGuidanceTone.info,
      );
    default:
      return null;
  }
}
