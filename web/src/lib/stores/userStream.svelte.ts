type Handler<T = any> = (payload: T) => void;

export type UserStream = ReturnType<typeof userStream>;

/**
 * SSE stream scoped to the authenticated user.
 *
 * Takes a getter for the token so each reconnect reads the latest value.
 * EventSource cannot set headers, so the token is appended as ?token=.
 * If getToken() returns null, connect() no-ops until start() is called again.
 */
export function userStream(getToken: () => string | null) {
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
		const token = getToken();
		if (!token) return;
		es = new EventSource(`/api/users/events?token=${encodeURIComponent(token)}`);
		es.onopen = () => {
			backoff = 500;
		};
		es.onerror = () => {
			// Close manually so we control reconnect timing and re-read the token.
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
			if (typeof document !== 'undefined') {
				document.removeEventListener('visibilitychange', onVis);
				window.removeEventListener('online', onOnline);
				window.removeEventListener('offline', onOffline);
			}
		}
	};
}
