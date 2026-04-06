package collect

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sync"
)

// 收集场景和资源文件.
// 参数:
//   - root: WebGAL 项目 game 目录
func Collect(root string, archiver chan<- string) error {
	return errors.Join(collectCommons(root, archiver), CollectScenes(root, archiver))
}

// 默认打包资源 (目录或文件).
var commonResources = []string{
	"template",
	"tex",
	"config.txt",
	// "config.json",
	"userStyleSheet.css",
}

// 收集默认打包资源.
func collectCommons(root string, archiver chan<- string) error {
	wg := sync.WaitGroup{}

	// 并发错误聚合
	var errs []error
	errCh := make(chan error)
	wg.Go(func() {
		for err := range errCh {
			errs = append(errs, err)
		}
	})

	// 收集资源
	for _, path := range commonResources {
		wg.Go(func() {
			if err := filepath.WalkDir(filepath.Join(root, path), func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					errCh <- err
					return nil
				}

				if d.IsDir() {
					return nil
				}

				archiver <- path
				return nil
			}); err != nil {
				errCh <- err
			}
		})
	}

	// 等待结果
	close(errCh)
	wg.Wait()
	if len(errs) > 0 {
		return errors.Join(errs...)
	} else {
		return nil
	}
}
