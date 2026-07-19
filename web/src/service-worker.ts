/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

declare const self: ServiceWorkerGlobalScope;

const CACHE = `cache-${version}`;
const ASSETS = [...build, ...files];

self.addEventListener('install', (event) => {
	event.waitUntil(caches.open(CACHE).then((cache) => cache.addAll(ASSETS)));
	self.skipWaiting();
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(keys.filter((key) => key !== CACHE).map((key) => caches.delete(key)))
			)
	);
	self.clients.claim();
});

self.addEventListener('fetch', (event) => {
	if (event.request.method !== 'GET') return;
	event.respondWith(
		caches
			.open(CACHE)
			.then((cache) => cache.match(event.request).then((cached) => cached ?? fetch(event.request)))
	);
});

self.addEventListener('push', (event) => {
	if (!event.data) return;

	let title = 'OpenPadel';
	let body = 'Something happened!';
	let url = '/';

	try {
		const data = event.data.json();
		title = data.title ?? title;
		body = data.body ?? body;
		url = data.url ?? url;
	} catch {
		body = event.data.text();
	}

	event.waitUntil(
		(async () => {
			await self.registration.showNotification(title, {
				body,
				icon: '/icon-192.png',
				badge: '/icon-192.png',
				data: { url }
			});
			// Bump the app-icon badge. `setAppBadge()` without a count shows a
			// generic dot on platforms that support it (no-op elsewhere).
			await self.navigator.setAppBadge?.().catch(() => {});
		})()
	);
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	const url = event.notification.data?.url ?? '/';
	event.waitUntil(
		(async () => {
			// Opening the notification acknowledges it — clear the app-icon badge.
			await self.navigator.clearAppBadge?.().catch(() => {});
			const clients = await self.clients.matchAll({ type: 'window' });
			const existing = clients.find((c) => c.url.includes(url) && 'focus' in c);
			if (existing) return existing.focus();
			return self.clients.openWindow(url);
		})()
	);
});
