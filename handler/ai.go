package handler

import (
	"bbsDemo/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	aiService *service.AIService
}

func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// AskAI 处理AI问答请求
func (h *AIHandler) AskAI(c *gin.Context) {
	var req service.AIQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 获取相关文档
	documents, err := h.aiService.GetRelevantDocuments(req.Question, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get relevant documents"})
		return
	}

	// 生成回答
	answer, err := h.aiService.GenerateAnswer(req.Question, documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate answer"})
		return
	}

	c.JSON(http.StatusOK, service.AIAnswerResponse{
		Answer: answer,
	})
}

// StreamAskAI 处理AI流式问答请求
func (h *AIHandler) StreamAskAI(c *gin.Context) {
	var req service.AIQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 获取相关文档
	documents, err := h.aiService.GetRelevantDocuments(req.Question, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get relevant documents"})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 流式生成回答
	stream, err := h.aiService.StreamGenerateAnswer(req.Question, documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate answer"})
		return
	}

	// 发送流式数据
	for chunk := range stream {
		c.SSEvent("message", chunk)
		c.Writer.Flush()
	}

	// 发送结束事件
	c.SSEvent("end", "")
}
