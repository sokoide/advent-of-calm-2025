---
id: index
title: Welcome to CALM Documentation
sidebar_position: 1
slug: /
---

# Welcome to CALM Documentation

This documentation is generated from the **CALM Architecture-as-Code** model.

## High Level Architecture
```mermaid
C4Deployment

    Deployment_Node(deployment, "Architecture", ""){
        Person(customer, "Customer", "A user who browses and purchases products.")
        Person(admin, "Admin", "A staff member who manages products and orders.")
        Deployment_Node(ecommerce-system, "E-Commerce Platform", "The overall e-commerce system containing microservices."){
            Container(api-gateway, "API Gateway", "", "Entry point for all client requests.")
            Container(order-service, "Order Service", "", "Handles order creation and lifecycle management.")
            Container(inventory-service, "Inventory Service", "", "Manages product stock levels.")
            Container(payment-service, "Payment Service", "", "Integrates with external payment providers.")
            Container(order-db, "Order Database", "", "Relational database for orders.")
            Container(inventory-db, "Inventory Database", "", "Stores stock levels.")
        }
    }

    Rel(customer,api-gateway,"Interacts With")
    Rel(admin,api-gateway,"Interacts With")
    Rel(api-gateway,order-service,"Connects To")
    Rel(api-gateway,inventory-service,"Connects To")
    Rel(order-service,order-db,"Connects To")
    Rel(order-service,payment-service,"Connects To")
    Rel(order-service,inventory-service,"Connects To")
    Rel(inventory-service,inventory-db,"Connects To")

    UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="2")
```
## Nodes
    - [Customer](nodes/customer)
    - [Admin](nodes/admin)
    - [E-Commerce Platform](nodes/ecommerce-system)
    - [API Gateway](nodes/api-gateway)
    - [Order Service](nodes/order-service)
    - [Inventory Service](nodes/inventory-service)
    - [Payment Service](nodes/payment-service)
    - [Order Database](nodes/order-db)
    - [Inventory Database](nodes/inventory-db)

## Relationships
    - [Customer Interacts Gateway](relationships/customer-interacts-gateway)
    - [Admin Interacts Gateway](relationships/admin-interacts-gateway)
    - [Gateway Connects Order](relationships/gateway-connects-order)
    - [Gateway Connects Inventory](relationships/gateway-connects-inventory)
    - [Order Connects Db](relationships/order-connects-db)
    - [Order Connects Payment](relationships/order-connects-payment)
    - [Order Connects Inventory](relationships/order-connects-inventory)
    - [Inventory Connects Db](relationships/inventory-connects-db)
    - [Ecommerce System Composition](relationships/ecommerce-system-composition)


## Flows
    - [Customer Order Processing](flows/order-processing-flow)
    - [Inventory Stock Check](flows/inventory-check-flow)

## Controls
| Requirement URL               | Category    | Scope        | Applied To                |
|-------------------------------|-----------|--------------|---------------------------|
|https://internal-policy.example.com/performance/rate-limiting|performance|Node|api-gateway|
|https://www.pcisecuritystandards.org/documents/PCI-DSS-v4.0|compliance|Node|payment-service|

## Metadata
  <div className="table-container">
      <table>
          <thead>
          <tr>
              <th>Key</th>
              <th>Value</th>
          </tr>
          </thead>
          <tbody>
          <tr>
              <td>
                  <b>Owner</b>
              </td>
              <td>
                  Architecture Team
                      </td>
          </tr>
          <tr>
              <td>
                  <b>Version</b>
              </td>
              <td>
                  1.0.0
                      </td>
          </tr>
          <tr>
              <td>
                  <b>Created</b>
              </td>
              <td>
                  2025-12-26
                      </td>
          </tr>
          <tr>
              <td>
                  <b>Description</b>
              </td>
              <td>
                  E-commerce order processing platform architecture demo.
                      </td>
          </tr>
          <tr>
              <td>
                  <b>Tags</b>
              </td>
              <td>
                  <ul>
                      <li>ecommerce</li>
                      <li>microservices</li>
                      <li>orders</li>
                  </ul>
              </td>
          </tr>
          </tbody>
      </table>
  </div>

## Adrs
- [docs/adr/0001-use-message-queue-for-async-processing.md](docs/adr/0001-use-message-queue-for-async-processing.md)
- [docs/adr/0002-use-oauth2-for-api-authentication.md](docs/adr/0002-use-oauth2-for-api-authentication.md)
