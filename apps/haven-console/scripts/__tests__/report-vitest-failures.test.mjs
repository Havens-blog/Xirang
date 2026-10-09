import { describe, expect, it } from 'vitest'
import { parseFailures, renderAnnotations, renderHumanReport } from '../report-vitest-failures.mjs'

const JUNIT = `<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="vitest tests" tests="2" failures="1" errors="0">
  <testsuite name="src/components/AppSwitcher.test.ts" tests="2" failures="1">
    <testcase classname="AppSwitcher.test.ts" name="触发开合面板" file="src/components/AppSwitcher.test.ts" time="0.02"/>
    <testcase classname="PolicyList.test.ts" name="提交 > 保存时校验 JSON" file="src/views/policies/PolicyList.test.ts" time="0.03">
      <failure message="expected &quot;Deny&quot; to be &quot;Allow&quot;" type="AssertionError"><![CDATA[AssertionError: expected 'Deny' to be 'Allow']]></failure>
    </testcase>
  </testsuite>
</testsuites>`

describe('parseFailures', () => {
  it('提取失败 testcase（file 属性直取，message 解实体）', () => {
    const f = parseFailures(JUNIT)
    expect(f).toHaveLength(1)
    expect(f[0].file).toBe('src/views/policies/PolicyList.test.ts')
    expect(f[0].name).toBe('提交 > 保存时校验 JSON')
    expect(f[0].message).toContain('expected "Deny" to be "Allow"')
  })

  it('无 file 属性时回退 classname 首段', () => {
    const xml = '<testsuite><testcase classname="a > b.test.ts > d" name="t"><failure message="m"/></testcase></testsuite>'
    const f = parseFailures(xml)
    expect(f).toHaveLength(1)
    expect(f[0].file).toBe('a')
    expect(f[0].name).toBe('t')
  })

  it('无失败返回空数组', () => {
    expect(parseFailures('<testsuites><testsuite><testcase classname="a" name="b"/></testsuite></testsuites>')).toEqual([])
  })

  it('error 节点同样计入失败', () => {
    const xml = '<testsuite><testcase classname="x.ts" name="boom"><error message="Crash"/></testcase></testsuite>'
    expect(parseFailures(xml)).toHaveLength(1)
  })
})

describe('renderAnnotations', () => {
  it('每条失败一条 ::error，文件路径反斜杠归一', () => {
    const lines = renderAnnotations([{ file: 'src\\a\\b.test.ts', name: 't1', message: 'm1' }])
    expect(lines[0]).toContain('::error file=src/a/b.test.ts,line=1::t1 — m1')
  })

  it('超过 10 条归并进第 11 条', () => {
    const many = Array.from({ length: 13 }, (_, i) => ({ file: 'x.ts', name: `t${i}`, message: '' }))
    const lines = renderAnnotations(many)
    expect(lines).toHaveLength(11)
    expect(lines[10]).toContain('其余 3 个失败')
  })
})

describe('renderHumanReport', () => {
  it('无失败为空；有失败列出文件与断言', () => {
    expect(renderHumanReport([])).toEqual([])
    const out = renderHumanReport([{ file: 'a.ts', name: 'x', message: 'boom' }])
    expect(out[0]).toContain('Vitest 失败 1 条')
    expect(out.join('\n')).toContain('a.ts')
    expect(out.join('\n')).toContain('boom')
  })
})