package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gapi-platform/internal/middleware"
	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPhase3Test(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.Channel{},
		&model.QuotaTransaction{},
		&model.UsageLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	channelRepo := repository.NewChannelRepository(db)

	tokenService := service.NewTokenService(tokenRepo)
	tokenService.SetUserRepo(userRepo, nil)
	channelService := service.NewChannelService(channelRepo)

	apiHandler := NewAPIHandler(tokenService, channelService, userRepo, nil, nil)

	r := gin.New()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/chat/completions", middleware.TokenAuth(tokenService), apiHandler.ChatCompletions)
		v1.POST("/completions", middleware.TokenAuth(tokenService), apiHandler.Completions)
	}

	return r, db
}

func createPhase3UserAndToken(t *testing.T, db *gorm.DB) string {
	t.Helper()
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	tokenService := service.NewTokenService(tokenRepo)
	tokenService.SetUserRepo(userRepo, nil)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &model.User{
		Username:     "testuser",
		Email:        "test@phase3.com",
		PasswordHash: string(hash),
		Level:        "free",
		Status:       "active",
		FreeQuota:    100000,
	}
	userRepo.Create(user)

	apiToken, err := tokenService.Create(user.ID, "Test Token", nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}
	return apiToken.TokenKey
}

func TestCompletions_StringPrompt(t *testing.T) {
	r, db := setupPhase3Test(t)
	token := createPhase3UserAndToken(t, db)

	body := map[string]interface{}{
		"model":  "gpt-3.5-turbo",
		"prompt": "Hello world",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// No channel available → 502, but should NOT be 400 (request parsing should succeed)
	if w.Code == http.StatusBadRequest {
		t.Errorf("expected request to parse successfully, got 400: %s", w.Body.String())
	}

	// Verify response format is JSON
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("response should be valid JSON: %v", err)
	}
}

func TestCompletions_ArrayPrompt(t *testing.T) {
	r, db := setupPhase3Test(t)
	token := createPhase3UserAndToken(t, db)

	body := map[string]interface{}{
		"model":  "gpt-3.5-turbo",
		"prompt": []string{"Hello", "World"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code == http.StatusBadRequest {
		t.Errorf("array prompt should parse successfully, got 400: %s", w.Body.String())
	}
}

func TestCompletions_MissingModel(t *testing.T) {
	r, db := setupPhase3Test(t)
	token := createPhase3UserAndToken(t, db)

	body := map[string]interface{}{
		"prompt": "Hello",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing model, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		if errObj["code"] != "missing_model" {
			t.Errorf("expected error code 'missing_model', got '%v'", errObj["code"])
		}
	} else {
		t.Error("expected error object in response")
	}
}

func TestChatCompletions_ArrayContent(t *testing.T) {
	r, db := setupPhase3Test(t)
	token := createPhase3UserAndToken(t, db)

	// OpenAI spec: content can be an array of content blocks
	body := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": "Hello"},
				},
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should NOT return 400 (unmarshal error)
	if w.Code == http.StatusBadRequest {
		t.Errorf("array content should be accepted, got 400: %s", w.Body.String())
	}
}

func TestChatCompletions_MixedContent(t *testing.T) {
	r, db := setupPhase3Test(t)
	token := createPhase3UserAndToken(t, db)

	body := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "plain text"},
			{"role": "assistant", "content": []map[string]interface{}{
				{"type": "text", "text": "array response"},
			}},
			{"role": "user", "content": "follow up"},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code == http.StatusBadRequest {
		t.Errorf("mixed content should be accepted, got 400: %s", w.Body.String())
	}
}

func TestAPIError_Format(t *testing.T) {
	r, db := setupPhase3Test(t)
	token := createPhase3UserAndToken(t, db)

	tests := []struct {
		name         string
		body         map[string]interface{}
		expectedCode string
	}{
		{
			name:         "missing model",
			body:         map[string]interface{}{"messages": []map[string]string{{"role": "user", "content": "hi"}}},
			expectedCode: "missing_model",
		},
		{
			name:         "empty messages",
			body:         map[string]interface{}{"model": "gpt-3.5-turbo"},
			expectedCode: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response should be valid JSON: %v", err)
			}

			errObj, ok := resp["error"].(map[string]interface{})
			if !ok {
				t.Fatal("expected error object in response")
			}

			// Verify OpenAI error format: type + code + message
			if errObj["type"] == nil || errObj["type"] == "" {
				t.Error("error should have 'type' field")
			}
			if errObj["code"] == nil || errObj["code"] == "" {
				t.Error("error should have 'code' field")
			}
			if errObj["message"] == nil || errObj["message"] == "" {
				t.Error("error should have 'message' field")
			}

			if errObj["code"] != tt.expectedCode {
				t.Errorf("expected error code '%s', got '%v'", tt.expectedCode, errObj["code"])
			}
		})
	}
}
