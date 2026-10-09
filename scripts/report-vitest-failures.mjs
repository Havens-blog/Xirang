// report-vitest-failures.mjs — CI 故障自报（无网络依赖，纯 stdlib）
//
// 输入 vitest junit 报告（--reporter=junit --outputFile.junit=<path>）。
// 对每个失败 testcase 输出 GitHub workflow command `::error …::<测试名> — <失败信息>`
// → CI 检查页 annotation，公开仓库可经 check-runs API 匿名读取失败文件/断言。
// commit 时间轴/文件定位信息由 junit 的 classname 承载。
//
// 用法: node scripts/report-vitest-failures.mjs <junit-xml> [--annote-only]
//   --annote-only：仅在失败时输出 ::error（供测试断言，不打印人读报告）

import { readFileSync, existsSync } from 'node:fs'

/** 尽力提取 junit 失败项（正则，无 XML 依赖）。
 *  切分策略：按 `<testcase` 分段，段内（至下一 `<testcase` 前）出现 `<failure`/`<error` 即失败；
 *  属性值用 `key="…"` 按段独立匹配，天然容忍 classname/name 值内出现 `>`。 */
export function parseFailures(xml) {
  const out = []
  const segs = xml.split('<testcase').slice(1)
  for (const seg of segs) {
    if (!seg.includes('<failure') && !seg.includes('<error')) continue
    const attr = (key) => {
      // \s 前置词边界：防止 classname 尾部的 `name=` 被误匹配为 name 属性
      const m = seg.match(new RegExp(`(?:^|\\s)${key}="([^"]*)"`))
      return m ? m[1] : ''
    }
    const file = attr('file') || attr('classname').split(' > ')[0] || '(unknown)'
    const name = attr('name') || '(unnamed)'
    const fm = seg.match(/<failure[^>]*message="([^"]*)"/) ?? seg.match(/<error[^>]*message="([^"]*)"/)
    const msg = fm ? decodeXml(fm[1]) : seg.trim().slice(0, 240)
    out.push({ file, name, message: msg })
  }
  return out
}

function decodeXml(s) {
  return s
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'")
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&amp;/g, '&')
}

/**
 * 渲染 workflow commands：每个失败项一条 ::error。
 * annotation 上限（每 job 10 条）内归并：超出部分合并进第一条 message。
 */
export function renderAnnotations(failures, limit = 10) {
  if (failures.length === 0) return []
  const lines = []
  const shown = failures.slice(0, limit)
  for (const f of shown) {
    const file = (f.file || '').replace(/\\/g, '/').replace(/\r?\n/g, ' ')
    const msg = `${f.name} — ${(f.message || '').slice(0, 220)}`
    lines.push(`::error file=${file},line=1::${msg}`)
  }
  if (failures.length > limit) {
    const rest = failures
      .slice(limit)
      .map((f) => f.name)
      .join('、')
    lines.push(`::error file=vitest-junit.xml,line=1::其余 ${failures.length - limit} 个失败：${rest}`)
  }
  return lines
}

export function renderHumanReport(failures) {
  if (failures.length === 0) return []
  const lines = [`❌ Vitest 失败 ${failures.length} 条（详见 job annotations）`]
  for (const f of failures) {
    lines.push(`- [${f.file}] ${f.name}`)
    if (f.message) lines.push(`    ${f.message.slice(0, 300)}`)
  }
  return lines
}

export function main(argv = process.argv) {
  const xmlPath = argv[2]
  const annoteOnly = argv.includes('--annote-only')
  if (!xmlPath || !existsSync(xmlPath)) {
    console.error('用法: node report-vitest-failures.mjs <junit-xml> [--annote-only]')
    return 1
  }
  const xml = readFileSync(xmlPath, 'utf8')
  const failures = parseFailures(xml)
  renderAnnotations(failures).forEach((l) => console.log(l))
  if (!annoteOnly) renderHumanReport(failures).forEach((l) => console.log(l))
  return failures.length === 0 ? 0 : 1
}

const isDirect = process.argv[1] && import.meta.url === new URL(`file://${process.argv[1].replace(/\\/g, '/')}`).href
if (isDirect) process.exit(main())