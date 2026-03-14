package service

import (
	"bbsDemo/config"
	"bbsDemo/database"
	"bbsDemo/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type AIService struct {
	config *config.Config
}

func NewAIService() *AIService {
	logger.Info("Initializing AI Service")

	// 读取配置
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Error("Failed to load AI config", zap.Error(err))
		// 如果配置加载失败，返回一个基本实现
		return &AIService{}
	}

	logger.Info("AI Service initialized successfully",
		zap.String("model", cfg.AI.Model),
		zap.String("api_base", cfg.AI.APIBase),
		zap.Int("timeout", cfg.AI.Timeout),
		zap.Int("max_tokens", cfg.AI.MaxTokens),
		zap.Float64("temperature", cfg.AI.Temperature))

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
	logger.Info("Searching for relevant documents",
		zap.String("query", query),
		zap.Int("limit", limit))

	docs, err := database.SearchDocuments("eyuforum", query, limit)
	if err != nil {
		logger.Error("Failed to search documents", zap.Error(err))
		return nil, err
	}

	logger.Info("Found documents", zap.Int("count", len(docs)))

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

	logger.Info("Documents formatted successfully", zap.Int("formatted_count", len(documents)))
	return documents, nil
}

// GenerateAnswer 生成AI回答
func (s *AIService) GenerateAnswer(question string, documents []string) (string, error) {
	logger.Info("Generating AI answer",
		zap.String("question", question),
		zap.Int("documents_count", len(documents)))

	// 构建RAG提示词
	prompt := `你是一个智能论坛助手，根据以下论坛内容回答用户问题。

`

	if len(documents) > 0 {
		prompt += "相关论坛内容：\n"
		for i, doc := range documents {
			prompt += fmt.Sprintf("%d. %s\n\n", i+1, doc)
		}
		logger.Info("Added documents to prompt", zap.Int("count", len(documents)))
	} else {
		prompt += "未找到相关论坛内容。\n\n"
		logger.Warn("No relevant documents found")
	}

	prompt += fmt.Sprintf("用户问题：%s\n\n", question)
	prompt += "请基于论坛内容，给出详细、准确的回答。如果没有相关内容，请基于你的知识给出合理的回答。"

	logger.Info("Prompt built successfully", zap.Int("prompt_length", len(prompt)))

	// 优先使用OpenAI兼容API调用大模型
	if s.config != nil && s.config.AI.APIKey != "" {
		logger.Info("Attempting to call AI API",
			zap.String("model", s.config.AI.Model),
			zap.String("api_base", s.config.AI.APIBase))

		answer, err := s.callOpenAICompatibleAPI(prompt)
		if err == nil {
			logger.Info("AI API call successful", zap.Int("answer_length", len(answer)))
			return answer, nil
		}
		// API调用失败时记录错误，但仍返回备用回答
		logger.Error("AI API call failed, using fallback", zap.Error(err))
	} else {
		logger.Warn("AI API key not configured, using fallback")
	}

	// 备用实现（仅当API调用失败时使用）
	logger.Info("Using fallback answer generation")
	fallbackAnswer := s.fallbackGenerateAnswer(question, documents)
	logger.Info("Fallback answer generated", zap.Int("answer_length", len(fallbackAnswer)))
	return fallbackAnswer, nil
}

// StreamGenerateAnswer 流式生成AI回答
func (s *AIService) StreamGenerateAnswer(question string, documents []string) (<-chan string, error) {
	logger.Info("Starting streaming AI answer generation",
		zap.String("question", question),
		zap.Int("documents_count", len(documents)))

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
			logger.Info("Added documents to streaming prompt", zap.Int("count", len(documents)))
		} else {
			prompt += "未找到相关论坛内容。\n\n"
			logger.Warn("No relevant documents found for streaming")
		}

		prompt += fmt.Sprintf("用户问题：%s\n\n", question)
		prompt += "请基于论坛内容，给出详细、准确的回答。如果没有相关内容，请基于你的知识给出合理的回答。"

		logger.Info("Streaming prompt built successfully", zap.Int("prompt_length", len(prompt)))

		// 优先使用OpenAI兼容API流式调用大模型
		if s.config != nil && s.config.AI.APIKey != "" {
			logger.Info("Attempting to call streaming AI API",
				zap.String("model", s.config.AI.Model),
				zap.String("api_base", s.config.AI.APIBase))

			err := s.streamOpenAICompatibleAPI(prompt, ch)
			if err == nil {
				logger.Info("Streaming AI API call completed successfully")
				return
			}
			// API调用失败时记录错误，但仍使用备用实现
			logger.Error("Streaming AI API call failed, using fallback", zap.Error(err))
		} else {
			logger.Warn("AI API key not configured for streaming, using fallback")
		}

		// 备用实现（仅当API调用失败时使用）
		logger.Info("Using fallback streaming answer generation")
		s.fallbackStreamGenerateAnswer(question, documents, ch)
		logger.Info("Fallback streaming answer generation completed")
	}()

	return ch, nil
}

// callOpenAICompatibleAPI 调用OpenAI兼容的API（如DeepSeek）
func (s *AIService) callOpenAICompatibleAPI(prompt string) (string, error) {
	logger.Info("Preparing AI API request",
		zap.String("model", s.config.AI.Model),
		zap.Int("max_tokens", s.config.AI.MaxTokens),
		zap.Int("prompt_length", len(prompt)))

	reqBody := openAIRequest{
		Model: s.config.AI.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: s.config.AI.MaxTokens,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("Failed to marshal request body", zap.Error(err))
		return "", err
	}

	logger.Info("Request body marshaled successfully", zap.Int("request_body_length", len(data)))

	// 使用DeepSeek的API端点
	apiEndpoint := s.config.AI.APIBase + "/chat/completions"
	logger.Info("Sending request to AI API", zap.String("endpoint", apiEndpoint))

	req, err := http.NewRequest("POST", apiEndpoint, strings.NewReader(string(data)))
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.AI.APIKey)

	logger.Info("HTTP request created successfully",
		zap.String("method", req.Method),
		zap.String("content_type", req.Header.Get("Content-Type")))

	client := &http.Client{
		Timeout: time.Duration(s.config.AI.Timeout) * time.Second,
	}

	logger.Info("Sending HTTP request", zap.Int("timeout_seconds", s.config.AI.Timeout))

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send HTTP request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	logger.Info("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.String("status", resp.Status))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return "", err
	}

	logger.Info("Response body read successfully", zap.Int("response_body_length", len(body)))

	var result openAIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Failed to unmarshal response body", zap.Error(err))
		return "", err
	}

	logger.Info("Response unmarshaled successfully", zap.Int("choices_count", len(result.Choices)))

	if len(result.Choices) > 0 {
		answer := result.Choices[0].Message.Content
		logger.Info("AI answer extracted successfully", zap.Int("answer_length", len(answer)))
		return answer, nil
	}

	logger.Error("No choices in AI response")
	return "", fmt.Errorf("no response from AI")
}

// streamOpenAICompatibleAPI 流式调用OpenAI兼容的API
func (s *AIService) streamOpenAICompatibleAPI(prompt string, ch chan string) error {
	logger.Info("Preparing streaming AI API request",
		zap.String("model", s.config.AI.Model),
		zap.Int("max_tokens", s.config.AI.MaxTokens),
		zap.Int("prompt_length", len(prompt)))

	reqBody := openAIRequest{
		Model: s.config.AI.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: s.config.AI.MaxTokens,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("Failed to marshal streaming request body", zap.Error(err))
		return err
	}

	logger.Info("Streaming request body marshaled successfully", zap.Int("request_body_length", len(data)))

	// 使用DeepSeek的API端点
	apiEndpoint := s.config.AI.APIBase + "/chat/completions"
	logger.Info("Sending streaming request to AI API", zap.String("endpoint", apiEndpoint))

	req, err := http.NewRequest("POST", apiEndpoint, strings.NewReader(string(data)))
	if err != nil {
		logger.Error("Failed to create streaming HTTP request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.AI.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	logger.Info("Streaming HTTP request created successfully",
		zap.String("method", req.Method),
		zap.String("content_type", req.Header.Get("Content-Type")),
		zap.String("accept", req.Header.Get("Accept")))

	client := &http.Client{
		Timeout: time.Duration(s.config.AI.Timeout) * time.Second,
	}

	logger.Info("Sending streaming HTTP request", zap.Int("timeout_seconds", s.config.AI.Timeout))

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send streaming HTTP request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	logger.Info("Streaming HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.String("status", resp.Status))

	// 简单的流式处理
	reader := resp.Body
	buffer := make([]byte, 1024)
	totalBytes := 0
	chunkCount := 0

	logger.Info("Starting to read streaming response")

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				logger.Info("Streaming response completed",
					zap.Int("total_bytes", totalBytes),
					zap.Int("chunk_count", chunkCount))
				break
			}
			logger.Error("Error reading streaming response", zap.Error(err))
			return err
		}

		// 简单处理流式输出
		if n > 0 {
			ch <- string(buffer[:n])
			totalBytes += n
			chunkCount++
		}
	}

	logger.Info("Streaming response processing completed successfully",
		zap.Int("total_bytes", totalBytes),
		zap.Int("chunk_count", chunkCount))

	return nil
}

// fallbackGenerateAnswer 备用的回答生成实现
func (s *AIService) fallbackGenerateAnswer(question string, documents []string) string {
	logger.Info("Starting fallback answer generation",
		zap.String("question", question),
		zap.Int("documents_count", len(documents)))

	// 只返回简单的错误消息，不添加任何额外内容
	result := "抱歉，AI服务暂时不可用，请稍后再试。"

	logger.Info("Fallback answer generation completed", zap.Int("answer_length", len(result)))
	return result
}

// fallbackStreamGenerateAnswer 备用的流式回答生成实现
func (s *AIService) fallbackStreamGenerateAnswer(question string, documents []string, ch chan string) {
	logger.Info("Starting fallback streaming answer generation",
		zap.String("question", question),
		zap.Int("documents_count", len(documents)))

	// 只返回简单的错误消息，不添加任何额外内容
	chunk := "抱歉，AI服务暂时不可用，请稍后再试。"

	logger.Info("Starting to send streaming chunks", zap.Int("total_chunks", 1))

	ch <- chunk
	logger.Debug("Sent streaming chunk",
		zap.Int("chunk_index", 1),
		zap.Int("chunk_length", len(chunk)),
		zap.Int("total_chunks", 1))

	logger.Info("Fallback streaming answer generation completed",
		zap.Int("total_chunks_sent", 1),
		zap.Int("total_characters", len(chunk)))
}
