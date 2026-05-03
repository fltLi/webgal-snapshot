package collect

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sync"
)

// collector 表示资源收集函数.
// 参数:
//   - root: WebGAL 项目根目录
//   - archiver: 资源发送管道
type collector func(root string, archiver chan<- Resource) error

// 资源收集函数注册表.
var collectors = []collector{
	collectCommons,
	CollectConfig,
	CollectScenes,
}

// Collect 收集场景和资源文件.
func Collect(root string, archiver chan<- Resource) error {
	wg := sync.WaitGroup{}

	// 并发错误聚合
	var errs []error
	errCh := make(chan error)
	wg.Go(func() {
		for err := range errCh {
			errs = append(errs, err)
		}
	})

	for _, c := range collectors {
		wg.Go(func() {
			if err := c(root, archiver); err != nil {
				errCh <- err
			}
		})
	}

	// 等待结果
	close(errCh)
	wg.Wait()
	return errors.Join(errs...)
}

// 默认打包资源 (目录或文件).
var commonResources = []string{
	// 引擎文件
	"assets/",
	"icons/",
	"lib/",
	"index.html",
	"manifest.json",
	"webgal-serviceworker.js",

	// 游戏文件
	"game/template",
	"game/tex",
	"game/userStyleSheet.css",
}

// collectCommons 收集默认打包资源.
func collectCommons(root string, archiver chan<- Resource) error {
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

				archiver <- Resource{Path: path}
				return nil
			}); err != nil {
				errCh <- err
			}
		})
	}

	// 等待结果
	close(errCh)
	wg.Wait()
	return errors.Join(errs...)
}
