---
id: ecommerce-system-composition
title: Ecommerce System Composition
---

## Details
<div className="table-container">
| Field               | Value                    |
|---------------------|--------------------------|
| **Unique ID**       | ecommerce-system-composition                   |
| **Description**      |  The E-Commerce Platform comprises its core services and databases.   |
</div>

## Related Nodes
```mermaid
graph TD;
ecommerce-system -- Composed Of --> api-gateway;
ecommerce-system -- Composed Of --> order-service;
ecommerce-system -- Composed Of --> inventory-service;
ecommerce-system -- Composed Of --> payment-service;
ecommerce-system -- Composed Of --> order-db;
ecommerce-system -- Composed Of --> inventory-db;

```

## Controls
    _No controls defined._

## Metadata
  _No Metadata defined._
