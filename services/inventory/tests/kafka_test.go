package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	testkafka "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/kafka"
	"google.golang.org/protobuf/proto"
)

func TestKafkaProductCreatedIndependentConsumers(t *testing.T) {
	db := newInventoryDB(t)
	svc := service.New(db)
	event := creationEvent(proto.Int32(5))
	payload, err := proto.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	topic := messaging.EventTypeProductCreated
	testkafka.ProduceKafkaRecord(t, t.Context(), brokers, topic, payload)
	cancel, runErr := testkafka.StartKafkaConsumer(t, t.Context(), brokers, "inventory-created-"+uuid.NewString(), []string{topic}, svc.KafkaProductCreatedHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, runErr, cancel, 30*time.Second, 200*time.Millisecond, "creation was not consumed", func() (bool, error) {
		var count int
		err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM inventory_inbox WHERE message_id=$1", topic+"/"+event.EventId).Scan(&count)
		return count == 1, err
	})
	assertStock(t, svc, uuid.MustParse(event.ProductId), 5, 0)
	// A later independent group must still receive the original creation event.
	observed := make(chan []byte, 1)
	observerCancel, observerErr := testkafka.StartKafkaConsumer(t, t.Context(), brokers, "catalog-observer-"+uuid.NewString(), []string{topic}, func(_ context.Context, msg messaging.KafkaMessage) error {
		select {
		case observed <- msg.Value:
		default:
		}
		return nil
	})
	defer observerCancel()
	testkafka.WaitForKafkaCondition(t, observerErr, observerCancel, 30*time.Second, 200*time.Millisecond, "independent group did not receive creation", func() (bool, error) {
		select {
		case value := <-observed:
			var got productsv1.ProductCreated
			if err := messaging.UnmarshalKafkaProtobuf(value, &got); err != nil {
				return false, err
			}
			if !proto.Equal(event, &got) {
				return false, fmt.Errorf("independent consumer received different event")
			}
			return true, nil
		default:
			return false, nil
		}
	})
}
