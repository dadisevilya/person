package stringFormat

import "strings"

func HumanizeString(text string) string {
	text = strings.NewReplacer("_", " ").Replace(text)
	originalSentences := strings.Split(text, ". ")
	sentences := []string{}
	for _, sentence := range originalSentences {
		sentences = append(sentences, strings.ToUpper(string(sentence[0]))+sentence[1:])
	}
	return strings.Join(sentences, ". ")
}
