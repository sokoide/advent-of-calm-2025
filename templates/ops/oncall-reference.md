# On-Call Quick Reference

**Architecture:** {{metadata.name}}
**Generated:** {{now}}

## Service Contacts

| Service | Owner | On-Call Channel | Tier |
|---------|-------|-----------------|------|
{{#each nodes}}
{{#if (eq node-type "service")}}
| {{name}} | {{metadata.owner}} | {{metadata.oncall-slack}} | {{metadata.tier}} |
{{/if}}
{{/each}}

## Database Contacts

| Database | DBA Contact | Backup Schedule | Restore Time |
|----------|-------------|-----------------|--------------|
{{#each nodes}}
{{#if (eq node-type "database")}}
| {{name}} | {{metadata.dba-contact}} | {{metadata.backup-schedule}} | {{metadata.restore-time}} |
{{/if}}
{{/each}}

## Critical Flows & Business Impact

{{#each flows}}

### {{name}}

- **Business Impact:** {{metadata.business-impact}}
- **SLA:** {{metadata.sla}}
- **Degraded Behavior:** {{metadata.degraded-behavior}}
- **Customer Communication:** {{metadata.customer-communication}}

**Flow Path:**
{{#each transitions}}
{{@index}}. {{relationship-unique-id}}
{{/each}}

---

{{/each}}

## Monitoring Links

{{#if metadata.monitoring}}

| Resource | Link |
|----------|------|
| Grafana Dashboard | {{metadata.monitoring.grafana-dashboard}} |
| Kibana Logs | {{metadata.monitoring.kibana-logs}} |
| PagerDuty | {{metadata.monitoring.pagerduty-service}} |
| Status Page | {{metadata.monitoring.statuspage}} |
{{/if}}

## Escalation Matrix

| Tier | Response Time | Escalation Path |
|------|---------------|-----------------|
| tier-1 | 15 minutes | Page immediately, all-hands |
| tier-2 | 30 minutes | Page on-call, notify manager |
| tier-3 | 2 hours | Slack notification, next business day OK |
