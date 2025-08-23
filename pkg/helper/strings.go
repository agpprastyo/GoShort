package helper

// stringToPtr converts a string to a string pointer
func StringToPtr(s string) *string {
	return &s
}

func StringJoin(messages []string, s string) string {
	result := ""
	for i, msg := range messages {
		if i > 0 {
			result += s
		}
		result += msg
	}
	return result
}
