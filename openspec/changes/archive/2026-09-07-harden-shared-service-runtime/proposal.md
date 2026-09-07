# Harden shared service runtime

Keep the existing domain service boundaries and improve shared operational behavior. Kafka batch failures currently continue polling, allowing subsequent commits to acknowledge failed records. HTTP shutdown errors are hidden, gRPC draining is unbounded, and CI does not execute shared library tests.

Scope: safe Kafka restart from committed offsets, bounded server shutdown, HTTP header/idle timeouts that preserve streaming responses, and shared test/dependency coverage in CI. No API, schema, or service-boundary changes.
