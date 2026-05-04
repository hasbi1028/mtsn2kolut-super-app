import {
  BACKEND_URL,
  normalizeRuntimeConfigPatch,
  WORKER_API_TIMEOUT_MS,
  WORKER_API_KEY,
  WORKER_ID,
} from './config.js';
import type {
  AttendanceRecord,
  ClaimedJob,
  RuntimeConfig,
} from './types.js';

const FAIL_REPORT_ATTEMPTS = 3;
const FAIL_REPORT_RETRY_MS = 300;
const ERROR_BODY_TIMEOUT_MS = 1000;
const ERROR_BODY_MAX_BYTES = 4096;

export class WorkerApiTimeoutError extends Error {
  constructor(operation: string) {
    super(`worker API request timeout: ${operation}`);
    this.name = 'WorkerApiTimeoutError';
  }
}

export class WorkerApiCancelledError extends Error {
  constructor(operation: string) {
    super(`worker API request cancelled: ${operation}`);
    this.name = 'WorkerApiCancelledError';
  }
}

function workerHeaders(): Record<string, string> {
  return {
    'content-type': 'application/json',
    'x-worker-key': WORKER_API_KEY,
  };
}

function timeoutSignal(ms: number, externalSignal?: AbortSignal): {
  signal: AbortSignal;
  cleanup: () => void;
  timedOut: () => boolean;
} {
  const controller = new AbortController();
  let didTimeout = false;
  const timeout = setTimeout(() => {
    didTimeout = true;
    controller.abort();
  }, ms);
  timeout.unref?.();

  const abortFromExternal = () => controller.abort();
  if (externalSignal) {
    if (externalSignal.aborted) {
      abortFromExternal();
    } else {
      externalSignal.addEventListener('abort', abortFromExternal, { once: true });
    }
  }

  return {
    signal: controller.signal,
    cleanup: () => {
      clearTimeout(timeout);
      externalSignal?.removeEventListener('abort', abortFromExternal);
    },
    timedOut: () => didTimeout,
  };
}

function isAbortError(error: unknown): boolean {
  const name = (error as Error | undefined)?.name;
  return name === 'AbortError' || name === 'TimeoutError';
}

async function workerFetch(operation: string, url: string, init: RequestInit): Promise<Response> {
  const externalSignal = init.signal ?? undefined;
  const timeout = timeoutSignal(WORKER_API_TIMEOUT_MS, externalSignal);
  try {
    return await fetch(url, {
      ...init,
      signal: timeout.signal,
    });
  } catch (error) {
    if (isAbortError(error)) {
      if (!timeout.timedOut() && externalSignal?.aborted) {
        throw new WorkerApiCancelledError(operation);
      }
      throw new WorkerApiTimeoutError(operation);
    }
    throw error;
  } finally {
    timeout.cleanup();
  }
}

async function readBoundedErrorBody(response: Response): Promise<string> {
  if (!response.body) {
    return '';
  }

  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  let truncated = false;
  const timeout = new Promise<Uint8Array[]>((_, reject) => {
    setTimeout(() => reject(new WorkerApiTimeoutError('read error body')), ERROR_BODY_TIMEOUT_MS).unref?.();
  });

  const read = (async () => {
    while (size < ERROR_BODY_MAX_BYTES) {
      const { value, done } = await reader.read();
      if (done) {
        break;
      }
      const remaining = ERROR_BODY_MAX_BYTES - size;
      const chunk = value.byteLength > remaining ? value.slice(0, remaining) : value;
      chunks.push(chunk);
      size += chunk.byteLength;
      if (value.byteLength > remaining) {
        truncated = true;
        break;
      }
    }
    if (size >= ERROR_BODY_MAX_BYTES) {
      truncated = true;
    }
    return chunks;
  })();

  try {
    const bodyChunks = await Promise.race([read, timeout]);
    const body = new TextDecoder().decode(Buffer.concat(bodyChunks));
    return truncated ? `${body}... [truncated]` : body;
  } catch (error) {
    if (error instanceof WorkerApiTimeoutError) {
      return '[error body read timeout]';
    }
    return '[error body unavailable]';
  } finally {
    reader.cancel().catch(() => {});
  }
}

async function responseError(prefix: string, response: Response): Promise<Error> {
  const body = await readBoundedErrorBody(response);
  return new Error(`${prefix}: ${response.status}${body ? ` ${body}` : ''}`);
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function claimJob(signal?: AbortSignal): Promise<ClaimedJob | null> {
  const response = await workerFetch(
    'claim job',
    `${BACKEND_URL}/api/pusaka/worker/claim`,
    {
      method: 'POST',
      headers: workerHeaders(),
      body: JSON.stringify({ worker_id: WORKER_ID }),
      signal,
    },
  );
  if (response.status === 204) {
    return null;
  }
  if (!response.ok) {
    throw await responseError('claim failed', response);
  }
  const payload = (await response.json()) as { data: ClaimedJob };
  return payload.data ?? null;
}

export async function fetchRuntimeConfig(): Promise<Partial<RuntimeConfig>> {
  const response = await workerFetch(
    'fetch runtime config',
    `${BACKEND_URL}/api/pusaka/worker/config`,
    {
      method: 'GET',
      headers: workerHeaders(),
    },
  );
  if (!response.ok) {
    throw await responseError('config fetch failed', response);
  }

  const payload = (await response.json()) as { data?: unknown };
  return normalizeRuntimeConfigPatch(payload.data);
}

export async function sendHeartbeat(args: {
  consumerCount: number;
  targetConcurrency: number;
  headless: boolean;
  lastSyncAt: string;
}): Promise<void> {
  const response = await workerFetch(
    'send heartbeat',
    `${BACKEND_URL}/api/pusaka/worker/heartbeat`,
    {
      method: 'POST',
      headers: workerHeaders(),
      body: JSON.stringify({
        worker_id: WORKER_ID,
        active_consumers: args.consumerCount,
        target_concurrency: args.targetConcurrency,
        headless: args.headless,
        last_sync_at: args.lastSyncAt,
      }),
    },
  );
  if (!response.ok) {
    throw await responseError('heartbeat failed', response);
  }
}

export async function completeJob(
  jobId: string,
  record: AttendanceRecord,
): Promise<void> {
  const response = await workerFetch(
    'complete job',
    `${BACKEND_URL}/api/pusaka/worker/jobs/${jobId}/complete`,
    {
      method: 'POST',
      headers: workerHeaders(),
      body: JSON.stringify(record),
    },
  );
  if (!response.ok) {
    throw await responseError('complete failed', response);
  }
}

export async function failJob(jobId: string, error: string): Promise<boolean> {
  let lastError: unknown;
  for (let attempt = 1; attempt <= FAIL_REPORT_ATTEMPTS; attempt++) {
    try {
      const response = await workerFetch(
        'fail job',
        `${BACKEND_URL}/api/pusaka/worker/jobs/${jobId}/fail`,
        {
          method: 'POST',
          headers: workerHeaders(),
          body: JSON.stringify({ error }),
        },
      );
      if (response.ok) {
        return true;
      }
      lastError = await responseError('fail report failed', response);
    } catch (fetchError) {
      lastError = fetchError;
    }

    if (attempt < FAIL_REPORT_ATTEMPTS) {
      await sleep(FAIL_REPORT_RETRY_MS * attempt);
    }
  }

  throw lastError instanceof Error
    ? lastError
    : new Error(`fail report failed: ${String(lastError)}`);
}
