import { describe, expect, it } from 'vitest'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { audit, classifyPath, renderReport } from '../repo-hygiene.mjs'

describe('classifyPath', () => {
  it('干净路径零违规', () => {
    expect(classifyPath('apps/haven-console/src/main.ts')).toEqual([])
    expect(classifyPath('docs/features/x.md')).toEqual([])
    expect(classifyPath('scripts/repo-hygiene.mjs')).toEqual([])
  })
  it('禁用目录任意层级命中', () => {
    expect(classifyPath('apps/e-cam-web/dist/index.html')).toContain('禁用目录')
    expect(classifyPath('node_modules/x.js')).toContain('禁用目录')
    expect(classifyPath('services/x/.claude/y')).toContain('禁用目录')
    expect(classifyPath('services/x/build/f')).toContain('禁用目录')
  })
  it('logs 视图目录不违禁（只有 *.log 文件违禁）', () => {
    expect(classifyPath('apps/e-cam-web/src/views/logs/csv.ts')).toEqual([])
    expect(classifyPath('apps/e-cam-web/src/views/audit/logs/index.vue')).toEqual([])
    expect(classifyPath('services/eiam/logs/eiam.log')).toContain('二进制产物')
  })
  it('工具残留目录命中', () => {
    for (const d of ['.forge/a', '.kiro/b', '.superpowers/c', '.idea/d', 'services/x/build/f']) {
      expect(classifyPath(d).length).toBeGreaterThan(0)
    }
  })
  it('.vscode 目录不违禁（合法推荐扩展配置）', () => {
    expect(classifyPath('apps/e-cam-web/.vscode/extensions.json')).toEqual([])
  })
  it('二进制产物命中', () => {
    expect(classifyPath('tools/mcp-server.exe')).toContain('二进制产物')
    expect(classifyPath('x/y.dll')).toContain('二进制产物')
  })
})

describe('audit + renderReport', () => {
  let dir
  beforeEach(() => {
    dir = mkdtempSync(join(tmpdir(), 'hygiene-'))
  })
  afterEach(() => {
    rmSync(dir, { recursive: true, force: true })
  })

  it('小文件干净通过', () => {
    writeFileSync(join(dir, 'a.txt'), 'hello')
    const issues = audit(['a.txt'], dir)
    expect(issues).toEqual([])
    expect(renderReport(issues)).toContain('✅')
  })

  it('超过 3MB 文件被检出', () => {
    mkdirSync(join(dir, 'x'))
    writeFileSync(join(dir, 'x', 'big.bin'), Buffer.alloc(3 * 1024 * 1024 + 1))
    const issues = audit(['x/big.bin'], dir)
    expect(issues.some((i) => i.includes('超大文件'))).toBe(true)
    expect(renderReport(issues)).toContain('❌')
  })

  it('路径与大小违规并报', () => {
    mkdirSync(join(dir, 'logs'))
    writeFileSync(join(dir, 'logs', 'run.log'), 'x')
    const issues = audit(['logs/run.log'], dir)
    expect(issues.length).toBeGreaterThanOrEqual(1)
  })
})