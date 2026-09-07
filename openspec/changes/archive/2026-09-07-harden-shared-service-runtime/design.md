# Design

Return Kafka handler and commit errors before polling again. The runtime recreates consumers after a cancellation-aware five-second delay; existing call sites close each consumer before returning. Successful messages from an unsuccessful batch may replay, so existing inbox/idempotency handling remains essential. Poison messages require operational intervention; a durable dead-letter policy is a separate domain decision.

Drain gRPC for a configurable timeout (30 seconds by default), then force stop. Wait for draining before releasing service dependencies. HTTP uses a five-second header timeout and 60-second idle timeout; no global write timeout is imposed on streaming endpoints. If HTTP draining expires, close connections and return the error.

Conservatively run every service test for any shared/workspace/CI change. This trades some CI time for correct transitive dependency coverage. Add an independent race-enabled shared test job.

Deferred: readiness tied to dependency/consumer health, unified process error propagation instead of fatal exits, bounded telemetry cleanup, Kafka rebalance ownership and poison-message policy, workload-driven database/index tuning. These require separate validation beyond this runtime refactor.
