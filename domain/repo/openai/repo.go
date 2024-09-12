package openai

// API DEMO。如果需要可自行实现openai的接口定义
type API interface {
	ChatCompletions(content string) string
}

// GetInstance ..
var GetInstance func() API
