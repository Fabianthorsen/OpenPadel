type Handler<T = any> = (payload: T) => void;

export type SessionStream = ReturnType<typeof sessionStream>;

export function sessionStream(sessionId: string) {
	let es: EventSource | null = null;
	let backoff = 500;
	let closed = false;
	const handlers = new Map<string, Set<Handler>>();

	function attach(type: string) {
		es?.addEventListener(type, (e: MessageEvent) => {
			const env = JSON.parse(e.data) as { payload?: unknown };
			handlers.get(type)?.forEach((fn) => fn(env.payload));
		});
	}

	function connect() {
		if (closed || es) return;
		es = new EventSource(`/api/sessions/${sessionId}/events`);
		es.onopen = () => {
			backoff = 500;
		};
		es.onerror = () => {
			es?.close();
			es = null;
			if (closed || document.hidden || !navigator.onLine) return;
			setTimeout(connect, backoff);
			backoff = Math.min(backoff * 2, 30_000);
		};
		for (const type of handlers.keys()) attach(type);
	}

	const onVis = () => {
		if (document.hidden) {
			es?.close();
			es = null;
		} else {
			backoff = 500;
			connect();
		}
	};
	const onOnline = () => {
		backoff = 500;
		connect();
	};
	const onOffline = () => {
		es?.close();
		es = null;
	};

	return {
		start() {
			if (typeof document === 'undefined') return;
			document.addEventListener('visibilitychange', onVis);
			window.addEventListener('online', onOnline);
			window.addEventListener('offline', onOffline);
			connect();
		},

		onEvent<T>(type: string, fn: Handler<T>): () => void {
			if (!handlers.has(type)) {
				handlers.set(type, new Set());
				if (es) attach(type);
			}
			handlers.get(type)!.add(fn as Handler);
			return () => handlers.get(type)?.delete(fn as Handler);
		},

		close() {
			closed = true;
			es?.close();
			es = null;
			document.removeEventListener('visibilitychange', onVis);
			window.removeEventListener('online', onOnline);
			window.removeEventListener('offline', onOffline);
		}
	};
}
