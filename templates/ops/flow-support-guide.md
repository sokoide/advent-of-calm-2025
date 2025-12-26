# Business Flow Support Guide

**Architecture:** {{metadata.name}}
**Generated:** {{now}}

This guide documents each business flow, the services involved, and troubleshooting steps for support teams.

---

{{#each flows}}

## {{name}}

**Description:** {{description}}

### Business Impact

| Aspect | Details |
|--------|---------|
| **Impact** | {{metadata.business-impact}} |
| **SLA** | {{metadata.sla}} |
| **Degraded Mode** | {{metadata.degraded-behavior}} |
| **Customer Message** | {{metadata.customer-communication}} |

### Flow Path

This flow traverses the following relationships:

| Step | Relationship | Description |
|------|--------------|-------------|
{{#each transitions}}
| {{sequence-number}} | `{{relationship-unique-id}}` | {{description}} |
{{/each}}

### Troubleshooting Checklist

When this flow is degraded:

1. Check the health endpoints for each service in the flow
2. Review circuit breaker status between services
3. Check message broker queue depths (if async)
4. Review recent deployments to services in this flow
5. Check database replication lag

### Escalation

If this flow is critical (tier-1), escalate immediately to the service owners.

---

{{/each}}

## Quick Reference: All Flows

| Flow | Business Impact | SLA |
|------|-----------------|-----|
{{#each flows}}
| {{name}} | {{metadata.business-impact}} | {{metadata.sla}} |
{{/each}}
