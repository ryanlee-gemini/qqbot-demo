package qqbot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/ryanlee-gemini/qqbot-demo/domain/repo/qqbot"
)

const (
	accessTokenURL    = "https://bots.qq.com/app/getAppAccessToken"
	timeoutRetrieveAT = 3 * time.Second
)

var (
	onceInitAccessTokenAPI sync.Once
	instanceAccessTokenAPI qqbot.AccessTokenAPI
)

type retrieveTokenReq struct {
	AppID        string `json:"appId"`
	ClientSecret string `json:"clientSecret"`
}

type retrieveTokenRsp struct {
	Code        int    `json:"code"`
	Message     string `json:"dto"`
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

// WireUpAccessTokenAPI ..
func WireUpAccessTokenAPI() {
	onceInitAccessTokenAPI.Do(func() {
		client := resty.New().SetTimeout(timeoutRetrieveAT).SetHeader("Content-Type", "application/json")
		instanceAccessTokenAPI = &accessTokenAPI{client: client}
	})
	qqbot.GetAccessTokenApiInstance = func() qqbot.AccessTokenAPI {
		return instanceAccessTokenAPI
	}
}

type accessTokenAPI struct {
	client *resty.Client
}

// request 每个请求，都需要创建一个 request
func (a accessTokenAPI) request(ctx context.Context) *resty.Request {
	return a.client.R().SetContext(ctx)
}

func (a accessTokenAPI) RetrieveAccessToken(ctx context.Context, appID, secret string) (*qqbot.AccessToken, error) {
	retrieveReq := retrieveTokenReq{
		AppID:        appID,
		ClientSecret: secret,
	}
	rsp, err := a.request(ctx).SetBody(retrieveReq).SetResult(&retrieveTokenRsp{}).Post(accessTokenURL)
	if err != nil {
		log.Println("retrieve token failed ", err)
		return nil, err
	}
	log.Println("retrieve rsp:", rsp.Result())
	retrieveRsp, ok := rsp.Result().(*retrieveTokenRsp)
	if !ok {
		log.Println("retrieve token failed rsp data type not match")
		return nil, err
	}
	log.Println("retrieve token rsp ", retrieveRsp)
	if retrieveRsp.Code != 0 {
		return nil, fmt.Errorf("query accessToken failed %v.%v", retrieveRsp.Code, retrieveRsp.Message)
	}
	rdata := &qqbot.AccessToken{
		Token:      retrieveRsp.AccessToken,
		UpdateTime: time.Now(),
	}
	rdata.ExpiresIn, err = strconv.ParseInt(retrieveRsp.ExpiresIn, 10, 64)
	if err != nil {
		log.Println("parse expire_in failed ", err)
		return nil, err
	}
	return rdata, nil
}
