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

// pendingRequest represents a pending input request from a user
type pendingRequest struct {
	mu        sync.Mutex
	response  *tele.Message // the response message
	completed bool          // whether the request has been completed
	canceled  bool          // whether the request has been canceled (completed will be true in this case)
}

// InputManager is a manager for input requests
type InputManager struct {
	storage  storage.StateStorage
	requests sync.Map
}

// InputOptions contains options for the input manager
type InputOptions struct {
	// Storage for storing user states (default: in memory)
	Storage storage.StateStorage
}

// NewInputManager creates a new input manager
//
// NOTE:
//   - If no storage is provided, by default storage.MemoryStorage will be used
func NewInputManager(opts InputOptions) *InputManager {
	if opts.Storage == nil {
		opts.Storage = storage.NewMemoryStorage()
	}
	return &InputManager{
		storage: opts.Storage,
	}
}

// Handler returns a OnText handler function for telebot, that you need to set in your bot, for handling input requests
//
// Example: b.Handle(tele.OnText, b.InputManager.Handler())
func (h *InputManager) Handler() tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Message() == nil {
			return nil
		}

		userID := c.Sender().ID

		// Check if we're waiting for input from this user
		state, err := h.storage.Get(userID)
		if err != nil || state != stateWaitingInput {
			return nil
		}

		// Get or create pending request
		value, _ := h.requests.LoadOrStore(userID, &pendingRequest{})
		req := value.(*pendingRequest)

		// Set response and mark as completed
		req.mu.Lock()
		req.response = c.Message()
		req.completed = true
		req.mu.Unlock()

		// Clean up storage
		h.storage.Delete(userID)

		return nil
	}
}

// Cancel cancels the input request for the given user
func (h *InputManager) Cancel(userID int64) {
	value, ok := h.requests.Load(userID)
	if !ok {
		return
	}

	req := value.(*pendingRequest)
	req.mu.Lock()
	req.canceled = true
	req.completed = true
	req.mu.Unlock()

	h.storage.Delete(userID)
	h.requests.Delete(userID)
}

// Get waits for user input and returns it. If timeout is 0, waits indefinitely.
//
// NOTE:
//   - This function is blocking, so make sure to call it in a separate goroutine
//   - It will return canceled=true and error if the context is canceled or context deadline is exceeded
//   - It will return ErrTimeout if the timeout is exceeded
//   - It will return canceled=true and nil error if input is canceled by Cancel
func (h *InputManager) Get(ctx context.Context, userID int64, timeout time.Duration) (response *tele.Message, canceled bool, err error) {
	// Create request
	req := &pendingRequest{}
	h.requests.Store(userID, req)

	// Set the state
	if err := h.storage.Set(userID, stateWaitingInput, timeout); err != nil {
		h.requests.Delete(userID)
		return nil, false, err
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
			return nil, true, ctx.Err()
		default:
			req.mu.Lock()
			if req.completed {
				canceled = req.canceled
				response = req.response
				req.mu.Unlock()
				return response, canceled, nil
			}
			req.mu.Unlock()

			if timeout > 0 && time.Since(start) > timeout {
				return nil, false, ErrTimeout
			}

			// Small sleep to prevent CPU spinning
			time.Sleep(100 * time.Millisecond)
		}
	}
}
