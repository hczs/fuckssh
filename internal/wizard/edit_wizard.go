package wizard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// EditInput 为 edit 向导表单的输入，预填现有 Host 条目的值。
type EditInput struct {
	Alias        string
	HostName     string
	User         string
	Port         string
	IdentityFile string
	Remark       string
}

// RunEditWizard 启动编辑向导，预填现有 entry 的值，用户修改后返回 EditInput。
// configPath 为 ssh config 路径（别名冲突检测用）。
func RunEditWizard(configPath string, entry config.HostEntry) (*EditInput, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, fmt.Errorf("%w: config path must not be empty", ErrInvalidInput)
	}

	// 保存原始别名，用于别名冲突检测（允许与自身相同）。
	originalAlias := entry.Alias

	// 首次使用原始值填充，之后用用户修改过的 draft 填充。
	var draft *EditInput

	for {
		in, err := collectEditInput(configPath, originalAlias, draft, entry)
		if err != nil {
			return nil, mapWizardAbort(err)
		}

		err = confirmEditRun(in, configPath)
		if err == nil {
			return &in, nil
		}
		// 用户在确认页选择"返回修改"，保留当前值作为 draft 重新打开表单。
		if errors.Is(err, ErrWizardRetryForm) {
			draft = &in
			continue
		}
		return nil, mapWizardAbort(err)
	}
}

// collectEditInput 收集编辑表单输入，预填现有值。
// draft 非空时使用 draft（用户之前修改过的值），否则使用 entry 的原始值。
func collectEditInput(configPath, originalAlias string, draft *EditInput, entry config.HostEntry) (EditInput, error) {
	var in EditInput
	if draft != nil {
		in = *draft
	} else {
		in = EditInput{
			Alias:        entry.Alias,
			HostName:     entry.HostName,
			User:         entry.User,
			Port:         entry.Port,
			IdentityFile: entry.IdentityFile,
			Remark:       entry.Remark,
		}
	}

	// 别名字段需要特殊处理：允许与自身别名相同。
	aliasValidate := func(raw string) error {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return fmt.Errorf("%s", i18n.T(i18n.KeyWizardErrEmpty))
		}
		// 如果新别名与原始别名相同（不区分大小写），允许通过。
		if strings.EqualFold(raw, originalAlias) {
			return nil
		}
		// 否则检查是否与已有别名冲突。
		exists, err := config.HostAliasExists(configPath, raw)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: %q", config.ErrHostExists, raw)
		}
		return nil
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(1, i18n.KeyWizardHostIP)).
				Inline(true).
				Placeholder("1.2.3.4").
				Value(&in.HostName).
				Validate(nonEmpty(i18n.T(i18n.KeyWizardErrEmpty))),
		),
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(2, i18n.KeyWizardAlias)).
				Inline(true).
				Placeholder(i18n.T(i18n.KeyWizardAliasEmptyHint)).
				Value(&in.Alias).
				Validate(aliasValidate),
		),
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(3, i18n.KeyWizardUser)).
				Inline(true).
				Placeholder("root").
				Value(&in.User).
				Validate(nonEmpty(i18n.T(i18n.KeyWizardErrEmpty))),
		),
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(4, i18n.KeyWizardPort)).
				Inline(true).
				Placeholder("22").
				Value(&in.Port).
				Validate(validatePort),
		),
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(5, i18n.KeyWizardIdentityFile)).
				Inline(true).
				Placeholder("~/.ssh/id_ed25519").
				Value(&in.IdentityFile),
		),
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(6, i18n.KeyWizardRemark)).
				Inline(true).
				Placeholder(i18n.T(i18n.KeyWizardRemarkEmptyHint)).
				Value(&in.Remark),
		),
	).WithTheme(WizardTheme()).
		WithKeyMap(wizardFormKeyMap()).
		WithLayout(huh.LayoutStack).
		WithShowErrors(false)

	if err := form.Run(); err != nil {
		return EditInput{}, err
	}

	// Port 留空时默认 22。
	if strings.TrimSpace(in.Port) == "" {
		in.Port = "22"
	}

	return in, nil
}

// confirmEditRun 展示编辑确认页，用户确认后才执行更新。
func confirmEditRun(in EditInput, configPath string) error {
	summary := buildEditConfirmSummary(in, configPath)
	ok := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(confirmTitle()).
				Description(summary),
			huh.NewConfirm().
				Title(i18n.T(i18n.KeyWizardConfirmTitle)).
				Affirmative(i18n.T(i18n.KeyWizardConfirmYes)).
				Negative(i18n.T(i18n.KeyWizardConfirmNo)).
				Value(&ok),
		),
	).WithTheme(WizardTheme()).
		WithShowErrors(false)

	if err := form.Run(); err != nil {
		return err
	}
	if !ok {
		SetFormRetryHint(i18n.T(i18n.KeyWizardRetryHint))
		return ErrWizardRetryForm
	}
	return nil
}

// buildEditConfirmSummary 构建编辑确认页的摘要文本。
func buildEditConfirmSummary(in EditInput, configPath string) string {
	port := in.Port
	if port == "" || port == "22" {
		port = "22"
	}
	summary := i18n.T(
		i18n.KeyEditConfirmSummary,
		safeTTYString(in.Alias),
		in.User,
		strings.TrimSpace(in.HostName),
		port,
		safeTTYString(configPath),
	)
	if in.Remark != "" {
		summary += "\n" + i18n.T(i18n.KeyWizardConfirmRemark, in.Remark)
	}
	return summary
}
