# On-Call Quick Reference

**Architecture:** 
**Generated:** 

## Service Contacts

| Service | Owner | On-Call Channel | Tier |
|---------|-------|-----------------|------|
| Load Balancer | platform-team | #oncall-platform | tier-1 |
| API Gateway Instance 1 | platform-team | #oncall-platform | tier-1 |
| API Gateway Instance 2 | platform-team | #oncall-platform | tier-1 |
| Order Service | orders-team | #oncall-orders | tier-1 |
| Inventory Service | inventory-team | #oncall-inventory | tier-2 |
| Payment Service | payments-team | #oncall-payments | tier-1 |

## Database Contacts

| Database | DBA Contact | Backup Schedule | Restore Time |
|----------|-------------|-----------------|--------------|
| Order Database (Primary) | dba-team@example.com | daily at 02:00 UTC | 60 minutes |
| Order Database (Replica) | dba-team@example.com | daily at 02:00 UTC | 60 minutes |
| Inventory Database | dba-team@example.com | weekly at Sunday 03:00 UTC | 30 minutes |

## Critical Flows & Business Impact


### Customer Order Processing

- **Business Impact:** Customers cannot complete purchases - direct revenue loss
- **SLA:** 99.9% availability, 30s p99 latency
- **Degraded Behavior:** Orders queue in message broker; processed when service recovers
- **Customer Communication:** Display &#x27;Order processing delayed&#x27; message

**Flow Path:**
0. customer-interacts-lb
1. lb-connects-gateway-1
2. gateway-1-connects-order
3. order-publishes-to-queue
4. payment-subscribes-to-queue

---


### Inventory Stock Check

- **Business Impact:** Stock levels may be inaccurate - risk of overselling
- **SLA:** 99.5% availability, 500ms p99 latency
- **Degraded Behavior:** Fall back to cached inventory; flag orders for manual review
- **Customer Communication:** Display &#x27;Stock availability being confirmed&#x27;

**Flow Path:**
0. admin-interacts-lb
1. lb-connects-gateway-2
2. gateway-2-connects-inventory
3. inventory-connects-db
4. inventory-connects-db

---


## Monitoring Links


| Resource | Link |
|----------|------|
| Grafana Dashboard | https://grafana.example.com/d/ecommerce-overview |
| Kibana Logs | https://kibana.example.com/app/discover#/ecommerce-* |
| PagerDuty | https://pagerduty.example.com/services/ECOMMERCE |
| Status Page | https://status.example.com |

## Escalation Matrix

| Tier | Response Time | Escalation Path |
|------|---------------|-----------------|
| tier-1 | 15 minutes | Page immediately, all-hands |
| tier-2 | 30 minutes | Page on-call, notify manager |
| tier-3 | 2 hours | Slack notification, next business day OK |
