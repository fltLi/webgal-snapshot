package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fltLi/webgal-snapshot/pkg/collect"
	"github.com/fltLi/webgal-snapshot/pkg/version"
)

// ----- 样式 -----
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500")).MarginLeft(1)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).MarginLeft(1)
	goodStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	badStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).MarginLeft(1).MarginTop(1)

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262")).
			Padding(0, 1).
			MarginTop(1).
			MarginLeft(1).
			MarginRight(1)
)

// ----- 消息 -----
type taskStartedMsg struct {
	index int
	total int
	root  string
}

type fileMsg struct {
	task  int
	index int
	path  string
	err   error
}

type taskFinishedMsg struct {
	index int
}

type resultCount struct {
	Success int
	Failure int
}

// ----- 模型 -----
type model struct {
	roots   []string
	results []resultCount

	current  int
	total    int
	progress progress.Model
	viewport viewport.Model
	logs     []string

	// 打包消息通道
	msgChan chan tea.Msg
	done    chan struct{}

	allDone  bool // 标记所有任务是否已完成
	quitting bool
	err      error
	width    int
	height   int
	page     int
}

func initialModel(roots []string) model {
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 60

	vp := viewport.New(80, 15)
	vp.Style = logBoxStyle

	return model{
		roots:    roots,
		results:  make([]resultCount, len(roots)),
		current:  -1,
		total:    len(roots),
		progress: prog,
		viewport: vp,
		logs:     []string{},
		msgChan:  make(chan tea.Msg, 100),
		done:     make(chan struct{}),
		allDone:  false,
		page:     0,
	}
}

func (m model) Init() tea.Cmd {
	if m.total == 0 {
		return tea.Quit
	}
	return tea.Batch(
		tea.EnterAltScreen,
		startNextTaskCmd(0, m.roots[0], m.total, m.msgChan, m.done),
		listenForMessages(m.msgChan, m.done),
	)
}

func startNextTaskCmd(index int, root string, total int, msgChan chan<- tea.Msg, done <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		msgChan <- taskStartedMsg{index: index, total: total, root: root}

		parent := filepath.Dir(root)
		name := filepath.Base(root)
		dst := filepath.Join(parent, name+".zip")

		converter := func(src string) string {
			rel, err := filepath.Rel(root, src)
			if err != nil {
				return ""
			}
			return rel
		}

		inspector := func(i int, path string, err error) {
			select {
			case <-done:
				return
			case msgChan <- fileMsg{task: index, index: i, path: path, err: err}:
			}
		}

		archiver, wait, err := collect.NewArchiver(dst, converter, inspector)
		if err != nil {
			msgChan <- taskFinishedMsg{index: index}
			return nil
		}

		_ = collect.Collect(root, archiver)

		wait()
		msgChan <- taskFinishedMsg{index: index}
		return nil
	}
}

func listenForMessages(msgChan <-chan tea.Msg, done <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-msgChan:
			return msg
		case <-done:
			return nil
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		if msg.Height > 12 {
			m.viewport.Height = msg.Height - 12
		} else {
			m.viewport.Height = 1
		}
		m.progress.Width = msg.Width - 8
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.page > 0 {
				m.page--
			}
			return m, nil
		case "right":
			h := m.viewport.Height
			if h <= 0 {
				h = 1
			}
			totalPages := (len(m.logs) + h - 1) / h
			if totalPages <= 0 {
				totalPages = 1
			}
			if m.page < totalPages-1 {
				m.page++
			}
			return m, nil
		case "esc":
			m.quitting = true
			close(m.done)
			return m, tea.Quit
		}

	case progress.FrameMsg:
		progModel, cmd := m.progress.Update(msg)
		m.progress = progModel.(progress.Model)
		return m, cmd

	case taskStartedMsg:
		m.current = msg.index
		m.addLog(fmt.Sprintf("▶ 开始打包 %d/%d: %s", msg.index+1, msg.total, msg.root))
		progressCmd := m.progress.SetPercent(0.0)
		return m, tea.Batch(progressCmd, listenForMessages(m.msgChan, m.done))

	case fileMsg:
		if msg.err != nil {
			m.results[msg.task].Failure++
			m.addLog(badStyle.Render(fmt.Sprintf("✗ %d. %s: %v", msg.index, msg.path, msg.err)))
		} else {
			m.results[msg.task].Success++
			m.addLog(fmt.Sprintf("  %d. %s", msg.index, msg.path))
		}
		h := m.viewport.Height
		if h <= 0 {
			h = 1
		}
		totalPages := (len(m.logs) + h - 1) / h
		if totalPages <= 0 {
			totalPages = 1
		}
		m.page = totalPages - 1
		return m, listenForMessages(m.msgChan, m.done)

	case taskFinishedMsg:
		overall := float64(msg.index+1) / float64(m.total)
		progressCmd := m.progress.SetPercent(overall)

		m.addLog(goodStyle.Render(fmt.Sprintf("✓ 完成 %d/%d: 成功 %d, 失败 %d",
			msg.index+1, m.total, m.results[msg.index].Success, m.results[msg.index].Failure)))

		next := msg.index + 1
		if next < m.total {
			return m, tea.Batch(
				progressCmd,
				startNextTaskCmd(next, m.roots[next], m.total, m.msgChan, m.done),
				listenForMessages(m.msgChan, m.done),
			)
		} else {
			close(m.done)
			m.allDone = true
			return m, progressCmd
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "正在退出...\n"
	}

	// 标题
	title := titleStyle.Render(fmt.Sprintf("WebGAL Snapshot v%s", version.Version))
	repo := infoStyle.Render("github.com/fltLi/webgal-snapshot")

	// 进度条
	progView := m.progress.View()

	// 当前任务信息
	var taskLine string
	if m.allDone {
		taskLine = goodStyle.Render("全部任务已完成！")
	} else if m.current >= 0 && m.current < m.total {
		taskLine = infoStyle.Render(fmt.Sprintf("正在打包 (%d/%d): %s",
			m.current+1, m.total, m.roots[m.current]))
	} else {
		taskLine = infoStyle.Render("准备中...")
	}

	// 统计
	var totalSuccess, totalFailure int
	for i := 0; i < m.total; i++ {
		totalSuccess += m.results[i].Success
		totalFailure += m.results[i].Failure
	}
	stats := infoStyle.Render(fmt.Sprintf("累计: 成功 %d, 失败 %d", totalSuccess, totalFailure))

	// 日志视口（分页展示）
	h := m.viewport.Height
	if h <= 0 {
		h = 1
	}
	totalPages := (len(m.logs) + h - 1) / h
	if totalPages <= 0 {
		totalPages = 1
	}
	if m.page < 0 {
		m.page = 0
	}
	if m.page > totalPages-1 {
		m.page = totalPages - 1
	}
	start := min(m.page*h, len(m.logs))
	end := min(start+h, len(m.logs))
	m.viewport.SetContent(strings.Join(m.logs[start:end], "\n"))
	logView := m.viewport.View()

	help := helpStyle.Render("按 ←/→ 翻页，esc 退出")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		repo,
		"",
		lipgloss.NewStyle().MarginLeft(1).Render(progView),
		taskLine,
		stats,
		logView,
		help,
	)
}

func (m *model) addLog(line string) {
	m.logs = append(m.logs, line)
}

func main() {
	roots := os.Args[1:]
	if len(roots) == 0 {
		fmt.Println("请提供至少一个要打包的目录路径")
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(roots), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("运行出错: %v\n", err)
		os.Exit(1)
	}
}
