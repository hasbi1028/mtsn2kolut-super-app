import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter(),
    csrf: {
      trustedOrigins: ['https://bs.mtsn2kolut.sch.id', 'http://localhost:5173', 'http://localhost:8222']
    }
  }
};

export default config;
