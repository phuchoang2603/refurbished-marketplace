package trace

import (
	"context"
	"testing"
)

func TestInitMergesDefaultResourceSchema(t *testing.T) {
	shutdown, err := Init(context.Background(), Config{
		ServiceName: "web",
		Endpoint:    "localhost:4317",
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(func() {
		_ = shutdown(context.Background())
	})
}
