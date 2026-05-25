package wizard

import "strings"

// revealState 控制堆叠表单中已「解锁」展示的分组数量，避免未填项提前出现、已填项消失。
type revealState struct {
	n int
}

func (r *revealState) showThrough(index int) {
	if index+1 > r.n {
		r.n = index + 1
	}
}

// lockThrough 在测连失败等场景收回后续步骤（例如隐藏尚未完成的别名项）。
func (r *revealState) lockThrough(index int) {
	max := index + 1
	if r.n > max {
		r.n = max
	}
}

func hideUntilRevealed(index int, r *revealState) func() bool {
	return func() bool { return index >= r.n }
}

// seedReveal 根据已填字段恢复堆叠可见范围（密码未通过测连时不展示别名步）。
func seedReveal(r *revealState, in PasswordModeInput) {
	if strings.TrimSpace(in.HostName) != "" {
		r.showThrough(1)
	}
	if strings.TrimSpace(in.User) != "" {
		r.showThrough(3)
	}
	if strings.TrimSpace(in.Password) != "" {
		r.showThrough(4)
	}
	if strings.TrimSpace(in.Alias) != "" {
		r.showThrough(5)
	}
}

// seedRevealKey 密钥模式：私钥测连通过后才展示别名步。
func seedRevealKey(r *revealState, in KeyModeInput) {
	if strings.TrimSpace(in.HostName) != "" {
		r.showThrough(1)
	}
	if strings.TrimSpace(in.User) != "" {
		r.showThrough(3)
	}
	if strings.TrimSpace(in.IdentityFile) != "" {
		r.showThrough(4)
	}
	if strings.TrimSpace(in.Alias) != "" {
		r.showThrough(5)
	}
}
