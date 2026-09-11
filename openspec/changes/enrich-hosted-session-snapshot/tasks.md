## 1. Contract

- [ ] 1.1 Add nested buyer/merchant snapshot messages (`id`, optional buyer `email`) and `name` on `HostedPaymentLineItem`
- [ ] 1.2 Regenerate payment protobufs

## 2. Payment persistence

- [ ] 2.1 Migration + sqlc: persist nested buyer/merchant JSON; persist `name` inside `line_items` JSON
- [ ] 2.2 Require buyer id, merchant id, and usable shipping on create; do not call users
- [ ] 2.3 Insert session+tx with the snapshot; PENDING create stays idempotent without rewriting the snapshot

## 3. Web

- [ ] 3.1 Render postal shipping fields on each merchant-group checkout form
- [ ] 3.2 Reject checkout before place-order when line1, city, postal code, or country is missing
- [ ] 3.3 Send nested buyer/merchant ids, optional JWT `eml`, shipping, and product names on `CreateHostedPaymentSession`
- [ ] 3.4 Leave hosted-pay query string unchanged (no PII)

## 4. Existing tests

- [ ] 4.1 Update payment create tests for required shipping and nested snapshot persistence (no users client)
- [ ] 4.2 Update web checkout tests for shipping validation and nested session request fields
