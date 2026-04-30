import {
  BACKEND_URL,
  WORKER_API_KEY,
  WORKER_ID,
} from './config.js';
import { log } from './logger.js';
import type {
  AttendanceRecord,
  ClaimedJob,
  RuntimeConfig,
} from './types.js';

function workerHeaders(): Record<string, string> {
  return {
    'content-type': 'application/json',
    'x-worker-key': WORKER_API_KEY,
  };
}

export async function claimJob(): Promise<ClaimedJob | null> {
  const response = await fetch(`${BACKEND_URL}/api/pusaka/worker/claim`, {
    method: 'POST',
    headers: workerHeaders(),
    body: JSON.stringify({ worker_id: WORKER_ID }),
  });
  if (response.status === 204) {
    return null;
  }
  if (!response.ok) {
    throw new Error(`claim failed: ${response.status} ${await response.text()}`);
  }
  const payload = (await response.json()) as { data: ClaimedJob };
  return payload.data ?? null;
}

export async function fetchRuntimeConfig(): Promise<Partial<RuntimeConfig>> {
  const response = await fetch(`${BACKEND_URL}/api/pusaka/worker/config`, {
    method: 'GET',
    headers: workerHeaders(),
  });
  if (!response.ok) {
    throw new Error(
      `config fetch failed: ${response.status} ${await response.text()}`,
    );
  }

  const payload = (await response.json()) as {
    data?: { max_concurrent?: string; headless?: string };
  };
  const data = payload.data ?? {};
  const next: Partial<RuntimeConfig> = {};

  if (data.max_concurrent) {
    const parsed = Number(data.max_concurrent);
    if (Number.isFinite(parsed) && parsed >= 1) {
      next.maxConcurrent = Math.floor(parsed);
    }
  }
  if (data.headless) {
    next.headless = ['1', 'true'].includes(data.headless.toLowerCase());
  }

  return next;
}

export async function sendHeartbeat(args: {
  consumerCount: number;
  targetConcurrency: number;
  headless: boolean;
  lastSyncAt: string;
}): Promise<void> {
  const response = await fetch(`${BACKEND_URL}/api/pusaka/worker/heartbeat`, {
    method: 'POST',
    headers: workerHeaders(),
    body: JSON.stringify({
      worker_id: WORKER_ID,
      active_consumers: args.consumerCount,
      target_concurrency: args.targetConcurrency,
      headless: args.headless,
      last_sync_at: args.lastSyncAt,
    }),
  });
  if (!response.ok) {
    throw new Error(
      `heartbeat failed: ${response.status} ${await response.text()}`,
    );
  }
}

export async function completeJob(
  jobId: string,
  record: AttendanceRecord,
): Promise<void> {
  const response = await fetch(
    `${BACKEND_URL}/api/pusaka/worker/jobs/${jobId}/complete`,
    {
      method: 'POST',
      headers: workerHeaders(),
      body: JSON.stringify(record),
    },
  );
  if (!response.ok) {
    throw new Error(
      `complete failed: ${response.status} ${await response.text()}`,
    );
  }
}

export async function failJob(jobId: string, error: string): Promise<void> {
  const response = await fetch(
    `${BACKEND_URL}/api/pusaka/worker/jobs/${jobId}/fail`,
    {
      method: 'POST',
      headers: workerHeaders(),
      body: JSON.stringify({ error }),
    },
  );
  if (!response.ok) {
    log('WARN', 'fail report error', { jobId, status: response.status });
  }
}
