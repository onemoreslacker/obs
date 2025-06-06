package models_test

import (
	"bytes"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/stretchr/testify/require"
)

func TestDeserialize(t *testing.T) {
	tests := map[string]struct {
		update  string
		wantErr bool
	}{
		"invalid chat id field name": {
			update:  `{"chat_id": 123, "url": "http://example.com", "description": "something happened"}`,
			wantErr: true,
		},
		"invalid url field name": {
			update:  `{"chatId": 123, "urll": "http://example.com", "description": "something happened"}`,
			wantErr: true,
		},
		"invalid description field name": {
			update:  `{"chatId": 123, "url": "http://example.com", "deescription": "valid update"}`,
			wantErr: true,
		},
		"valid update": {
			update:  `{"chatId": 123, "url": "http://example.com", "description": "valid update"}`,
			wantErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var update models.KafkaUpdate
			deserializer := models.NewDeserializer()

			err := deserializer.Deserialize(bytes.NewReader([]byte(test.update)), &update)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
