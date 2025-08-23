package template

import (
	"bytes"
	_ "embed" // Required for the //go:embed directive
	"html/template"
)

//go:embed html/registration_verification.html
var registrationHTML string

// RegistrationData holds the dynamic data for the registration verification email.
// The field names must be exported (start with a capital letter) to be accessible by the html/template package.
type RegistrationData struct {
	Username string
	Token    string
}

// GenerateRegistrationHTML parses the embedded registration template and executes it with the provided data.
// It returns the generated HTML content as a string, ready to be sent as an email body.
func GenerateRegistrationHTML(data RegistrationData) (string, error) {
	// Parse the embedded HTML string into a new template.
	// We give it a name, "registration", for identification in case of errors.
	t, err := template.New("registration").Parse(registrationHTML)
	if err != nil {
		// If parsing fails, it's a developer error (e.g., syntax error in the template).
		// We return an error to be logged.
		return "", err
	}

	// The 'buffer' will hold the result of the template execution.
	var body bytes.Buffer

	// Execute combines the parsed template with the data struct.
	// It writes the resulting HTML into the buffer.
	if err := t.Execute(&body, data); err != nil {
		// This error could happen if there's a mismatch between the template placeholders and the data struct.
		return "", err
	}

	// Return the content of the buffer as a string.
	return body.String(), nil
}
