import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			// Proxy API requests to the Go backend
			// Adjust the target port if your Go backend runs on a different port
			'/api': {
        target: process.env.VITE_API_BACKEND_URL || 'http://localhost:8080',
                changeOrigin: true,
			}
		}
	}
});
