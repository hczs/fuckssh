# fuckssh 安装脚本（Windows PowerShell）
#
# 用法（在 PowerShell 中）：
#   irm https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.ps1 | iex
#   irm https://.../install.ps1 | iex; Install-Fuckssh -Version v0.1.0
#
# 或先下载再执行：
#   Set-ExecutionPolicy -Scope Process Bypass
#   .\scripts\install.ps1 -Version v0.1.0

# 将整个脚本包装为 ScriptBlock，兼容 irm | iex 和本地直接执行两种方式
$___install = {
    param(
        [string]$Version = "",
        [string]$BinDir = "",
        [string]$Owner = "hczs",
        [string]$Repo = "fuckssh"
    )

    $ErrorActionPreference = "Stop"
    $BinName = "fuckssh"
    $AliasName = "fs"

    function Write-Info([string]$Message) { Write-Host "==> $Message" }
    function Write-Warn([string]$Message) { Write-Warning $Message }

    function Get-AssetName {
        if ($env:OS -notmatch "Windows") {
            throw "此脚本仅适用于 Windows；macOS / Linux 请使用 scripts/install.sh"
        }
        $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
        switch ($arch) {
            "X64" { return "fuckssh_windows_x86_64.zip" }
            "Arm64" { return "fuckssh_windows_arm64.zip" }
            default { throw "不支持的 CPU 架构: $arch" }
        }
    }

    function Get-ReleaseTag {
        if ($Version) { return $Version }
        $uri = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
        $release = Invoke-RestMethod -Uri $uri -Headers @{ "User-Agent" = "fuckssh-installer" }
        if (-not $release.tag_name) { throw "无法获取最新版本，请使用 -Version 指定标签" }
        return $release.tag_name
    }

    function Setup-ShortAlias {
        param([string]$BinDir)

        $aliasPath = Join-Path $BinDir "$AliasName.cmd"

        # 检查别名是否已被其他命令占用
        $existing = Get-Command $AliasName -ErrorAction SilentlyContinue
        if ($existing) {
            # 如果已经是我们的包装脚本，则跳过
            if ($existing.Source -eq $aliasPath) {
                Write-Info "短别名已存在: $AliasName -> $BinName"
                return
            }
            Write-Warn "命令 '$AliasName' 已被占用 ($($existing.Source))，无法创建短别名。"
            Write-Info "请使用全称: $BinName"
            return
        }

        # 创建 .cmd 包装脚本
        $cmdContent = "@echo off`r`n$BinName.exe %*"
        Set-Content -Path $aliasPath -Value $cmdContent -Encoding ASCII
        Write-Info "短别名: $AliasName -> $BinName"
        Write-Info "可使用 '$AliasName' 或 '$BinName' 命令"
    }

    function Install-Fuckssh {
        $asset = Get-AssetName
        Write-Info "平台产物: $asset"

        $tag = Get-ReleaseTag
        Write-Info "版本: $tag"

        if (-not $BinDir) {
            $BinDir = Join-Path $env:USERPROFILE ".local\bin"
        }
        New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

        $tmp = Join-Path $env:TEMP ("fuckssh_" + [guid]::NewGuid().ToString("n"))
        New-Item -ItemType Directory -Force -Path $tmp | Out-Null
        try {
            $archive = Join-Path $tmp $asset
            $url = "https://github.com/$Owner/$Repo/releases/download/$tag/$asset"
            Write-Info "下载 $url"
            Invoke-WebRequest -Uri $url -OutFile $archive -UseBasicParsing

            $extract = Join-Path $tmp "extract"
            Expand-Archive -Path $archive -DestinationPath $extract -Force

            $src = Join-Path $extract "$BinName.exe"
            if (-not (Test-Path $src)) { throw "归档中未找到 $BinName.exe" }

            $dest = Join-Path $BinDir "$BinName.exe"
            Copy-Item -Path $src -Destination $dest -Force
            Write-Info "已安装: $dest"
        }
        finally {
            Remove-Item -Recurse -Force -Path $tmp -ErrorAction SilentlyContinue
        }

        Setup-ShortAlias -BinDir $BinDir

        $pathParts = $env:Path -split ';' | Where-Object { $_ -ne "" }
        if ($pathParts -contains $BinDir) {
            Write-Info "完成。可直接运行: $BinName version"
            return
        }

        Write-Warn "$BinDir 不在 PATH 中。"
        Write-Host ""
        Write-Host "请将下面目录加入用户 PATH（设置 → 系统 → 关于 → 高级系统设置 → 环境变量），然后重新打开终端："
        Write-Host ""
        Write-Host "  $BinDir"
        Write-Host ""
        Write-Host "或在当前 PowerShell 会话临时生效："
        Write-Host ""
        Write-Host ('  $env:Path = "' + $BinDir + ';" + $env:Path')
        Write-Host ""
        Write-Host "验证:"
        Write-Host ""
        Write-Host "  $BinName version"
    }

    Install-Fuckssh
}

# 本地执行时将参数透传给 ScriptBlock
& $___install @args
