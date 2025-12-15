# Dockerfile for Frontend (Vue.js with Vite)
FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY frontend/package.json frontend/package-lock.json* ./

# Install dependencies
RUN npm install

# Install serve for static file serving
RUN npm install -g serve

# Copy frontend source code
COPY frontend/ ./

# Build the production bundle
RUN npm run build

# Expose port 3000
EXPOSE 3000

# Serve the built files with serve (supports SPA routing with -s flag)
CMD ["serve", "-s", "dist", "-l", "3000"]
