package template

import (
	"bytes"
	_ "embed" // Required for the //go:embed directive
	"html/template"
)

//go:embed html/reset_password.html
var resetPasswordHTML string

// ResetPasswordData holds the dynamic data for the password reset email.
type ResetPasswordData struct {
	Username string
	Token    string // The token the user will enter to reset their password
}

// GenerateResetPasswordHTML parses the embedded reset password template and executes it with the provided data.
// It returns the generated HTML content as a string, ready to be sent as an email body.
func GenerateResetPasswordHTML(data ResetPasswordData) (string, error) {
	// Parse the embedded HTML string into a new template.
	t, err := template.New("reset-password").Parse(resetPasswordHTML)
	if err != nil {
		// This indicates a syntax error in the HTML template itself.
		return "", err
	}

	// The 'buffer' will hold the result of the template execution.
	var body bytes.Buffer

	// Execute combines the parsed template with the data struct.
	if err := t.Execute(&body, data); err != nil {
		// This error could happen if there's a mismatch between template placeholders and the data struct.
		return "", err
	}

	// Return the content of the buffer as a string.
	return body.String(), nil
}
