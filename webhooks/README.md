# Saan System - Webhook Services

This directory contains the webhook microservices for the Saan System. Each service handles webhooks from specific external platforms and publishes events to Kafka for further processing.

## Architecture Overview

```
External APIs → Webhook Services → Kafka → Event Processors
```

## Services

### 1. Loyverse Webhook (`loyverse-webhook`)
- **Port**: 8093
- **Container**: `loyverse-webhook`
- **Purpose**: Handles webhooks from Loyverse POS system
- **Endpoints**: 
  - `POST /webhook/loyverse` - Receive Loyverse webhooks
  - `GET /health` - Health check
  - `GET /ready` - Readiness check

### 2. Chat Webhook (`chat-webhook`)
- **Port**: 8094
- **Container**: `chat-webhook`
- **Purpose**: Handles chat platform webhooks (Facebook Messenger, LINE)
- **Endpoints**:
  - `GET/POST /webhook/facebook` - Facebook Messenger webhooks
  - `POST /webhook/line` - LINE messaging webhooks
  - `GET /health` - Health check
  - `GET /ready` - Readiness check

### 3. Delivery Webhook (`delivery-webhook`)
- **Port**: 8095
- **Container**: `delivery-webhook`
- **Purpose**: Handles delivery platform webhooks (Grab, LineMan)
- **Endpoints**:
  - `POST /webhook/grab` - Grab delivery status webhooks
  - `POST /webhook/lineman` - LineMan delivery status webhooks
  - `GET /health` - Health check
  - `GET /ready` - Readiness check

### 4. Payment Webhook (`payment-webhook`)
- **Port**: 8096
- **Container**: `payment-webhook`
- **Purpose**: Handles payment gateway webhooks (Omise, 2C2P)
- **Endpoints**:
  - `POST /webhook/omise` - Omise payment webhooks
  - `POST /webhook/2c2p` - 2C2P payment webhooks
  - `GET /health` - Health check
  - `GET /ready` - Readiness check

## Common Features

All webhook services implement:

1. **Signature Verification**: Validates webhook authenticity using platform-specific secrets
2. **Async Processing**: Returns 200 OK immediately, processes events asynchronously
3. **Event Publishing**: Publishes domain events to Kafka topics
4. **Caching**: Stores webhook payloads in Redis for debugging
5. **Health Checks**: Standard health and readiness endpoints
6. **Structured Logging**: Comprehensive logging for monitoring and debugging

## Configuration

Environment variables are used for configuration:

```bash
# Copy example environment file
cp .env.webhook.example .env

# Edit with your actual values
nano .env
```

## Development

### Building Services

```bash
# Build specific service
cd webhooks/loyverse-webhook
make build

# Build all services using Docker Compose
docker-compose build loyverse-webhook chat-webhook delivery-webhook payment-webhook
```

### Running Services

**⚠️ Always use Docker Compose - never run services directly!**

```bash
# Start all webhook services
docker-compose up loyverse-webhook chat-webhook delivery-webhook payment-webhook

# Start specific service
docker-compose up loyverse-webhook

# View logs
docker-compose logs -f loyverse-webhook
```

### Testing Webhooks

```bash
# Health check
curl http://localhost:8093/health

# Test Loyverse webhook (replace with actual payload)
curl -X POST http://localhost:8093/webhook/loyverse \
  -H "Content-Type: application/json" \
  -H "X-Loyverse-Signature: your_signature" \
  -d '{"type":"receipt_created","data":{}}'

# Test Facebook webhook verification
curl "http://localhost:8094/webhook/facebook?hub.mode=subscribe&hub.verify_token=your_token&hub.challenge=challenge"
```

## Event Flow

1. **Webhook Received**: External platform sends webhook to appropriate service
2. **Signature Verification**: Service validates webhook authenticity
3. **Payload Parsing**: Service parses and validates webhook structure
4. **Immediate Response**: Service returns 200 OK to external platform
5. **Async Processing**: Service processes webhook in background
6. **Event Creation**: Service transforms webhook to domain event
7. **Kafka Publishing**: Service publishes event to appropriate Kafka topic
8. **Caching**: Service stores raw webhook in Redis for debugging

## Kafka Topics

- `loyverse-webhooks` - Loyverse POS events
- `chat-messages` - Chat platform events (Facebook, LINE)
- `delivery-updates` - Delivery status updates (Grab, LineMan)
- `payment-events` - Payment gateway events (Omise, 2C2P)

## Monitoring

### Health Checks

```bash
# Check service health
curl http://localhost:8093/health
curl http://localhost:8094/health
curl http://localhost:8095/health
curl http://localhost:8096/health

# Check service readiness (includes dependencies)
curl http://localhost:8093/ready
curl http://localhost:8094/ready
curl http://localhost:8095/ready
curl http://localhost:8096/ready
```

### Logs

```bash
# View service logs
docker-compose logs -f loyverse-webhook
docker-compose logs -f chat-webhook
docker-compose logs -f delivery-webhook
docker-compose logs -f payment-webhook

# View all webhook logs
docker-compose logs -f loyverse-webhook chat-webhook delivery-webhook payment-webhook
```

### Debugging

Webhook payloads are cached in Redis with keys:
- `loyverse:webhook:{type}:{timestamp}`
- `facebook:message:{sender_id}:{timestamp}`
- `line:event:{user_id}:{timestamp}`

## Security

- All webhooks require signature verification
- Secrets are stored in environment variables
- Services run as non-root users in containers
- HTTPS required for production webhooks

## Production Deployment

1. Set up proper SSL certificates for webhook endpoints
2. Configure monitoring and alerting
3. Set up log aggregation
4. Configure auto-scaling based on webhook volume
5. Set up backup webhook endpoints for high availability

## Troubleshooting

### Common Issues

1. **Invalid Signature**: Check webhook secrets in environment variables
2. **Kafka Connection Failed**: Ensure Kafka service is running
3. **Redis Connection Failed**: Ensure Redis service is running
4. **Service Not Ready**: Check dependencies (Kafka, Redis, PostgreSQL)

### Debug Commands

```bash
# Check service status
docker-compose ps

# Restart specific service
docker-compose restart loyverse-webhook

# View container logs
docker logs loyverse-webhook

# Access container shell
docker exec -it loyverse-webhook sh
```
