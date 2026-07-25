import adapter from '@sveltejs/adapter-node';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
		runes: ({ filename }) => (filename.split(/[/\\\\]/).includes('node_modules') ? undefined : true)
	},
	kit: {
		adapter: adapter(),
		paths: {
			base: '',
			relative: false
		},
		csrf: {
			checkOrigin: true,
			trustedOrigins: [
				'https://mtsn2kolut.sch.id',
				'https://www.mtsn2kolut.sch.id',
				'http://localhost:8021',
			]
		},
		experimental: {
			remoteFunctions: true
		}
	}
};

export default config;
