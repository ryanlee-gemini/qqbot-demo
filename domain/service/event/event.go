package event

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/ryanlee-gemini/qqbot-demo/domain/repo/openai"
	"github.com/ryanlee-gemini/qqbot-demo/domain/repo/qqbot"
	"github.com/ryanlee-gemini/qqbot-demo/domain/repo/qqbot/dto"
)

// 自定义Http请求头
const (
	sigMethodEd25519 = "Ed25519"

	OpTypeDispatchEvent = 0
	OpTypeValidation    = 13

	C2CMessageCreate     = "C2C_MESSAGE_CREATE"
	GroupATMessageCreate = "GROUP_AT_MESSAGE_CREATE"
)

// Payload 回调请求
type Payload struct {
	Op   int             `json:"op"`
	Id   string          `json:"id"`
	Data json.RawMessage `json:"d"`
	Type string          `json:"t"`
}

// ValidationRequest 机器人回调验证请求Data
type ValidationRequest struct {
	PlainToken string `json:"plain_token"`
	EventTs    string `json:"event_ts"`
}

// ValidationResponse 机器人回调验证响应结果
type ValidationResponse struct {
	PlainToken string `json:"plain_token"`
	Signature  string `json:"signature"`
}

func VerifySignature(timestamp string, payloadBytes []byte, secret string, signature string) bool {
	if timestamp == "" {
		log.Println("timestamp empty")
		return false
	}
	if signature == "" {
		log.Println("signature empty")
		return false
	}
	sig, err := hex.DecodeString(signature)
	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		return false
	}
	if err != nil {
		log.Println("decode signature failed ", err)
		return false
	}
	var msg bytes.Buffer
	msg.WriteString(timestamp)
	msg.Write(payloadBytes)
	seed := secret
	for len(seed) < ed25519.SeedSize {
		seed = strings.Repeat(seed, 2)
	}
	rand := strings.NewReader(seed[:ed25519.SeedSize])
	publicKey, _, err := ed25519.GenerateKey(rand)
	if err != nil {
		log.Println("ed25519.GenerateKey failed ", err)
		return false
	}
	return ed25519.Verify(publicKey, msg.Bytes(), sig)
}

func HandleEvent(ctx context.Context, payload *Payload) {
	//校验签名
	api := qqbot.GetMessageApiInstance()
	switch payload.Type {
	case C2CMessageCreate, GroupATMessageCreate:
		{
			recMsg := &dto.Message{}
			err := json.Unmarshal(payload.Data, recMsg)
			if err != nil {
				log.Println("unmarshal payload.data failed ", err)
				return
			}
			switch payload.Type {
			case C2CMessageCreate:
				replyMsg := generateDemoMessage(recMsg)
				_, err = api.SendC2CMessage(ctx, recMsg.Author.ID, replyMsg)
				if err != nil {
					log.Println("send c2c msg failed ", err)
				}
			case GroupATMessageCreate:
				replyMsg := generateDemoMessage(recMsg)
				_, err = api.SendGroupMessage(ctx, recMsg.GroupID, replyMsg)
				if err != nil {
					log.Println("send group msg failed ", err)
				}
			}
		}
	default:
		{
			log.Println("receive unsupported event ", string(payload.Data))
			return
		}
	}

}

func GenerateSignature(ctx context.Context, method, timestamp string, body []byte, botSecret string) (string, error) {
	switch method {
	case sigMethodEd25519:
		// 获取seed生成公钥、私钥
		seed, err := getSeed(botSecret)
		if err != nil {
			return "", err
		}
		reader := strings.NewReader(seed)
		// GenerateKey 方法会返回公钥、私钥，这里只需要私钥进行签名生成不需要返回公钥
		_, privateKey, err := ed25519.GenerateKey(reader)
		if err != nil {
			log.Println(ctx, "ED26619-GenerateSign-GenerateKeyFail && error", err)
			return "", err
		}
		var msg bytes.Buffer
		msg.WriteString(timestamp)
		msg.Write(body)
		return hex.EncodeToString(ed25519.Sign(privateKey, msg.Bytes())), nil
	}
	return "", errors.New("unsupported signature")
}

func getSeed(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("secret invalid")
	}
	seed := secret
	for len(seed) < ed25519.SeedSize {
		seed = strings.Repeat(seed, 2)
	}
	return seed[:ed25519.SeedSize], nil
}

func generateDemoMessage(recMsg *dto.Message) dto.APIMessage {
	api := openai.GetInstance()
	rsp := api.ChatCompletions(recMsg.Content)
	return &dto.MessageToCreate{
		Timestamp: time.Now().UnixMilli(),
		MsgType:   dto.TextMsg,
		Content:   rsp,
		MessageReference: &dto.MessageReference{
			// 引用这条消息
			MessageID:             recMsg.ID,
			IgnoreGetMessageError: true,
		},
		MsgID: recMsg.ID,
	}
}
