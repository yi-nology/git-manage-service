import { ref, onMounted, onUnmounted } from 'vue'

export interface WSMessage {
  topic: string
  type: string
  payload: any
}

export function useWebSocket(url?: string) {
  const wsUrl = url || `ws://${window.location.hostname}:12346/api/v1/ws/reviews`
  const connected = ref(false)
  const lastMessage = ref<WSMessage | null>(null)
  const listeners = new Map<string, Set<(payload: any) => void>>()
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  function connect() {
    if (ws?.readyState === WebSocket.OPEN) return

    try {
      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        connected.value = true
        if (ws) {
          ws.send(JSON.stringify({
            action: 'subscribe',
            topics: ['review']
          }))
        }
      }

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)
          lastMessage.value = msg
          const topicListeners = listeners.get(msg.topic)
          if (topicListeners) {
            topicListeners.forEach(fn => fn(msg.payload))
          }
          const allListeners = listeners.get('*')
          if (allListeners) {
            allListeners.forEach(fn => fn(msg))
          }
        } catch {}
      }

      ws.onclose = () => {
        connected.value = false
        scheduleReconnect()
      }

      ws.onerror = () => {
        connected.value = false
      }
    } catch {
      scheduleReconnect()
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) return
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, 5000)
  }

  function on(topic: string, callback: (payload: any) => void) {
    if (!listeners.has(topic)) {
      listeners.set(topic, new Set())
    }
    listeners.get(topic)!.add(callback)
    return () => {
      listeners.get(topic)?.delete(callback)
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    ws?.close()
    ws = null
    connected.value = false
  }

  onMounted(() => connect())
  onUnmounted(() => disconnect())

  return {
    connected,
    lastMessage,
    on,
    disconnect
  }
}
