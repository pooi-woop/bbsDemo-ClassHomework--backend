package service

import (
	"bbsDemo/database"
	"fmt"

	"github.com/cloudwego/eino/adk"
)

type AIService struct {
	einoClient *adk.ChatModelAgent
}

func NewAIService() *AIService {
	// 注意：实际使用时需要从EinoExt仓库导入具体的ChatModel实现
	// 这里暂时使用nil，实际部署时需要替换为真实的ChatModel
	// 例如：chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{...})

	// 由于缺少具体的ChatModel实现，暂时返回nil
	// 实际使用时需要取消注释以下代码并添加正确的导入
	/*
		chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
			Model:  "gpt-3.5-turbo",
			APIKey: "your-api-key-here",
		})
		if err != nil {
			return nil
		}

		agent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
			Model: chatModel,
		})
		if err != nil {
			return nil
		}

		return &AIService{
			einoClient: agent,
		}
	*/

	// 暂时返回一个简单的实现
	return &AIService{}
}

type AIQuestionRequest struct {
	Question string `json:"question" binding:"required"`
}

type AIAnswerResponse struct {
	Answer string `json:"answer"`
}

// GetRelevantDocuments 从Elasticsearch获取相关文档
func (s *AIService) GetRelevantDocuments(query string, limit int) ([]string, error) {
	docs, err := database.SearchDocuments("eyuforum", query, limit)
	if err != nil {
		return nil, err
	}

	var documents []string
	for _, doc := range docs {
		var content string
		if doc.Type == "post" {
			content = fmt.Sprintf("标题: %s\n内容: %s", doc.Title, doc.Content)
		} else {
			content = fmt.Sprintf("评论: %s", doc.Content)
		}
		documents = append(documents, content)
	}

	return documents, nil
}

// GenerateAnswer 生成AI回答
func (s *AIService) GenerateAnswer(question string, documents []string) (string, error) {
	// 构建回答
	answer := fmt.Sprintf("基于论坛内容的回答：\n\n")
	if len(documents) > 0 {
		answer += "相关内容：\n"
		for i, doc := range documents {
			answer += fmt.Sprintf("%d. %s\n\n", i+1, doc)
		}
	} else {
		answer += "未找到相关内容\n"
	}
	answer += fmt.Sprintf("问题：%s\n", question)
	answer += "这是一个基于论坛内容的智能回答。"

	return answer, nil
}

// StreamGenerateAnswer 流式生成AI回答
func (s *AIService) StreamGenerateAnswer(question string, documents []string) (<-chan string, error) {
	ch := make(chan string)

	go func() {
		defer close(ch)

		// 模拟流式输出
		chunks := []string{
			"基于论坛内容的回答：",
			"\n\n相关内容：",
		}

		for i, doc := range documents {
			chunks = append(chunks, fmt.Sprintf("\n%d. %s", i+1, doc))
		}

		chunks = append(chunks, fmt.Sprintf("\n\n问题：%s", question))
		chunks = append(chunks, "\n这是一个基于论坛内容的智能回答。")

		for _, chunk := range chunks {
			ch <- chunk
		}
	}()

	return ch, nil
}
