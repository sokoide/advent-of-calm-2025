# CALM Controls Guide

## Purpose

Controls document non-functional requirements (NFRs) including security, compliance, performance, and operational needs in architecture.

## Control Domains

| Domain | Purpose | Example Requirements |
|--------|---------|---------------------|
| security | Data protection, access control | Encryption, TLS, authentication |
| compliance | Regulatory adherence | PCI-DSS, GDPR, SOC2 |
| performance | Non-functional requirements | SLAs, rate limits, availability |
| operational | Runtime concerns | Logging, monitoring, backup |

## Controls in This Architecture

### Architecture-Level Controls

**Security**

- Encryption at rest: <https://internal-policy.example.com/security/encryption-at-rest> (inline config)
- TLS 1.3 minimum: <https://configs.example.com/security/tls-config.yaml>

**Performance**

- Response time SLA: <https://internal-policy.example.com/performance/response-time-sla> (inline config)
- Availability target: <https://configs.example.com/infra/ha-config.yaml>

### Node-Level Controls

**Payment Service - Compliance**

- PCI-DSS: <https://www.pcisecuritystandards.org/documents/PCI-DSS-v4.0>
- Configuration: <https://configs.example.com/compliance/pci-dss-config.json>

**API Gateway - Performance**

- Rate limiting: <https://configs.example.com/gateway/rate-limits.yaml>
- Caching policy: <https://internal-policy.example.com/performance/caching-policy> (inline config)

## Benefits

1. **Audit Trail:** Links architecture to compliance requirements
2. **NFR Tracking:** Makes non-functional requirements explicit and measurable
3. **Traceability:** Connects technical implementation to policy
4. **SLA Documentation:** Makes performance requirements explicit and trackable
