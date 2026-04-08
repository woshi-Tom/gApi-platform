package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// sensitivePatterns contains regex patterns for sensitive data
var sensitivePatterns = []*regexp.Regexp{
	// Email addresses (partial masking)
	regexp.MustCompile(`([a-zA-Z0-9._%+-]+)@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`),
	// JWT tokens
	regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`),
	// API keys (sk-xxx format)
	regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`),
	// Passwords in JSON
	regexp.MustCompile(`"password"\s*:\s*"[^"]*"`),
	// Tokens in JSON
	regexp.MustCompile(`"token"\s*:\s*"[^"]*"`),
	// Secret keys
	regexp.MustCompile(`"secret"\s*:\s*"[^"]*"`),
	// API keys in JSON
	regexp.MustCompile(`"api_key"\s*:\s*"[^"]*"`),
	// Private keys
	regexp.MustCompile(`"private_key"\s*:\s*"[^"]*"`),
}

// emailPattern matches email for partial masking
var emailPattern = regexp.MustCompile(`([a-zA-Z0-9._%+-])@([a-zA-Z0-9.-]+\.[a-zA-Z0-9]{2,})`)

// Logger is a structured logger with built-in redaction
type Logger struct {
	*log.Logger
}

// Default logger instance
var Default = &Logger{Logger: log.New(os.Stdout, "[GAPI] ", log.LstdFlags|log.Lshortfile)}

// Info logs info level messages
func Info(v ...interface{}) {
	Default.Print("[INFO] ", fmt.Sprint(v...))
}

// Infof logs formatted info messages
func Infof(format string, v ...interface{}) {
	Default.Printf("[INFO] "+format, v...)
}

// Warn logs warning level messages
func Warn(v ...interface{}) {
	Default.Print("[WARN] ", fmt.Sprint(v...))
}

// Warnf logs formatted warning messages
func Warnf(format string, v ...interface{}) {
	Default.Printf("[WARN] "+format, v...)
}

// Error logs error messages
func Error(v ...interface{}) {
	Default.Print("[ERROR] ", fmt.Sprint(v...))
}

// Errorf logs formatted error messages
func Errorf(format string, v ...interface{}) {
	Default.Printf("[ERROR] "+format, v...)
}

// Debug logs debug messages (only in development)
func Debug(v ...interface{}) {
	Default.Print("[DEBUG] ", fmt.Sprint(v...))
}

// Debugf logs formatted debug messages
func Debugf(format string, v ...interface{}) {
	Default.Printf("[DEBUG] "+format, v...)
}

// RedactEmail masks email address partially (e.g., u***@example.com)
func RedactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "[EMAIL_REDACTED]"
	}
	local := parts[0]
	domain := parts[1]
	if len(local) <= 1 {
		return "[EMAIL_REDACTED]@" + domain
	}
	return local[:1] + "***@" + domain
}

// RedactToken masks a token/token key
func RedactToken(token string) string {
	if len(token) <= 8 {
		return "[TOKEN_REDACTED]"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// RedactAPIKey masks an API key
func RedactAPIKey(key string) string {
	if len(key) <= 8 {
		return "[API_KEY_REDACTED]"
	}
	return key[:6] + "..." + key[len(key)-4:]
}

// RedactMessage removes or masks sensitive data from log messages
func RedactMessage(msg string) string {
	if msg == "" {
		return msg
	}

	// Replace email addresses with redacted versions
	msg = emailPattern.ReplaceAllStringFunc(msg, func(match string) string {
		parts := strings.Split(match, "@")
		if len(parts) == 2 {
			return RedactEmail(parts[0] + "@" + parts[1])
		}
		return "[EMAIL_REDACTED]"
	})

	// Replace JWT tokens
	msg = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`).ReplaceAllString(msg, "[JWT_REDACTED]")

	// Replace sk- format API keys
	msg = regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`).ReplaceAllString(msg, "[API_KEY_REDACTED]")

	// Replace password values in JSON-like strings
	msg = regexp.MustCompile(`"password"\s*:\s*"[^"]*"`).ReplaceAllString(msg, `"password":"[REDACTED]"`)

	// Replace token values in JSON-like strings
	msg = regexp.MustCompile(`"token"\s*:\s*"[^"]*"`).ReplaceAllString(msg, `"token":"[REDACTED]"`)

	// Replace secret values
	msg = regexp.MustCompile(`"secret"\s*:\s*"[^"]*"`).ReplaceAllString(msg, `"secret":"[REDACTED]"`)

	// Replace api_key values
	msg = regexp.MustCompile(`"api_key"\s*:\s*"[^"]*"`).ReplaceAllString(msg, `"api_key":"[REDACTED]"`)

	// Replace private_key values
	msg = regexp.MustCompile(`"private_key"\s*:\s*"[^"]*"`).ReplaceAllString(msg, `"private_key":"[REDACTED]"`)

	return msg
}

// RedactJSON redacts sensitive fields from a JSON string
func RedactJSON(data string) string {
	if data == "" {
		return data
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		// Not valid JSON, try RedactMessage instead
		return RedactMessage(data)
	}

	sensitiveFields := map[string]bool{
		"password":       true,
		"password_hash":  true,
		"api_key":        true,
		"token":          true,
		"token_key":      true,
		"token_key_full": true,
		"secret":         true,
		"private_key":    true,
		"public_key":     true,
		"reset_token":    true,
		"email_code":     true,
		"code":           true, // verification codes
	}

	for key := range obj {
		if sensitiveFields[strings.ToLower(key)] {
			switch obj[key].(type) {
			case string:
				obj[key] = "[REDACTED]"
			case float64:
				obj[key] = "[REDACTED]"
			}
		}
	}

	result, err := json.Marshal(obj)
	if err != nil {
		return RedactMessage(data)
	}
	return string(result)
}
