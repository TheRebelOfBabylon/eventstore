package binary

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"

	"github.com/mailru/easyjson"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/test_common"
)

func BenchmarkBinaryDecoding(b *testing.B) {
	events := make([][]byte, len(test_common.NormalEvents))
	gevents := make([][]byte, len(test_common.NormalEvents))
	for i, jevt := range test_common.NormalEvents {
		evt := &nostr.Event{}
		json.Unmarshal([]byte(jevt), evt)
		bevt, _ := Marshal(evt)
		events[i] = bevt

		var buf bytes.Buffer
		gob.NewEncoder(&buf).Encode(evt)
		gevents[i] = buf.Bytes()
	}

	b.Run("easyjson.Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, jevt := range test_common.NormalEvents {
				evt := &nostr.Event{}
				err := easyjson.Unmarshal([]byte(jevt), evt)
				if err != nil {
					b.Fatalf("failed to unmarshal: %s", err)
				}
			}
		}
	})

	b.Run("gob.Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, gevt := range gevents {
				evt := &nostr.Event{}
				buf := bytes.NewBuffer(gevt)
				evt = &nostr.Event{}
				gob.NewDecoder(buf).Decode(evt)
			}
		}
	})

	b.Run("binary.Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, bevt := range events {
				evt := &nostr.Event{}
				err := Unmarshal(bevt, evt)
				if err != nil {
					b.Fatalf("failed to unmarshal: %s", err)
				}
			}
		}
	})

	b.Run("easyjson.Unmarshal+sig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, nevt := range test_common.NormalEvents {
				evt := &nostr.Event{}
				err := easyjson.Unmarshal([]byte(nevt), evt)
				if err != nil {
					b.Fatalf("failed to unmarshal: %s", err)
				}
				evt.CheckSignature()
			}
		}
	})

	b.Run("binary.Unmarshal+sig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, bevt := range events {
				evt := &nostr.Event{}
				err := Unmarshal(bevt, evt)
				if err != nil {
					b.Fatalf("failed to unmarshal: %s", err)
				}
				evt.CheckSignature()
			}
		}
	})
}
