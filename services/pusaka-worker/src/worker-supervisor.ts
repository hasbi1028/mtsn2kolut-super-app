import process from 'process';
import { setTimeout as delay } from 'timers/promises';

import {
  CONFIG_SYNC_MS,
  normalizeRuntimeConfigPatch,
  POLL_MS,
  WORKER_CONCURRENCY_CAP,
  WORKER_HEARTBEAT_MS,
  WORKER_ID,
} from './config.js';
import {
  claimJob,
  completeJob,
  failJob,
  fetchRuntimeConfig,
  sendHeartbeat,
  WorkerApiCancelledError,
  WorkerApiFinalStateError,
  WorkerApiUncertainCompletionError,
} from './api-client.js';
import { log } from './logger.js';
import { processClaimedJob } from './pusaka-runner.js';
import type { WorkerHealthSnapshot } from './health-server.js';
import type { ConsumerState, RuntimeConfig } from './types.js';

type WorkerSupervisorDeps = {
  claimJob: typeof claimJob;
  completeJob: typeof completeJob;
  failJob: typeof failJob;
  fetchRuntimeConfig: typeof fetchRuntimeConfig;
  sendHeartbeat: typeof sendHeartbeat;
  processClaimedJob: typeof processClaimedJob;
  log: typeof log;
  delay: typeof delay;
  setInterval: typeof globalThis.setInterval;
  clearInterval: typeof globalThis.clearInterval;
  now: typeof Date.now;
  processOn: typeof process.on;
  processExit: typeof process.exit;
};

const defaultDeps: WorkerSupervisorDeps = {
  claimJob,
  completeJob,
  failJob,
  fetchRuntimeConfig,
  sendHeartbeat,
  processClaimedJob,
  log,
  delay,
  setInterval: globalThis.setInterval,
  clearInterval: globalThis.clearInterval,
  now: Date.now,
  processOn: process.on.bind(process),
  processExit: process.exit.bind(process),
};

export class WorkerSupervisor {
  private readonly consumers = new Map<number, ConsumerState>();
  private configTimer?: NodeJS.Timeout;
  private heartbeatTimer?: NodeJS.Timeout;
  private nextConsumerNumber = 1;
  private lastConfigSyncAt = '';
  private started = false;
  private shuttingDown = false;
  private readonly startedAt = new Date().toISOString();
  private readonly shutdownController = new AbortController();
  private readonly shutdownReportedJobIds = new Set<string>();

  constructor(
    private readonly runtimeConfig: RuntimeConfig,
    private readonly deps: WorkerSupervisorDeps = defaultDeps,
  ) {}

  start(): void {
    this.started = true;
    void this.syncRuntimeConfig();
    void this.heartbeatLoop();
    this.configTimer = this.deps.setInterval(() => {
      void this.syncRuntimeConfig();
    }, CONFIG_SYNC_MS);
    this.heartbeatTimer = this.deps.setInterval(() => {
      void this.heartbeatLoop();
    }, WORKER_HEARTBEAT_MS);
    this.reconcileConsumers();
    this.installSignalHandlers();
  }

  getStatus(): WorkerHealthSnapshot {
    return {
      workerId: WORKER_ID,
      startedAt: this.startedAt,
      status: !this.started ? 'starting' : this.shuttingDown ? 'shutting_down' : 'running',
      activeConsumers: Array.from(this.consumers.values()).filter(
        (state) => !state.stopRequested,
      ).length,
      targetConcurrency: this.runtimeConfig.maxConcurrent,
      headless: this.runtimeConfig.headless,
      lastConfigSyncAt: this.lastConfigSyncAt,
    };
  }

  private async consumerLoop(state: ConsumerState): Promise<void> {
    this.deps.log('INFO', 'consumer started', { consumerId: state.consumerId });
    while (!state.stopRequested) {
      try {
        const job = await this.deps.claimJob(this.shutdownController.signal);
        if (!job) {
          await this.sleep(POLL_MS);
          continue;
        }

        this.deps.log('INFO', 'job claimed', {
          consumerId: state.consumerId,
          job_id: job.id,
          run_type: job.run_type,
          employee_id: job.employee_id,
        });

        state.activeJobIds.add(job.id);
        const jobController = new AbortController();
        const abortActiveJob = () => jobController.abort();
        this.shutdownController.signal.addEventListener('abort', abortActiveJob, { once: true });
        state.activeJobControllers.set(job.id, jobController);
        try {
          const record = await this.deps.processClaimedJob(
            job,
            this.runtimeConfig,
            jobController.signal,
          );
          if (jobController.signal.aborted) {
            throw new Error('worker shutdown interrupted job before completion report');
          }
          await this.deps.completeJob(job.id, record);
          this.deps.log('INFO', 'job success', {
            consumerId: state.consumerId,
            job_id: job.id,
            record,
          });
        } catch (error) {
          if (error instanceof WorkerApiUncertainCompletionError) {
            this.deps.log('WARN', 'job completion report uncertain - fail report skipped', {
              consumerId: state.consumerId,
              job_id: job.id,
              error: error.message,
            });
            continue;
          }
          if (error instanceof WorkerApiFinalStateError) {
            this.deps.log('WARN', 'job completion report reached final-state response - fail report skipped', {
              consumerId: state.consumerId,
              job_id: job.id,
              status: error.status,
              error: error.message,
            });
            continue;
          }
          const errorMessage = String((error as Error)?.message ?? error);
          if (this.shutdownReportedJobIds.has(job.id)) {
            this.deps.log('WARN', 'job failed after shutdown timeout report', {
              consumerId: state.consumerId,
              job_id: job.id,
              error: errorMessage,
            });
            continue;
          }
          try {
            await this.deps.failJob(job.id, errorMessage);
            this.deps.log('WARN', 'job failed - reported to backend', {
              consumerId: state.consumerId,
              job_id: job.id,
              error: errorMessage,
            });
          } catch (reportError) {
            this.deps.log('ERROR', 'job failed - fail report failed', {
              consumerId: state.consumerId,
              job_id: job.id,
              error: errorMessage,
              reportError: (reportError as Error)?.message ?? String(reportError),
            });
          }
        } finally {
          this.shutdownController.signal.removeEventListener('abort', abortActiveJob);
          state.activeJobControllers.delete(job.id);
          state.activeJobIds.delete(job.id);
        }
      } catch (error) {
        if (this.shuttingDown && error instanceof WorkerApiCancelledError) {
          break;
        }
        this.deps.log('ERROR', 'consumer loop error', {
          consumerId: state.consumerId,
          error: (error as Error)?.message ?? String(error),
        });
        await this.sleep(Math.max(1000, Math.floor(POLL_MS / 2)));
      }
    }
    this.deps.log('INFO', 'consumer stopped', { consumerId: state.consumerId });
  }

  private async syncRuntimeConfig(): Promise<void> {
    if (this.shuttingDown) {
      return;
    }
    try {
      const next = normalizeRuntimeConfigPatch(await this.deps.fetchRuntimeConfig());
      const previous = { ...this.runtimeConfig };

      if (typeof next.maxConcurrent === 'number') {
        // Opsi D (hybrid): WORKER_CONCURRENCY (default 3) = hard cap per
        // instance. Backend tidak boleh menaikkan di atas cap; kalau backend
        // menurunkan di bawah cap, ikuti backend (kontrol darurat via UI).
        this.runtimeConfig.maxConcurrent = Math.min(
          next.maxConcurrent,
          WORKER_CONCURRENCY_CAP,
        );
      }
      if (typeof next.headless === 'boolean') {
        this.runtimeConfig.headless = next.headless;
      }
      if (next.geo) {
        this.runtimeConfig.geo = {
          ...this.runtimeConfig.geo,
          ...next.geo,
        };
      }

      if (
        previous.maxConcurrent !== this.runtimeConfig.maxConcurrent ||
        previous.headless !== this.runtimeConfig.headless ||
        JSON.stringify(previous.geo) !== JSON.stringify(this.runtimeConfig.geo)
      ) {
        this.deps.log('INFO', 'runtime config updated', {
          previous,
          current: this.runtimeConfig,
        });
      }
      this.lastConfigSyncAt = new Date(this.deps.now()).toISOString();
      this.reconcileConsumers();
    } catch (error) {
      this.deps.log('WARN', 'runtime config sync failed', {
        error: (error as Error)?.message ?? String(error),
      });
    }
  }

  private reconcileConsumers(): void {
    if (this.shuttingDown) {
      return;
    }

    while (this.consumers.size < this.runtimeConfig.maxConcurrent) {
      const id = this.nextConsumerNumber++;
      const state: ConsumerState = {
        id,
        consumerId: `${WORKER_ID}-c${id}`,
        stopRequested: false,
        activeJobIds: new Set<string>(),
        activeJobControllers: new Map<string, AbortController>(),
      };
      this.consumers.set(id, state);
      this.consumerLoop(state)
        .catch((error) => {
          this.deps.log('ERROR', 'consumer crashed', {
            consumerId: state.consumerId,
            error: (error as Error)?.message ?? String(error),
          });
        })
        .finally(() => {
          this.consumers.delete(id);
          this.reconcileConsumers();
        });
    }

    if (this.consumers.size <= this.runtimeConfig.maxConcurrent) {
      return;
    }

    const active = Array.from(this.consumers.values()).sort((a, b) => b.id - a.id);
    let activeCount = this.consumers.size;
    for (const state of active) {
      if (activeCount <= this.runtimeConfig.maxConcurrent) {
        break;
      }
      if (state.stopRequested) {
        continue;
      }
      state.stopRequested = true;
      activeCount -= 1;
      this.deps.log('INFO', 'consumer stop requested', {
        consumerId: state.consumerId,
        targetConcurrency: this.runtimeConfig.maxConcurrent,
      });
    }
  }

  private async heartbeatLoop(): Promise<void> {
    if (this.shuttingDown) {
      return;
    }
    try {
      await this.deps.sendHeartbeat({
        consumerCount: Array.from(this.consumers.values()).filter(
          (state) => !state.stopRequested,
        ).length,
        targetConcurrency: this.runtimeConfig.maxConcurrent,
        headless: this.runtimeConfig.headless,
        lastSyncAt: this.lastConfigSyncAt,
      });
    } catch (error) {
      this.deps.log('WARN', 'worker heartbeat failed', {
        error: (error as Error)?.message ?? String(error),
      });
    }
  }

  private installSignalHandlers(): void {
    for (const signal of ['SIGINT', 'SIGTERM'] as const) {
      this.deps.processOn(signal, () => {
        void this.gracefulShutdown(signal);
      });
    }
  }

  private async gracefulShutdown(signal: string): Promise<void> {
    if (this.shuttingDown) {
      return;
    }
    this.shuttingDown = true;
    this.shutdownController.abort();
    if (this.configTimer) {
      this.deps.clearInterval(this.configTimer);
    }
    if (this.heartbeatTimer) {
      this.deps.clearInterval(this.heartbeatTimer);
    }

    this.deps.log('INFO', 'shutdown requested', {
      signal,
      activeConsumers: this.consumers.size,
    });

    for (const state of this.consumers.values()) {
      state.stopRequested = true;
      for (const controller of state.activeJobControllers?.values() ?? []) {
        controller.abort();
      }
    }

    const deadline = this.deps.now() + 15_000;
    while (this.consumers.size > 0 && this.deps.now() < deadline) {
      await this.deps.delay(250);
    }

    if (this.consumers.size > 0) {
      await this.failActiveJobsOnShutdownTimeout();
      this.deps.log('WARN', 'shutdown timeout reached', {
        remainingConsumers: this.consumers.size,
      });
    }

    try {
      await this.deps.sendHeartbeat({
        consumerCount: 0,
        targetConcurrency: this.runtimeConfig.maxConcurrent,
        headless: this.runtimeConfig.headless,
        lastSyncAt: this.lastConfigSyncAt,
      });
    } catch {
      // best-effort
    }

    this.deps.log('INFO', 'worker shutdown complete', {
      remainingConsumers: this.consumers.size,
    });
    this.deps.processExit(0);
  }

  private sleep(ms: number): Promise<void> {
    if (this.shutdownController.signal.aborted) {
      return Promise.resolve();
    }

    return new Promise((resolve) => {
      const cleanup = () => this.shutdownController.signal.removeEventListener('abort', onAbort);
      const timeout = setTimeout(() => {
        cleanup();
        resolve();
      }, ms);
      const onAbort = () => {
        clearTimeout(timeout);
        cleanup();
        resolve();
      };
      this.shutdownController.signal.addEventListener('abort', onAbort, { once: true });
      timeout.unref?.();
    });
  }

  private async failActiveJobsOnShutdownTimeout(): Promise<void> {
    const activeJobIds = Array.from(this.consumers.values()).flatMap((state) =>
      Array.from(state.activeJobIds),
    );
    const uniqueJobIds = Array.from(new Set(activeJobIds));

    await Promise.allSettled(
      uniqueJobIds.map(async (jobId) => {
        try {
          await this.deps.failJob(jobId, 'worker shutdown timeout');
          this.shutdownReportedJobIds.add(jobId);
          this.deps.log('WARN', 'active job failed during shutdown timeout', {
            job_id: jobId,
          });
        } catch (error) {
          this.deps.log('ERROR', 'active job shutdown fail report failed', {
            job_id: jobId,
            error: (error as Error)?.message ?? String(error),
          });
        }
      }),
    );
  }
}
