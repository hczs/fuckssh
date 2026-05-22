package cmd

import (
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// applyLocalizedHelp 在语言确定后刷新 Cobra 帮助文案。
func applyLocalizedHelp() {
	rootCmd.Short = i18n.T(i18n.KeyRootShort)
	rootCmd.Long = i18n.T(i18n.KeyRootLong)
	addCmd.Short = i18n.T(i18n.KeyAddShort)
	addCmd.Long = i18n.T(i18n.KeyAddLong)
	listCmd.Short = i18n.T(i18n.KeyListShort)
	listCmd.Long = i18n.T(i18n.KeyListLong)
	searchCmd.Short = i18n.T(i18n.KeySearchShort)
	searchCmd.Long = i18n.T(i18n.KeySearchLong)
	_ = rootCmd.PersistentFlags().Lookup("config")
	if f := rootCmd.PersistentFlags().Lookup("config"); f != nil {
		f.Usage = i18n.T(i18n.KeyConfigFlag)
	}
}
