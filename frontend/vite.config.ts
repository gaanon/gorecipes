import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			// Proxy API requests to the Go backend
			// Adjust the target port if your Go backend runs on a different port
			'/api': {
				target: 'http://localhost:8080', // Your Go backend address
				changeOrigin: true, // Recommended for virtual hosted sites
				// You might not need rewrite if your Go API already expects /api prefix
				// rewrite: (path) => path.replace(/^\/api/, ''), // Example: if Go doesn't expect /api
			}
		}
	}
});
