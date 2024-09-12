package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ryanlee-gemini/qqbot-demo/domain/service/event"
	"github.com/ryanlee-gemini/qqbot-demo/infra/repoimpl/openai"
	"github.com/ryanlee-gemini/qqbot-demo/infra/repoimpl/qqbot"
	"gopkg.in/yaml.v3"
)

const (
	HeadSignatureMethod    = "X-Signature-Method"
	HeadSignatureEd25519   = "X-Signature-Ed25519"
	HeadSignatureTimestamp = "X-Signature-Timestamp"
	HeadBotAppID           = "X-Bot-AppID"
	sigMethodEd25519       = "Ed25519"

	host_ = "0.0.0.0"
	port_ = 9000
)

type appConfig struct {
	QQBot struct {
		AppID  string `yaml:"appid"`
		Secret string `yaml:"secret"`
	} `yaml:"qq_bot"`
	TencentCloud struct {
		SecretID  string `yaml:"secret_id"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"tencent_cloud"`
}

var (
	config appConfig
)

func init() {
	// 读取 YAML 文件
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}
	// 解析 YAML 文件
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}
	qqbot.WireUpMessageAPI(config.QQBot.AppID, config.QQBot.Secret)
	qqbot.WireUpAccessTokenAPI()
	openai.WireUp(config.TencentCloud.SecretID, config.TencentCloud.SecretKey)
	log.Println("wireup complete")
}

func main() {
	http.HandleFunc("/qqbot", handle)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host_, port_), nil)
	if err != nil {
		log.Fatal("setup server fatal:", err)
	}
	log.Println("setup server success")
}

func handle(rw http.ResponseWriter, r *http.Request) {
	log.Println("req header:", r.Header)
	headerAppid := r.Header.Get(HeadBotAppID)
	if headerAppid != config.QQBot.AppID {
		log.Println("unknown bot ", headerAppid)
		return
	}
	httpBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("read http body err", err)
		return
	}
	log.Println("req body:", string(httpBody))
	if !verifySignature(r, httpBody, config.QQBot.Secret) {
		log.Println("signature not match")
		return
	}
	log.Println("pass signature validation")
	payload := &event.Payload{}
	if err = json.Unmarshal(httpBody, payload); err != nil {
		log.Println("parse http payload err", err)
		return
	}
	switch payload.Op {
	case event.OpTypeDispatchEvent:
		{
			//处理事件
			go event.HandleEvent(context.Background(), payload)
		}
	case event.OpTypeValidation:
		{
			method := r.Header.Get(HeadSignatureMethod)
			//处理回调验证
			handleValidation(r.Context(), payload.Data, method, config.QQBot.Secret, rw)
		}
	}
}

// verifySignature 校验签名
func verifySignature(r *http.Request, payloadBytes []byte, secret string) bool {
	if method := r.Header.Get(HeadSignatureMethod); method != sigMethodEd25519 {
		log.Println("unsupported signature method ", method)
		return false
	}
	signature := r.Header.Get(HeadSignatureEd25519)
	timestamp := r.Header.Get(HeadSignatureTimestamp)
	// 按照timestamp+Body顺序组成签名体
	return event.VerifySignature(timestamp, payloadBytes, secret, signature)
}

// handleValidation 回调地址校验
func handleValidation(ctx context.Context, payload []byte, sigMethod, botSecret string, rw http.ResponseWriter) {
	req := &event.ValidationRequest{}
	if err := json.Unmarshal(payload, req); err != nil {
		log.Println("parse http payload err", err)
		return
	}
	// GenerateKey 方法会返回公钥、私钥，这里只需要私钥进行签名生成不需要返回公钥
	signature, err := event.GenerateSignature(ctx, sigMethod, req.EventTs, []byte(req.PlainToken), botSecret)
	if err != nil {
		log.Println("generate signature failed ", err)
		return
	}
	rsp, err := json.Marshal(
		&event.ValidationResponse{
			PlainToken: req.PlainToken,
			Signature:  signature,
		})
	if err != nil {
		log.Println("handle validation failed:", err)
		return
	}
	_, err = rw.Write(rsp)
	if err != nil {
		log.Println("write rsp failed:", err)
		return
	}
	return
}
