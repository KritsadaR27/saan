package events

import (
	"encoding/json"
	"fmt"
	"time"
)

type Transformer struct{}

func NewTransformer() *Transformer {
	return &Transformer{}
}

func (t *Transformer) TransformProduct(input []byte) (DomainEvent, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return DomainEvent{}, fmt.Errorf("invalid product payload: %w", err)
	}

	event := DomainEvent{
		ID:            fmt.Sprintf("loyverse_product_%v", time.Now().UnixNano()),
		Type:          EventProductUpdated,
		AggregateID:   fmt.Sprintf("%v", data["id"]),
		AggregateType: "product",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          input,
		Source:        "loyverse",
	}
	return event, nil
}
