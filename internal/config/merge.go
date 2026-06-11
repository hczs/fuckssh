package config

import (
	"fmt"
	"strings"
)

// ConflictAction 冲突解决动作。
type ConflictAction int

const (
	ConflictOverwrite ConflictAction = iota // 覆盖
	ConflictSkip                            // 跳过
	ConflictRename                          // 重命名
)

// ConflictInfo 描述一个 Host 别名冲突。
type ConflictInfo struct {
	Alias    string         // 冲突的别名
	Existing HostSummary    // 已有的 Host 信息
	Incoming HostSummary    // 待导入的 Host 信息
	Action   ConflictAction // 解决动作
	NewAlias string         // 重命名时的新别名
}

// HostSummary Host 摘要信息（用于冲突展示）。
type HostSummary struct {
	HostName     string
	User         string
	Port         string
	IdentityFile string
}

// RenameInfo 记录一次重命名操作的旧别名和新别名。
type RenameInfo struct {
	OldAlias        string
	NewAlias        string
	OldIdentityFile string // 重命名前的 IdentityFile，用于定位密钥文件
}

// MergeResult 合并结果。
type MergeResult struct {
	Imported  []string     // 成功导入的 Host 别名
	Skipped   []string     // 跳过的 Host 别名
	Overwrite []string     // 覆盖的 Host 别名
	Renamed   []string     // 重命名的 Host 别名（格式：old→new，用于展示）
	Renames   []RenameInfo // 重命名详情（用于后续密钥重命名等逻辑）
}

// FindConflicts 找出 incoming 和 existing 之间的所有别名冲突。
func FindConflicts(existing []HostEntry, incoming []HostEntry) []ConflictInfo {
	existingMap := make(map[string]HostEntry)
	for _, e := range existing {
		existingMap[e.Alias] = e
	}

	var conflicts []ConflictInfo
	for _, inc := range incoming {
		if ex, found := existingMap[inc.Alias]; found {
			conflicts = append(conflicts, ConflictInfo{
				Alias: inc.Alias,
				Existing: HostSummary{
					HostName:     ex.HostName,
					User:         ex.User,
					Port:         ex.Port,
					IdentityFile: ex.IdentityFile,
				},
				Incoming: HostSummary{
					HostName:     inc.HostName,
					User:         inc.User,
					Port:         inc.Port,
					IdentityFile: inc.IdentityFile,
				},
			})
		}
	}

	return conflicts
}

// MergeHosts 将 incoming 合并到 existing 中，按 conflicts 中的策略解决冲突。
func MergeHosts(existing []HostEntry, incoming []HostEntry, conflicts map[string]ConflictInfo) ([]HostEntry, *MergeResult) {
	result := &MergeResult{}

	// 建立现有 Host 的索引
	existingIdx := make(map[string]int)
	for i, e := range existing {
		existingIdx[e.Alias] = i
	}

	merged := make([]HostEntry, len(existing))
	copy(merged, existing)

	for _, inc := range incoming {
		idx, hasConflict := existingIdx[inc.Alias]
		if !hasConflict {
			// 无冲突，直接追加
			merged = append(merged, inc)
			result.Imported = append(result.Imported, inc.Alias)
			continue
		}

		// 有冲突，查解决策略
		ci, ok := conflicts[inc.Alias]
		if !ok {
			// 没有提供策略，默认跳过
			result.Skipped = append(result.Skipped, inc.Alias)
			continue
		}

		switch ci.Action {
		case ConflictOverwrite:
			merged[idx] = inc
			result.Overwrite = append(result.Overwrite, inc.Alias)
		case ConflictSkip:
			result.Skipped = append(result.Skipped, inc.Alias)
		case ConflictRename:
			newAlias := ci.NewAlias
			if newAlias == "" {
				result.Skipped = append(result.Skipped, inc.Alias)
				continue
			}
			renamed := inc
			renamed.Alias = newAlias
			if len(renamed.Aliases) > 0 {
				renamed.Aliases[0] = newAlias
			}
			merged = append(merged, renamed)
			result.Renamed = append(result.Renamed, fmt.Sprintf("%s→%s", inc.Alias, newAlias))
			result.Renames = append(result.Renames, RenameInfo{
				OldAlias:        inc.Alias,
				NewAlias:        newAlias,
				OldIdentityFile: inc.IdentityFile,
			})
		}
	}

	return merged, result
}

// FormatConflictSummary 格式化冲突摘要，用于终端展示。
func FormatConflictSummary(ci ConflictInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  ⚠️  [%s] 已存在同名 Host\n", ci.Alias)
	fmt.Fprintf(&b, "      现有: %s@%s:%s (%s)\n", ci.Existing.User, ci.Existing.HostName, ci.Existing.Port, ci.Existing.IdentityFile)
	fmt.Fprintf(&b, "      导入: %s@%s:%s (%s)\n", ci.Incoming.User, ci.Incoming.HostName, ci.Incoming.Port, ci.Incoming.IdentityFile)
	return b.String()
}
