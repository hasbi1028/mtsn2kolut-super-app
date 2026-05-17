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

  final normalizedMessage = error.message.trim().toLowerCase();
  switch (error.statusCode) {
    case 404:
      return 'Token Ujian tidak ditemukan. Periksa kembali Token Ujian dari kartu peserta atau minta bantuan pengawas.';
    case 400:
      if (normalizedMessage.contains('room token')) {
        return 'Token Ruang wajib diisi. Minta Token Ruang kepada pengawas.';
      }
      return error.message;
    case 403:
      if (normalizedMessage.contains('room token')) {
        return 'Token Ruang tidak sesuai. Pastikan Anda berada di ruang ujian yang benar dan minta pengawas memeriksa Token Ruang.';
      }
      return 'Sesi ujian belum aktif atau sudah berakhir. Hubungi pengawas untuk memastikan jadwal sesi.';
    case 409:
      return 'Token Ujian ini sudah terhubung dengan perangkat lain. Gunakan perangkat yang sama atau minta bantuan pengawas.';
    case 423:
      return 'Akses ujian sedang dikunci oleh pengawas atau sistem keamanan. Tetap di tempat dan tunggu pengawas membuka akses kembali.';
    default:
      return error.message.isNotEmpty
          ? error.message
          : 'Login belum berhasil. Tunjukkan layar ini kepada pengawas agar dapat diperiksa.';
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
        title: 'Token Ujian terhubung ke perangkat lain',
        message:
            'Jangan terus mencoba login dari perangkat ini. Gunakan perangkat yang sama seperti sebelumnya atau minta pengawas memeriksa Token Ujian.',
        tone: ExamGuidanceTone.danger,
      );
    case 423:
      return const ExamGuidanceNotice(
        title: 'Akses dikunci pengawas',
        message:
            'Peserta tidak perlu mencoba berulang-ulang. Pengawas perlu memeriksa panel ruang dan membuka akses jika sudah dinyatakan aman.',
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
      return 'Sesi lama belum ditemukan. Login ulang dengan Token Ujian aktif dari pengawas jika sesi masih berlangsung.';
    case 403:
      return 'Sesi lama tidak bisa dipulihkan karena ujian belum aktif lagi atau sudah ditutup. Periksa status sesi dengan pengawas.';
    case 409:
      return 'Sesi lama terikat ke perangkat lain. Gunakan perangkat yang sama seperti sebelumnya atau minta bantuan pengawas.';
    case 423:
      return 'Sesi lama belum bisa dipulihkan karena akses peserta sedang dikunci. Minta pengawas memeriksa status peserta di panel ruang.';
    default:
      return 'Sesi lama tidak bisa dipulihkan. ${error.message}';
  }
}

ExamGuidanceNotice? restoreFailureNotice(ExamApiException error) {
  if (error.statusCode == null) {
    return const ExamGuidanceNotice(
      title: 'Pulihkan sesi tertunda karena koneksi',
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
            'Peserta tidak perlu terus mencoba memulihkan sesi di perangkat ini. Pengawas sebaiknya mengarahkan peserta kembali ke perangkat awal atau memeriksa Token Ujian.',
        tone: ExamGuidanceTone.danger,
      );
    case 423:
      return const ExamGuidanceNotice(
        title: 'Pulihkan sesi ditahan karena akses terkunci',
        message:
            'Pengawas perlu memeriksa alasan penguncian terlebih dahulu sebelum peserta mencoba masuk kembali.',
        tone: ExamGuidanceTone.danger,
      );
    default:
      return null;
  }
}

bool shouldClearSnapshotAfterRestoreFailure(ExamApiException error) {
  return error.statusCode == 403 || error.statusCode == 404;
}

String statusFailureMessage(ExamApiException error) {
  if (error.statusCode == null) {
    return 'Status server belum bisa diperbarui karena perangkat belum terhubung. Tetap di layar ujian dan minta pengawas memeriksa koneksi.';
  }

  switch (error.statusCode) {
    case 401:
      return 'Sesi perangkat perlu diperiksa. Minta pengawas memeriksa Token Ujian dan perangkat sebelum melanjutkan.';
    case 403:
      return 'Sesi ujian tidak lagi aktif menurut server. Tunggu arahan pengawas sebelum melanjutkan.';
    case 409:
      return 'Token Ujian ini terdeteksi aktif di perangkat lain. Jangan lanjutkan dari perangkat ini sebelum pengawas memeriksa.';
    case 423:
      return 'Akses peserta sedang dikunci. Tetap di layar ini dan tunggu pengawas memeriksa status ruang.';
    default:
      return 'Status server belum bisa diperbarui. ${error.message}';
  }
}

ExamGuidanceNotice? statusFailureNotice(ExamApiException error) {
  if (error.statusCode == null) {
    return const ExamGuidanceNotice(
      title: 'Status belum tersinkron',
      message:
          'Pengawas perlu memastikan perangkat kembali menjangkau server sebelum peserta keluar dari mode aman atau mengirim ujian.',
      tone: ExamGuidanceTone.warning,
    );
  }

  switch (error.statusCode) {
    case 401:
      return const ExamGuidanceNotice(
        title: 'Konteks peserta tidak sah',
        message:
            'Data sesi peserta belum cocok. Pengawas sebaiknya memeriksa Token Ujian, perangkat yang dipakai, dan status buka akses.',
        tone: ExamGuidanceTone.danger,
      );
    case 403:
      return const ExamGuidanceNotice(
        title: 'Sesi tidak aktif',
        message:
            'Server menolak status karena sesi belum aktif atau sudah ditutup. Peserta sebaiknya menunggu keputusan pengawas.',
        tone: ExamGuidanceTone.warning,
      );
    case 409:
      return const ExamGuidanceNotice(
        title: 'Perangkat berbeda terdeteksi',
        message:
            'Jangan lanjutkan ujian dari perangkat ini sebelum pengawas memastikan apakah akses perlu dibuka ulang atau peserta kembali ke perangkat awal.',
        tone: ExamGuidanceTone.danger,
      );
    case 423:
      return const ExamGuidanceNotice(
        title: 'Peserta terkunci',
        message:
            'Server menahan akses peserta. Pengawas perlu membuka kunci dari panel ruang jika peserta boleh melanjutkan.',
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
    case 401:
      return 'Sesi perangkat perlu diperiksa sebelum mengirim jawaban. Jawaban tetap disimpan di perangkat sambil menunggu pemeriksaan pengawas.';
    case 403:
      return 'Waktu ujian sudah berakhir. Jawaban tetap disimpan di perangkat ini, tetapi pengawas perlu memastikan apakah sesi masih bisa dipulihkan.';
    case 409:
      return 'Ujian ini sudah dinyatakan selesai di server. Jawaban baru tidak bisa dikirim lagi.';
    case 423:
      return 'Jawaban belum dikirim karena akses peserta sedang dikunci. Jawaban tetap tersimpan di perangkat sampai pengawas membuka akses.';
    default:
      return '${error.message} Jawaban tetap disimpan di perangkat dan akan dicoba sinkron ulang.';
  }
}

ExamGuidanceNotice? answerFailureNotice(ExamApiException error) {
  if (error.statusCode == null) {
    return const ExamGuidanceNotice(
      title: 'Jawaban tersimpan di perangkat',
      message:
          'Perangkat belum bisa menjangkau server, tetapi jawaban peserta masih aman di perangkat ini. Pengawas perlu membantu memulihkan koneksi sebelum sinkron ulang.',
      tone: ExamGuidanceTone.warning,
    );
  }

  switch (error.statusCode) {
    case 401:
      return const ExamGuidanceNotice(
        title: 'Sesi perangkat perlu diverifikasi',
        message:
            'Jawaban tetap disimpan di perangkat, tetapi data sesi peserta belum cocok. Pengawas perlu memeriksa Token Ujian dan perangkat.',
        tone: ExamGuidanceTone.danger,
      );
    case 403:
      return const ExamGuidanceNotice(
        title: 'Waktu ujian sudah berakhir',
        message:
            'Jawaban masih aman di perangkat ini, tetapi pengawas perlu memastikan apakah sesi masih bisa dipulihkan atau harus diakhiri.',
        tone: ExamGuidanceTone.warning,
      );
    case 409:
      return const ExamGuidanceNotice(
        title: 'Ujian sudah selesai di server',
        message:
            'Perangkat ini tidak dapat mengirim jawaban baru lagi. Pengawas sebaiknya mengecek apakah jawaban akhir sebelumnya sudah tercatat.',
        tone: ExamGuidanceTone.danger,
      );
    case 423:
      return const ExamGuidanceNotice(
        title: 'Sinkron jawaban ditahan',
        message:
            'Jawaban tetap aman di perangkat ini. Pengawas perlu menyelesaikan status penguncian sebelum perangkat mencoba sinkron ulang.',
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
        ? 'Kirim otomatis belum berhasil karena perangkat kehilangan koneksi ke server ujian. Jawaban tetap aman; segera minta pengawas memeriksa jaringan.'
        : 'Perangkat belum bisa terhubung ke server ujian. Jangan tinggalkan layar ini sebelum pengawas memastikan koneksi kembali.';
  }

  switch (error.statusCode) {
    case 401:
      return 'Sesi perangkat perlu diperiksa sebelum jawaban akhir dikirim. Tetap di layar ini dan minta pengawas memeriksa Token Ujian atau membuka akses ulang.';
    case 403:
      return autoSubmit
          ? 'Waktu ujian sudah habis, tetapi server belum menerima kiriman otomatis. Jawaban tetap aman; segera minta pengawas memeriksa koneksi dan status sesi.'
          : 'Waktu ujian sudah berakhir menurut server. Hubungi pengawas untuk memastikan status kirim ujian.';
    case 409:
      return 'Ujian ini sudah tercatat selesai di server. Tidak perlu menekan kirim lagi.';
    case 423:
      return autoSubmit
          ? 'Kirim otomatis ditahan karena akses peserta sedang dikunci. Segera minta pengawas memeriksa panel ruang.'
          : 'Jawaban akhir belum bisa dikirim karena akses peserta sedang dikunci. Tunggu pengawas membuka akses atau memberi instruksi.';
    default:
      return error.message.isNotEmpty
          ? error.message
          : 'Jawaban akhir belum bisa dikirim. Tunjukkan layar ini kepada pengawas untuk diperiksa.';
  }
}

ExamGuidanceNotice? submitFailureNotice(
  ExamApiException error, {
  required bool autoSubmit,
}) {
  if (error.statusCode == null) {
    return ExamGuidanceNotice(
      title: autoSubmit
          ? 'Kirim otomatis tertunda karena koneksi'
          : 'Jawaban akhir belum bisa dikirim',
      message: autoSubmit
          ? 'Pengawas perlu segera memeriksa jaringan perangkat dan memastikan server dapat dijangkau sebelum peserta meninggalkan sesi.'
          : 'Koneksi ke server ujian belum tersedia. Pengawas perlu membantu memulihkan jaringan sebelum peserta menekan kirim lagi.',
      tone: ExamGuidanceTone.warning,
    );
  }

  switch (error.statusCode) {
    case 401:
      return const ExamGuidanceNotice(
        title: 'Kirim jawaban akhir ditahan',
        message:
            'Data sesi peserta belum cocok. Pengawas perlu memeriksa Token Ujian, perangkat, atau membuka akses ulang sebelum peserta mencoba lagi.',
        tone: ExamGuidanceTone.danger,
      );
    case 403:
      return ExamGuidanceNotice(
        title: autoSubmit
            ? 'Kirim otomatis belum diterima server'
            : 'Waktu ujian sudah berakhir',
        message: autoSubmit
            ? 'Pengawas perlu segera memeriksa koneksi perangkat dan memastikan status sesi di server sebelum peserta meninggalkan ujian.'
            : 'Server menilai waktu sesi sudah selesai. Pengawas perlu memastikan apakah ujian perlu ditutup manual atau cukup diverifikasi.',
        tone: ExamGuidanceTone.warning,
      );
    case 409:
      return const ExamGuidanceNotice(
        title: 'Jawaban akhir sudah tercatat',
        message:
            'Server sudah menganggap ujian ini selesai. Pengawas cukup memverifikasi status akhir, tidak perlu mengirim ulang.',
        tone: ExamGuidanceTone.info,
      );
    case 423:
      return const ExamGuidanceNotice(
        title: 'Kirim jawaban akhir ditahan karena akses terkunci',
        message:
            'Pengawas perlu membuka kunci peserta atau menutup sesi secara resmi sebelum peserta meninggalkan ruang.',
        tone: ExamGuidanceTone.danger,
      );
    default:
      return null;
  }
}
