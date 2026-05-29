import { describe, expect, it, vi, beforeEach } from 'vitest';

const mocks = vi.hoisted(() => {
	const sessionExtractor = vi.fn();
	const roomExtractor = vi.fn();
	return {
		sessionExtractor,
		roomExtractor,
		proctoringEventStream: vi.fn(() => new Response('stream', {
			headers: { 'content-type': 'text/event-stream; charset=utf-8' }
		})),
		handleRouteError: vi.fn((error: unknown, route: string) => new Response(JSON.stringify({
			error: error instanceof Error ? error.message : 'failed',
			route
		}), { status: 503 }))
	};
});

vi.mock('$lib/server/api', () => ({
	apiPath: (strings: TemplateStringsArray, ...values: Array<string | number | boolean>) => {
		let path = strings[0] ?? '';
		for (let index = 0; index < values.length; index += 1) {
			path += encodeURIComponent(String(values[index]));
			path += strings[index + 1] ?? '';
		}
		return path;
	},
	apiPathWithQuery: (path: string, params: URLSearchParams | string) => {
		const query = typeof params === 'string' ? params.replace(/^\?/, '') : params.toString();
		return query ? `${path}?${query}` : path;
	},
	requiredRouteParam: (value: string | undefined, name: string) => {
		if (!value) throw new Error(`${name} tidak valid`);
		return value;
	},
	handleRouteError: mocks.handleRouteError
}));

vi.mock('$lib/server/cbt-backend-proxy/proctoring-stream', () => ({
	extractSessionEvents: mocks.sessionExtractor,
	extractRoomDashboardEvents: mocks.roomExtractor,
	proctoringEventStream: mocks.proctoringEventStream
}));

function createEvent(params: Record<string, string | undefined>) {
	return {
		params,
		request: new Request('http://localhost/api/asesmen/sessions/session-1/proctoring/stream'),
		url: new URL('http://localhost/api/asesmen/sessions/session-1/proctoring/stream')
	};
}

describe('proctoring SSE BFF routes', () => {
	beforeEach(() => {
		mocks.proctoringEventStream.mockClear();
		mocks.handleRouteError.mockClear();
	});

	it('bridges session proctoring stream through polling events instead of backend SSE', async () => {
		const mod = await import('../../routes/api/asesmen/sessions/[id]/proctoring/stream/+server');
		const event = createEvent({ id: 'session 1' });

		const response = await mod.GET(event as never);

		expect(response.headers.get('content-type')).toContain('text/event-stream');
		expect(mocks.proctoringEventStream).toHaveBeenCalledWith(event, {
			name: 'session-proctoring-events',
			path: '/api/asesmen/sessions/session%201/proctoring/events?limit=100',
			extractEvents: mocks.sessionExtractor
		});
	});

	it('bridges room proctoring stream through the room dashboard poller', async () => {
		const mod = await import('../../routes/api/asesmen/sessions/[id]/rooms/[rid]/proctoring/stream/+server');
		const event = createEvent({ id: 'session 1', rid: 'room A' });

		const response = await mod.GET(event as never);

		expect(response.headers.get('content-type')).toContain('text/event-stream');
		expect(mocks.proctoringEventStream).toHaveBeenCalledWith(event, {
			name: 'room-proctoring-events',
			path: '/api/asesmen/sessions/session%201/rooms/room%20A/proctoring',
			extractEvents: mocks.roomExtractor
		});
	});
});
