# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY web/package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY web/ ./

# Build the application
RUN npm run build

# Final stage
FROM node:18-alpine

WORKDIR /app

# Install only production dependencies
COPY web/package*.json ./
RUN npm ci --only=production && npm cache clean --force

# Copy built application
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/next.config.ts ./
COPY --from=builder /app/package.json ./

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -u 1001 -S nextjs -G nodejs

# Change ownership of the app directory
RUN chown -R nextjs:nodejs /app
USER nextjs

EXPOSE 3000

ENV NODE_ENV=production
ENV PORT=3000

CMD ["npm", "start"]