package parse

import "strings"

// WebGAL 语句.
type Sentence struct {
	Command   string
	Content   string
	Arguments map[string]string
	Comment   string
}

// 解析语句.
func ParseSentence(sentence string) Sentence {
	sentence, comment := splitComment(sentence)
	command, sentence := splitCommand(sentence)
	content, arguments := splitArguments(sentence)

	return Sentence{
		Command:   command,
		Content:   content,
		Arguments: arguments,
		Comment:   comment,
	}
}

// 分离语句注释.
func splitComment(sentence string) (string, string) {
	for i, c := range sentence {
		if c == ';' && (i > 0 && sentence[i-1] != '\\') {
			return sentence[:i], strings.TrimSpace(sentence[i+1:])
		}
	}
	return sentence, ""
}

// 分离语句类型.
func splitCommand(sentence string) (string, string) {
	command, sentence, ok := strings.Cut(sentence, ":")
	if !ok {
		return "", sentence
	} else {
		return strings.TrimSpace(command), sentence
	}
}

// 分割并解析语句参数.
func splitArguments(sentence string) (string, map[string]string) {
	content := ""
	arguments := make(map[string]string)

	begin := 0
	sentence += " -" // 简化末项处理
	for i, c := range sentence {
		if !(c == '-' && (i > 0 && sentence[i-1] == ' ')) {
			continue
		}

		// 记录参数
		if begin == 0 {
			content = strings.TrimSpace(sentence[:i-1])
		} else {
			name, value := parseArgument(sentence[begin : i-1])
			arguments[name] = value
		}

		begin = i + 1
	}

	return content, arguments
}

// 解析语句参数.
func parseArgument(argument string) (string, string) {
	name, value, _ := strings.Cut(argument, "=")
	return strings.TrimSpace(name), strings.TrimSpace(value)
}
