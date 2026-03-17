package handler

import (
	"net/http"

	"bbsDemo/service"

	"github.com/gin-gonic/gin"
)

// WeatherHandler 天气处理器
type WeatherHandler struct {
	weatherService *service.WeatherService
}

// NewWeatherHandler 创建天气处理器
func NewWeatherHandler(weatherService *service.WeatherService) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
	}
}

// GetWeather 获取当前天气信息
// @Summary 获取当前天气信息
// @Description 根据用户IP获取当前位置的天气信息，包括温度、湿度、天气状况等
// @Tags 天气
// @Accept json
// @Produce json
// @Success 200 {object} service.WeatherInfo "天气信息"
// @Failure 500 {object} map[string]string "获取天气失败"
// @Router /api/weather [get]
func (h *WeatherHandler) GetWeather(c *gin.Context) {
	// 获取客户端IP
	clientIP := c.ClientIP()

	// 获取天气信息
	weather, err := h.weatherService.GetWeatherByIP(clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get weather information",
		})
		return
	}

	c.JSON(http.StatusOK, weather)
}

// GetWeatherByIP 根据指定IP获取天气信息
// @Summary 根据IP获取天气信息
// @Description 根据指定的IP地址或客户端IP获取该位置的天气信息
// @Tags 天气
// @Accept json
// @Produce json
// @Param ip query string false "IP地址（可选，不传则使用客户端IP）"
// @Success 200 {object} service.WeatherInfo "天气信息"
// @Failure 500 {object} map[string]string "获取天气失败"
// @Router /api/weather/by-ip [get]
func (h *WeatherHandler) GetWeatherByIP(c *gin.Context) {
	// 优先从query参数获取IP，如果没有则从请求头中解析客户端IP
	ip := c.Query("ip")
	if ip == "" {
		ip = c.ClientIP()
	}

	// 获取天气信息
	weather, err := h.weatherService.GetWeatherByIP(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get weather information",
		})
		return
	}

	c.JSON(http.StatusOK, weather)
}
