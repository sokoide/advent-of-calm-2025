---
architecture: ../architectures/ecommerce-platform.json
---
# Relationship Details: {{metadata.description}}

{{#each relationships}}
### {{unique-id}}

{{#if relationship-type.connects}}
**Type:** Connection

| Property | Value |
|----------|-------|
| Source | `{{relationship-type.connects.source.node}}` |
| Destination | `{{relationship-type.connects.destination.node}}` |
{{#if relationship-type.connects.source.interfaces}}
| Source Interfaces | {{#each relationship-type.connects.source.interfaces}}`{{this}}`{{#unless @last}}, {{/unless}}{{/each}} |
{{/if}}
{{#if relationship-type.connects.destination.interfaces}}
| Dest Interfaces | {{#each relationship-type.connects.destination.interfaces}}`{{this}}`{{#unless @last}}, {{/unless}}{{/each}} |
{{/if}}
{{/if}}

{{#if relationship-type.interacts}}
**Type:** Interaction

- **Actor:** `{{relationship-type.interacts.actor}}`
- **Interacts with:** {{#each relationship-type.interacts.nodes}}`{{this}}`{{#unless @last}}, {{/unless}}{{/each}}
{{/if}}

{{#if relationship-type.composed-of}}
**Type:** Composition

- **Container:** `{{relationship-type.composed-of.container}}`
- **Contains:** {{#each relationship-type.composed-of.nodes}}`{{this}}`{{#unless @last}}, {{/unless}}{{/each}}
{{/if}}

---
{{/each}}
