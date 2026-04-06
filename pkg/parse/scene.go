package parse

import (
	"bufio"
	"os"
)

// 解析场景文件.
func ParseScene(scene string) ([]Sentence, error) {
	file, err := os.Open(scene)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sentences []Sentence
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		sentence := ParseSentence(line)
		sentences = append(sentences, sentence)
	}

	return sentences, scanner.Err()
}
