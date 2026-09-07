## Implementation

- [x] Stop Kafka polling after batch errors and restart closed consumers with delay.
- [x] Bound gRPC draining and propagate HTTP shutdown failures.
- [x] Add HTTP header and idle timeouts without limiting streaming response duration.
- [x] Cover shared modules and transitive dependency changes in CI.
- [x] Add regression tests for Kafka batch failures and gRPC cancellation.
- [x] Execute shared regression tests with the race detector (passed).
- [ ] Compile all services and tests (execution permission declined; not run).
