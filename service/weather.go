package service

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"bbsDemo/logger"

	"go.uber.org/zap"
)

// WeatherService 天气服务
type WeatherService struct {
	httpClient *http.Client
}

// WeatherInfo 天气信息
type WeatherInfo struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feels_like"`
	Humidity    int     `json:"humidity"`
	Weather     string  `json:"weather"`
	WindSpeed   float64 `json:"wind_speed"`
	UpdatedAt   string  `json:"updated_at"`
}

// IPInfo IP信息响应
type IPInfo struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Country     string  `json:"country_name"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// OpenMeteoResponse Open-Meteo API响应
type OpenMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
		WeatherCode int     `json:"weathercode"`
		Time        string  `json:"time"`
	} `json:"current_weather"`
	Hourly struct {
		Temperature2m       []float64 `json:"temperature_2m"`
		RelativeHumidity2m  []int     `json:"relative_humidity_2m"`
		ApparentTemperature []float64 `json:"apparent_temperature"`
		Time                []string  `json:"time"`
	} `json:"hourly"`
}

// NewWeatherService 创建天气服务
func NewWeatherService() *WeatherService {
	return &WeatherService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetWeatherByIP 根据IP获取天气信息
func (s *WeatherService) GetWeatherByIP(ip string) (*WeatherInfo, error) {
	log := logger.Log

	// 如果IP为空或为本地地址，使用公网IP
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		publicIP, err := s.getPublicIP()
		if err != nil {
			log.Error("Failed to get public IP", zap.Error(err))
			return nil, err
		}
		ip = publicIP
	}

	log.Info("Getting weather for IP", zap.String("ip", ip))

	// 获取IP地理位置
	ipInfo, err := s.getIPLocation(ip)
	if err != nil {
		log.Error("Failed to get IP location", zap.Error(err), zap.String("ip", ip))
		return nil, err
	}

	log.Info("Got IP location",
		zap.String("city", ipInfo.City),
		zap.String("country", ipInfo.Country),
		zap.Float64("lat", ipInfo.Latitude),
		zap.Float64("lon", ipInfo.Longitude))

	// 获取天气数据
	weather, err := s.getWeatherData(ipInfo.Latitude, ipInfo.Longitude)
	if err != nil {
		log.Error("Failed to get weather data", zap.Error(err))
		return nil, err
	}

	// 组合天气信息
	weatherInfo := &WeatherInfo{
		IP:          ip,
		City:        ipInfo.City,
		Country:     ipInfo.Country,
		Temperature: weather.CurrentWeather.Temperature,
		FeelsLike:   s.getCurrentFeelsLike(weather),
		Humidity:    s.getCurrentHumidity(weather),
		Weather:     s.getWeatherDescription(weather.CurrentWeather.WeatherCode),
		WindSpeed:   weather.CurrentWeather.Windspeed,
		UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	log.Info("Weather info retrieved successfully",
		zap.String("city", weatherInfo.City),
		zap.Float64("temperature", weatherInfo.Temperature))

	return weatherInfo, nil
}

// getPublicIP 获取公网IP
func (s *WeatherService) getPublicIP() (string, error) {
	resp, err := s.httpClient.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.IP, nil
}

// getIPLocation 获取IP地理位置
func (s *WeatherService) getIPLocation(ip string) (*IPInfo, error) {
	url := fmt.Sprintf("https://ipapi.co/%s/json/", ip)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ipInfo IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		return nil, err
	}

	return &ipInfo, nil
}

// getWeatherData 获取天气数据
func (s *WeatherService) getWeatherData(lat, lon float64) (*OpenMeteoResponse, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current_weather=true&hourly=temperature_2m,relative_humidity_2m,apparent_temperature&timezone=auto",
		lat, lon,
	)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var weather OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, err
	}

	return &weather, nil
}

// getCurrentFeelsLike 获取当前体感温度
func (s *WeatherService) getCurrentFeelsLike(weather *OpenMeteoResponse) float64 {
	if len(weather.Hourly.ApparentTemperature) > 0 {
		return weather.Hourly.ApparentTemperature[0]
	}
	return weather.CurrentWeather.Temperature
}

// getCurrentHumidity 获取当前湿度
func (s *WeatherService) getCurrentHumidity(weather *OpenMeteoResponse) int {
	if len(weather.Hourly.RelativeHumidity2m) > 0 {
		return weather.Hourly.RelativeHumidity2m[0]
	}
	return 0
}

// getWeatherDescription 根据天气代码获取天气描述
func (s *WeatherService) getWeatherDescription(code int) string {
	weatherCodes := map[int]string{
		0:  "晴朗",
		1:  "主要晴朗",
		2:  "部分多云",
		3:  "多云",
		45: "雾",
		48: "雾凇",
		51: "毛毛雨（轻微）",
		53: "毛毛雨（中等）",
		55: "毛毛雨（密集）",
		56: "冻雨（轻微）",
		57: "冻雨（密集）",
		61: "雨（轻微）",
		63: "雨（中等）",
		65: "雨（大雨）",
		66: "冻雨（轻微）",
		67: "冻雨（大雨）",
		71: "雪（轻微）",
		73: "雪（中等）",
		75: "雪（大雪）",
		77: "雪粒",
		80: "阵雨（轻微）",
		81: "阵雨（中等）",
		82: "阵雨（猛烈）",
		85: "阵雪（轻微）",
		86: "阵雪（猛烈）",
		95: "雷雨（轻微或中等）",
		96: "雷雨（轻微冰雹）",
		99: "雷雨（猛烈冰雹）",
	}

	if desc, ok := weatherCodes[code]; ok {
		return desc
	}
	return "未知"
}

// GetClientIP 从请求中获取客户端IP
func GetClientIP(c *http.Request) string {
	// 尝试从X-Forwarded-For获取
	xff := c.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// 尝试从X-Real-IP获取
	xri := c.Header.Get("X-Real-Ip")
	if xri != "" {
		return xri
	}

	// 从RemoteAddr获取
	host, _, err := net.SplitHostPort(c.RemoteAddr)
	if err != nil {
		return c.RemoteAddr
	}
	return host
}
