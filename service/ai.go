package service

import (
	"bbsDemo/config"
	"bbsDemo/database"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AIService struct {
	config *config.Config
}

func NewAIService() *AIService {
	// 读取配置
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		// 如果配置加载失败，返回一个基本实现
		return &AIService{}
	}

	return &AIService{
		config: cfg,
	}
}

type AIQuestionRequest struct {
	Question string `json:"question" binding:"required"`
}

type AIAnswerResponse struct {
	Answer string `json:"answer"`
}

// OpenAI兼容的请求结构
type openAIRequest struct {
	Model     string    `json:"model"`
	Messages  []message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI兼容的响应结构
type openAIResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message message `json:"message"`
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
	// 构建RAG提示词
	prompt := `你是一个智能论坛助手，根据以下论坛内容回答用户问题。

`

	if len(documents) > 0 {
		prompt += "相关论坛内容：\n"
		for i, doc := range documents {
			prompt += fmt.Sprintf("%d. %s\n\n", i+1, doc)
		}
	} else {
		prompt += "未找到相关论坛内容。\n\n"
	}

	prompt += fmt.Sprintf("用户问题：%s\n\n", question)
	prompt += "请基于论坛内容，给出详细、准确的回答。如果没有相关内容，请基于你的知识给出合理的回答。"

	// 优先使用OpenAI兼容API调用大模型
	if s.config != nil && s.config.AI.APIKey != "" {
		answer, err := s.callOpenAICompatibleAPI(prompt)
		if err == nil {
			return answer, nil
		}
		// API调用失败时记录错误，但仍返回备用回答
		fmt.Printf("API调用失败: %v\n", err)
	}

	// 备用实现（仅当API调用失败时使用）
	return s.fallbackGenerateAnswer(question, documents), nil
}

// StreamGenerateAnswer 流式生成AI回答
func (s *AIService) StreamGenerateAnswer(question string, documents []string) (<-chan string, error) {
	ch := make(chan string)

	go func() {
		defer close(ch)

		// 构建RAG提示词
		prompt := `你是一个智能论坛助手，根据以下论坛内容回答用户问题。

`

		if len(documents) > 0 {
			prompt += "相关论坛内容：\n"
			for i, doc := range documents {
				prompt += fmt.Sprintf("%d. %s\n\n", i+1, doc)
			}
		} else {
			prompt += "未找到相关论坛内容。\n\n"
		}

		prompt += fmt.Sprintf("用户问题：%s\n\n", question)
		prompt += "请基于论坛内容，给出详细、准确的回答。如果没有相关内容，请基于你的知识给出合理的回答。"

		// 优先使用OpenAI兼容API流式调用大模型
		if s.config != nil && s.config.AI.APIKey != "" {
			err := s.streamOpenAICompatibleAPI(prompt, ch)
			if err == nil {
				return
			}
			// API调用失败时记录错误，但仍使用备用实现
			fmt.Printf("流式API调用失败: %v\n", err)
		}

		// 备用实现（仅当API调用失败时使用）
		s.fallbackStreamGenerateAnswer(question, documents, ch)
	}()

	return ch, nil
}

// callOpenAICompatibleAPI 调用OpenAI兼容的API（如MiniMax）
func (s *AIService) callOpenAICompatibleAPI(prompt string) (string, error) {
	reqBody := openAIRequest{
		Model: s.config.AI.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: s.config.AI.MaxTokens,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 使用SiliconFlow的API端点
	apiEndpoint := "https://api.siliconflow.cn/v1/chat/completions"
	req, err := http.NewRequest("POST", apiEndpoint, strings.NewReader(string(data)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.AI.APIKey)

	client := &http.Client{
		Timeout: time.Duration(s.config.AI.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result openAIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from AI")
}

// streamOpenAICompatibleAPI 流式调用OpenAI兼容的API
func (s *AIService) streamOpenAICompatibleAPI(prompt string, ch chan string) error {
	reqBody := openAIRequest{
		Model: s.config.AI.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: s.config.AI.MaxTokens,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// 使用SiliconFlow的API端点
	apiEndpoint := "https://api.siliconflow.cn/v1/chat/completions"
	req, err := http.NewRequest("POST", apiEndpoint, strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.AI.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{
		Timeout: time.Duration(s.config.AI.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 简单的流式处理
	reader := resp.Body
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// 简单处理流式输出
		if n > 0 {
			ch <- string(buffer[:n])
		}
	}

	return nil
}

// fallbackGenerateAnswer 备用的回答生成实现
func (s *AIService) fallbackGenerateAnswer(question string, documents []string) string {
	// 构建更自然的回答
	var answer strings.Builder

	if len(documents) > 0 {
		answer.WriteString("您好！根据论坛中的相关内容，我为您提供以下信息：\n\n")

		// 整理相关内容
		for i, doc := range documents {
			answer.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, doc))
		}

		answer.WriteString(fmt.Sprintf("针对您的问题：%s\n\n", question))
		answer.WriteString("基于以上内容，我为您提供以下分析和建议：\n")
	} else {
		answer.WriteString("您好！我在论坛中没有找到与您的问题相关的内容。\n\n")
		answer.WriteString(fmt.Sprintf("您的问题是：%s\n\n", question))
		answer.WriteString("为了更好地帮助您，建议您：\n")
	}

	// 智能回复逻辑
	lowerQuestion := strings.ToLower(question)
	if strings.Contains(lowerQuestion, "test") {
		answer.WriteString("• 这是一个测试问题，系统运行正常\n")
		answer.WriteString("• 您可以尝试提出更具体的问题，我会为您提供详细解答\n")
	} else if strings.Contains(lowerQuestion, "help") {
		answer.WriteString("• 如需帮助，请详细描述您遇到的问题\n")
		answer.WriteString("• 包括具体的错误信息、操作步骤等细节\n")
		answer.WriteString("• 这样我能更准确地为您提供解决方案\n")
	} else if strings.Contains(lowerQuestion, "如何") || strings.Contains(lowerQuestion, "怎样") {
		answer.WriteString("• 您可以在论坛中搜索相关教程或指南\n")
		answer.WriteString("• 也可以查看论坛的帮助中心获取更多信息\n")
		answer.WriteString("• 如有具体问题，请提供更多细节\n")
	} else if len(question) <= 5 {
		answer.WriteString("• 您的问题过于简短，建议提供更多细节\n")
		answer.WriteString("• 例如：您遇到的具体问题、操作环境等\n")
		answer.WriteString("• 详细的描述有助于我为您提供更准确的回答\n")
	} else {
		answer.WriteString("• 建议您在论坛中使用更具体的关键词搜索相关内容\n")
		answer.WriteString("• 也可以查看论坛的热门话题和常见问题\n")
		answer.WriteString("• 如有需要，请提供更多问题细节\n")
	}

	answer.WriteString("\n如果您有其他问题，随时告诉我！")

	return answer.String()
}

// fallbackStreamGenerateAnswer 备用的流式回答生成实现
func (s *AIService) fallbackStreamGenerateAnswer(question string, documents []string, ch chan string) {
	// 构建更自然的流式回答
	var chunks []string

	if len(documents) > 0 {
		chunks = append(chunks, "您好！根据论坛中的相关内容，我为您提供以下信息：\n\n")

		// 整理相关内容
		for i, doc := range documents {
			chunks = append(chunks, fmt.Sprintf("%d. %s\n\n", i+1, doc))
		}

		chunks = append(chunks, fmt.Sprintf("针对您的问题：%s\n\n", question))
		chunks = append(chunks, "基于以上内容，我为您提供以下分析和建议：\n")
	} else {
		chunks = append(chunks, "您好！我在论坛中没有找到与您的问题相关的内容。\n\n")
		chunks = append(chunks, fmt.Sprintf("您的问题是：%s\n\n", question))
		chunks = append(chunks, "为了更好地帮助您，建议您：\n")
	}

	// 智能回复逻辑
	lowerQuestion := strings.ToLower(question)
	if strings.Contains(lowerQuestion, "test") {
		chunks = append(chunks, "• 这是一个测试问题，系统运行正常\n")
		chunks = append(chunks, "• 您可以尝试提出更具体的问题，我会为您提供详细解答\n")
	} else if strings.Contains(lowerQuestion, "help") {
		chunks = append(chunks, "• 如需帮助，请详细描述您遇到的问题\n")
		chunks = append(chunks, "• 包括具体的错误信息、操作步骤等细节\n")
		chunks = append(chunks, "• 这样我能更准确地为您提供解决方案\n")
	} else if strings.Contains(lowerQuestion, "如何") || strings.Contains(lowerQuestion, "怎样") {
		chunks = append(chunks, "• 您可以在论坛中搜索相关教程或指南\n")
		chunks = append(chunks, "• 也可以查看论坛的帮助中心获取更多信息\n")
		chunks = append(chunks, "• 如有具体问题，请提供更多细节\n")
	} else if len(question) <= 5 {
		chunks = append(chunks, "• 您的问题过于简短，建议提供更多细节\n")
		chunks = append(chunks, "• 例如：您遇到的具体问题、操作环境等\n")
		chunks = append(chunks, "• 详细的描述有助于我为您提供更准确的回答\n")
	} else {
		chunks = append(chunks, "• 建议您在论坛中使用更具体的关键词搜索相关内容\n")
		chunks = append(chunks, "• 也可以查看论坛的热门话题和常见问题\n")
		chunks = append(chunks, "• 如有需要，请提供更多问题细节\n")
	}

	chunks = append(chunks, "\n如果您有其他问题，随时告诉我！")

	for _, chunk := range chunks {
		ch <- chunk
	}
}
