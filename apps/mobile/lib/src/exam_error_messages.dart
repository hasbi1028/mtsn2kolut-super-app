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
  if (error.statusCode == null) {
    return 'Perangkat belum bisa terhubung ke server ujian. Periksa alamat server dan koneksi yang sedang dipakai.';
  }

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

ExamGuidanceNotice? loginFailureNotice(ExamApiException error) {
  if (error.statusCode == null) {
    return const ExamGuidanceNotice(
      title: 'Server ujian belum terjangkau',
      message:
          'Peserta tidak perlu terus menekan login. Periksa koneksi perangkat atau alamat server, lalu coba lagi setelah pengawas memastikan jaringan siap.',
      tone: ExamGuidanceTone.warning,
    );
  }

  switch (error.statusCode) {
    case 403:
      return const ExamGuidanceNotice(
        title: 'Sesi belum bisa dimasuki',
        message:
            'Peserta sebaiknya menunggu arahan pengawas. Login ulang hanya perlu dilakukan setelah jadwal sesi dipastikan aktif.',
        tone: ExamGuidanceTone.warning,
      );
    case 409:
      return const ExamGuidanceNotice(
        title: 'Token sudah terikat ke perangkat lain',
        message:
            'Jangan terus mencoba login dari perangkat ini. Gunakan perangkat yang sama seperti sebelumnya atau minta pengawas memverifikasi token.',
        tone: ExamGuidanceTone.danger,
      );
    default:
      return null;
  }
}

String restoreFailureMessage(ExamApiException error) {
  if (error.statusCode == null) {
    return 'Sesi lama belum bisa dipulihkan karena perangkat belum terhubung ke server ujian. Coba lagi setelah koneksi membaik atau hubungi pengawas.';
  }

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

ExamGuidanceNotice? restoreFailureNotice(ExamApiException error) {
  if (error.statusCode == null) {
    return const ExamGuidanceNotice(
      title: 'Restore tertunda karena koneksi',
      message:
          'Pengawas perlu memastikan perangkat sudah kembali terhubung ke server sebelum peserta mencoba memulihkan sesi lama lagi.',
      tone: ExamGuidanceTone.warning,
    );
  }

  switch (error.statusCode) {
    case 403:
      return const ExamGuidanceNotice(
        title: 'Sesi lama belum bisa dipulihkan',
        message:
            'Pengawas perlu memastikan apakah ujian memang belum aktif lagi atau sudah resmi ditutup sebelum peserta mencoba masuk ulang.',
        tone: ExamGuidanceTone.warning,
      );
    case 409:
      return const ExamGuidanceNotice(
        title: 'Sesi lama aktif di perangkat lain',
        message:
            'Peserta tidak perlu terus mencoba restore di perangkat ini. Pengawas sebaiknya mengarahkan peserta kembali ke perangkat awal atau memeriksa status token.',
        tone: ExamGuidanceTone.danger,
      );
    default:
      return null;
  }
}

String answerFailureMessage(ExamApiException error) {
  if (error.statusCode == null) {
    return 'Perangkat sedang kehilangan koneksi ke server ujian. Jawaban tetap disimpan di perangkat dan akan dicoba sinkron ulang.';
  }

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
  if (error.statusCode == null) {
    return const ExamGuidanceNotice(
      title: 'Jawaban tersimpan lokal',
      message:
          'Perangkat belum bisa menjangkau server, tetapi jawaban peserta masih aman di perangkat ini. Pengawas perlu membantu memulihkan koneksi sebelum sinkron ulang.',
      tone: ExamGuidanceTone.warning,
    );
  }

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
  if (error.statusCode == null) {
    return autoSubmit
        ? 'Submit otomatis belum bisa dikirim karena perangkat kehilangan koneksi ke server ujian. Segera minta pengawas memeriksa jaringan.'
        : 'Perangkat belum bisa terhubung ke server ujian. Jangan tinggalkan layar ini sebelum pengawas memastikan koneksi kembali.';
  }

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
  if (error.statusCode == null) {
    return ExamGuidanceNotice(
      title: autoSubmit
          ? 'Submit otomatis tertunda karena koneksi'
          : 'Submit belum bisa dikirim',
      message: autoSubmit
          ? 'Pengawas perlu segera memeriksa jaringan perangkat dan memastikan server dapat dijangkau sebelum peserta meninggalkan sesi.'
          : 'Koneksi ke server ujian belum tersedia. Pengawas perlu membantu memulihkan jaringan sebelum peserta menekan kirim lagi.',
      tone: ExamGuidanceTone.warning,
    );
  }

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
