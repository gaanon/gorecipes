import adapter from '@sveltejs/adapter-static'; // Changed from adapter-auto
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),

	kit: {
		adapter: adapter({
			// default options are fine for most SPAs
			pages: 'build', // Default output directory for HTML files
			assets: 'build', // Default output directory for static assets
			fallback: 'index.html', // Default fallback for SPAs (can also be 200.html or 404.html)
			precompress: false // Optionally enable precompression (brotli, gzip)
		}),
		// Ensure paths are correctly handled for static export if you have a base path
		// paths: {
		//  base: process.env.NODE_ENV === 'production' ? '/your-repo-name' : ''
		// }
	}
};

export default config;
