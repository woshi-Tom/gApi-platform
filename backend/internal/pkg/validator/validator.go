package validator

import (
	"net"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()
	emailRe  = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	tokenRe  = regexp.MustCompile(`^sk-ap-[a-f0-9]{32}$`)
)

func init() {
	// Register custom validators
	validate.RegisterValidation("email_format", validateEmail)
	validate.RegisterValidation("token_format", validateTokenKey)
	validate.RegisterValidation("password_strength", validatePasswordStrength)
}

// Validate validates a struct
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	return emailRe.MatchString(email)
}

// ValidateTokenKey validates API token key format
func ValidateTokenKey(key string) bool {
	return tokenRe.MatchString(key)
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	return emailRe.MatchString(email)
}

func validateTokenKey(fl validator.FieldLevel) bool {
	key := fl.Field().String()
	if key == "" {
		return true // optional field
	}
	return tokenRe.MatchString(key)
}

func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	// At least one letter and one number
	hasLetter := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasNumber := strings.ContainsAny(password, "0123456789")
	return hasLetter && hasNumber
}

// GetValidationErrors returns formatted validation errors
func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := strings.ToLower(fe.Field())
			switch fe.Tag() {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = "invalid email format"
			case "min":
				errors[field] = field + " must be at least " + fe.Param() + " characters"
			case "max":
				errors[field] = field + " must be at most " + fe.Param() + " characters"
			default:
				errors[field] = field + " is invalid"
			}
		}
	}

	return errors
}

var privateIPBlocks = []*net.IPNet{
	parseCIDR("127.0.0.0/8"),
	parseCIDR("10.0.0.0/8"),
	parseCIDR("172.16.0.0/12"),
	parseCIDR("192.168.0.0/16"),
	parseCIDR("169.254.0.0/16"),
	parseCIDR("0.0.0.0/8"),
	parseCIDR("100.64.0.0/10"),
	parseCIDR("192.0.0.0/24"),
	parseCIDR("192.0.2.0/24"),
	parseCIDR("198.51.100.0/24"),
	parseCIDR("203.0.113.0/24"),
	parseCIDR("fc00::/7"),
	parseCIDR("fe80::/10"),
}

func parseCIDR(cidr string) *net.IPNet {
	_, ipNet, _ := net.ParseCIDR(cidr)
	return ipNet
}

func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func ValidateURLForSSRF(urlStr string) error {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return nil
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	host, _, err := net.SplitHostPort(urlStr)
	if err != nil {
		host = urlStr
		if !strings.Contains(urlStr, "://") {
			host = strings.Split(urlStr, "/")[0]
		}
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return &SSRFError{message: "URL points to private/internal IP address"}
		}
	}

	return nil
}

type SSRFError struct {
	message string
}

func (e *SSRFError) Error() string {
	return e.message
}

var allowedSortColumns = map[string]bool{
	"id": true, "created_at": true, "updated_at": true,
	"username": true, "email": true, "level": true,
	"status": true, "quota": true, "vip_expired_at": true,
	"requests": true, "failed": true, "failure_rate": true,
	"tokens": true, "amount": true, "order_no": true,
	"name": true, "price": true, "type": true,
}

func ValidateSortColumn(table, column string) bool {
	if allowedSortColumns[column] {
		return true
	}
	return false
}
