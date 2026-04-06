package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// 创建归档器.
// 参数:
//   - path: 输出路径
//   - converter: 输入路径转换器
//   - inspector: 每项资源归档结果通知
//
// 返回:
//   - chan<- string: 资源发送管道
//   - func() (int, int): 等待执行完毕, 获取成功、失败数量
//   - error: 创建失败
func newArchiver(
	path string,
	converter func(string) string,
	inspector func(int, string, error),
) (chan<- string, func() (int, int), error) {
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

	wg := &sync.WaitGroup{}
	archiver := make(chan string)
	success := 0
	failure := 0

	// 启动输出线程
	wg.Go(func() {
		defer file.Close()
		defer zip.Close()

		i := 0
		history := make(map[string]struct{})

		for src := range archiver {
			dst := converter(src)
			src, err = filepath.Abs(src)
			if err != nil {
				inspector(i, dst, err)
				failure++
				continue
			}

			// 去重
			if _, ok := history[src]; ok {
				continue
			}
			history[src] = struct{}{}

			i++
			err := write(dst, src)
			inspector(i, dst, err)

			if err != nil {
				failure++
			} else {
				success++
			}
		}
	})

	return archiver, func() (int, int) {
		wg.Wait()
		return success, failure
	}, nil
}
