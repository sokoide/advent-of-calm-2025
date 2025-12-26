# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the e-commerce platform.

## Format

We use **MADR (Markdown Any Decision Records)** format, based on [the MADR template](https://adr.github.io/madr/):

- Title and date
- Status (Proposed, Accepted, Deprecated, Superseded)
- Context (the situation and problem)
- Decision (what was decided)
- Consequences (positive, negative, and mitigations)

## Linking to CALM

All ADRs in this directory are linked from `architectures/ecommerce-platform.json` in the `adrs` array. This creates traceability between decisions and implementation.

## Index

### Active

| ADR | Title | Date |
|-----|-------|------|
| [0001](0001-use-message-queue-for-async-processing.md) | Use Message Queue for Asynchronous Order Processing | 2024-12-15 |
| [0002](0002-use-oauth2-for-api-authentication.md) | Use OAuth2 for API Authentication | 2024-12-15 |

### Superseded
None yet.

## Creating New ADRs

Use the numbering sequence: 0003, 0004, etc.

Filename format: `NNNN-short-title-with-hyphens.md`

Link the ADR in `architectures/ecommerce-platform.json` in the `adrs` array.

## Benefits

1. **Traceability:** Link decisions to architecture implementation
2. **Onboarding:** New team members understand "why" not just "what"
3. **Auditing:** Decision history for compliance and reviews
4. **Evolution:** Track how architecture decisions change over time
