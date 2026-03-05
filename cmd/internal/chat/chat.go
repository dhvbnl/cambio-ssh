package chat

import "fmt"

type Message struct {
	Sender  string
	Content string
}

func (m Message) String() string {
	return fmt.Sprintf("%s: %s", m.Sender, m.Content)
}

type Chat struct {
	Messages []Message
}

func NewChat() *Chat {
	return &Chat{
		Messages: []Message{},
	}
}

func (c *Chat) AddMessage(sender, content string) {
	c.Messages = append(c.Messages, Message{Sender: sender, Content: content})
}

func (c *Chat) GetMessages() []Message {
	return c.Messages
}

func (c *Chat) GetFormattedMessages() []string {
	formatted := make([]string, len(c.Messages))
	for i, msg := range c.Messages {
		formatted[i] = msg.String()
	}
	return formatted
}
