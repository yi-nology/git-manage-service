import { onUnmounted } from 'vue'

export interface PollOptions {
  /** 轮询间隔（毫秒），默认 2000 */
  interval?: number
  /**
   * 每轮执行的工作函数。返回 true 继续轮询，false 停止。
   * 抛出异常时走 onError 决定去留。
   */
  tick: () => Promise<boolean> | boolean
  /**
   * tick 抛错时的钩子：返回 true 继续重试，false 停止。默认遇错即停。
   */
  onError?: (err: unknown) => boolean
}

/**
 * 统一的轮询 composable。
 *
 * 替代各处手写的 setInterval 轮询，修复两个共性问题：
 * 1. 上一轮请求未返回时跳过本轮（setInterval 不等异步完成，慢请求会堆积）；
 * 2. 组件卸载时自动停止，调用方不再需要各自的 onUnmounted 清理。
 */
export function usePolling({ interval = 2000, tick, onError }: PollOptions) {
  let timer: ReturnType<typeof setInterval> | null = null
  let inFlight = false

  function stop() {
    if (timer !== null) {
      clearInterval(timer)
      timer = null
    }
  }

  function start() {
    stop()
    timer = setInterval(async () => {
      if (inFlight) return
      inFlight = true
      try {
        if (!(await tick())) stop()
      } catch (err) {
        if (!onError?.(err)) stop()
      } finally {
        inFlight = false
      }
    }, interval)
  }

  onUnmounted(stop)
  return { start, stop }
}
