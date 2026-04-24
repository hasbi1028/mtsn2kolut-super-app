import { enqueueAllActive } from '$lib/server/queue';
import { logError, logInfo } from '$lib/server/logger';
import { db } from '$lib/server/db';

const TZ = 'Asia/Makassar';

function nowMakassar() {
  const d = new Date();
  const date = new Intl.DateTimeFormat('en-CA', {
    timeZone: TZ,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).format(d);
  const time = new Intl.DateTimeFormat('en-GB', {
    timeZone: TZ,
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  }).format(d);
  return { date, time };
}

export function schedulerTick() {
  const { date, time } = nowMakassar();
  const schedules = db.prepare('SELECT * FROM schedules WHERE is_enabled = 1').all();

  if (!globalThis.__scheduleRunCache) globalThis.__scheduleRunCache = new Set();

  let totalEnqueued = 0;
  for (const s of schedules) {
    if (s.run_time !== time) continue;
    const key = `${date}:${s.id}:${s.run_time}`;
    if (globalThis.__scheduleRunCache.has(key)) continue;

    const result = enqueueAllActive(s.run_type, 3);
    totalEnqueued += result.inserted;
    globalThis.__scheduleRunCache.add(key);
    logInfo('scheduler enqueued jobs', { schedule_id: s.id, run_type: s.run_type, ...result, at: `${date} ${time}` });
  }

  return totalEnqueued;
}

export function initScheduler() {
  if (globalThis.__schedulerInitialized) return;
  globalThis.__schedulerInitialized = true;

  setInterval(() => {
    try {
      schedulerTick();
    } catch (e) {
      logError('scheduler tick failed', { error: e?.message || String(e) });
    }
  }, 30_000);

  logInfo('scheduler initialized');
}
