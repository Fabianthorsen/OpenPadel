import { api } from '$lib/api/client';

function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = atob(base64);
  return Uint8Array.from([...rawData].map((c) => c.charCodeAt(0)));
}

export async function subscribeToPush(token: string): Promise<boolean> {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    throw new Error('Push not supported in this browser');
  }

  const permission = await Notification.requestPermission();
  if (permission !== 'granted') return false;

  const reg = await Promise.race([
    navigator.serviceWorker.ready,
    new Promise<never>((_, reject) => setTimeout(() => reject(new Error('Service worker timed out')), 5000)),
  ]) as ServiceWorkerRegistration;

  const { public_key } = await api.push.getVapidKey();

  const sub = await reg.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(public_key),
  });

  const json = sub.toJSON();
  const keys = json.keys ?? {};
  await api.push.subscribe(token, json.endpoint!, keys['p256dh']!, keys['auth']!);
  return true;
}

export async function unsubscribeFromPush(token: string): Promise<boolean> {
  if (!('serviceWorker' in navigator)) return false;
  const reg = await navigator.serviceWorker.ready;
  const sub = await reg.pushManager.getSubscription();
  if (!sub) return true;
  await api.push.unsubscribe(token, sub.endpoint);
  await sub.unsubscribe();
  return true;
}
