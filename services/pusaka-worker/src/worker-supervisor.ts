import process from 'process';
import { setTimeout as delay } from 'timers/promises';

import { CONFIG_SYNC_MS, POLL_MS, WORKER_ID } from './config.js';
import {
  claimJob,
  completeJob,
  failJob,
  fetchRuntimeConfig,
  sendHeartbeat,
} from './api-client.js';
import { log } from './logger.js';
import { processClaimedJob } from './pusaka-runner.js';
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
  private shuttingDown = false;

  constructor(
    private readonly runtimeConfig: RuntimeConfig,
    private readonly deps: WorkerSupervisorDeps = defaultDeps,
  ) {}

  start(): void {
    void this.syncRuntimeConfig();
    this.configTimer = this.deps.setInterval(() => {
      void this.syncRuntimeConfig();
    }, CONFIG_SYNC_MS);
    this.heartbeatTimer = this.deps.setInterval(() => {
      void this.heartbeatLoop();
    }, CONFIG_SYNC_MS);
    this.reconcileConsumers();
    this.installSignalHandlers();
  }

  private async consumerLoop(state: ConsumerState): Promise<void> {
    this.deps.log('INFO', 'consumer started', { consumerId: state.consumerId });
    while (!state.stopRequested) {
      try {
        const job = await this.deps.claimJob();
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

        try {
          const record = await this.deps.processClaimedJob(job, this.runtimeConfig);
          await this.deps.completeJob(job.id, record);
          this.deps.log('INFO', 'job success', {
            consumerId: state.consumerId,
            job_id: job.id,
            record,
          });
        } catch (error) {
          const errorMessage = String((error as Error)?.message ?? error);
          await this.deps.failJob(job.id, errorMessage);
          this.deps.log('WARN', 'job failed — reported to frontend', {
            consumerId: state.consumerId,
            job_id: job.id,
            error: errorMessage,
          });
        }
      } catch (error) {
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
      const next = await this.deps.fetchRuntimeConfig();
      const previous = { ...this.runtimeConfig };

      if (typeof next.maxConcurrent === 'number') {
        this.runtimeConfig.maxConcurrent = next.maxConcurrent;
      }
      if (typeof next.headless === 'boolean') {
        this.runtimeConfig.headless = next.headless;
      }

      if (
        previous.maxConcurrent !== this.runtimeConfig.maxConcurrent ||
        previous.headless !== this.runtimeConfig.headless
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
    }

    const deadline = this.deps.now() + 15_000;
    while (this.consumers.size > 0 && this.deps.now() < deadline) {
      await this.deps.delay(250);
    }

    if (this.consumers.size > 0) {
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
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}
