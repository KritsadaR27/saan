 🛠 SAAN PROJECT RULES
## ❌ DO NOT:
- Run services directly with go run or npm run dev
- Install dependencies on your host machine
- Use localhost in your code or API calls
- Create or modify Dockerfiles without team review
## ✅ ALWAYS USE:
- docker-compose up to run the project
- docker-compose logs -f [service] to view service logs
- docker exec -it [container_name] sh to enter a container
- Service names (not localhost) for internal URLs
---
## 🧩 SERVICES & PORTS
| Service             | Description                        | Port      | Docker Container Name |
|---------------------|------------------------------------|-----------|------------------------|
| Web App             | Customer web frontend (Next.js)    | 3008      | web                  |
| Admin Dashboard     | Internal admin panel               | 3010      | admin                |
| Chatbot Service     | AI / Rule-based reply engine       | 8090      | chatbot              |
| Webhook Listener    | Facebook / LINE webhook endpoint   | 8091      | webhook              |
| Order Service       | Manages all order logic            | 8081      | order-service        |
| Inventory Service   | Manages stock and warehouse        | 8082      | inventory-service    |
| Product Service     | Catalog / SKU management           | 8083      | product-service      |
| Sale Service        | Sale entry & revenue API           | 8084      | sale-service         |
| Finance Service     | Profit / accounting                | 8085      | finance-service      |
| Shipping Service    | Delivery & routing logic           | 8086      | shipping-service     |
| Payment Service     | Payment verification & QR          | 8087      | payment-service      |
| User Service        | Customer profile & member system   | 8088      | user-service         |
| Reporting Service   | Analytics & dashboard data         | 8089      | reporting-service    |
| Notification Service| LINE, FB, Email push notifications | 8092      | notification-service |
| PostgreSQL Database | Shared relational database         | 5532      | postgres             |
| Kafka (Message Bus) | Event queue system                 | 9092      | kafka                |
---
## 🔄 INTERNAL COMMUNICATION RULES
Use service names as hostnames for API calls and DB access:
### ✅ Examples:
- http://order-service:8081/api/orders
- http://inventory-service:8082/api/stock
- http://chatbot:8090/process
- postgres://postgres:5432 (not localhost:5532)
- Kafka: kafka:9092
---
## ⚙️ DEVELOPMENT TIPS
- Use .env files per service for environment variables
- Use air or reflex in Go services for hot-reload
- Map ports only when external access is needed
- Keep docker-compose.override.yml for local dev tweaks
---
> ✅ Stick to these rules and we guarantee a smooth Dev → Test → Prod transition.
