package mongo

import (
	"context"
	"net/url"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const DefaultDatabase = "catalog"

func SetupReplicaSet(t *testing.T) *mongo.Database {
	t.Helper()

	ctx := context.Background()
	container, err := mongodb.Run(ctx, "mongo:8.0.9", mongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start mongodb replica set: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("terminate mongodb container: %v", err)
		}
	})

	rawURI, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("mongodb connection string: %v", err)
	}

	// Replica-set hello advertises the container's docker IP. From the host
	// (including Colima on macOS) that address is unreachable, so pin the
	// mapped endpoint with directConnection while still using a 1-node RS.
	u, err := url.Parse(rawURI)
	if err != nil {
		t.Fatalf("parse mongodb uri: %v", err)
	}
	q := u.Query()
	q.Del("replicaSet")
	q.Set("directConnection", "true")
	u.RawQuery = q.Encode()

	client, err := mongo.Connect(options.Client().ApplyURI(u.String()))
	if err != nil {
		t.Fatalf("connect mongodb: %v", err)
	}

	t.Cleanup(func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Fatalf("disconnect mongodb: %v", err)
		}
	})

	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping mongodb: %v", err)
	}

	return client.Database(DefaultDatabase)
}
