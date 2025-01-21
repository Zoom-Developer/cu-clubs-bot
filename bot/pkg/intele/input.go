package intele

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/intele/storage"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	stateWaitingInput = "waiting_input"
)

type pendingRequest struct {
	mu        sync.Mutex
	response  string
	completed bool
}

// Input is a manager for input requests
type Input struct {
	storage       storage.StateStorage
	requests      sync.Map
	timeout       time.Duration
	maxConcurrent int
}

type InputOptions struct {
	// Storage for storing user states (default: in memory)
	Storage storage.StateStorage
	// Timeout for input requests (default: no timeout).
	Timeout time.Duration
	// MaxConcurrent limits the number of one time input requests (default: no limit).
	MaxConcurrent int
}

// NewInput creates a new input manager
func NewInput(opts InputOptions) *Input {
	if opts.Storage == nil {
		opts.Storage = storage.NewMemoryStorage()
	}

	return &Input{
		storage:       opts.Storage,
		timeout:       opts.Timeout,       // default no timeout
		maxConcurrent: opts.MaxConcurrent, // default max concurrent requests
	}
}

// Handler returns a middleware function for telebot, that you need to set in your bot, for handling input requests
func (h *Input) Handler() tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Message() == nil {
			return nil
		}

		userID := c.Sender().ID

		// Check if we're waiting for input from this user
		state, err := h.storage.Get(userID)
		if err != nil {
			return err
		}
		if state != stateWaitingInput {
			return nil
		}

		// Get or create pending request
		value, _ := h.requests.LoadOrStore(userID, &pendingRequest{})
		req := value.(*pendingRequest)

		// Set response and mark as completed
		req.mu.Lock()
		req.response = c.Message().Text
		req.completed = true
		req.mu.Unlock()

		// Clean up storage
		h.storage.Delete(userID)

		return nil
	}
}

// Get waits for user input and returns it
func (h *Input) Get(ctx context.Context, userID int64) (string, error) {
	// Check concurrent limit
	if h.maxConcurrent != 0 {
		var count int
		h.requests.Range(func(key, value interface{}) bool {
			count++
			return count < h.maxConcurrent
		})
		if count >= h.maxConcurrent {
			return "", ErrTooManyConcurrent
		}
	}

	// Create request
	req := &pendingRequest{}
	h.requests.Store(userID, req)

	// Set the state
	if err := h.storage.Set(userID, stateWaitingInput); err != nil {
		h.requests.Delete(userID)
		return "", err
	}

	// Clean up when we're done
	defer func() {
		h.storage.Delete(userID)
		h.requests.Delete(userID)
	}()

	// Wait for response with polling
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			req.mu.Lock()
			if req.completed {
				response := req.response
				req.mu.Unlock()
				return response, nil
			}
			req.mu.Unlock()

			if h.timeout > 0 && time.Since(start) > h.timeout {
				return "", ErrTimeout
			}

			// Small sleep to prevent CPU spinning
			time.Sleep(100 * time.Millisecond)
		}
	}
}
