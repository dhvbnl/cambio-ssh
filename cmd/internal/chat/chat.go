package chat

type Message struct {
	Sender  string
	Content string
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
