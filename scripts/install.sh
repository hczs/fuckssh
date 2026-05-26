#!/bin/sh
# fuckssh 安装脚本（macOS / Linux）
#
# 用法：
#   curl -fsSL https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.sh | sh
#   curl -fsSL ... | sh -s -- --version v0.1.0
#   curl -fsSL ... | sh -s -- --bin-dir "$HOME/bin"
#
# 设计要点（与 zoxide / golangci-lint 等常见脚本一致）：
# 1. 用 /bin/sh + set -eu，兼容 curl | sh
# 2. 检测 OS/架构，拼出与 GoReleaser 一致的压缩包名
# 3. 从 GitHub Releases 下载并解压到用户目录（默认 ~/.local/bin）
# 4. 检查 PATH，提示用户如何加入（不擅自改 shell 配置）

set -eu

GITHUB_OWNER="${FUCKSSH_INSTALL_OWNER:-hczs}"
GITHUB_REPO="${FUCKSSH_INSTALL_REPO:-fuckssh}"
BIN_NAME="fuckssh"
DEFAULT_BIN_DIR="${HOME}/.local/bin"

# --- 小工具 ---

info() { printf '==> %s\n' "$*"; }
warn() { printf 'warning: %s\n' "$*" >&2; }
err() { printf 'error: %s\n' "$*" >&2; exit 1; }

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || err "需要命令: $1"
}

# curl | bash 若网络中断，未下完的脚本往往语法不完整；用花括号包一层 main，避免执行半截逻辑。
main() {
	parse_args "$@"

	archive_name="$(asset_name)" || exit $?
	info "平台产物: ${archive_name}"

	tag="$(resolve_tag "${VERSION_ARG}")" || exit $?
	info "版本: ${tag}"

	bin_dir="${BIN_DIR:-$DEFAULT_BIN_DIR}"
	mkdir -p "${bin_dir}"

	tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t fuckssh)" || err "无法创建临时目录"
	# shellcheck disable=SC2064
	trap 'rm -rf "${tmpdir}"' EXIT INT HUP

	archive="${tmpdir}/${archive_name}"
	download_release "${tag}" "${archive_name}" "${archive}"

	extract_archive "${archive}" "${tmpdir}/extract"
	install_binary "${tmpdir}/extract" "${bin_dir}/${BIN_NAME}"

	info "已安装: ${bin_dir}/${BIN_NAME}"
	print_path_hint "${bin_dir}"
}

parse_args() {
	VERSION_ARG=""
	BIN_DIR=""
	while [ $# -gt 0 ]; do
		case "$1" in
		--version)
			[ $# -ge 2 ] || err "--version 需要参数"
			VERSION_ARG="$2"
			shift 2
			;;
		--version=*)
			VERSION_ARG="${1#*=}"
			shift
			;;
		--bin-dir)
			[ $# -ge 2 ] || err "--bin-dir 需要参数"
			BIN_DIR="$2"
			shift 2
			;;
		--bin-dir=*)
			BIN_DIR="${1#*=}"
			shift
			;;
		-h | --help)
			usage
			exit 0
			;;
		*)
			err "未知参数: $1（使用 --help 查看用法）"
			;;
		esac
	done
}

usage() {
	cat <<EOF
fuckssh 安装脚本（macOS / Linux）

用法:
  install.sh [选项]

选项:
  --version <tag>   指定版本，如 v0.1.0；默认 latest
  --bin-dir <dir>   安装目录，默认: ${DEFAULT_BIN_DIR}
  -h, --help        显示帮助

示例:
  curl -fsSL https://raw.githubusercontent.com/${GITHUB_OWNER}/${GITHUB_REPO}/master/scripts/install.sh | sh
  curl -fsSL .../install.sh | sh -s -- --version v0.1.0
EOF
}

# 将 uname 结果映射为 GoReleaser 压缩包名（见 .goreleaser.yaml name_template）。
asset_name() {
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"
	arch="$(uname -m)"
	case "${arch}" in
	x86_64 | amd64) arch_suffix="x86_64" ;;
	aarch64 | arm64) arch_suffix="arm64" ;;
	*) err "不支持的 CPU 架构: $(uname -m)" ;;
	esac

	case "${os}" in
	linux) printf '%s\n' "fuckssh_linux_${arch_suffix}.tar.gz" ;;
	darwin)
		# 优先通用包（GoReleaser Arch=all）；旧版 release 可能只有分架构包
		printf '%s\n' "fuckssh_macos_all.tar.gz"
		;;
	*)
		err "此脚本仅支持 macOS / Linux；Windows 请使用 scripts/install.ps1"
		;;
	esac
}

resolve_tag() {
	requested="${1:-}"
	if [ -n "${requested}" ]; then
		printf '%s\n' "${requested}"
		return 0
	fi
	need_cmd curl
	tag="$(
		curl -fsSL "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest" |
			grep '"tag_name"' |
			head -n 1 |
			sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/'
	)" || true
	[ -n "${tag}" ] || err "无法获取最新版本，请指定 --version 或稍后再试"
	printf '%s\n' "${tag}"
}

# macOS 通用包不存在时（如 v0.1.0）回退到与 uname -m 对应的分架构包。
asset_name_fallback() {
	primary="$1"
	case "${primary}" in
	fuckssh_macos_all.tar.gz)
		arch="$(uname -m)"
		case "${arch}" in
		x86_64 | amd64) printf '%s\n' "fuckssh_macos_x86_64.tar.gz" ;;
		aarch64 | arm64) printf '%s\n' "fuckssh_macos_arm64.tar.gz" ;;
		*) return 1 ;;
		esac
		;;
	*)
		return 1
		;;
	esac
}

download_release() {
	tag="$1"
	name="$2"
	dest="$3"
	url="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${tag}/${name}"

	if fetch_url "${url}" "${dest}"; then
		return 0
	fi

	fallback="$(asset_name_fallback "${name}" 2>/dev/null || true)"
	if [ -n "${fallback}" ]; then
		warn "未找到 ${name}，改用 ${fallback}"
		name="${fallback}"
		url="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${tag}/${name}"
	fi

	fetch_url "${url}" "${dest}" || err "下载失败: ${url}"
}

fetch_url() {
	url="$1"
	dest="$2"
	if command -v curl >/dev/null 2>&1; then
		info "下载 ${url}"
		curl -fsSL -o "${dest}" "${url}"
	elif command -v wget >/dev/null 2>&1; then
		info "下载 ${url}"
		wget -qO "${dest}" "${url}"
	else
		err "需要 curl 或 wget"
	fi
}

extract_archive() {
	archive="$1"
	dest="$2"
	mkdir -p "${dest}"
	case "${archive}" in
	*.tar.gz | *.tgz)
		need_cmd tar
		tar -xzf "${archive}" -C "${dest}"
		;;
	*.zip)
		need_cmd unzip
		unzip -q "${archive}" -d "${dest}"
		;;
	*)
		err "不支持的压缩格式: ${archive}"
		;;
	esac
}

install_binary() {
	extract_dir="$1"
	dest="$2"
	# GoReleaser 默认把二进制放在归档根目录
	if [ -f "${extract_dir}/${BIN_NAME}" ]; then
		src="${extract_dir}/${BIN_NAME}"
	elif [ -f "${extract_dir}/${BIN_NAME}.exe" ]; then
		src="${extract_dir}/${BIN_NAME}.exe"
	else
		err "归档中未找到 ${BIN_NAME}"
	fi
	cp "${src}" "${dest}"
	chmod +x "${dest}"
}

print_path_hint() {
	bin_dir="$1"
	if printf '%s' ":${PATH}:" | grep -Fq ":${bin_dir}:"; then
		info "完成。可直接运行: ${BIN_NAME} version"
		return 0
	fi
	warn "${bin_dir} 不在 PATH 中，安装后还无法直接运行 ${BIN_NAME}。"
	cat <<EOF

请将下面一行加入 shell 配置文件（~/.bashrc、~/.zshrc 等），然后重新打开终端：

  export PATH="${bin_dir}:\$PATH"

验证:

  ${BIN_NAME} version
EOF
}

{
	main "$@"
}
