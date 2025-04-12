package bili

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

func FormatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := duration / time.Hour
	minutes := (duration % time.Hour) / time.Minute
	secs := (duration % time.Minute) / time.Second

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}
func GenerateBase64RandomString(minLength, maxLength int) string {
	// Initialize random source
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Calculate random length
	length := minLength
	if maxLength > minLength {
		length += r.Intn(maxLength - minLength + 1)
	}

	// Generate random bytes
	randomBytes := make([]byte, length)
	for i := 0; i < length; i++ {
		// Generate values between 0x20 and 0x7F
		randomBytes[i] = byte(r.Intn(0x60) + 0x20)
	}

	// Encode to base64 and return
	return base64.StdEncoding.EncodeToString(randomBytes)
}
