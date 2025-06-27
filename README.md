# 🚀 Saan System

**Modern Business Management Platform with Chat-to-Order Flow**

> "สาน" (Saan) แปลว่า "เชื่อมโยง" หรือ "สานต่อ" - เชื่อมโยงการสื่อสาร กับการสั่งซื้อ และการจัดส่ง

## 🏗️ Architecture Overview

```
Chat → Order → Delivery
  ↓      ↓       ↓
💬    📋     🚚
```

### 🧠 Core Services (Go Microservices)
- **Chat Service** (8001) - Real-time messaging & AI responses
- **Order Service** (8002) - Order management & workflow
- **Inventory Service** (8003) - Stock tracking & management
- **Delivery Service** (8004) - Shipping & logistics
- **Finance Service** (8005) - Payments & accounting

### 💻 Frontend Applications (Next.js 15)
- **Web App** - Main customer interface
- **Admin Dashboard** - Management interface
- **POS System** - Point of sale interface

### 📦 Shared Packages
- **UI Components** - Design system based on shadcn/ui
- **TypeScript Types** - Shared interfaces & types
- **Utilities** - Common functions & helpers

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.23+
- Node.js 18+
- Make

### Development Setup

```bash
# Clone and setup
git clone <repo-url>
cd saan

# Start all services
make dev

# Or manually
docker-compose up -d
```

### Available Services
- 🌐 Web App: http://localhost:3000
- 🔧 Admin: http://localhost:3001
- 💬 Chat API: http://localhost:8001
- 📋 Order API: http://localhost:8002
- 📦 Inventory API: http://localhost:8003

## 📁 Project Structure

```
saan/
├── services/              # 🧠 Go Microservices
│   ├── chat/             # Real-time messaging
│   ├── order/            # Order management
│   ├── inventory/        # Stock tracking
│   ├── delivery/         # Shipping & logistics
│   ├── finance/          # Payments & accounting
│   └── shared/           # Common utilities
│
├── apps/                 # 💻 Frontend Applications
│   ├── web/              # Main customer app (Next.js)
│   ├── admin/            # Management dashboard
│   └── pos/              # Point of sale system
│
├── packages/             # 📦 Shared Libraries
│   ├── ui/               # Design system components
│   ├── types/            # TypeScript interfaces
│   └── utils/            # Common utilities
│
├── infrastructure/       # 🛠️ DevOps & Deployment
│   ├── docker/           # Dockerfiles
│   ├── k8s/              # Kubernetes manifests
│   └── terraform/        # Infrastructure as Code
│
├── docker-compose.yml    # 🐳 Local development
├── Makefile              # ⚙️ Build & deployment commands
└── .env.local            # 🔐 Environment variables
```

## 🎯 Key Features

### 💬 Smart Chat System
- Multi-platform messaging (LINE, Facebook, WhatsApp)
- AI-powered responses
- Automatic order conversion
- Real-time notifications

### 📋 Streamlined Orders
- Chat-to-order conversion
- Automated inventory checks
- Smart supplier recommendations
- Real-time status tracking

### 🚚 Integrated Delivery
- Route optimization
- Real-time tracking
- Delivery notifications
- Performance analytics

### 📊 Business Intelligence
- Real-time dashboards
- Predictive analytics
- Performance metrics
- Financial reporting

## 🛠️ Development

### Make Commands
```bash
make dev          # Start development environment
make build        # Build all services
make test         # Run tests
make lint         # Code linting
make deploy       # Deploy to production
```

### Service Commands
```bash
# Individual service management
make chat-dev     # Start chat service in dev mode
make order-build  # Build order service
make web-start    # Start web frontend
```

## 🔧 Environment Configuration

Copy `.env.example` to `.env.local` and configure:

```env
# Database
DATABASE_URL=postgres://saan:password@localhost:5432/saan_db

# Redis
REDIS_URL=redis://localhost:6379

# External APIs
LINE_CHANNEL_SECRET=your_line_secret
FACEBOOK_APP_SECRET=your_fb_secret

# Services
CHAT_SERVICE_URL=http://chat:8001
ORDER_SERVICE_URL=http://order:8002
```

## 📚 Documentation

- [Architecture Guide](./docs/architecture.md)
- [API Documentation](./docs/api.md)
- [Deployment Guide](./docs/deployment.md)
- [Contributing](./docs/contributing.md)

## 🚀 Production Deployment

### Docker Swarm
```bash
make deploy-swarm
```

### Kubernetes
```bash
make deploy-k8s
```

### Cloud Providers
- AWS ECS/EKS
- Google Cloud Run/GKE
- Azure Container Instances/AKS

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ by the Saan Team**
