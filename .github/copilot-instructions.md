<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# Saan System - Copilot Instructions

## Project Overview
This is the Saan System - a modern business management platform with microservices architecture using Go backend services and Next.js frontend applications.

## Architecture
- **Backend**: Go microservices with clean architecture pattern
- **Frontend**: Next.js 15 with App Router and TypeScript
- **Database**: PostgreSQL with Redis for caching
- **Infrastructure**: Docker Compose for development, Kubernetes for production

## Code Style Guidelines

### Go Services (services/)
- Follow Clean Architecture pattern with layers: transport, application, domain, infrastructure
- Use standard Go naming conventions (PascalCase for exported, camelCase for unexported)
- Always include proper error handling and logging
- Use interfaces for dependency injection
- Structure: cmd/main.go, internal/transport, internal/application, internal/domain, internal/infrastructure

### Next.js Apps (apps/)
- Use TypeScript for all files
- Follow Next.js 15 App Router conventions
- Use Server Components by default, Client Components when needed
- Implement proper error boundaries and loading states
- Use shared packages from packages/ directory

### Shared Packages (packages/)
- Export only necessary interfaces and functions
- Use proper TypeScript types and interfaces
- Follow semantic versioning for internal packages
- Include proper JSDoc documentation

## Naming Conventions
- Services: kebab-case (chat-service, order-service)
- Files: kebab-case for Go packages, camelCase for TypeScript
- Functions: camelCase in TypeScript, PascalCase for exported Go functions
- Constants: UPPER_SNAKE_CASE
- Environment variables: UPPER_SNAKE_CASE

## Development Patterns
- Use dependency injection in Go services
- Implement proper error handling with custom error types
- Use React Query for API state management in frontend
- Implement proper logging with structured logs
- Use middleware for cross-cutting concerns (CORS, auth, logging)

## API Design
- RESTful APIs with consistent response format
- Use proper HTTP status codes
- Include request/response validation
- Implement rate limiting and authentication
- Use WebSocket for real-time features (chat service)

## Database
- Use migrations for schema changes
- Follow PostgreSQL naming conventions (snake_case)
- Implement proper indexing for performance
- Use transactions for data consistency

## Testing
- Unit tests for business logic
- Integration tests for API endpoints
- Use test containers for database testing
- Mock external dependencies

## Security
- Never commit secrets or API keys
- Use environment variables for configuration
- Implement proper authentication and authorization
- Validate and sanitize all inputs
- Use HTTPS in production

## Performance
- Implement caching strategies with Redis
- Use database connection pooling
- Optimize Docker images with multi-stage builds
- Implement proper monitoring and metrics

When generating code, please follow these conventions and patterns to maintain consistency across the Saan System codebase.
