package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type SliderCaptchaService struct {
	redisClient *redis.Client
}

type CaptchaData struct {
	Token       string    `json:"token"`
	BgImage     string    `json:"bg_image"`
	SliderImage string    `json:"slider_image"`
	XPosition   int       `json:"x_position"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type CaptchaResult struct {
	Token    string  `json:"token"`
	Track    []int   `json:"track"`
	Duration int64   `json:"duration"`
	Score    float64 `json:"score"`
}

func NewSliderCaptchaService(redisClient *redis.Client) *SliderCaptchaService {
	return &SliderCaptchaService{redisClient: redisClient}
}

func (s *SliderCaptchaService) Generate() (*CaptchaData, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	xPosition := 50 + (int(tokenBytes[0]) % 200)

	data := &CaptchaData{
		Token:       token,
		BgImage:     "/api/v1/captcha/bg/" + token,
		SliderImage: "/api/v1/captcha/slider/" + token,
		XPosition:   xPosition,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	dataBytes, err := json.Marshal(map[string]interface{}{
		"x_position": xPosition,
		"expires_at": data.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	ctx := s.redisClient.Context()
	if err := s.redisClient.Set(ctx, "captcha:"+token, dataBytes, 5*time.Minute).Err(); err != nil {
		return nil, fmt.Errorf("failed to save captcha: %w", err)
	}

	return data, nil
}

func (s *SliderCaptchaService) Verify(token string, track []int, duration int64) (bool, float64, error) {
	ctx := s.redisClient.Context()

	dataBytes, err := s.redisClient.Get(ctx, "captcha:"+token).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, 1.0, fmt.Errorf("captcha expired or invalid")
		}
		return false, 1.0, fmt.Errorf("failed to get captcha: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return false, 1.0, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	s.redisClient.Del(ctx, "captcha:"+token)

	expectedX := int(data["x_position"].(float64))

	if len(track) == 0 {
		return false, 1.0, fmt.Errorf("empty track")
	}

	finalPosition := track[len(track)-1]
	positionDiff := abs(finalPosition - expectedX)

	var score float64
	if positionDiff <= 5 {
		score = 0.1
	} else if positionDiff <= 10 {
		score = 0.3
	} else if positionDiff <= 20 {
		score = 0.5
	} else {
		score = 0.9
	}

	if duration < 200 || duration > 10000 {
		score += 0.3
	}

	if len(track) < 10 {
		score += 0.2
	}

	if score > 1.0 {
		score = 1.0
	}

	passed := score < 0.7

	return passed, score, nil
}

func (s *SliderCaptchaService) ValidateToken(token string) bool {
	ctx := s.redisClient.Context()
	exists := s.redisClient.Exists(ctx, "captcha:"+token)
	return exists.Val() > 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
