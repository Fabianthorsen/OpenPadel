import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  plugins: [
    tailwindcss(),
    sveltekit(),
    VitePWA({
      strategies: 'generateSW',
      injectRegister: null,
      selfDestroying: false,
      manifest: {
        name: 'OpenPadel',
        short_name: 'OpenPadel',
        description: 'Padel, organised.',
        theme_color: '#4A7856',
        background_color: '#f7f6f3',
        display: 'standalone',
        orientation: 'portrait',
        start_url: '/',
        icons: [
          { src: '/android-chrome-192x192.png', sizes: '192x192', type: 'image/png' },
          { src: '/android-chrome-512x512.png', sizes: '512x512', type: 'image/png' },
          { src: '/android-chrome-512x512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
    }),
  ],
});
