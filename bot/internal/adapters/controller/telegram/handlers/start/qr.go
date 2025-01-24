package start

import (
	"context"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) qr(c tele.Context, qrCodeID string) error {
	h.logger.Infof("(user: %d) scan QR code", c.Sender().ID)

	user, err := h.userService.GetByQRCodeID(context.Background(), qrCodeID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	return c.Send(
		user.FIO,
	)

}
