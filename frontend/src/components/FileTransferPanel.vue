<template>
  <div class="fixed inset-0 z-[1000] flex items-center justify-center" v-if="show">
    <div class="fixed inset-0 bg-black/50"></div>
    <div class="relative bg-white rounded-xl w-[480px] max-h-[80vh] flex flex-col shadow-panel">
      <div class="flex items-center justify-between px-5 py-4">
        <h3 class="text-base font-semibold text-text-primary m-0">发送文件</h3>
        <button v-if="canClose" class="w-7 h-7 border-none bg-transparent rounded-md text-xl text-gray-400 cursor-pointer flex items-center justify-center hover:bg-gray-100" @click="close">×</button>
        <button v-else class="w-7 h-7 border-none bg-transparent rounded-md text-base text-danger cursor-pointer flex items-center justify-center hover:bg-red-50" @click="cancelTransfer">✕</button>
      </div>
      <div class="p-5 overflow-y-auto flex-1">
        <div v-if="currentStep === 'select'">
          <p class="text-sm text-text-secondary mb-4">选择要发送的文件</p>
          <button class="w-full py-8 bg-gray-50 border-2 border-dashed border-gray-300 rounded-lg text-sm text-gray-500 cursor-pointer transition-all hover:bg-gray-100 hover:border-primary hover:text-primary" @click="selectFile">
            选择文件
          </button>
          <div v-if="selectedFile" class="flex items-center p-3 bg-gray-50 rounded-lg mt-4">
            <img :src="fileIcon" class="w-8 h-8 mr-3 shrink-0" alt="">
            <div class="flex-1">
              <div class="text-sm font-medium text-text-primary break-all mb-1">{{ selectedFile.name }}</div>
              <div class="text-xs text-text-tertiary">{{ formatFileSize(selectedFile.size) }}</div>
            </div>
          </div>
        </div>

        <div v-if="currentStep === 'peer'">
          <p class="text-sm text-text-secondary mb-4">选择接收设备</p>
          <div class="flex flex-col gap-2">
            <div
              v-for="device in onlineDevices"
              :key="device.device_id"
              class="flex items-center p-3 bg-gray-50 rounded-lg cursor-pointer transition-colors hover:bg-gray-200"
              @click="selectPeer(device)"
            >
              <div class="mr-3">
                <img
                  alt="avatar"
                  src="@/assets/images/avatar/win.svg"
                  class="rounded-full w-10 h-10"
                  v-if="device.platform === 'windows'"
                />
                <img
                  alt="avatar"
                  src="@/assets/images/avatar/mac.svg"
                  class="rounded-full w-10 h-10"
                  v-else-if="device.platform === 'darwin'"
                />
                <img
                  alt="avatar"
                  src="@/assets/images/avatar/linux.svg"
                  class="rounded-full w-10 h-10"
                  v-else
                />
              </div>
             
              <div class="flex-1">
                <div class="text-sm font-medium text-text-primary mb-1">{{ device.username }}@{{ device.device_name }}</div>
                <div class="text-xs text-text-tertiary">{{ device.ip }}</div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="currentStep === 'waiting'">
          <p class="text-sm text-text-secondary mb-4">等待对方接收...</p>
          <div class="flex items-center p-3 bg-gray-50 rounded-lg mt-4">
            <img :src="fileIcon" class="w-8 h-8 mr-3 shrink-0" alt="">
            <div class="flex-1">
              <div class="text-sm font-medium text-text-primary break-all mb-1">{{ selectedFile.name }}</div>
              <div class="text-xs text-text-tertiary">{{ formatFileSize(selectedFile.size) }}</div>
            </div>
          </div>
          <div class="text-center text-[13px] text-text-tertiary mt-4">对方接受后将自动开始传输</div>
        </div>

        <div v-if="currentStep === 'sending'">
          <p class="text-sm text-text-secondary mb-4">正在发送文件...</p>
          <div class="flex justify-between items-center mb-2">
            <div class="text-[13px] font-medium text-gray-700 break-all flex-1 mr-3">{{ selectedFile?.name }}</div>
            <div class="text-xs text-text-tertiary whitespace-nowrap">{{ formatFileSize(transferredBytes) }} / {{ formatFileSize(selectedFile?.size) }}</div>
          </div>
          <div class="py-1 pb-3">
            <div class="h-2 bg-gray-200 rounded-full overflow-hidden mb-2">
              <div class="h-full bg-primary rounded-full transition-[width] duration-300" :style="{width: progress + '%'}"></div>
            </div>
            <div class="text-center text-sm text-text-secondary">{{ Math.round(progress) }}%</div>
          </div>
        </div>

        <div v-if="currentStep === 'completed'" class="text-center py-3">
          <div class="mb-3 flex justify-center">
            <img src="@/assets/images/success.svg" class="w-8 h-8">
          </div>
          <p class="text-base font-medium text-text-primary mb-2">文件发送完成</p>
          <div class="text-[13px] text-text-tertiary break-all">{{ selectedFile?.name }}</div>
        </div>

        <div v-if="currentStep === 'failed'" class="text-center py-3">
          <div class="mb-3 flex justify-center">
            <img src="@/assets/images/error.svg" class="w-8 h-8">
          </div>
          <p class="text-base font-medium text-text-primary mb-2">发送失败</p>
          <div class="text-[13px] text-text-tertiary">{{ errorMsg || '文件传输过程中出错' }}</div>
        </div>

        <div v-if="currentStep === 'cancelled'" class="text-center py-3">
          <div class="mb-3 flex justify-center">
            <img src="@/assets/images/cancelled.svg" class="w-8 h-8">
          </div>
          <p class="text-base font-medium text-text-primary mb-2">已取消发送</p>
          <div class="text-[13px] text-text-tertiary break-all">{{ selectedFile?.name }}</div>
        </div>
      </div>
      <div class="flex justify-end gap-3 px-5 py-4" v-if="currentStep === 'select'">
        <button class="px-5 py-2 border-none rounded-md text-sm cursor-pointer transition-colors bg-gray-100 text-gray-600 hover:bg-gray-200" @click="close">取消</button>
        <button class="px-5 py-2 border-none rounded-md text-sm cursor-pointer transition-colors bg-primary text-white hover:bg-primary-hover disabled:bg-gray-300 disabled:cursor-not-allowed" @click="nextStep" :disabled="!selectedFile">下一步</button>
      </div>
      <div class="flex justify-end gap-3 px-5 py-4 " v-if="currentStep === 'waiting' || currentStep === 'sending'">
        <button class="px-5 py-2 border-none rounded-md text-sm cursor-pointer transition-colors bg-danger text-white hover:bg-danger-hover" @click="cancelTransfer">取消发送</button>
      </div>
      <div class="flex justify-end gap-3 px-5 py-4" v-if="currentStep === 'completed' || currentStep === 'failed' || currentStep === 'cancelled'">
        <button class="px-5 py-2 border-none rounded-md text-sm cursor-pointer transition-colors bg-primary text-white hover:bg-primary-hover" @click="close">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import {ref, computed, onUnmounted} from 'vue'
import {useDevicesStore} from '../stores/devices'
import {SelectFile, SendFile, CancelFile, GetFileInfo} from '../../wailsjs/go/app/App'
import {EventsOn, EventsOff} from '../../wailsjs/runtime/runtime'
import {getFileIcon, formatFileSize} from '../utils/file'

const props = defineProps({
  show: Boolean
})

const emit = defineEmits(['close'])

const devicesStore = useDevicesStore()
const onlineDevices = computed(() => devicesStore.onlineDevices)

const currentStep = ref('select')
const selectedFile = ref(null)
const selectedPeer = ref(null)
const progress = ref(0)
const transferredBytes = ref(0)
const errorMsg = ref('')
const currentTransferId = ref(null)

function close() {
  emit('close')
  reset()
}

const canClose = computed(() => {
  return currentStep.value !== 'waiting' && currentStep.value !== 'sending'
})

const fileIcon = computed(() => getFileIcon(selectedFile.value?.name))

async function cancelTransfer() {
  if (currentTransferId.value) {
    try {
      await CancelFile(currentTransferId.value)
    } catch (err) {
      console.error('Failed to cancel transfer:', err)
    }
  }
  currentStep.value = 'cancelled'
}

function reset() {
  currentStep.value = 'select'
  selectedFile.value = null
  selectedPeer.value = null
  progress.value = 0
  transferredBytes.value = 0
  errorMsg.value = ''
  currentTransferId.value = null
  EventsOff('file:transfer_started')
  EventsOff('file:transfer_progress')
  EventsOff('file:transfer_completed')
  EventsOff('file:transfer_failed')
  EventsOff('file:transfer_rejected')
  EventsOff('file:transfer_cancelled')
}

async function selectFile() {
  try {
    const filePath = await SelectFile()
    if (filePath) {
      selectedFile.value = {
        path: filePath,
        name: filePath.split('/').pop().split('\\').pop(),
        size: 0
      }
      try {
        const info = await GetFileInfo(filePath)
        if (selectedFile.value) {
          selectedFile.value.size = info.size
        }
      } catch (e) {
        console.error('Failed to get file info:', e)
      }
    }
  } catch (err) {
    console.error('Failed to select file:', err)
  }
}

function nextStep() {
  if (currentStep.value === 'select' && selectedFile.value) {
    currentStep.value = 'peer'
  }
}

async function selectPeer(device) {
  selectedPeer.value = device
  currentStep.value = 'waiting'

  try {
    EventsOn('file:transfer_started', (data) => {
      if (data.transfer_id) {
        currentTransferId.value = data.transfer_id
      }
      if (data.file_size && selectedFile.value) {
        selectedFile.value.size = data.file_size
      }
    })
    await SendFile(device.device_id, selectedFile.value.path)
    listenForProgress()
  } catch (err) {
    console.error('Failed to send file:', err)
    errorMsg.value = err.message || '发送文件失败'
    currentStep.value = 'failed'
  }
}

function listenForProgress() {
  EventsOn('file:transfer_progress', (data) => {
    if (currentTransferId.value && data.transfer_id !== currentTransferId.value) return
    progress.value = data.progress || 0
    transferredBytes.value = data.transferred || 0
    if (currentStep.value === 'waiting') {
      currentStep.value = 'sending'
    }
  })

  EventsOn('file:transfer_completed', (data) => {
    if (currentTransferId.value && data.transfer_id !== currentTransferId.value) return
    progress.value = 100
    currentStep.value = 'completed'
    EventsOff('file:transfer_progress')
    EventsOff('file:transfer_completed')
    EventsOff('file:transfer_failed')
  })

  EventsOn('file:transfer_failed', (data) => {
    if (currentTransferId.value && data.transfer_id !== currentTransferId.value) return
    currentStep.value = 'failed'
    errorMsg.value = '传输失败'
    EventsOff('file:transfer_progress')
    EventsOff('file:transfer_completed')
    EventsOff('file:transfer_failed')
  })

  EventsOn('file:transfer_rejected', (data) => {
    if (currentTransferId.value && data.transfer_id !== currentTransferId.value) return
    currentStep.value = 'failed'
    errorMsg.value = '对方拒绝了文件传输'
    EventsOff('file:transfer_progress')
    EventsOff('file:transfer_completed')
    EventsOff('file:transfer_failed')
    EventsOff('file:transfer_rejected')
    EventsOff('file:transfer_cancelled')
  })

  EventsOn('file:transfer_cancelled', (data) => {
    if (currentTransferId.value && data.transfer_id !== currentTransferId.value) return
    currentStep.value = 'cancelled'
    EventsOff('file:transfer_progress')
    EventsOff('file:transfer_completed')
    EventsOff('file:transfer_failed')
    EventsOff('file:transfer_rejected')
    EventsOff('file:transfer_cancelled')
  })
}

onUnmounted(() => {
  EventsOff('file:transfer_started')
  EventsOff('file:transfer_progress')
  EventsOff('file:transfer_completed')
  EventsOff('file:transfer_failed')
  EventsOff('file:transfer_rejected')
  EventsOff('file:transfer_cancelled')
})
</script>
