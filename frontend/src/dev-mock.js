const DEVICE_ID = 'dev-local-001'

const listeners = {}

const mockRuntime = {
  EventsOnMultiple(eventName, callback, maxCallbacks) {
    if (!listeners[eventName]) listeners[eventName] = []
    listeners[eventName].push({ callback, maxCallbacks, count: 0 })
    return () => {
      listeners[eventName] = (listeners[eventName] || []).filter(l => l.callback !== callback)
    }
  },
  EventsOff(eventName, ...additionalEventNames) {
    const names = [eventName, ...additionalEventNames]
    names.forEach(name => { delete listeners[name] })
  },
  EventsEmit(eventName, ...args) {
    const cbs = listeners[eventName] || []
    const remaining = []
    for (const l of cbs) {
      l.callback(...args)
      l.count++
      if (l.maxCallbacks === -1 || l.count < l.maxCallbacks) {
        remaining.push(l)
      }
    }
    listeners[eventName] = remaining
  },
  LogPrint: console.log,
  LogDebug: console.debug,
  LogInfo: console.info,
  LogWarning: console.warn,
  LogError: console.error,
}

const now = Date.now()
const mockMessages = [
  { id: '1', type: 'chat', content: '你好，这是测试消息', sender_id: 'peer-001', sender_name: 'user1@Win-PC', timestamp: now - 300000 },
  { id: '2', type: 'chat', content: '收到，我来发个文件', sender_id: DEVICE_ID, sender_name: 'sun@MacBook', timestamp: now - 240000 },
  { id: '3', type: 'chat', content: '好的，请发送', sender_id: 'peer-002', sender_name: 'user2@Win-PC2', timestamp: now - 180000 },
  { id: '4', type: 'file', content: '发送了文件', sender_id: 'peer-001', sender_name: 'user1@Win-PC', timestamp: now - 120000, file_name: '项目文档.pdf', file_size: 2048576, file_path: '/Users/sun/Downloads/项目文档.pdf' },
  { id: '5', type: 'file', content: '发送了文件', sender_id: DEVICE_ID, sender_name: 'sun@MacBook', timestamp: now - 60000, file_name: '设计稿.psd', file_size: 5242880, file_path: '/Users/sun/Downloads/设计稿.psd' },
  { id: '6', type: 'chat', content: '文件收到了，谢谢！', sender_id: DEVICE_ID, sender_name: 'sun@MacBook', timestamp: now - 30000 },
]

const mockDevices = [
  { device_id: 'peer-001', device_name: 'Win-PC', username: 'user1', ip: '192.168.1.101', port: 19999, platform: 'windows', status: 1 },
  { device_id: 'peer-002', device_name: 'Win-PC2', username: 'user2', ip: '192.168.1.102', port: 19999, platform: 'windows', status: 1 },
]

const mockConfig = {
  device_id: DEVICE_ID,
  device_name: 'MacBook',
  username: 'sun',
  tcp_listen_port: 19999,
}

let msgCounter = 100

const mockApp = {
  GetConfig: () => Promise.resolve(mockConfig),
  GetDevices: () => Promise.resolve(mockDevices),
  GetOnlineDevices: () => Promise.resolve(mockDevices),
  GetOnlineCount: () => Promise.resolve(mockDevices.length),
  GetMessages: () => Promise.resolve([...mockMessages]),
  GetTransfers: () => Promise.resolve([]),

  GetFileInfo: (filePath) => {
    const name = filePath.split('/').pop().split('\\').pop()
    return Promise.resolve({ name, size: 1048576 })
  },

  SendMessage: (content) => {
    const msg = {
      id: String(++msgCounter),
      type: 'chat',
      content,
      sender_id: DEVICE_ID,
      sender_name: 'sun@MacBook',
      timestamp: Date.now(),
    }
    mockMessages.push(msg)
    setTimeout(() => mockRuntime.EventsEmit('chat:message', msg), 100)
    return Promise.resolve()
  },

  SelectFile: () => {
    console.log('[Mock] SelectFile called')
    return Promise.resolve('/Users/sun/Downloads/test-file.pdf')
  },
  SelectSavePath: (fileName) => {
    console.log('[Mock] SelectSavePath:', fileName)
    return Promise.resolve(`/Users/sun/Downloads/${fileName}`)
  },
  OpenFileLocation: (filePath) => {
    console.log('[Mock] OpenFileLocation:', filePath)
    window.open(`file://${filePath}`)
    return Promise.resolve()
  },

  SendFile: (peerID, filePath) => {
    console.log('[Mock] SendFile:', peerID, filePath)
    const transferId = 'mock-transfer-' + Date.now()
    setTimeout(() => {
      mockRuntime.EventsEmit('file:transfer_started', { transfer_id: transferId, file_size: 1024000 })
      let progress = 0
      const interval = setInterval(() => {
        progress += Math.random() * 15
        if (progress >= 100) {
          progress = 100
          clearInterval(interval)
          mockRuntime.EventsEmit('file:transfer_completed', { transfer_id: transferId })
        } else {
          mockRuntime.EventsEmit('file:transfer_progress', {
            transfer_id: transferId,
            transferred: Math.floor(1024000 * progress / 100),
            total: 1024000,
            progress,
          })
        }
      }, 300)
    }, 500)
    return Promise.resolve()
  },
  AcceptFile: (transferID, savePath) => {
    console.log('[Mock] AcceptFile:', transferID, savePath)
    let progress = 0
    setTimeout(() => {
      const interval = setInterval(() => {
        progress += Math.random() * 12
        if (progress >= 100) {
          progress = 100
          clearInterval(interval)
          mockRuntime.EventsEmit('file:transfer_completed', { transfer_id: transferID })
        } else {
          mockRuntime.EventsEmit('file:transfer_progress', {
            transfer_id: transferID,
            transferred: Math.floor(2048576 * progress / 100),
            total: 2048576,
            progress,
          })
        }
      }, 300)
    }, 500)
    return Promise.resolve()
  },
  RejectFile: (transferID) => {
    console.log('[Mock] RejectFile:', transferID)
    return Promise.resolve()
  },
  CancelFile: (transferID) => {
    console.log('[Mock] CancelFile:', transferID)
    setTimeout(() => {
      mockRuntime.EventsEmit('file:transfer_cancelled', { transfer_id: transferID, direction: 0 })
    }, 200)
    return Promise.resolve()
  },
  AddPeerManually: () => Promise.resolve(),
}

if (!window.go) {
  window.go = { app: { App: mockApp } }
  console.log('[Mock] window.go injected')
}

if (!window.runtime) {
  window.runtime = mockRuntime
  console.log('[Mock] window.runtime injected')
}

if (!window.__DEVICE_ID__) {
  window.__DEVICE_ID__ = DEVICE_ID
}

window.__mock_emit__ = mockRuntime.EventsEmit
