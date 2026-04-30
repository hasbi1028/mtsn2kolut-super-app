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

export class WorkerSupervisor {
  private readonly consumers = new Map<number, ConsumerState>();
  private configTimer?: NodeJS.Timeout;
  private heartbeatTimer?: NodeJS.Timeout;
  private nextConsumerNumber = 1;
  private lastConfigSyncAt = '';
  private shuttingDown = false;

  constructor(private readonly runtimeConfig: RuntimeConfig) {}

  start(): void {
    void this.syncRuntimeConfig();
    this.configTimer = setInterval(() => {
      void this.syncRuntimeConfig();
    }, CONFIG_SYNC_MS);
    this.heartbeatTimer = setInterval(() => {
      void this.heartbeatLoop();
    }, CONFIG_SYNC_MS);
    this.reconcileConsumers();
    this.installSignalHandlers();
  }

  private async consumerLoop(state: ConsumerState): Promise<void> {
    log('INFO', 'consumer started', { consumerId: state.consumerId });
    while (!state.stopRequested) {
      try {
        const job = await claimJob();
        if (!job) {
          await this.sleep(POLL_MS);
          continue;
        }

        log('INFO', 'job claimed', {
          consumerId: state.consumerId,
          job_id: job.id,
          run_type: job.run_type,
          employee_id: job.employee_id,
        });

        try {
          const record = await processClaimedJob(job, this.runtimeConfig);
          await completeJob(job.id, record);
          log('INFO', 'job success', {
            consumerId: state.consumerId,
            job_id: job.id,
            record,
          });
        } catch (error) {
          const errorMessage = String((error as Error)?.message ?? error);
          await failJob(job.id, errorMessage);
          log('WARN', 'job failed — reported to frontend', {
            consumerId: state.consumerId,
            job_id: job.id,
            error: errorMessage,
          });
        }
      } catch (error) {
        log('ERROR', 'consumer loop error', {
          consumerId: state.consumerId,
          error: (error as Error)?.message ?? String(error),
        });
        await this.sleep(Math.max(1000, Math.floor(POLL_MS / 2)));
      }
    }
    log('INFO', 'consumer stopped', { consumerId: state.consumerId });
  }

  private async syncRuntimeConfig(): Promise<void> {
    if (this.shuttingDown) {
      return;
    }
    try {
      const next = await fetchRuntimeConfig();
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
        log('INFO', 'runtime config updated', {
          previous,
          current: this.runtimeConfig,
        });
      }
      this.lastConfigSyncAt = new Date().toISOString();
      this.reconcileConsumers();
    } catch (error) {
      log('WARN', 'runtime config sync failed', {
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
          log('ERROR', 'consumer crashed', {
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
      log('INFO', 'consumer stop requested', {
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
      await sendHeartbeat({
        consumerCount: Array.from(this.consumers.values()).filter(
          (state) => !state.stopRequested,
        ).length,
        targetConcurrency: this.runtimeConfig.maxConcurrent,
        headless: this.runtimeConfig.headless,
        lastSyncAt: this.lastConfigSyncAt,
      });
    } catch (error) {
      log('WARN', 'worker heartbeat failed', {
        error: (error as Error)?.message ?? String(error),
      });
    }
  }

  private installSignalHandlers(): void {
    for (const signal of ['SIGINT', 'SIGTERM'] as const) {
      process.on(signal, () => {
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
      clearInterval(this.configTimer);
    }
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
    }

    log('INFO', 'shutdown requested', {
      signal,
      activeConsumers: this.consumers.size,
    });

    for (const state of this.consumers.values()) {
      state.stopRequested = true;
    }

    const deadline = Date.now() + 15_000;
    while (this.consumers.size > 0 && Date.now() < deadline) {
      await delay(250);
    }

    if (this.consumers.size > 0) {
      log('WARN', 'shutdown timeout reached', {
        remainingConsumers: this.consumers.size,
      });
    }

    try {
      await sendHeartbeat({
        consumerCount: 0,
        targetConcurrency: this.runtimeConfig.maxConcurrent,
        headless: this.runtimeConfig.headless,
        lastSyncAt: this.lastConfigSyncAt,
      });
    } catch {
      // best-effort
    }

    log('INFO', 'worker shutdown complete', {
      remainingConsumers: this.consumers.size,
    });
    process.exit(0);
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}
