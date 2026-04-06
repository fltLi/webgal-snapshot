package collect

import (
	"path/filepath"

	"github.com/fltLi/webgal-snapshot/pkg/figure"
	"github.com/fltLi/webgal-snapshot/pkg/parse"
)

// 语句处理函数.
// 参数:
//   - sentence: 语句
//   - root: 根目录
//   - archiver: 资源发送管道
type handler func(sentence parse.Sentence, root string, archiver chan<- string) error

func handleNop(parse.Sentence, string, chan<- string) error {
	return nil
}

// 具名语句处理函数注册表
// 静默处理的语句需要用 handleNop 注册.
var handlers = map[string]handler{
	"say":                 handleSay, // 非语法糖识别 (加了效果一样)
	"changeBg":            handleChangeBg,
	"changeFigure":        handleChangeFigure,
	"bgm":                 contentHandler(catBgm),
	"playVideo":           contentHandler(catVideo),
	"pixiPerform":         handleNop,
	"pixiInit":            handleNop,
	"intro":               handleIntro,
	"miniAvatar":          contentHandler(catFigure),
	"changeScene":         handleNop, // 暂时用不着可达性检测 (TODO: 在场景开头注释 `ignore`)
	"choose":              handleNop,
	"end":                 handleNop,
	"setComplexAnimation": handleNop,
	"label":               handleNop,
	"jumpLabel":           handleNop,
	"setVar":              handleNop,
	"callScene":           handleNop,
	"showVars":            handleNop,
	"unlockCg":            contentHandler(catBackground),
	"unlockBgm":           contentHandler(catBgm),
	"filmMode":            handleNop,
	"setTextbox":          handleNop,
	"setAnimation":        contentHandler(catAnimation),
	"playEffect":          contentHandler(catVocal),
	"setTempAnimation":    handleNop,
	"setTransform":        handleNop,
	"setTransition":       handleSetTransition,
	"getUserInput":        handleNop,
	"applyStyle":          handleNop, // UI 已全部包含
	"wait":                handleNop,
}

// 收集语句资源
// 此操作不读取场景.
func handle(sentence parse.Sentence, root string, archiver chan<- string) error {
	handler, ok := handlers[sentence.Command]
	if ok {
		return handler(sentence, root, archiver)
	} else {
		return handleSay(sentence, root, archiver) // 默认视为对话语句
	}
}

//////// common ////////

func collectContent(content, root, category string, archiver chan<- string) {
	if content != "" && content != "none" {
		archiver <- filepath.Join(root, category, content)
	}
}

func collectArguments(
	arguments map[string]string,
	root, category string,
	targets map[string]struct{},
	archiver chan<- string,
) {
	for name, value := range arguments {
		if _, ok := targets[name]; ok {
			archiver <- filepath.Join(root, category, value)
		}
	}
}

func contentHandler(category string) handler {
	return func(sentence parse.Sentence, root string, archiver chan<- string) error {
		collectContent(sentence.Content, root, category, archiver)
		return nil
	}
}

//////// say ////////

// say 语句非音频参数表.
var sayArguments = map[string]struct{}{
	"center":   {},
	"clear":    {},
	"concat":   {},
	"figureId": {},
	"fontSize": {},
	"id":       {},
	"left":     {},
	"notend":   {},
	"right":    {},
	"speaker":  {},
	"when":     {},
}

func handleSay(sentence parse.Sentence, root string, archiver chan<- string) error {
	for name, value := range sentence.Arguments {
		if _, ok := sayArguments[name]; ok {
			continue
		}

		if name == catVocal {
			archiver <- filepath.Join(root, catVocal, value)
		} else if value != "" {
			archiver <- filepath.Join(root, catVocal, value)
		}
	}

	return nil
}

//////// changeBg ////////

func handleChangeBg(sentence parse.Sentence, root string, archiver chan<- string) error {
	collectContent(sentence.Content, root, catBackground, archiver)
	collectArguments(sentence.Arguments, root, catAnimation, animationArguments, archiver)
	return nil
}

//////// changeFigure ////////

// 图像立绘参数列表.
var imageFigureArguments = map[string]struct{}{
	"mouthOpen":     {},
	"mouthHalfOpen": {},
	"mouthClose":    {},
	"eyesOpen":      {},
	"eyesClose":     {},
}

func handleChangeFigure(sentence parse.Sentence, root string, archiver chan<- string) error {
	collectArguments(sentence.Arguments, root, catAnimation, animationArguments, archiver)
	collectArguments(sentence.Arguments, root, catFigure, imageFigureArguments, archiver)
	if figure := sentence.Content; figure != "" && figure != "none" {
		return collectFigure(filepath.Join(root, catFigure, figure), archiver)
	}
	return nil
}

// 收集模型关联资源.
func collectFigure(path string, archiver chan<- string) error {
	assets, err := figure.Assets(path)
	for _, asset := range assets {
		archiver <- asset
	}
	return err
}

//////// intro ////////

func handleIntro(sentence parse.Sentence, root string, archiver chan<- string) error {
	if bg, ok := sentence.Arguments["backgroundImage"]; ok {
		archiver <- filepath.Join(root, catBackground, bg)
	}
	return nil
}

//////// animation ////////

// 进出场参数列表.
var animationArguments = map[string]struct{}{"enter": {}, "exit": {}}

//////// setTransition ////////

func handleSetTransition(sentence parse.Sentence, root string, archiver chan<- string) error {
	collectArguments(sentence.Arguments, root, catAnimation, animationArguments, archiver)
	return nil
}
