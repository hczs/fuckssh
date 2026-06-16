package config

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/fuckssh/fuckssh/internal/vault"
)

// ErrImportKeyMissing 表示备份 archive 中找不到待导入 Host 需要的私钥。
var ErrImportKeyMissing = errors.New("config: private key missing from backup archive")

// KeyImportPlan 记录 import 前对 archive 密钥的规范化结果。
type KeyImportPlan struct {
	KeysRenamed []string // 如 shared→id_ed25519_fuckssh_staging
}

// KeyBasename 从 Host 条目推导私钥文件名（不含目录）。
func KeyBasename(entry HostEntry) string {
	if base := filepath.Base(entry.IdentityFile); base != "" && base != "." {
		return base
	}
	priv, _ := keys.KeyPaths(entry.Alias)
	return priv
}

// IncomingHostsToImport 返回实际会写入 config 的 incoming Host（排除 Skip，含重命名后的别名）。
func IncomingHostsToImport(incoming []HostEntry, mergeResult *MergeResult) []HostEntry {
	if mergeResult == nil {
		return append([]HostEntry(nil), incoming...)
	}

	skipped := make(map[string]bool, len(mergeResult.Skipped))
	for _, alias := range mergeResult.Skipped {
		skipped[alias] = true
	}

	renameByOld := make(map[string]RenameInfo, len(mergeResult.Renames))
	for _, r := range mergeResult.Renames {
		renameByOld[r.OldAlias] = r
	}

	var out []HostEntry
	for _, inc := range incoming {
		if skipped[inc.Alias] {
			continue
		}
		if r, ok := renameByOld[inc.Alias]; ok {
			h := inc
			h.Alias = r.NewAlias
			if len(h.Aliases) > 0 {
				h.Aliases[0] = r.NewAlias
			}
			out = append(out, h)
			continue
		}
		out = append(out, inc)
	}
	return out
}

// HostsReferencingKey 返回 merged config 中引用指定私钥文件名的 Host 别名列表。
func HostsReferencingKey(entries []HostEntry, keyBasename string) []string {
	var aliases []string
	for _, e := range entries {
		if identityKeyBasename(e) == keyBasename {
			aliases = append(aliases, e.Alias)
		}
	}
	return aliases
}

// BuildKeyWriteContext 为 vault 写盘构建密钥引用与归属信息。
func BuildKeyWriteContext(merged []HostEntry, toImport []HostEntry) vault.KeyWriteContext {
	ctx := vault.KeyWriteContext{
		KeyOwner: make(map[string]string),
		KeyRefs:  make(map[string][]string),
	}

	for _, e := range merged {
		base := identityKeyBasename(e)
		if base == "" {
			continue
		}
		ctx.KeyRefs[base] = append(ctx.KeyRefs[base], e.Alias)
	}

	for _, h := range toImport {
		targetPriv, _ := keys.KeyPaths(h.Alias)
		ctx.KeyOwner[targetPriv] = h.Alias
	}

	return ctx
}

// PrepareImportKeys 按 Host 别名规范化 archive 密钥、更新 merged IdentityFile，并过滤无需写入的密钥。
func PrepareImportKeys(
	files *[]vault.ExtractedFile,
	merged *[]HostEntry,
	incoming []HostEntry,
	mergeResult *MergeResult,
) (*KeyImportPlan, error) {
	plan := &KeyImportPlan{}
	toImport := IncomingHostsToImport(incoming, mergeResult)
	if len(toImport) == 0 {
		filterArchiveKeys(files, nil)
		return plan, nil
	}

	renameByNew := make(map[string]RenameInfo)
	if mergeResult != nil {
		for _, r := range mergeResult.Renames {
			renameByNew[r.NewAlias] = r
		}
	}

	keepKeys := make(map[string]bool, len(toImport))

	for _, host := range toImport {
		targetPriv, _ := keys.KeyPaths(host.Alias)
		rename := renameByNew[host.Alias]

		renamedFrom, err := ensureArchiveKey(files, candidateSourceKeys(host, rename), targetPriv)
		if err != nil {
			return nil, fmt.Errorf(
				"host %q: %w（若使用自定义密钥路径，请在源机器用新版 fuckssh 重新 export 后再导入）",
				host.Alias, err,
			)
		}
		if renamedFrom != "" && renamedFrom != targetPriv {
			plan.KeysRenamed = append(plan.KeysRenamed, fmt.Sprintf("%s→%s", renamedFrom, targetPriv))
		}

		if err := updateMergedIdentityFile(merged, host.Alias, targetPriv); err != nil {
			return nil, err
		}
		keepKeys[targetPriv] = true
	}

	filterArchiveKeys(files, keepKeys)
	return plan, nil
}

func identityKeyBasename(entry HostEntry) string {
	if base := filepath.Base(entry.IdentityFile); base != "" && base != "." {
		return base
	}
	priv, _ := keys.KeyPaths(entry.Alias)
	return priv
}

func candidateSourceKeys(host HostEntry, rename RenameInfo) []string {
	seen := make(map[string]bool)
	var candidates []string
	add := func(name string) {
		if name == "" || name == "." || seen[name] {
			return
		}
		seen[name] = true
		candidates = append(candidates, name)
	}

	if rename.NewAlias != "" && host.Alias == rename.NewAlias {
		add(filepath.Base(rename.OldIdentityFile))
		oldPriv, _ := keys.KeyPaths(rename.OldAlias)
		add(oldPriv)
	}
	add(KeyBasename(host))
	return candidates
}

// ensureArchiveKey 确保 archive 中存在 targetPriv 对应私钥；必要时从候选源文件名重命名。
// 返回实际使用的源文件名（与 target 相同时表示 archive 中已就位）。
func ensureArchiveKey(files *[]vault.ExtractedFile, sourceCandidates []string, targetPriv string) (string, error) {
	targetPath := "ssh/keys/" + targetPriv
	if archiveHasPath(*files, targetPath) {
		return targetPriv, nil
	}

	for _, sourceKey := range sourceCandidates {
		if sourceKey == "" || sourceKey == targetPriv {
			continue
		}
		sourcePath := "ssh/keys/" + sourceKey
		if !archiveHasPath(*files, sourcePath) {
			continue
		}
		if err := moveArchiveKeyEntry(files, sourcePath, targetPath); err != nil {
			return "", err
		}
		return sourceKey, nil
	}

	return "", fmt.Errorf("%w: %s", ErrImportKeyMissing, targetPriv)
}

func archiveHasPath(files []vault.ExtractedFile, archivePath string) bool {
	for _, f := range files {
		if f.ArchivePath == archivePath {
			return true
		}
	}
	return false
}

func moveArchiveKeyEntry(files *[]vault.ExtractedFile, oldPath, newPath string) error {
	if archiveHasPath(*files, newPath) {
		return nil
	}
	for i, f := range *files {
		if f.ArchivePath != oldPath {
			continue
		}
		(*files)[i].ArchivePath = newPath
		return nil
	}
	return fmt.Errorf("%w: %s", ErrImportKeyMissing, filepath.Base(newPath))
}

func updateMergedIdentityFile(merged *[]HostEntry, alias, targetPriv string) error {
	keysDir, err := platform.KeysDir()
	if err != nil {
		return err
	}
	absKey := filepath.Join(keysDir, targetPriv)
	identityRef, err := platform.IdentityFileRef(absKey)
	if err != nil {
		return err
	}

	for i, entry := range *merged {
		if entry.Alias != alias {
			continue
		}
		(*merged)[i].IdentityFile = identityRef
		return nil
	}
	return nil
}

func filterArchiveKeys(files *[]vault.ExtractedFile, keepKeys map[string]bool) {
	list := *files
	n := 0
	for _, f := range list {
		if f.ArchivePath == "ssh/config" {
			list[n] = f
			n++
			continue
		}
		if filepath.Dir(f.ArchivePath) != "ssh/keys" {
			list[n] = f
			n++
			continue
		}
		base := filepath.Base(f.ArchivePath)
		if keepKeys != nil && keepKeys[base] {
			list[n] = f
			n++
		}
	}
	*files = list[:n]
}
