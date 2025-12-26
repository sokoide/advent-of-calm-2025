# Service Runbooks

Generated from architecture: {{metadata.name}}
Generated on: {{now}}

---

{{#each nodes}}
{{#if (eq node-type "service")}}

## {{name}}

**Unique ID:** `{{unique-id}}`
**Type:** {{node-type}}

### Ownership

| Field | Value |
|-------|-------|
| Owner | {{metadata.owner}} |
| On-Call Slack | {{metadata.oncall-slack}} |
| Tier | {{metadata.tier}} |
| Runbook | {{metadata.runbook}} |

### Health & Monitoring

- **Health Endpoint:** `{{metadata.health-endpoint}}`
- **Dashboard:** {{metadata.dashboard}}
- **Log Query:** `{{metadata.log-query}}`

### Dependencies

{{#if metadata.dependencies}}
This service depends on:
{{#each metadata.dependencies}}

- {{this}}
{{/each}}
{{else}}
No dependencies documented.
{{/if}}

### Known Failure Modes

{{#if metadata.failure-modes}}
{{#each metadata.failure-modes}}

#### {{symptom}}

| Aspect | Details |
|--------|---------|
| **Likely Cause** | {{likely-cause}} |
| **How to Check** | {{check}} |
| **Remediation** | {{remediation}} |
| **Escalation** | {{escalation}} |

{{/each}}
{{else}}
No failure modes documented yet.
{{/if}}

---

{{/if}}
{{/each}}

## Quick Links

| Service | Health Check | Dashboard | Runbook |
|---------|--------------|-----------|---------|
{{#each nodes}}
{{#if (eq node-type "service")}}
| {{name}} | `{{metadata.health-endpoint}}` | [Dashboard]({{metadata.dashboard}}) | [Runbook]({{metadata.runbook}}) |
{{/if}}
{{/each}}
