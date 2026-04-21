package fileindex

import "strings"

func chunkText(text, locatorType, locatorValue, heading string) []Chunk {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}

	rawParts := strings.Split(trimmed, "\n\n")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		if cleaned := strings.TrimSpace(part); cleaned != "" {
			parts = append(parts, cleaned)
		}
	}
	if len(parts) == 0 {
		return nil
	}

	chunks := make([]Chunk, 0, len(parts))
	for _, part := range parts {
		chunks = append(chunks, Chunk{
			Text:          part,
			TokenEstimate: len(strings.Fields(part)),
			LocatorType:   locatorType,
			LocatorValue:  locatorValue,
			Heading:       heading,
		})
	}
	return chunks
}
