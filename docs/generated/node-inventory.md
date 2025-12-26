---
architecture: ../architectures/ecommerce-platform.json
---
# Node Inventory: E-commerce order processing platform architecture demo.

| Name | Type | ID | Description |
|------|------|-------|-------------|
| Customer | Actor | `customer` | A user who browses and purchases products. |
| Admin | Actor | `admin` | A staff member who manages products and orders. |
| E-Commerce Platform | System | `ecommerce-system` | The overall e-commerce system containing microservices. |
| API Gateway | Service | `api-gateway` | Entry point for all client requests. |
| Order Service | Service | `order-service` | Handles order creation and lifecycle management. |
| Inventory Service | Service | `inventory-service` | Manages product stock levels. |
| Payment Service | Service | `payment-service` | Integrates with external payment providers. |
| Order Database | Database | `order-db` | Relational database for orders. |
| Inventory Database | Database | `inventory-db` | Stores stock levels. |

## Services Only


- **API Gateway** (`api-gateway`): Entry point for all client requests.

- **Order Service** (`order-service`): Handles order creation and lifecycle management.

- **Inventory Service** (`inventory-service`): Manages product stock levels.

- **Payment Service** (`payment-service`): Integrates with external payment providers.

## Databases Only

---
architecture: ../architectures/ecommerce-platform.json
---
# Relationship Details: 

- **Order Database** (`order-db`): Relational database for orders.
---
architecture: ../architectures/ecommerce-platform.json
---
# Relationship Details: 

- **Inventory Database** (`inventory-db`): Stores stock levels.

## Actors


- **Customer**: A user who browses and purchases products.

- **Admin**: A staff member who manages products and orders.
