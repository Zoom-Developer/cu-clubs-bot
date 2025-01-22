package collector

import tele "gopkg.in/telebot.v3"

// MessageCollector is a collector for messages, it may be used to storage messages and delete them all afterward
type MessageCollector struct {
	messages []*tele.Message
}

func New() *MessageCollector {
	return &MessageCollector{
		messages: make([]*tele.Message, 0),
	}
}

// Collect adds message to the collector
func (mc *MessageCollector) Collect(m *tele.Message) {
	mc.messages = append(mc.messages, m)
}

// Send sends message to the context chat and collects it
func (mc *MessageCollector) Send(c tele.Context, what interface{}, opts ...interface{}) error {
	message, errSend := c.Bot().Send(c.Chat(), what, opts...)
	if errSend != nil {
		return errSend
	}

	mc.Collect(message)
	return nil
}

// GetMessages returns collected messages
func (mc *MessageCollector) GetMessages() []*tele.Message {
	return mc.messages
}

// Clear deletes all collected messages and cleans the collector
//
// If ignoreErrors is true, it will ignore all errors that occurred during deletion
func (mc *MessageCollector) Clear(c tele.Context, ignoreErrors bool) error {
	for _, message := range mc.messages {
		err := c.Bot().Delete(message)
		if err != nil && !ignoreErrors {
			return err
		}
	}

	mc.messages = make([]*tele.Message, 0)
	return nil
}
