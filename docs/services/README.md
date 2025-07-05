# 🔧 SAAN Services Documentation

This directory contains current API documentation, integration guides, and troubleshooting information for all SAAN microservices.

## � **Documentation Structure**

Each service has its own directory with three key documents:

### **Per-Service Documentation**
```
services/
├── {service-name}/
│   ├── API.md              # API endpoints, request/response examples
│   ├── INTEGRATION.md      # How to integrate with this service
│   └── TROUBLESHOOTING.md  # Common issues and solutions
└── SERVICE_INTEGRATION_MAP.md  # Cross-service communication map
```

### **Available Services**
- **[order/](./order/)** - Central order orchestration and workflow management
- **[product/](./product/)** - Product catalog, pricing, and availability  
- **[customer/](./customer/)** - Customer management and VIP handling
- **[payment/](./payment/)** - Payment processing and Loyverse integration
- **[shipping/](./shipping/)** - Delivery management and third-party logistics
- **[inventory/](./inventory/)** - Stock management and availability tracking
- **[finance/](./finance/)** - Financial reporting and transaction recording
- **[chat/](./chat/)** - AI chat and customer interaction

## 🗺️ **Service Integration Overview**

See **[SERVICE_INTEGRATION_MAP.md](./SERVICE_INTEGRATION_MAP.md)** for:
- Service communication patterns
- Event-driven workflows  
- Redis caching strategies
- Error handling patterns
- Troubleshooting procedures

## 🎯 **Documentation Standards**
### **API.md Requirements**
- Base URL and port information
- All endpoint definitions with examples
- Request/response schemas
- Error codes and messages
- Authentication requirements
- Rate limiting information

### **INTEGRATION.md Requirements**  
- Services that this service depends on
- Services that call this service
- Event publishing and consumption
- Caching patterns and Redis keys
- Error handling and fallback strategies
- Performance considerations

### **TROUBLESHOOTING.md Requirements**
- Common error scenarios with solutions
- Performance debugging steps
- Cache and database troubleshooting
- Service dependency issues
- Emergency recovery procedures

## 📊 **Service Status & Implementation**

| Service | Port | Status | API Docs | Integration | Troubleshooting | Module Name |
|---------|------|--------|----------|-------------|-----------------|-------------|
| Order | 8081 | ✅ Live | ✅ [API](./order/API.md) | ✅ [Integration](./order/INTEGRATION.md) | ✅ [Troubleshooting](./order/TROUBLESHOOTING.md) | `order` |
| Product | 8083 | ✅ Live | ✅ [API](./product/API.md) | 🔄 In Progress | 🔄 In Progress | `product` |
| Customer | 8110 | ✅ Live | ✅ [API](./customer/API.md) | 🔄 In Progress | 🔄 In Progress | `customer` |
| Payment | 8087 | ✅ Live | 🔄 In Progress | 🔄 In Progress | 🔄 In Progress | `payment` |
| Shipping | 8086 | ✅ Live | 🔄 In Progress | 🔄 In Progress | 🔄 In Progress | `shipping` |
| Inventory | 8082 | ✅ Live | 🔄 In Progress | 🔄 In Progress | 🔄 In Progress | `inventory` |
| Finance | 8088 | ✅ Live | 🔄 In Progress | 🔄 In Progress | 🔄 In Progress | `finance` |
| Chat | 8090 | ✅ Live | 🔄 In Progress | 🔄 In Progress | 🔄 In Progress | `chat` |

## 🏗️ **Module Name Standardization**

All services use simplified module names for better developer experience:

```go
// Before: Long, complex module names
import "github.com/saan/payment-service/internal/domain/entity"

// After: Clean, simple module names  
import "payment/internal/domain/entity"
```

### **Benefits:**
- **Cleaner imports**: Shorter, more readable import statements
- **Faster development**: Less typing, easier to remember
- **Consistent structure**: All services follow the same naming pattern
- **Local development friendly**: Works seamlessly in monorepo setup

## 🔄 **Adding New Services**

When implementing new services:

1. **Create service directory**: `mkdir services/{service-name}`
2. **Add documentation**: Create API.md, INTEGRATION.md, TROUBLESHOOTING.md
3. **Update integration map**: Add to SERVICE_INTEGRATION_MAP.md
4. **Update status table**: Add row to table above
5. **Follow naming convention**: Use simple module name in go.mod

### **Documentation Template**
```bash
# Copy from existing service for consistency
cp -r services/order services/new-service
# Update content for new service specifics
```

## � **Archived Documentation**

Historical implementation plans and completion reports have been moved to:
- **[/docs/archive/implementation-plans/](../archive/implementation-plans/)** - Original planning documents
- **[/docs/archive/service-reports/](../archive/service-reports/)** - Implementation completion reports

## 🔍 **Quick Navigation**

### **For Developers**
- 🔗 [Service Integration Map](./SERVICE_INTEGRATION_MAP.md) - Understanding service communication
- 📝 [Order API](./order/API.md) - Most commonly used service
- 🛠️ [Order Troubleshooting](./order/TROUBLESHOOTING.md) - Common issues and fixes

### **For DevOps/SRE** 
- 🚨 [Troubleshooting Guides](./*/TROUBLESHOOTING.md) - Service-specific debugging
- 🔄 [Integration Dependencies](./*/INTEGRATION.md) - Service dependency mapping
- 📊 [Health Check Endpoints](./SERVICE_INTEGRATION_MAP.md#monitoring--alerts) - Service monitoring

### **For Product/Business**
- 📖 [API Documentation](./*/API.md) - Understanding service capabilities
- 🗺️ [Integration Map](./SERVICE_INTEGRATION_MAP.md) - Business workflow visualization

---

> 🔧 **Current documentation focuses on operational excellence, integration, and maintenance rather than implementation planning**
