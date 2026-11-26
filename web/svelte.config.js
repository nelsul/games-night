import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
export default {
	preprocess: vitePreprocess(),
	kit: {
		// A single-page app: the Go API serves index.html for every route.
		adapter: adapter({ fallback: 'index.html' })
	}
};
