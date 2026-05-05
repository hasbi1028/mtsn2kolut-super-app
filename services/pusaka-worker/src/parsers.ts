import type { AttendanceRecord, GeoCoords } from './types.js';

export function randomGeo(
  baseLat: number,
  baseLng: number,
  radiusMeters = 50,
): GeoCoords {
  const earthRadius = 6371000;
  const latOffset =
    (Math.random() - 0.5) * 2 * (radiusMeters / earthRadius) * (180 / Math.PI);
  const lngOffset =
    ((Math.random() - 0.5) *
      2 *
      (radiusMeters / earthRadius) *
      (180 / Math.PI)) /
    Math.cos((baseLat * Math.PI) / 180);
  return {
    latitude: baseLat + latOffset,
    longitude: baseLng + lngOffset,
  };
}

export function parseTodayFromText(
  text: string,
  todayLabel: string,
): AttendanceRecord | null {
  const lines = text
    .replace(/\r/g, '')
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean);
  const candidates = lines.reduce<number[]>((accumulator, line, index) => {
    if (line === todayLabel) {
      accumulator.push(index);
    }
    return accumulator;
  }, []);
  if (!candidates.length) {
    return null;
  }

  const blockLines: string[] = [];
  const index = candidates[candidates.length - 1];
  for (let cursor = index + 1; cursor < lines.length; cursor++) {
    if (
      /^(Senin|Selasa|Rabu|Kamis|Jumat|Sabtu|Minggu),\s+\d{2}\s+\w+\s+\d{4}$/.test(
        lines[cursor],
      )
    ) {
      break;
    }
    blockLines.push(lines[cursor]);
  }

  const block = blockLines.join('\n');
  const jamMasuk = extractJam(block, 'Jam Masuk');
  const jamPulang = extractJam(block, 'Jam Pulang');

  return {
    tanggal: toISODateMakassar(),
    jam_masuk: jamMasuk === '-' ? '' : jamMasuk,
    jam_pulang: jamPulang === '-' ? '' : jamPulang,
  };
}

export function extractJam(block: string, label: string): string {
  const lines = block
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean);
  for (let index = 0; index < lines.length; index++) {
    if (!lines[index].toLowerCase().includes(label.toLowerCase())) {
      continue;
    }
    const inline = lines[index].match(
      /([0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*(?:WITA|WIB|WIT))?|-)/i,
    );
    if (inline) {
      return inline[1].replace(/\s+/g, ' ').trim();
    }
    for (
      let nextIndex = index + 1;
      nextIndex < Math.min(lines.length, index + 4);
      nextIndex++
    ) {
      const next = lines[nextIndex].match(
        /([0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*(?:WITA|WIB|WIT))?|-)/i,
      );
      if (next) {
        return next[1].replace(/\s+/g, ' ').trim();
      }
    }
  }
  return '';
}

export function getTodayLabelID(): string {
  return new Intl.DateTimeFormat('id-ID', {
    weekday: 'long',
    day: '2-digit',
    month: 'long',
    year: 'numeric',
    timeZone: 'Asia/Makassar',
  }).format(new Date());
}

export function toISODateMakassar(): string {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Makassar',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date());
}

export function toTimeWITA(): string {
  return new Intl.DateTimeFormat('id-ID', {
    timeZone: 'Asia/Makassar',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date());
}
