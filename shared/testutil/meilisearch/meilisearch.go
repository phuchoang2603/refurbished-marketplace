package meilisearch

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	Image     = "getmeili/meilisearch:v1.53.1"
	MasterKey = "integration-master-key"
)

type Container struct {
	testcontainers.Container
	URL       string
	MasterKey string
}

func Setup(t *testing.T) *Container {
	t.Helper()

	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        Image,
			ExposedPorts: []string{"7700/tcp"},
			Env: map[string]string{
				"MEILI_MASTER_KEY":   MasterKey,
				"MEILI_NO_ANALYTICS": "true",
			},
			WaitingFor: wait.ForHTTP("/health").WithPort("7700/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start meilisearch container: %v", err)
	}

	t.Cleanup(func() {
		if err := c.Terminate(ctx); err != nil {
			t.Fatalf("terminate meilisearch container: %v", err)
		}
	})

	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("meilisearch host: %v", err)
	}
	port, err := c.MappedPort(ctx, "7700/tcp")
	if err != nil {
		t.Fatalf("meilisearch port: %v", err)
	}

	return &Container{
		Container: c,
		URL:       fmt.Sprintf("http://%s:%s", host, port.Port()),
		MasterKey: MasterKey,
	}
}
