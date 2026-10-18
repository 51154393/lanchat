<template>
  <div class="flex h-screen w-screen overflow-hidden bg-bg-chat">
    <div class="w-[280px] shrink-0 bg-[#2e2e2e]">
      <DeviceList />
    </div>
    <div class="flex-1 min-w-0 flex flex-col bg-bg-chat">
      <ChatView ref="chatView" />
    </div>
    <FileRequestDialog
      :show="showFileRequest"
      :transfer="incomingTransfer"
      :sender-name="incomingSenderName"
      :sender-ip="incomingSenderIP"
      :sender-platform="incomingSenderPlatform"
      @close="showFileRequest = false"
      @accepted="onFileAccepted"
    />
    <FileReceivePanel
      ref="receivePanelRef"
      :show="showReceivePanel"
      :transfer-id="receiveTransferId"
      :file-name="receiveFileName"
      :file-size="receiveFileSize"
      :save-path="receiveSavePath"
      @close="showReceivePanel = false"
    />
  </div>
</template>

<script setup>
import {onMounted, ref, nextTick} from 'vue'
import {EventsOn} from '../wailsjs/runtime/runtime'
import {GetDevices, GetConfig} from '../wailsjs/go/app/App'
import DeviceList from './components/DeviceList.vue'
import ChatView from './components/ChatView.vue'
import FileRequestDialog from './components/FileRequestDialog.vue'
import FileReceivePanel from './components/FileReceivePanel.vue'
import {useDevicesStore} from './stores/devices'
import {useChatStore} from './stores/chat'

const devicesStore = useDevicesStore()
const chatStore = useChatStore()
const chatView = ref(null)

const showFileRequest = ref(false)
const incomingTransfer = ref(null)
const incomingSenderName = ref('')
const incomingSenderIP = ref('')
const incomingSenderPlatform = ref('')

const showReceivePanel = ref(false)
const receiveTransferId = ref('')
const receiveFileName = ref('')
const receiveFileSize = ref(0)
const receiveSavePath = ref('')
const receivePanelRef = ref(null)

async function loadDevices() {
  try {
    const devices = await GetDevices()
    devicesStore.setDevices(devices)
  } catch (err) {
    console.error('Failed to load devices:', err)
  }
}

async function initApp() {
  try {
    const config = await GetConfig()
    window.__DEVICE_ID__ = config.device_id
  } catch (err) {
    console.error('Failed to get config:', err)
  }
}

function onFileAccepted(savePath) {
  const t = incomingTransfer.value
  if (!t) return

  receiveTransferId.value = t.transfer_id
  receiveFileName.value = t.file_name
  receiveFileSize.value = t.file_size
  receiveSavePath.value = savePath || ''
  showFileRequest.value = false

  nextTick(() => {
    showReceivePanel.value = true
    receivePanelRef.value?.reset()
    receivePanelRef.value?.setupListeners()
  })
}

onMounted(async () => {
  await initApp()
  await loadDevices()

  EventsOn('device:joined', (device) => {
    devicesStore.addDevice(device)
  })

  EventsOn('device:left', (device) => {
    devicesStore.removeDevice(device.device_id)
  })

  EventsOn('device:offline', (device) => {
    devicesStore.updateDeviceStatus(device.device_id, 0)
  })

  EventsOn('device:online', (device) => {
    devicesStore.updateDeviceStatus(device.device_id, 1)
  })

  EventsOn('chat:message', (message) => {
    chatStore.addMessage(message)
    chatView.value?.scrollToBottom()
  })

  EventsOn('file:request', (data) => {
    incomingTransfer.value = data.transfer
    incomingSenderName.value = data.sender_name || '未知设备'
    incomingSenderIP.value = data.sender_ip || ''
    incomingSenderPlatform.value = data.sender_platform || ''
    showFileRequest.value = true
  })
})
</script>

<style>
/* Global resets handled by Tailwind base layer */
</style>
