import test from 'node:test';
import assert from 'node:assert/strict';

import { WorkerSupervisor } from './worker-supervisor.js';
import type { ClaimedJob, RuntimeConfig } from './types.js';

function createSupervisorHarness(overrides: Partial<any> = {}) {
  const runtimeConfig: RuntimeConfig = {
    maxConcurrent: 1,
    headless: true,
  };

  const calls = {
    claimJob: 0,
    completeJob: [] as Array<{ jobId: string; record: unknown }>,
    failJob: [] as Array<{ jobId: string; error: string }>,
    heartbeats: [] as Array<{
      consumerCount: number;
      targetConcurrency: number;
      headless: boolean;
      lastSyncAt: string;
    }>,
    logs: [] as Array<{ level: string; message: string; meta?: Record<string, unknown> }>,
    intervals: [] as Array<{ fn: () => void; ms: number }>,
    cleared: [] as unknown[],
    processOn: [] as string[],
    exitCodes: [] as number[],
    delayed: [] as number[],
  };

  let now = 1_700_000_000_000;
  const claimQueue: Array<ClaimedJob | null> = [];
  const configQueue: Array<Partial<RuntimeConfig>> = [];

  const deps = {
    claimJob: async () => {
      calls.claimJob += 1;
      return claimQueue.length > 0 ? claimQueue.shift() ?? null : null;
    },
    completeJob: async (jobId: string, record: unknown) => {
      calls.completeJob.push({ jobId, record });
    },
    failJob: async (jobId: string, error: string) => {
      calls.failJob.push({ jobId, error });
      return true;
    },
    fetchRuntimeConfig: async () => (configQueue.length > 0 ? configQueue.shift() ?? {} : {}),
    sendHeartbeat: async (payload: {
      consumerCount: number;
      targetConcurrency: number;
      headless: boolean;
      lastSyncAt: string;
    }) => {
      calls.heartbeats.push(payload);
    },
    processClaimedJob: async () => ({ tanggal: '2026-05-01', jam_masuk: '07:00', jam_pulang: '16:00' }),
    log: (level: string, message: string, meta?: Record<string, unknown>) => {
      calls.logs.push({ level, message, meta });
    },
    delay: async (ms: number) => {
      calls.delayed.push(ms);
    },
    setInterval: (fn: () => void, ms: number) => {
      calls.intervals.push({ fn, ms });
      return { fn, ms } as unknown as NodeJS.Timeout;
    },
    clearInterval: (timer: unknown) => {
      calls.cleared.push(timer);
    },
    now: () => now,
    processOn: (signal: string, _listener: () => void) => {
      calls.processOn.push(signal);
      return {} as any;
    },
    processExit: (code: number) => {
      calls.exitCodes.push(code);
      return undefined as never;
    },
    ...overrides,
  };

  const supervisor = new WorkerSupervisor(runtimeConfig, deps as any);

  return {
    supervisor,
    runtimeConfig,
    calls,
    claimQueue,
    configQueue,
    setNow(value: number) {
      now = value;
    },
  };
}

test('WorkerSupervisor syncRuntimeConfig updates runtime config and timestamp', async () => {
  const h = createSupervisorHarness();
  h.configQueue.push({ maxConcurrent: 3, headless: false });
  (h.supervisor as any).reconcileConsumers = () => {};

  await (h.supervisor as any).syncRuntimeConfig();

  assert.equal(h.runtimeConfig.maxConcurrent, 3);
  assert.equal(h.runtimeConfig.headless, false);
  assert.match((h.supervisor as any).lastConfigSyncAt, /^\d{4}-\d{2}-\d{2}T/);
});

test('WorkerSupervisor syncRuntimeConfig does not mutate invalid concurrency', async () => {
  const h = createSupervisorHarness();
  h.configQueue.push({ maxConcurrent: 1000, headless: false });
  (h.supervisor as any).reconcileConsumers = () => {};

  await (h.supervisor as any).syncRuntimeConfig();

  assert.equal(h.runtimeConfig.maxConcurrent, 1);
  assert.equal(h.runtimeConfig.headless, false);
});

test('WorkerSupervisor reconcileConsumers marks extra consumers for stop when concurrency shrinks', () => {
  const h = createSupervisorHarness();
  const consumers = (h.supervisor as any).consumers as Map<
    number,
    { id: number; consumerId: string; stopRequested: boolean }
  >;
  consumers.set(1, { id: 1, consumerId: 'w-c1', stopRequested: false });
  consumers.set(2, { id: 2, consumerId: 'w-c2', stopRequested: false });
  consumers.set(3, { id: 3, consumerId: 'w-c3', stopRequested: false });
  h.runtimeConfig.maxConcurrent = 1;

  (h.supervisor as any).reconcileConsumers();

  assert.equal(consumers.get(3)?.stopRequested, true);
  assert.equal(consumers.get(2)?.stopRequested, true);
  assert.equal(consumers.get(1)?.stopRequested, false);
});

test('WorkerSupervisor heartbeatLoop reports only non-stopped consumers', async () => {
  const h = createSupervisorHarness();
  const consumers = (h.supervisor as any).consumers as Map<
    number,
    { id: number; consumerId: string; stopRequested: boolean }
  >;
  consumers.set(1, { id: 1, consumerId: 'w-c1', stopRequested: false });
  consumers.set(2, { id: 2, consumerId: 'w-c2', stopRequested: true });
  h.runtimeConfig.maxConcurrent = 4;
  (h.supervisor as any).lastConfigSyncAt = '2026-05-01T12:00:00.000Z';

  await (h.supervisor as any).heartbeatLoop();

  assert.deepEqual(h.calls.heartbeats[0], {
    consumerCount: 1,
    targetConcurrency: 4,
    headless: true,
    lastSyncAt: '2026-05-01T12:00:00.000Z',
  });
});

test('WorkerSupervisor gracefulShutdown stops consumers, sends zero heartbeat, and exits', async () => {
  const h = createSupervisorHarness();
  const consumers = (h.supervisor as any).consumers as Map<
    number,
    { id: number; consumerId: string; stopRequested: boolean }
  >;
  consumers.set(1, { id: 1, consumerId: 'w-c1', stopRequested: false });
  (h.supervisor as any).configTimer = { kind: 'config' };
  (h.supervisor as any).heartbeatTimer = { kind: 'heartbeat' };
  h.setNow(10_000);
  let tick = 0;
  (h.supervisor as any).deps.now = () => {
    tick += 1;
    if (tick === 2) {
      consumers.clear();
    }
    return 10_000 + tick * 500;
  };

  await (h.supervisor as any).gracefulShutdown('SIGTERM');

  assert.equal(consumers.size, 0);
  assert.equal(h.calls.cleared.length, 2);
  assert.equal(h.calls.heartbeats.at(-1)?.consumerCount, 0);
  assert.deepEqual(h.calls.exitCodes, [0]);
});

test('WorkerSupervisor reports active claimed jobs when shutdown deadline expires', async () => {
  const h = createSupervisorHarness();
  const consumers = (h.supervisor as any).consumers as Map<number, any>;
  consumers.set(1, {
    id: 1,
    consumerId: 'w-c1',
    stopRequested: false,
    activeJobControllers: new Map(),
    activeJobIds: new Set(['job-active-1', 'job-active-2']),
  });
  h.setNow(10_000);
  let tick = 0;
  (h.supervisor as any).deps.now = () => {
    tick += 1;
    return tick === 1 ? 10_000 : 30_000;
  };

  await (h.supervisor as any).gracefulShutdown('SIGTERM');

  assert.deepEqual(h.calls.failJob, [
    { jobId: 'job-active-1', error: 'worker shutdown timeout' },
    { jobId: 'job-active-2', error: 'worker shutdown timeout' },
  ]);
  assert.equal(
    h.calls.logs.some((entry) => entry.message === 'shutdown timeout reached'),
    true,
  );
});

test('WorkerSupervisor aborts active job before shutdown timeout fail report', async () => {
  let activeSignal: AbortSignal | undefined;
  let releaseJob: (() => void) | undefined;
  const h = createSupervisorHarness({
    processClaimedJob: async (_job: ClaimedJob, _config: RuntimeConfig, signal?: AbortSignal) => {
      activeSignal = signal;
      await new Promise<void>((resolve) => {
        releaseJob = resolve;
      });
      throw new Error(signal?.aborted ? 'cancelled after abort' : 'not cancelled');
    },
  });
  const consumers = (h.supervisor as any).consumers as Map<number, any>;
  const state = {
    id: 1,
    consumerId: 'w-c1',
    stopRequested: false,
    activeJobIds: new Set<string>(),
    activeJobControllers: new Map<string, AbortController>(),
  };
  consumers.set(1, state);
  h.claimQueue.push({
    id: 'job-active-shutdown',
    employee_id: 'employee-1',
    run_type: 'checkin',
    attempts: 1,
    max_attempts: 3,
    pusaka_username: 'user-1',
    pusaka_password: 'secret',
  });

  const loop = (h.supervisor as any).consumerLoop(state);
  while (!activeSignal) {
    await Promise.resolve();
  }

  h.setNow(10_000);
  let tick = 0;
  (h.supervisor as any).deps.now = () => {
    tick += 1;
    return tick === 1 ? 10_000 : 30_000;
  };

  await (h.supervisor as any).gracefulShutdown('SIGTERM');
  assert.equal(activeSignal.aborted, true);
  assert.deepEqual(h.calls.failJob, [
    { jobId: 'job-active-shutdown', error: 'worker shutdown timeout' },
  ]);

  releaseJob?.();
  await loop;
  assert.deepEqual(h.calls.failJob, [
    { jobId: 'job-active-shutdown', error: 'worker shutdown timeout' },
  ]);
});
