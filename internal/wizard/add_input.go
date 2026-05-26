package wizard

// AddInput 为 add 向导全量表单的统一输入（可在测试中直接构造）。
type AddInput struct {
	Alias        string
	HostName     string
	User         string
	Port         string
	Mode         ConnectionMode
	Password     string
	IdentityFile string
	Remark       string
	// AuthTestOK 为 true 表示凭证字段已成功测连（提交表单前必填）。
	AuthTestOK bool
}

// ToPasswordModeInput 转为密码模式结构（供测连、finalize、执行链复用）。
func (a AddInput) ToPasswordModeInput() PasswordModeInput {
	return PasswordModeInput{
		HostName: a.HostName,
		User:     a.User,
		Password: a.Password,
		Port:     a.Port,
		Alias:    a.Alias,
		Remark:   a.Remark,
	}
}

// ToKeyModeInput 转为密钥模式结构。
func (a AddInput) ToKeyModeInput() KeyModeInput {
	return KeyModeInput{
		HostName:     a.HostName,
		User:         a.User,
		Port:         a.Port,
		Alias:        a.Alias,
		IdentityFile: a.IdentityFile,
		Remark:       a.Remark,
	}
}

// SyncFromPassword 将密码测连 scratch 的凭证写回 AddInput。
func (a *AddInput) SyncFromPassword(pw PasswordModeInput) {
	a.Password = pw.Password
}

// SyncFromKey 将私钥测连 scratch 的路径写回 AddInput。
func (a *AddInput) SyncFromKey(k KeyModeInput) {
	a.IdentityFile = k.IdentityFile
}

// clearCredentialsOnModeChange 切换连接方式时清空另一侧凭证并重置测连状态。
func clearCredentialsOnModeChange(in *AddInput, prev ConnectionMode) {
	if in.Mode == prev {
		return
	}
	in.Password = ""
	in.IdentityFile = ""
	in.AuthTestOK = false
}
