# Contributing to Saan System

Thank you for your interest in contributing to the Saan System! This document provides guidelines and information for contributors.

## 🤝 Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and constructive in all interactions.

## 🚀 Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.23+
- Node.js 18+
- Make
- Git

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/saan.git
   cd saan
   ```

2. **Set up environment**
   ```bash
   cp .env.example .env.local
   # Edit .env.local with your configuration
   ```

3. **Start development environment**
   ```bash
   make dev
   ```

## 📁 Project Structure

```
saan/
├── services/              # Go microservices
│   ├── order/            # Order management service
│   ├── chat/             # Chat service (future)
│   └── ...
├── apps/                 # Frontend applications
│   ├── web/              # Customer web app
│   ├── admin/            # Admin dashboard
│   └── ...
├── packages/             # Shared packages
│   ├── types/            # TypeScript types
│   └── ui/               # UI components
└── infrastructure/       # DevOps configs
```

## 🛠️ Development Guidelines

### Go Services (Clean Architecture)

#### Directory Structure
```
services/service-name/
├── cmd/                  # Application entrypoints
├── internal/
│   ├── domain/          # Business logic & entities
│   ├── application/     # Use cases & DTOs
│   ├── transport/       # HTTP/gRPC handlers
│   └── infrastructure/ # Database, external APIs
├── pkg/                 # Public packages
├── migrations/          # Database migrations
└── Dockerfile
```

#### Coding Standards
- Follow Clean Architecture principles
- Use dependency injection
- Write comprehensive tests
- Include proper error handling
- Add structured logging
- Document public APIs

#### Example Service Structure
```go
// Domain layer
type Order struct {
    ID       uuid.UUID
    Status   OrderStatus
    // ... other fields
}

func (o *Order) UpdateStatus(status OrderStatus) error {
    // Business logic here
}

// Application layer
type OrderService struct {
    repo domain.OrderRepository
}

func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
    // Use case implementation
}

// Transport layer
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // HTTP handler implementation
}
```

### Frontend Applications (Next.js 15)

#### Directory Structure
```
apps/app-name/
├── src/
│   ├── app/             # App Router pages
│   ├── components/      # React components
│   │   ├── ui/         # Base UI components
│   │   └── feature/    # Feature-specific components
│   ├── lib/            # Utilities & configurations
│   ├── hooks/          # Custom React hooks
│   ├── store/          # State management
│   └── types/          # TypeScript types
├── public/             # Static assets
└── package.json
```

#### Coding Standards
- Use TypeScript for all files
- Follow Next.js 15 App Router conventions
- Use Server Components by default
- Implement proper error boundaries
- Use shared packages from `packages/`
- Follow the design system

### Database

#### Migrations
```sql
-- migrations/001_create_orders.sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    -- ... other columns
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
```

#### Naming Conventions
- Tables: `snake_case` (e.g., `order_items`)
- Columns: `snake_case` (e.g., `customer_id`)
- Indexes: `idx_table_column` (e.g., `idx_orders_status`)

## 🧪 Testing

### Go Services
```bash
# Run tests
cd services/order
make test

# Run with coverage
make test-coverage

# Run specific test
go test ./internal/domain/...
```

### Frontend
```bash
# Run tests
cd apps/web
npm test

# Run with coverage
npm run test:coverage
```

### Integration Tests
```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make test-integration
```

## 📝 Commit Guidelines

### Commit Message Format
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples
```bash
feat(order): add order status validation
fix(web): resolve chat interface rendering issue
docs(readme): update installation instructions
test(order): add unit tests for order service
```

## 🔄 Pull Request Process

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**
   - Follow coding standards
   - Add tests for new features
   - Update documentation if needed

3. **Test Your Changes**
   ```bash
   make test
   make lint
   ```

4. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat(scope): description"
   ```

5. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **PR Requirements**
   - Descriptive title and description
   - Reference any related issues
   - Include screenshots for UI changes
   - Ensure all CI checks pass

## 🐛 Bug Reports

### Before Submitting
- Check existing issues
- Try to reproduce the bug
- Gather relevant information

### Bug Report Template
```markdown
**Bug Description**
A clear description of the bug.

**Steps to Reproduce**
1. Go to '...'
2. Click on '...'
3. See error

**Expected Behavior**
What you expected to happen.

**Screenshots**
If applicable, add screenshots.

**Environment**
- OS: [e.g., macOS, Ubuntu]
- Browser: [e.g., Chrome, Safari]
- Version: [e.g., 1.0.0]
```

## 💡 Feature Requests

### Feature Request Template
```markdown
**Feature Description**
A clear description of the feature.

**Problem Statement**
What problem does this solve?

**Proposed Solution**
How should this feature work?

**Alternatives Considered**
Any alternative solutions considered.

**Additional Context**
Any other context or screenshots.
```

## 📚 Resources

### Documentation
- [Architecture Guide](./docs/architecture.md)
- [API Documentation](./docs/api.md)
- [Deployment Guide](./docs/deployment.md)

### External Resources
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Next.js Documentation](https://nextjs.org/docs)
- [Go Documentation](https://golang.org/doc/)

## ❓ Getting Help

- **Discord**: [Join our Discord](#)
- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion

## 🎉 Recognition

Contributors will be recognized in our:
- README.md contributors section
- Release notes
- Annual contributor appreciation post

Thank you for contributing to the Saan System! 🙏
