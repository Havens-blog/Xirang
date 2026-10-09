// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { QueueStat } from '@/api/opsagent'
import TaskQueue from './TaskQueue.vue'

const QUEUE: QueueStat = { depth: 2, capacity: 10, inflight: 1, concurrency: 3, overflowTotal: 0 }

describe('TaskQueue 任务队列（原型 agent-management.html 四格）', () => {
    it('四格：待处理=depth / 处理中=inflight / 已完成 / 失败 （跨 Agent 合计）', () => {
        const w = mount(TaskQueue, { props: { queue: QUEUE, totals: { done: 702, failed: 3 } } })
        const tiles = w.findAll('.tile')
        expect(tiles).toHaveLength(4)
        expect(tiles[0]!.find('.tile-value').text()).toBe('2')
        expect(tiles[0]!.find('.tile-label').text()).toBe('待处理')
        expect(tiles[1]!.find('.tile-value').text()).toBe('1')
        expect(tiles[2]!.find('.tile-value').text()).toBe('702')
        expect(tiles[3]!.find('.tile-value').text()).toBe('3')
        expect(w.find('.card-header h3').text()).toBe('任务队列')
    })

    it('容量/并发/溢出真实元信息可见', () => {
        const w = mount(TaskQueue, { props: { queue: QUEUE, totals: { done: 0, failed: 0 } } })
        expect(w.find('.task-queue-meta').text()).toContain('容量 10 · 并发 3 · 溢出总量 0')
    })
})