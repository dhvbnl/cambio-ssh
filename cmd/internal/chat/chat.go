package chat

import (
	"fmt"
	"sync"
	"time"
)

type Message struct {
	Sender    string
	Content   string
	Timestamp time.Time
}

func (m Message) String() string {
	return fmt.Sprintf("%s: %s", m.Sender, m.Content)
}

type Chat struct {
	mu       sync.RWMutex
	Messages []Message
	version  uint64
}

func NewChat() *Chat {
	c := &Chat{
		Messages: []Message{},
	}
	go c.cleanupLoop()
	return c
}

func (c *Chat) AddMessage(sender, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Messages = append(c.Messages, Message{
		Sender:    sender,
		Content:   content,
		Timestamp: time.Now(),
	})
	c.version++
}

func (c *Chat) GetMessages() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	msgs := make([]Message, len(c.Messages))
	copy(msgs, c.Messages)
	return msgs
}

func (c *Chat) GetFormattedMessages() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	formatted := make([]string, len(c.Messages))
	for i, msg := range c.Messages {
		formatted[i] = msg.String()
	}
	return formatted
}

func (c *Chat) Version() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.version
}

func (c *Chat) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupOldMessages(1 * time.Minute)
	}
}

func (c *Chat) cleanupOldMessages(maxAge time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	newMessages := make([]Message, 0, len(c.Messages))

	for _, msg := range c.Messages {
		if msg.Timestamp.After(cutoff) {
			newMessages = append(newMessages, msg)
		}
	}

	if len(newMessages) != len(c.Messages) {
		c.Messages = newMessages
		c.version++
	}
}
