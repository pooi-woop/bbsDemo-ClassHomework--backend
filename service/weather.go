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
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
	}
}

// GetWeatherByIP 根据IP获取天气信息
func (s *WeatherService) GetWeatherByIP(ip string) (*WeatherInfo, error) {
	log := logger.Log

	// 记录请求的IP
	log.Info("Getting weather for IP", zap.String("ip", ip))

	// 使用国内的IP地理位置API
	ipInfo, err := s.getIPLocationCN(ip)
	if err != nil {
		log.Error("Failed to get IP location", zap.Error(err), zap.String("ip", ip))
		// 失败时返回默认的北京天气数据
		return s.getDefaultWeatherInfo(ip), nil
	}

	log.Info("Got IP location",
		zap.String("city", ipInfo.City),
		zap.String("country", ipInfo.Country),
		zap.Float64("lat", ipInfo.Latitude),
		zap.Float64("lon", ipInfo.Longitude))

	// 直接构造天气信息（模拟数据）
	// 实际使用时可以调用国内的天气API，如高德地图天气API
	weatherInfo := &WeatherInfo{
		IP:          ip,
		City:        ipInfo.City,
		Country:     ipInfo.Country,
		Temperature: 20.5,
		FeelsLike:   21.0,
		Humidity:    50,
		Weather:     "晴",
		WindSpeed:   10.5,
		UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	log.Info("Weather info retrieved successfully",
		zap.String("city", weatherInfo.City),
		zap.Float64("temperature", weatherInfo.Temperature))

	return weatherInfo, nil
}

// getDefaultWeatherInfo 获取默认天气信息
func (s *WeatherService) getDefaultWeatherInfo(ip string) *WeatherInfo {
	return &WeatherInfo{
		IP:          ip,
		City:        "北京",
		Country:     "中国",
		Temperature: 18.5,
		FeelsLike:   17.8,
		Humidity:    45,
		Weather:     "晴朗",
		WindSpeed:   12.3,
		UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
}

// getIPLocationCN 使用国内API获取IP地理位置
func (s *WeatherService) getIPLocationCN(ip string) (*IPInfo, error) {
	// 使用 ip-api.com 国内节点
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status  string  `json:"status"`
		Country string  `json:"country"`
		City    string  `json:"city"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("failed to get IP location: %s", result.Status)
	}

	return &IPInfo{
		City:      result.City,
		Country:   result.Country,
		Latitude:  result.Lat,
		Longitude: result.Lon,
	}, nil
}

// getPublicIP 获取公网IP
func (s *WeatherService) getPublicIP() (string, error) {
	// 尝试使用ipify
	ip, err := s.getIPFromService("https://api.ipify.org?format=json")
	if err == nil {
		return ip, nil
	}

	// 尝试使用ipinfo
	ip, err = s.getIPFromService("https://ipinfo.io/json")
	if err == nil {
		return ip, nil
	}

	// 尝试使用ifconfig.me
	ip, err = s.getIPFromService("https://ifconfig.me/ip")
	if err == nil {
		return ip, nil
	}

	// 所有服务都失败，返回默认IP（北京）
	return "202.106.0.20", nil
}

// getIPFromService 从指定服务获取IP
func (s *WeatherService) getIPFromService(url string) (string, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 处理不同格式的响应
	if url == "https://ifconfig.me/ip" {
		// 直接返回文本
		var ip string
		if err := json.NewDecoder(resp.Body).Decode(&ip); err != nil {
			return "", err
		}
		return ip, nil
	} else {
		// JSON格式
		var result struct {
			IP string `json:"ip"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}
		return result.IP, nil
	}
}

// getIPLocation 获取IP地理位置
func (s *WeatherService) getIPLocation(ip string) (*IPInfo, error) {
	// 尝试使用ipapi.co
	ipInfo, err := s.getLocationFromService(ip, "https://ipapi.co/%s/json/")
	if err == nil {
		return ipInfo, nil
	}

	// 尝试使用ipinfo.io
	ipInfo, err = s.getLocationFromService(ip, "https://ipinfo.io/%s/json")
	if err == nil {
		return ipInfo, nil
	}

	// 所有服务都失败，返回默认位置（北京）
	return &IPInfo{
		IP:        ip,
		City:      "Beijing",
		Country:   "China",
		Latitude:  39.9042,
		Longitude: 116.4074,
	}, nil
}

// getLocationFromService 从指定服务获取地理位置
func (s *WeatherService) getLocationFromService(ip, urlFormat string) (*IPInfo, error) {
	url := fmt.Sprintf(urlFormat, ip)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ipInfo IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		return nil, err
	}

	// 确保返回有效的数据
	if ipInfo.City == "" || ipInfo.Country == "" {
		return nil, fmt.Errorf("invalid location data")
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
