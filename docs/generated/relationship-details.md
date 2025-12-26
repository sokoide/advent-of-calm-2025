---
architecture: ../architectures/ecommerce-platform.json
---
# Relationship Details: E-commerce order processing platform architecture demo.

### customer-interacts-gateway


**Type:** Interaction

- **Actor:** `customer`
- **Interacts with:** `api-gateway`


---
### admin-interacts-gateway


**Type:** Interaction

- **Actor:** `admin`
- **Interacts with:** `api-gateway`


---
### gateway-connects-order

**Type:** Connection

| Property | Value |
|----------|-------|
| Source | `api-gateway` |
| Destination | `order-service` |
| Source Interfaces | `order-client` |
| Dest Interfaces | `order-api` |



---
### gateway-connects-inventory

**Type:** Connection

| Property | Value |
|----------|-------|
| Source | `api-gateway` |
| Destination | `inventory-service` |
| Source Interfaces | `inventory-client` |
| Dest Interfaces | `inventory-api` |



---
### order-connects-db

**Type:** Connection

| Property | Value |
|----------|-------|
| Source | `order-service` |
| Destination | `order-db` |
| Source Interfaces | `order-db-client` |
| Dest Interfaces | `order-sql` |



---
### order-connects-payment

**Type:** Connection

| Property | Value |
|----------|-------|
| Source | `order-service` |
| Destination | `payment-service` |
| Source Interfaces | `payment-client` |
| Dest Interfaces | `payment-api` |



---
### order-connects-inventory

**Type:** Connection

| Property | Value |
|----------|-------|
| Source | `order-service` |
| Destination | `inventory-service` |
| Source Interfaces | `order-inventory-client` |
| Dest Interfaces | `inventory-api` |



---
### inventory-connects-db

**Type:** Connection

| Property | Value |
|----------|-------|
| Source | `inventory-service` |
| Destination | `inventory-db` |
| Source Interfaces | `inventory-db-client` |
| Dest Interfaces | `inventory-sql` |



---
### ecommerce-system-composition



**Type:** Composition

- **Container:** `ecommerce-system`
- **Contains:** `api-gateway`, `order-service`, `inventory-service`, `payment-service`, `order-db`, `inventory-db`

---
