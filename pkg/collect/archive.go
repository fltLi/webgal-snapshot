package collect

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Resource 表示收集的资源.
type Resource struct {
	Path   string
	Expand func() ([]Resource, error) // 关联文件解析
}

// NewArchiver 创建归档器.
// 参数:
//   - path: 输出路径
//   - converter: 输入路径转换器
//   - inspector: 每项资源归档结果通知
//
// 返回:
//   - chan<- Resource: 资源发送管道
//   - func() (int, int): 关闭管道, 等待执行完毕, 获取成功、失败数量
//   - error: 创建失败
func NewArchiver(
	path string,
	converter func(string) string,
	inspector func(int, string, error),
) (chan<- Resource, func(), error) {
	// 打开压缩包
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	zip := zip.NewWriter(file)

	// 归档写入
	write := func(dst, src string) error {
		file, err := os.Open(src)
		if err != nil {
			return err
		}
		defer file.Close()

		w, err := zip.Create(dst)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, file)
		return nil
	}

	end := make(chan struct{}) // 结束输出
	archiver := make(chan Resource, 256)

	// 启动输出线程
	go func() {
		defer file.Close()
		defer zip.Close()

		i := 0
		history := make(map[string]struct{})

		handle := func(src string) (string, bool, error) {
			dst := converter(src)
			src, err := filepath.Abs(src)
			if err != nil {
				return dst, true, err
			}

			// 去重
			if _, ok := history[src]; ok {
				return dst, false, nil
			}
			history[src] = struct{}{}

			// 写入文件
			err = write(dst, src)
			return dst, true, err
		}

		for res := range archiver {
			tasks := []Resource{res}

			for len(tasks) != 0 {
				res := tasks[len(tasks)-1]
				tasks = tasks[:len(tasks)-1]

				// 处理一个资源
				dst, ok, err := handle(res.Path)
				if ok {
					i++
					inspector(i, dst, err)
				}

				if ok && res.Expand != nil {
					// 获取扩展资源
					exp, err := res.Expand()
					tasks = append(tasks, exp...) // 追加任务
					if err != nil {
						inspector(i, "", fmt.Errorf("关联资源解析出错: %w", err))
					}
				}
			}
		}

		end <- struct{}{}
	}()

	return archiver, func() {
		close(archiver)
		<-end
	}, nil
}
