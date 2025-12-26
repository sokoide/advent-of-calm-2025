---
id: inventory-check-flow
title: Inventory Stock Check
---

## Details
<div className="table-container">
| Field               | Value                    |
|---------------------|--------------------------|
| **Unique ID**       | inventory-check-flow                   |
| **Name**            | Inventory Stock Check                 |
| **Description**     | Admin checks and updates inventory stock levels          |
</div>

## Sequence Diagram
```mermaid
sequenceDiagram
            Admin ->> API Gateway: Admin requests inventory status
            API Gateway ->> Inventory Service: Route to inventory service
            Inventory Service ->> Inventory Database: Query current stock levels
            Inventory Database -->> Inventory Service: Return stock data
            Inventory Service -->> API Gateway: Return inventory report
```
## Controls
    _No controls defined._

## Metadata
  _No Metadata defined._
