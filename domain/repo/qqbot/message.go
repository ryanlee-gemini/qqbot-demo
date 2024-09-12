package qqbot

import (
	"context"

	"github.com/ryanlee-gemini/qqbot-demo/domain/repo/qqbot/dto"
)

// MessageAPI ..
type MessageAPI interface {
	SendC2CMessage(ctx context.Context, userID string, msg dto.APIMessage) (*dto.Message, error)
}

// GetMessageApiInstance ..
var GetMessageApiInstance func() MessageAPI
