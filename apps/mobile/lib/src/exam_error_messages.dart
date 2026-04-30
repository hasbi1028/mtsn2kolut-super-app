import 'exam_api.dart';

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
