# Business Flows

## Order Processing Flow

**ID:** order-processing-flow
**Purpose:** Track customer orders from placement to payment

### Steps

1. Customer submits order (Customer → API Gateway)
2. Route to order processing (API Gateway → Order Service)
3. Initiate payment (Order Service → Payment Service)

### Controls

- Transaction logging required for audit compliance

## Inventory Check Flow

**ID:** inventory-check-flow
**Purpose:** Admin checks and updates inventory stock levels

### Steps

1. Admin requests inventory status (Admin → API Gateway)
2. Route to inventory service (API Gateway → Inventory Service)
3. Query current stock (Inventory Service → Inventory Database)
4. Return stock data (response flow)
5. Return inventory report (response flow)

## Benefits

- **Business Alignment:** Maps technical architecture to business processes
- **Impact Analysis:** Understand which components are involved in each business capability
- **Compliance:** Attach specific controls to business-critical flows
- **Documentation:** Auto-generate flow diagrams and descriptions
