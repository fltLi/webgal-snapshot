package collect

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fltLi/webgal-snapshot/pkg/parse"
)

// 收集全部场景文件
// 此操作会解析并归档 `{root}/scene/` 下所有场景.
func CollectScenes(root string, archiver chan<- string) error {
	wg := sync.WaitGroup{}

	// 并发错误聚合
	var errs []error
	errCh := make(chan error)
	wg.Go(func() {
		for err := range errCh {
			errs = append(errs, err)
		}
	})

	// 收集场景资源
	collectScene := func(path string) {
		scentences, err := parse.ParseScene(path)
		if err != nil {
			errCh <- err
			return
		}

		// 并发处理每行
		for line, sentence := range scentences {
			wg.Go(func() {
				if err := handle(sentence, root, archiver); err != nil {
					errCh <- fmt.Errorf("%s:%d: %w", path, line, err)
				}
			})
		}
	}

	// 扫描场景
	if err := filepath.WalkDir(filepath.Join(root, "scene"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			errCh <- err
			return nil
		}

		if d.IsDir() || !strings.EqualFold(filepath.Ext(path), ".txt") {
			return nil
		}

		// 处理场景
		wg.Go(func() {
			archiver <- path
			collectScene(path)
		})
		return nil
	}); err != nil {
		errCh <- err
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
