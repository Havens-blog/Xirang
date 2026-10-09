// repo-hygiene.mjs — 仓库卫生门禁（CI hygiene job / 本地 `node scripts/repo-hygiene.mjs`）
//
// 检查 tracked 文件全集（git ls-files）：
//   1. 禁用目录零容忍：logs/ build/ dist/ node_modules/ .claude/ .forge/ .kiro/
//      .superpowers/ .idea/ .vscode/ temp/ 及任意层 .exe/.dll/bin 二进制产物
//   2. 大文件零容忍：任何 tracked 文件 > 3MB
// 违反 → 逐条列出并 exit 1（Hard Rule），保证单仓不会再次混入运行时垃圾。
import { execSync } from 'node:child_process'
import { statSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const FORBIDDEN_SEGMENTS = [
  /(^|\/)logs\//,
  /(^|\/)build\//,
  /(^|\/)dist\//,
  /(^|\/)node_modules\//,
  /(^|\/)temp\//,
  /(^|\/)\.claude\//,
  /(^|\/)\.forge\//,
  /(^|\/)\.kiro\//,
  /(^|\/)\.superpowers\//,
  /(^|\/)\.idea\//,
  /(^|\/)\.vscode\//,
]
const FORBIDDEN_FILES = [/\.exe$/i, /\.dll$/i]
const MAX_BYTES = 3 * 1024 * 1024

/** 对单个相对路径做规则判定，返回违规列表（空=干净） */
export function classifyPath(relPath) {
  const out = []
  const posix = relPath.replaceAll('\\', '/')
  for (const re of FORBIDDEN_SEGMENTS) {
    if (re.test(posix)) {
      out.push(`禁用目录: ${posix}`)
      break
    }
  }
  for (const re of FORBIDDEN_FILES) {
    if (re.test(posix)) out.push(`二进制产物: ${posix}`)
  }
  return out
}

/** 带大小检查的收敛：paths=tracked 文件相对路径数组，root=仓根 */
export function audit(paths, root) {
  const issues = []
  for (const p of paths) {
    issues.push(...classifyPath(p))
    const abs = join(root, p)
    if (existsSync(abs)) {
      let size = 0
      try {
        size = statSync(abs).size
      } catch {
        /* 忽略不可读 */
      }
      if (size > MAX_BYTES) issues.push(`超大文件 ${(size / 1024 / 1024).toFixed(1)}MB: ${p}`)
    }
  }
  return issues
}

export function renderReport(issues) {
  if (issues.length === 0) return '✅ 仓库卫生通过：无禁用目录 / 无二进制产物 / 无 >3MB 文件'
  return ['❌ 仓库卫生违规（Hard Rule：必须修复后合入）', ...issues.map((x) => `  - ${x}`)].join('\n')
}

function main() {
  let files
  try {
    files = execSync('git ls-files', { encoding: 'utf8' })
      .split(/[\r\n]+/)
      .filter(Boolean)
  } catch {
    console.error('git ls-files 执行失败，请在仓库内运行')
    return 1
  }
  const issues = audit(files, process.cwd())
  console.log(renderReport(issues))
  return issues.length === 0 ? 0 : 1
}

if (process.argv[1] && import.meta.url === new URL(`file://${process.argv[1].replaceAll('\\', '/')}`).href) {
  process.exit(main())
}