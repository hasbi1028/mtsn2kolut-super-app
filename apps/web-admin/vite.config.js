import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	resolve: {
		conditions: ['browser']
	},
	ssr: {
		noExternal: ['lucide-svelte', 'bits-ui', 'tailwind-variants', 'svelte-sonner', 'katex', '@tiptap/core', '@tiptap/pm', '@tiptap/starter-kit'],
	},
	test: {
		environment: 'jsdom',
		setupFiles: ['./src/test/setup.ts'],
		include: ['src/**/*.test.ts']
	}
});
