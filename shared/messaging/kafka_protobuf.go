package messaging

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

const confluentWireMagic byte = 0

func UnmarshalKafkaProtobuf(data []byte, m proto.Message) error {
	if len(data) == 0 {
		return fmt.Errorf("empty payload")
	}
	body := data
	if len(data) >= 5 && data[0] == confluentWireMagic {
		body = data[5:]
	}
	if err := proto.Unmarshal(body, m); err != nil {
		return fmt.Errorf("protobuf decode: %w", err)
	}
	return nil
}
