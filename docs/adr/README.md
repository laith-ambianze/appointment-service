# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the Appointment Service project.

## What is an ADR?

An ADR is a document that captures an important architectural decision made along with its context and consequences.

## ADR List

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [001](ADR-001-database-access-layer.md) | Database Access Layer | Accepted | 2026-01-31 |
| [002](ADR-002-api-framework.md) | API Framework | Accepted | 2026-01-31 |
| [003](ADR-003-logging.md) | Logging Strategy | Accepted | 2026-01-31 |
| [004](ADR-004-configuration.md) | Configuration Management | Accepted | 2026-01-31 |
| [005](ADR-005-testing.md) | Testing Strategy | Accepted | 2026-01-31 |

## Status Definitions

- **Proposed**: Decision is proposed and under review
- **Accepted**: Decision has been approved and is in effect
- **Deprecated**: Decision is no longer recommended
- **Superseded**: Decision has been replaced by a newer ADR

## Creating New ADRs

1. Copy `TEMPLATE.md`
2. Name it `ADR-XXX-title.md` (XXX = next number)
3. Fill in all sections
4. Add to this index
5. Commit with message: `docs: add ADR-XXX for [title]`
