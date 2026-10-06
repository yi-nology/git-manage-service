// API 错误标记：拦截器弹出错误提示后给错误对象打 handled 标记，
// 调用方 catch 到带标记的错误时不应再弹一次（配合 toastApiError 使用）。
export interface HandledError extends Error {
  handled?: boolean
}

export function handledError(err: Error): HandledError {
  ;(err as HandledError).handled = true
  return err as HandledError
}

export function isHandledError(err: unknown): boolean {
  return !!(err as HandledError | null | undefined)?.handled
}
