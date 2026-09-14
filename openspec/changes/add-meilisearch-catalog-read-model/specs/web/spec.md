## ADDED Requirements

### Requirement: Web catalog and seller lists use SearchProducts

The web service MUST render the public catalog and the authenticated seller product list from the search service SearchProducts. Public browse SHALL show catalog fields (name, price, detail links) and SHALL NOT show live stock. Seller list SHALL filter by the authenticated merchant on the search request rather than loading all listings and filtering in the web process. Product detail SHALL keep using products GetProductByID and inventory GetStock. Search unavailability SHALL use a localized catalog unavailable page and SHALL NOT fall back to products ListProducts.

#### Scenario: Public catalog is requested

- **WHEN** a browser requests the catalog route
- **THEN** the web service SHALL call search SearchProducts without a merchant filter and render catalog hits, or the catalog unavailable page if search cannot be served

#### Scenario: Browse cards omit stock

- **WHEN** the public catalog page renders listing cards
- **THEN** those cards SHALL NOT display available or reserved quantity

#### Scenario: Seller list is requested

- **WHEN** an authenticated browser requests the seller product list
- **THEN** the web service SHALL call search SearchProducts with that user's merchant filter and SHALL NOT call products ListProducts
