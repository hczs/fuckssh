package wizard

// revealState 控制堆叠表单中已「解锁」展示的分组数量，避免未填项提前出现、已填项消失。
type revealState struct {
	n int
}

func (r *revealState) showThrough(index int) {
	if index+1 > r.n {
		r.n = index + 1
	}
}

func hideUntilRevealed(index int, r *revealState) func() bool {
	return func() bool { return index >= r.n }
}
