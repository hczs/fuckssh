package i18n

var messagesEN = map[string]string{
	KeyLangSelectTitle: "Choose language",
	KeyLangZh:          "中文 (Chinese)",
	KeyLangEn:          "English",

	KeyRootShort:   "Manage VPS hosts in ~/.ssh/config",
	KeyRootLong:    "fuckssh is a cross-platform CLI for SSH config, host listing, and search.",
	KeyAddShort:    "Add a VPS host via interactive wizard",
	KeyAddLong:     "Run the wizard to generate keys, update ssh config, and optionally deploy a public key.",
	KeyListShort:   "List hosts from ssh config",
	KeyListLong:    "Parse and display all Host entries in your ssh config file. Multiple aliases on one Host line are shown comma-separated.",
	KeySearchShort: "Search hosts by alias, hostname, or IP",
	KeySearchLong:  "Match hosts by alias, HostName, or IP substring.",
	KeyConfigFlag:  "path to ssh config file (default: ~/.ssh/config)",
	KeyCmdElapsedMs: "Elapsed: %d ms\n",

	KeySSHMissingWarning: "Error: ssh client not found in PATH. Install OpenSSH before running fuckssh add.",

	KeyWizardWelcome:         "Add a VPS — backs up and updates ~/.ssh/config",
	KeyWizardWelcomeETA:      "About 7 steps · ~2 minutes",
	KeyWizardConnModeTitle:   "Connection mode",
	KeyWizardConnModeDesc:    "New cloud VM: first option. Already have key login: second option.",
	KeyWizardModePassword:    "New server · console password (recommended)",
	KeyWizardModePasswordSub: "Generate a key and set up passwordless login",
	KeyWizardModeKey:         "Existing private key · config only",
	KeyWizardModeKeySub:      "No new key; write ssh config only",
	KeyWizardStepTitle:       "Step %d/%d · %s",
	KeyWizardHostIP:          "Server IP or hostname",
	KeyWizardPort:            "SSH port",
	KeyWizardPortDesc:        "Press Enter for 22; change only for non-default ports",
	KeyWizardUser:            "SSH username",
	KeyWizardUserDesc:        "Common: root, ubuntu (see your cloud panel)",
	KeyWizardPassword:        "SSH password",
	KeyWizardIdentityFile:    "Private key path",
	KeyWizardIdentityDesc:    "Full path, e.g. ~/.ssh/id_ed25519; tests login without uploading the key",
	KeyWizardAlias:           "Host alias",
	KeyWizardAliasDesc:       "Press Enter to auto-generate from IP/hostname",
	KeyWizardAliasPreview:    "Will use alias: %s",
	KeyWizardHostKeyHint:     "Note: server Host Key is not verified yet (planned for a later release)",
	KeyWizardErrEmpty:        "cannot be empty",
	KeyWizardTestingConn:     "Testing connection…",
	KeyWizardTestingConnInline: "Testing connection",
	KeyWizardConnOK:            "Connection successful",
	KeyWizardConnOKMs:          "✓ Connection OK (%d ms)",
	KeyWizardConnOKContinue:    "Press Enter to set Host alias",
	KeyWizardConnFailInline:    "✗ Wrong password or login denied — try again",
	KeyWizardConnRefused:       "✗ Cannot reach server — check IP, port, and firewall",
	KeyWizardConnTimeout:       "✗ Connection timed out — check network, IP, and port",
	KeyWizardConnUnreachable:   "✗ Host unreachable — check IP or hostname",
	KeyWizardConnFailGeneric:   "✗ Connection failed — check network and try again",
	KeyWizardAuthFailed:      "Invalid username or password, try again",
	KeyWizardKeyAuthFailed:   "Could not connect with this private key; check path, permissions, and remote authorized_keys",
	KeyWizardPassphraseNA:    "Passphrase-protected keys are not supported yet; use an unencrypted key",
	KeyWizardConfirmStep:     "Confirm",
	KeyWizardConfirmTitle:    "Run these actions?",
	KeyWizardConfirmSummaryPW: "About to:\n" +
		"  · Back up SSH config\n" +
		"  · Generate Ed25519 key\n" +
		"  · Host alias: %s\n" +
		"  · Login: %s@%s:%s\n" +
		"  · Config file: %s\n" +
		"  · Deploy public key to server",
	KeyWizardConfirmSummaryKey: "About to:\n" +
		"  · Back up SSH config\n" +
		"  · Host alias: %s\n" +
		"  · Login: %s@%s:%s\n" +
		"  · Config file: %s\n" +
		"  · Private key: %s",
	KeyWizardConfirmYes:        "Run",
	KeyWizardConfirmNo:         "Go back",
	KeyWizardRetryHint:         "Back to the form — edit your answers and confirm again",
	KeyWizardAliasConflictNote: "Host alias %q already exists in config; enter a new alias",
	KeyWizardAliasNew:          "New Host alias",
	KeyWizardAliasStillExists:  "That alias still exists; try another",
	KeyWizardCancelled:         "Cancelled; SSH config was not changed",
	KeyWizardProgressStep:      "[%d/%d] %s",
	KeyWizardErrFillBasic:  "fill IP/hostname, username, and private key path",
	KeyWizardErrKeyMissing: "private key not found: %s",
	KeyWizardErrKeyRead:    "cannot read private key: %v",
	KeyWizardErrAliasGen:   "cannot generate alias from HostName",

	KeySummaryTitle:       "Completed:\n",
	KeySummaryHeadline:    "Done! Run:",
	KeySummaryReadyHint:   "(key-based login; no password needed)",
	KeySummaryListHint:    "Run fuckssh list to see all hosts",
	KeySummaryBackup:      "  · Backed up SSH config to %s\n",
	KeySummaryKeygen:      "  · Generated Ed25519 key: %s\n",
	KeySummaryHostWritten: "  · Wrote Host %s to %s\n",
	KeySummaryDeployed:    "  · Deployed public key to remote ~/.ssh/authorized_keys\n",
	KeySummaryExistingKey: "  · Using existing private key: %s\n",
	KeySummaryHostKey:     "  · Security: Host Key is not verified yet (MITM risk). known_hosts support is planned.\n",
	KeySummaryNextStep:    "Config written to %s. Run: ssh %s\n",

	KeyListReading:   "Reading: %s\n",
	KeyListTotal:     "%d host(s)\n\n",
	KeyListEmpty:     "No Host entries found.\n",
	KeyListEmptyCTA:  "Run fuckssh add to add your first VPS.\n",
	KeySearchNoMatch: "No Host entries match %q.\n",
	KeySearchHint:    "Check the keyword or run fuckssh list to see all hosts.\n",
	KeySearchMeta:    "Search: %s, %d match(es)\n\n",
	KeySearchEmptyQ:  "search: query must not be empty",

	KeyTableAlias:    "ALIAS",
	KeyTableHostname: "HOSTNAME",
	KeyTablePort:     "PORT",
	KeyTableUser:     "USER",
}
