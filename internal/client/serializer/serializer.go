package serializer

import (
	"encoding/json"
	"github.com/goletan/events-service/shared/types"
	"github.com/pkg/errors"
)

// Serializer handles event serialization and deserialization.
type Serializer struct{}

// NewSerializer creates a new Serializer instance.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// SerializeEvent encodes an Event into a JSON string.
func (s *Serializer) SerializeEvent(event *types.Event) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize event")
	}
	return data, nil
}

// DeserializeEvent decodes JSON data into an Event.
func (s *Serializer) DeserializeEvent(data []byte) (*types.Event, error) {
	var event types.Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, errors.Wrap(err, "failed to deserialize event")
	}
	return &event, nil
}
