package parse

import (
	"bufio"
	"os"
	"strings"
)

// 解析配置文件.
// 参数:
//   - config: 场景路径
func ParseConfig(config string) (map[string]string, error) {
	file, err := os.Open(config)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	items := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line, _, _ := strings.Cut(scanner.Text(), ";")
		name, value, _ := strings.Cut(line, ":")
		items[strings.TrimSpace(name)] = strings.TrimSpace(value)
	}

	return items, scanner.Err()
}
