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
// @Description 根据指定的IP地址获取该位置的天气信息
// @Tags 天气
// @Accept json
// @Produce json
// @Param ip query string true "IP地址"
// @Success 200 {object} service.WeatherInfo "天气信息"
// @Failure 400 {object} map[string]string "IP参数错误"
// @Failure 500 {object} map[string]string "获取天气失败"
// @Router /api/weather/by-ip [get]
func (h *WeatherHandler) GetWeatherByIP(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "IP address is required",
		})
		return
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
