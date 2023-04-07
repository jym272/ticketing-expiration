package utils

import "strings"

func ExtractStreamName(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) == 0 {
		panic("Subject is empty")
	}

	stream := parts[0]
	return stream
}

func GetDurableName(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) == 0 {
		panic("Subject is empty")
	}
	var upperCaseParts []string
	for _, part := range parts {
		upperCaseParts = append(upperCaseParts, strings.ToUpper(part))
	}

	return strings.Join(upperCaseParts, "_")

}
func CreateConsumerProps(subject string) (string, string, string) {
	durableName := GetDurableName(subject)
	return durableName, subject, subject

}
