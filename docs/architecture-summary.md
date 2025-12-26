---
architecture: ../architectures/ecommerce-platform.json
---

# Architecture Summary

## System Overview

{{block-architecture this}}

## Nodes

{{table nodes columns="unique-id,name,node-type,description"}}

## Order Processing Flow View

{{block-architecture this focus-flows="order-processing-flow"}}

## Payment Processing Components

{{block-architecture this focus-nodes="payment-service,order-service" highlight-nodes="payment-service" render-node-type-shapes=true}}

## Order Processing Sequence

{{flow-sequence this flow-id="order-processing-flow"}}

## API Gateway Connections

{{related-nodes node-id="api-gateway"}}
