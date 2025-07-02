# Build stage
FROM node:20-alpine AS builder

# Set build-time arguments for optimization
ARG NODE_ENV=production
ARG NEXT_TELEMETRY_DISABLED=1

WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache libc6-compat

# Copy package files first for better Docker layer caching
COPY package*.json ./
COPY yarn.lock* ./

# Install dependencies with proper flags for production build
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile --production=false; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i; \
  else echo "Warning: Lockfile not found. Consider using a lockfile for better reproducibility." && npm install; \
  fi

# Copy source code and shared packages
COPY . .

# Copy shared packages from monorepo
COPY ../../packages ./packages

# Disable Next.js telemetry during build
ENV NEXT_TELEMETRY_DISABLED 1

# Build the application with optimizations
RUN \
  if [ -f yarn.lock ]; then yarn build; \
  elif [ -f package-lock.json ]; then npm run build; \
  elif [ -f pnpm-lock.yaml ]; then pnpm build; \
  else npm run build; \
  fi

# Production stage
FROM node:18-alpine AS runner

WORKDIR /app

# Create non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy built application
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

# Expose port
EXPOSE 3001

ENV PORT 3001
ENV HOSTNAME "0.0.0.0"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3001/health || exit 1

# Start the application
CMD ["node", "server.js"]
