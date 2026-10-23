-- +goose Up
CREATE TABLE users_outbox (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL REFERENCES users(id),
    event_type TEXT NOT NULL,
    payload BYTEA NOT NULL,
    tracingspancontext TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementBegin
CREATE FUNCTION PUBLISH_ACCOUNT_CREATED() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO users_outbox (id, aggregate_id, event_type, payload)
    VALUES (
        NEW.id, NEW.id, 'users.account-created.v1',
        CONVERT_TO(
            JSON_BUILD_OBJECT(
                'user_id', NEW.id, 'created_at', NEW.created_at
            )::TEXT, 'UTF8'
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER users_account_created
AFTER INSERT ON users
FOR EACH ROW EXECUTE FUNCTION PUBLISH_ACCOUNT_CREATED();

INSERT INTO users_outbox (id, aggregate_id, event_type, payload)
SELECT
    id,
    id AS aggregate_id,
    'users.account-created.v1' AS event_type,
    CONVERT_TO(
        JSON_BUILD_OBJECT('user_id', id, 'created_at', created_at)::TEXT,
        'UTF8'
    ) AS payload
FROM users;

-- +goose Down
DROP TRIGGER users_account_created ON users;
DROP FUNCTION PUBLISH_ACCOUNT_CREATED();
DROP TABLE users_outbox;
