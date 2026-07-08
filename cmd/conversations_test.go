package cmd

import (
	"encoding/json"
	"testing"
)

// representativeThreadMessages mirrors the real HubSpot Conversations v3 shape
// for GET /conversations/v3/conversations/threads/{id}/messages, including
// deliveryIdentifier objects and the non-MESSAGE entry types that carry empty
// senders/recipients.
const representativeThreadMessages = `{
  "results": [
    {
      "id": "1",
      "type": "MESSAGE",
      "direction": "OUTGOING",
      "senders": [
        { "actorId": "A-123", "deliveryIdentifier": { "type": "HS_EMAIL_ADDRESS", "value": "support@edcontrols.nl" } }
      ],
      "recipients": [
        { "recipientField": "TO", "deliveryIdentifier": { "type": "HS_EMAIL_ADDRESS", "value": "customer@example.com" } }
      ],
      "text": "Hello from support",
      "channelId": "1002",
      "channelAccountId": "555"
    },
    {
      "id": "2",
      "type": "MESSAGE",
      "direction": "INCOMING",
      "senders": [
        { "actorId": "V-999", "deliveryIdentifier": { "type": "HS_EMAIL_ADDRESS", "value": "customer@example.com" } }
      ],
      "recipients": [
        { "recipientField": "TO", "deliveryIdentifier": { "type": "HS_EMAIL_ADDRESS", "value": "support@edcontrols.nl" } }
      ],
      "text": "Reply from customer"
    },
    {
      "id": "3",
      "type": "ASSIGNMENT",
      "senders": [],
      "recipients": []
    },
    {
      "id": "4",
      "type": "THREAD_STATUS_CHANGE",
      "senders": [],
      "recipients": []
    }
  ]
}`

func TestUnmarshalThreadMessages_DeliveryIdentifierObject(t *testing.T) {
	var resp threadMessagesResponse
	if err := json.Unmarshal([]byte(representativeThreadMessages), &resp); err != nil {
		t.Fatalf("failed to unmarshal thread messages: %v", err)
	}

	if len(resp.Results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(resp.Results))
	}

	outgoing := resp.Results[0]
	if got := outgoing.Senders[0].DeliveryIdentifier.Value; got != "support@edcontrols.nl" {
		t.Errorf("outgoing sender email = %q, want %q", got, "support@edcontrols.nl")
	}
	if got := outgoing.Recipients[0].DeliveryIdentifier.Value; got != "customer@example.com" {
		t.Errorf("outgoing recipient email = %q, want %q", got, "customer@example.com")
	}

	incoming := resp.Results[1]
	if got := incoming.Senders[0].DeliveryIdentifier.Value; got != "customer@example.com" {
		t.Errorf("incoming sender email = %q, want %q", got, "customer@example.com")
	}
}

// ASSIGNMENT and THREAD_STATUS_CHANGE entries have empty senders/recipients and
// must be handled without panicking when the read sites index into the slices.
func TestUnmarshalThreadMessages_NonMessageTypesHaveEmptySenders(t *testing.T) {
	var resp threadMessagesResponse
	if err := json.Unmarshal([]byte(representativeThreadMessages), &resp); err != nil {
		t.Fatalf("failed to unmarshal thread messages: %v", err)
	}

	for _, m := range resp.Results {
		if m.Type == "MESSAGE" {
			continue
		}
		if len(m.Senders) != 0 || len(m.Recipients) != 0 {
			t.Errorf("%s expected empty senders/recipients, got senders=%d recipients=%d",
				m.Type, len(m.Senders), len(m.Recipients))
		}
	}
}
