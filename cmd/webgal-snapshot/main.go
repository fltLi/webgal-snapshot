package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fltLi/webgal-snapshot/pkg/collect"
	"github.com/fltLi/webgal-snapshot/pkg/version"
)

func main() {
	fmt.Printf("webgal-snapshot v%s \nrepo: https://github.com/fltLi/webgal-snapshot\n", version.Version)

	roots := os.Args[1:]
	for i, root := range roots {
		fmt.Printf("\n开始打包 %d/%d: %s\n", i+1, len(roots), root)

		parent := filepath.Dir(root)
		name := filepath.Base(root)

		dst := filepath.Join(parent, name+".zip")
		converter := func(src string) string {
			dst, err := filepath.Rel(root, src)
			if err != nil {
				return ""
			} else {
				return dst
			}
		}
		inspector := func(i int, path string, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "%d. %s: 出错! %v\n", i, path, err)
			} else {
				fmt.Printf("%d. %s\n", i, path)
			}
		}

		// 启动打包
		archiver, wait, err := newArchiver(dst, converter, inspector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "归档压缩包打开失败: %v\n", err)
			continue
		}

		// 收集资源
		if err = collect.Collect(root, archiver); err != nil {
			fmt.Fprintf(os.Stderr, "解析时出错: %v\n", err)
		}

		close(archiver)
		success, failure := wait()
		fmt.Printf("打包结束, 成功: %d, 失败: %d\n", success, failure)
	}
}
