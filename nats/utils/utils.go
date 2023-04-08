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

	upperCaseParts := make([]string, 0, len(parts))

	for _, part := range parts {
		upperCaseParts = append(upperCaseParts, strings.ToUpper(part))
	}

	return strings.Join(upperCaseParts, "_")
}

func CreateConsumerProps(subject string) (durableName, queueGroupName, filterSubject string) {
	return GetDurableName(subject), subject, subject
}
