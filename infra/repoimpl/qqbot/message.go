package qqbot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/ryanlee-gemini/qqbot-demo/domain/repo/qqbot"
	"github.com/ryanlee-gemini/qqbot-demo/domain/repo/qqbot/dto"
	"github.com/tencent-connect/botgo/version"
)

const (
	apiDomain      = "https://api.sgroup.qq.com"
	timeoutSendMsg = 3 * time.Second
)

var (
	onceInitMessageAPI sync.Once
	instanceMessageAPI qqbot.MessageAPI
)

// WireUpMessageAPI 组装
func WireUpMessageAPI(appid, secret string) {
	onceInitMessageAPI.Do(func() {
		client := resty.New().
			SetTimeout(timeoutSendMsg).
			SetAuthScheme("QQBot").
			SetHeader("User-Agent", version.String()).
			SetHeader("X-Union-Appid", fmt.Sprint(appid))
		instanceMessageAPI = &messageAPI{
			client: client,
			appid:  appid,
			secret: secret,
		}
		qqbot.GetMessageApiInstance = func() qqbot.MessageAPI {
			return instanceMessageAPI
		}
	})
}

// messageAPI ..
type messageAPI struct {
	client *resty.Client
	appid  string
	secret string
}

// request 每个请求，都需要创建一个 request
func (o *messageAPI) request(ctx context.Context) *resty.Request {
	return o.client.R().SetContext(ctx)
}

// getURL 获取接口地址，会处理沙箱环境判断
func (o *messageAPI) getURL(path string) string {
	return fmt.Sprintf("%s%s", apiDomain, path)
}

// SendC2CMessage 发送单聊消息
func (o *messageAPI) SendC2CMessage(ctx context.Context, userID string, msg dto.APIMessage) (*dto.Message, error) {
	at, err := qqbot.GetAccessTokenApiInstance().RetrieveAccessToken(ctx, o.appid, o.secret)
	if err != nil {
		log.Println("retrieve access token failed ", err)
		return nil, err
	}
	resp, err := o.request(ctx).
		SetResult(&dto.Message{}).
		SetAuthToken(at.Token).
		SetPathParam("user_id", userID).
		SetBody(msg).
		Post(o.getURL("/v2/users/{user_id}/messages"))
	if err != nil {
		log.Println("send c2c message failed ", err)
		return nil, err
	}
	return resp.Result().(*dto.Message), nil
}
