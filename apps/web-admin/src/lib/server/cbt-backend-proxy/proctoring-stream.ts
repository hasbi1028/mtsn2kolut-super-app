import type { RequestEvent } from '@sveltejs/kit';
import { proxy } from '$lib/server/api';

const encoder = new TextEncoder();
const POLL_MS = 1500;
const PING_MS = 25000;
const MAX_TICKS = 60 * 30; // ~45 minutes at 1.5s/tick; EventSource reconnects automatically.

type EventLike = {
	id?: string;
	created_at?: string;
	[key: string]: unknown;
};

type StreamOptions = {
	path: string;
	extractEvents: (payload: unknown) => EventLike[];
	name?: string;
};

function eventKey(event: EventLike): string {
	return String(event.id ?? `${event.created_at ?? ''}:${event.participant_id ?? ''}:${event.event_type ?? ''}`);
}

function eventSortValue(event: EventLike): string {
	return String(event.created_at ?? event.id ?? '');
}

function serializeSse(eventName: string, data: unknown, id?: string): Uint8Array {
	const lines = [`event: ${eventName}`];
	if (id) lines.push(`id: ${id}`);
	lines.push(`data: ${JSON.stringify(data)}`, '', '');
	return encoder.encode(lines.join('\n'));
}

function sleep(ms: number, signal: AbortSignal): Promise<void> {
	return new Promise((resolve) => {
		if (signal.aborted) {
			resolve();
			return;
		}
		const timer = setTimeout(resolve, ms);
		signal.addEventListener('abort', () => {
			clearTimeout(timer);
			resolve();
		}, { once: true });
	});
}

export function proctoringEventStream(event: RequestEvent, options: StreamOptions): Response {
	const signal = event.request.signal;
	const seen = new Set<string>();
	let lastPing = 0;
	let primed = false;

	const stream = new ReadableStream<Uint8Array>({
		async start(controller) {
			controller.enqueue(serializeSse('ready', { mode: 'sse-poll-bridge', path: options.name ?? options.path }));
			for (let tick = 0; tick < MAX_TICKS && !signal.aborted; tick += 1) {
				try {
					const payload = await proxy(event).get<unknown>(options.path);
					const events = options.extractEvents(payload)
						.filter((item) => item && typeof item === 'object')
						.sort((a, b) => eventSortValue(a).localeCompare(eventSortValue(b)));

					for (const item of events) {
						const key = eventKey(item);
						if (!key || seen.has(key)) continue;
						seen.add(key);
						if (primed) controller.enqueue(serializeSse('proctor_event', item, key));
					}
					primed = true;

					const now = Date.now();
					if (now - lastPing >= PING_MS) {
						lastPing = now;
						controller.enqueue(serializeSse('ping', { time: new Date().toISOString() }));
					}
				} catch (error) {
					controller.enqueue(serializeSse('stream_error', {
						message: error instanceof Error ? error.message : 'Gagal membaca event pengawasan',
					}));
				}
				await sleep(POLL_MS, signal);
			}
			controller.close();
		},
		cancel() {
			// Request abort is observed by the polling loop.
		},
	});

	return new Response(stream, {
		headers: {
			'Content-Type': 'text/event-stream; charset=utf-8',
			'Cache-Control': 'no-cache, no-transform',
			Connection: 'keep-alive',
			'X-Accel-Buffering': 'no',
		},
	});
}

export function extractRoomDashboardEvents(payload: unknown): EventLike[] {
	if (!payload || typeof payload !== 'object') return [];
	const record = payload as { events?: unknown };
	return Array.isArray(record.events) ? record.events as EventLike[] : [];
}

export function extractSessionEvents(payload: unknown): EventLike[] {
	return Array.isArray(payload) ? payload as EventLike[] : [];
}
