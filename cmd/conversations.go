package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dutchview/hubspot-cli/internal/api"
)

type ConversationsCmd struct {
	Messages ConvMessagesCmd `cmd:"" help:"List messages in a thread."`
	Comment  ConvCommentCmd  `cmd:"" help:"Add an internal comment to a thread."`
	Reply    ConvReplyCmd    `cmd:"" help:"Send a reply to a customer."`
}

type ConvMessagesCmd struct {
	ThreadID string `arg:"" help:"Conversation thread ID."`
	JSON     bool   `short:"j" help:"Output as JSON."`
}

func (c *ConvMessagesCmd) Run(client *api.Client) error {
	data, err := client.GetThreadMessages(c.ThreadID)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	var resp struct {
		Results []struct {
			ID        string `json:"id"`
			Type      string `json:"type"`
			CreatedAt string `json:"createdAt"`
			Direction string `json:"direction"`
			Senders   []struct {
				ActorID             string `json:"actorId"`
				Name                string `json:"name"`
				DeliveryIdentifier  string `json:"deliveryIdentifier"`
			} `json:"senders"`
			Recipients []struct {
				RecipientField      string `json:"recipientField"`
				DeliveryIdentifier  string `json:"deliveryIdentifier"`
			} `json:"recipients"`
			Text string `json:"text"`
			Body *struct {
				Text    string `json:"text"`
				Content string `json:"content"`
			} `json:"body"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if len(resp.Results) == 0 {
		fmt.Println("No messages in this thread.")
		return nil
	}

	for _, msg := range resp.Results {
		direction := msg.Direction
		if direction == "" {
			direction = msg.Type
		}

		var sender string
		if len(msg.Senders) > 0 {
			s := msg.Senders[0]
			if s.Name != "" {
				sender = s.Name
			} else if s.DeliveryIdentifier != "" {
				sender = s.DeliveryIdentifier
			} else {
				sender = s.ActorID
			}
		}

		fmt.Printf("--- [%s] %s | %s | %s ---\n",
			msg.Type,
			formatTimestamp(msg.CreatedAt),
			direction,
			sender,
		)

		text := msg.Text
		if text == "" && msg.Body != nil {
			text = msg.Body.Text
		}
		if text != "" {
			fmt.Println(text)
		}
		fmt.Println()
	}

	return nil
}

type ConvCommentCmd struct {
	ThreadID string `arg:"" help:"Conversation thread ID."`
	Text     string `arg:"" help:"Comment text."`
	JSON     bool   `short:"j" help:"Output as JSON."`
}

func (c *ConvCommentCmd) Run(client *api.Client) error {
	data, err := client.AddComment(c.ThreadID, c.Text, "")
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	fmt.Printf("Comment added to thread %s\n", c.ThreadID)
	return nil
}

type ConvReplyCmd struct {
	ThreadID       string `arg:"" help:"Conversation thread ID."`
	Text           string `arg:"" help:"Reply text."`
	SenderActorID  string `help:"Sender actor ID (e.g. A-46807710). Auto-detected if not set."`
	Recipient      string `help:"Recipient email. Auto-detected from thread if not set."`
	JSON           bool   `short:"j" help:"Output as JSON."`
}

func (c *ConvReplyCmd) Run(client *api.Client) error {
	// Auto-detect sender, recipient, and channel from existing thread messages
	msgData, err := client.GetThreadMessages(c.ThreadID)
	if err != nil {
		return fmt.Errorf("failed to read thread for auto-detection: %w", err)
	}

	var msgs struct {
		Results []struct {
			Type      string `json:"type"`
			Direction string `json:"direction"`
			Senders   []struct {
				ActorID            string `json:"actorId"`
				DeliveryIdentifier string `json:"deliveryIdentifier"`
			} `json:"senders"`
			Recipients []struct {
				RecipientField     string `json:"recipientField"`
				DeliveryIdentifier string `json:"deliveryIdentifier"`
			} `json:"recipients"`
			ChannelID        string `json:"channelId"`
			ChannelAccountID string `json:"channelAccountId"`
		} `json:"results"`
	}
	if err := json.Unmarshal(msgData, &msgs); err != nil {
		return fmt.Errorf("parse thread messages: %w", err)
	}

	senderActorID := c.SenderActorID
	recipient := c.Recipient
	var channelID, channelAccountID string

	for _, msg := range msgs.Results {
		if msg.Type != "MESSAGE" {
			continue
		}
		// Get channel info from any MESSAGE
		if channelID == "" && msg.ChannelID != "" {
			channelID = msg.ChannelID
			channelAccountID = msg.ChannelAccountID
		}
		// Get sender from OUTGOING messages
		if senderActorID == "" && strings.EqualFold(msg.Direction, "OUTGOING") && len(msg.Senders) > 0 {
			senderActorID = msg.Senders[0].ActorID
		}
		// Get recipient from INCOMING messages (the customer)
		if recipient == "" && strings.EqualFold(msg.Direction, "INCOMING") && len(msg.Senders) > 0 {
			if email := msg.Senders[0].DeliveryIdentifier; email != "" {
				recipient = email
			}
		}
	}

	if senderActorID == "" {
		return fmt.Errorf("could not detect sender; use --sender-actor-id")
	}
	if recipient == "" {
		return fmt.Errorf("could not detect recipient; use --recipient")
	}
	if channelID == "" {
		channelID = "1002" // default to email
	}

	recipients := []map[string]string{
		{
			"recipientField":     "TO",
			"deliveryIdentifier": recipient,
		},
	}

	data, err := client.SendMessage(c.ThreadID, c.Text, "", senderActorID, channelID, channelAccountID, recipients)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	fmt.Printf("Reply sent to %s on thread %s\n", recipient, c.ThreadID)
	return nil
}
