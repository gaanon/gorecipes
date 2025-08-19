# Stage 1: Build the SvelteKit static application
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package.json and package-lock.json (or yarn.lock if you use Yarn)
COPY package.json package-lock.json* ./
# If you use yarn, replace the above with:
# COPY package.json yarn.lock ./

# Install dependencies
RUN npm install --frozen-lockfile
# If you use yarn, replace with:
# RUN yarn install --frozen-lockfile

# Copy the rest of the frontend application code
COPY . .

# Build the SvelteKit application using adapter-static
# The output is typically in the 'build' directory for adapter-static
RUN npm run build

# Stage 2: Serve the static assets with Nginx
FROM nginx:stable-alpine

# Remove default Nginx static assets
RUN rm -rf /usr/share/nginx/html/*

# Configure Nginx to handle large file uploads (for photo uploads)
RUN { \
    echo 'client_max_body_size 10M;'; \
    echo 'client_body_timeout 60s;'; \
    echo 'proxy_connect_timeout 300;'; \
    echo 'proxy_send_timeout 300;'; \
    echo 'proxy_read_timeout 300;'; \
    echo 'send_timeout 300;'; \
} > /etc/nginx/conf.d/custom.conf

# Copy static assets from the builder stage
# SvelteKit adapter-static usually outputs to 'build' directory
COPY --from=builder /app/build /usr/share/nginx/html

# Copy custom Nginx configuration
# This file will be created next and needs to handle SPA routing and API proxying
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80 for Nginx
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]