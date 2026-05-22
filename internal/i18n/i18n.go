package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/mattn/go-isatty"
)

// Lang 表示界面语言。
type Lang string

const (
	LangZH Lang = "zh"
	LangEN Lang = "en"
)

// settingsFile 为 ~/.config/fuckssh/settings.json 的结构。
type settingsFile struct {
	Lang Lang `json:"lang"`
}

var (
	mu       sync.RWMutex
	current  Lang = LangZH
	loadOnce sync.Once
)

// envLangKey 允许测试或 CI 跳过交互并固定语言。
const envLangKey = "FUCKSSH_LANG"

// settingsPathOverride 单测可注入设置文件路径。
var settingsPathOverride string

// isInteractiveOverride 单测可覆盖 TTY 检测。
var isInteractiveOverride func(io.Writer) bool

// pickLanguageFn 单测可替换语言选择表单。
var pickLanguageFn = pickLanguageInteractive

// Current 返回当前语言。
func Current() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// SetCurrent 设置当前语言（不写盘，供测试或 Load 使用）。
func SetCurrent(lang Lang) {
	mu.Lock()
	defer mu.Unlock()
	current = normalizeLang(lang)
}

// T 返回当前语言的文案；未知键返回 key 本身。
func T(key string, args ...any) string {
	mu.RLock()
	lang := current
	mu.RUnlock()

	msg := lookup(lang, key)
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

func lookup(lang Lang, key string) string {
	table := messagesZH
	if lang == LangEN {
		table = messagesEN
	}
	if s, ok := table[key]; ok {
		return s
	}
	return key
}

func normalizeLang(lang Lang) Lang {
	switch strings.ToLower(string(lang)) {
	case "en", "english":
		return LangEN
	default:
		return LangZH
	}
}

// Load 从 settings.json 或 FUCKSSH_LANG 加载语言；无配置时返回 false。
func Load() (bool, error) {
	if v := os.Getenv(envLangKey); v != "" {
		SetCurrent(Lang(v))
		return true, nil
	}

	path, err := settingsPath()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	var sf settingsFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return false, fmt.Errorf("i18n: parse settings: %w", err)
	}
	if sf.Lang == "" {
		return false, nil
	}
	SetCurrent(sf.Lang)
	return true, nil
}

// Save 将语言写入 settings.json。
func Save(lang Lang) error {
	if os.Getenv(envLangKey) != "" {
		SetCurrent(lang)
		return nil
	}
	if err := platform.MkdirSettingsDir(); err != nil {
		return err
	}
	path, err := settingsPath()
	if err != nil {
		return err
	}
	sf := settingsFile{Lang: normalizeLang(lang)}
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// EnsureInteractive 确保语言已设置：有配置则加载；无配置且 TTY 则弹出选择；否则默认 zh。
func EnsureInteractive(stderr io.Writer) error {
	loadOnce.Do(func() {})
	if v := os.Getenv(envLangKey); v != "" {
		SetCurrent(Lang(v))
		return nil
	}
	ok, err := Load()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	if isInteractive(stderr) {
		lang, err := pickLanguageFn()
		if err != nil {
			return err
		}
		SetCurrent(lang)
		return Save(lang)
	}
	SetCurrent(LangZH)
	return nil
}

func isInteractive(w io.Writer) bool {
	if isInteractiveOverride != nil {
		return isInteractiveOverride(w)
	}
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

func pickLanguageInteractive() (Lang, error) {
	var choice Lang = LangZH
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Lang]().
				Title(messagesZH[KeyLangSelectTitle]).
				Options(
					huh.NewOption(messagesZH[KeyLangZh], LangZH),
					huh.NewOption(messagesEN[KeyLangEn], LangEN),
				).
				Value(&choice),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return LangZH, nil
		}
		return "", err
	}
	return choice, nil
}

func settingsPath() (string, error) {
	if settingsPathOverride != "" {
		return settingsPathOverride, nil
	}
	return platform.SettingsPath()
}

// SetSettingsPathForTest 注入设置文件路径（单测）。
func SetSettingsPathForTest(path string) {
	settingsPathOverride = path
}

// SetInteractiveOverrideForTest 覆盖 TTY 检测（单测）。
func SetInteractiveOverrideForTest(fn func(io.Writer) bool) {
	isInteractiveOverride = fn
}

// ResetForTest 重置语言与单测钩子。
func ResetForTest() {
	settingsPathOverride = ""
	isInteractiveOverride = nil
	pickLanguageFn = pickLanguageInteractive
	SetCurrent(LangZH)
	loadOnce = sync.Once{}
}
