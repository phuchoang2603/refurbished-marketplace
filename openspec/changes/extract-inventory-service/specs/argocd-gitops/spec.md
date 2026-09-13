## ADDED Requirements

### Requirement: Inventory workload in marketplace chart

The marketplace Helm release SHALL deploy an `inventory` Service and Deployment in `ecommerce` with GHCR image tags matching other marketplace services.

#### Scenario: Inventory syncs with the chart

- **WHEN** the marketplace Application syncs
- **THEN** an inventory Deployment and Service exist in `ecommerce`
