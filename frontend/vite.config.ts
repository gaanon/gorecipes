import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
    plugins: [sveltekit()],
    server: {
        proxy: {
            // Proxy API requests to the Go backend
            '/api': {
                target: 'http://192.168.1.45:8080', // Your Go backend address
                changeOrigin: true,
            },
            // Add this new proxy rule for images
            '/uploads/images': {
                target: 'http://192.168.1.45:8080', // Your Go backend address
                changeOrigin: true,
                // No rewrite needed, as the backend serves from /uploads/images
            }
        }
    }
});
