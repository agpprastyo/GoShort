package helper

func GenerateShortCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Add time-based prefix (first 2 characters)
	timePrefix := make([]byte, 2)
	now := time.Now().UnixNano()
	timePrefix[0] = charset[now%int64(len(charset))]
	timePrefix[1] = charset[(now/int64(len(charset)))%int64(len(charset))]

	// Generate remaining random characters
	randomLength := length - len(timePrefix)
	randomBytes := make([]byte, randomLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i, b := range randomBytes {
		randomBytes[i] = charset[b%byte(len(charset))]
	}

	return string(timePrefix) + string(randomBytes), nil
}

// IsValidShortCode ensures the short code follows allowed format
func IsValidShortCode(code string) bool {
	regex := regexp.MustCompile("^[a-zA-Z0-9_-]{3,16}$")
	return regex.MatchString(code)
}
