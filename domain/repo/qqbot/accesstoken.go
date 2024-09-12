package qqbot

import (
	"context"
	"time"
)

// AccessToken AccessToken信息
type AccessToken struct {
	Token      string
	ExpiresIn  int64
	UpdateTime time.Time
}

// AccessTokenAPI ..
type AccessTokenAPI interface {
	// RetrieveAccessToken 获取AccessToken
	RetrieveAccessToken(ctx context.Context, appID, secret string) (*AccessToken, error)
}

// GetAccessTokenApiInstance ..
var GetAccessTokenApiInstance func() AccessTokenAPI
