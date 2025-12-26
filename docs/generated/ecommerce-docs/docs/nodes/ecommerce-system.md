---
id: ecommerce-system
title: E-Commerce Platform
---

## Details
<div className="table-container">
| Field               | Value                    |
|---------------------|--------------------------|
| **Unique ID**       | ecommerce-system                   |
| **Node Type**       | system             |
| **Name**            | E-Commerce Platform                 |
| **Description**     | The overall e-commerce system containing microservices.          |

</div>

## Interfaces
    _No interfaces defined._


## Related Nodes
```mermaid
graph TD;
ecommerce-system[ecommerce-system]:::highlight;
ecommerce-system -- Composed Of --> api-gateway;
ecommerce-system -- Composed Of --> order-service;
ecommerce-system -- Composed Of --> inventory-service;
ecommerce-system -- Composed Of --> payment-service;
ecommerce-system -- Composed Of --> order-db;
ecommerce-system -- Composed Of --> inventory-db;
classDef highlight fill:#f2bbae;

```
## Controls
    _No controls defined._

## Metadata
  _No Metadata defined._
