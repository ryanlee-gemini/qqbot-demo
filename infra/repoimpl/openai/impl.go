package openai

import (
	"log"
	"sync"

	"github.com/ryanlee-gemini/qqbot-demo/domain/repo/openai"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	hy "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901"
)

var (
	onceInitOpenAI sync.Once
	instanceOpenAI openai.API
)

// WireUp 组装
func WireUp(secretID, secretKey string) {
	onceInitOpenAI.Do(func() {
		credential := common.NewCredential(
			secretID,
			secretKey,
		)
		// 实例化一个client选项，可选的，没有特殊需求可以跳过
		cpf := profile.NewClientProfile()
		cpf.HttpProfile.Endpoint = "hunyuan.tencentcloudapi.com"
		// 实例化要请求产品的client对象,clientProfile是可选的
		client, _ := hy.NewClient(credential, "", cpf)
		instanceOpenAI = &hunYuanOpenAPI{client: client}
	})
	openai.GetInstance = func() openai.API {
		return instanceOpenAI
	}
}

// hunYuanOpenAPI ..
type hunYuanOpenAPI struct {
	client *hy.Client
}

func (api *hunYuanOpenAPI) ChatCompletions(content string) string {
	request := hy.NewChatCompletionsRequest()
	request.Model = common.StringPtr("hunyuan-lite")
	request.Messages = []*hy.Message{
		{
			Role:    common.StringPtr("user"),
			Content: common.StringPtr(content),
		},
	}
	// 返回的resp是一个ChatCompletionsResponse的实例，与请求对象对应
	response, err := api.client.ChatCompletions(request)
	if err != nil {
		log.Println("ChatCompletions failed: ", err)
		return err.Error()
	}
	log.Println("chat completion rsp ", *response.Response.Choices[0].Message.Content)
	return *response.Response.Choices[0].Message.Content
}
