<template>
  <div class="fixed inset-0 z-[1500] flex items-center justify-center" v-if="show">
    <div class="fixed inset-0 bg-black/40"></div>
    <div class="relative bg-white rounded-xl w-[420px] shadow-panel overflow-hidden">
      <div class="flex items-center justify-between px-5 py-4 border-b border-border-lighter">
        <h3 class="text-base font-semibold text-text-primary m-0">{{ panelTitle }}</h3>
        <button v-if="status === 'receiving'" class="w-7 h-7 border-none bg-transparent rounded-md text-base text-danger cursor-pointer flex items-center justify-center hover:bg-red-50" @click="cancelTransfer">✕</button>
        <button v-else class="w-7 h-7 border-none bg-transparent rounded-md text-xl text-gray-400 cursor-pointer flex items-center justify-center hover:bg-gray-100" @click="close">×</button>
      </div>
      <div class="p-5">
        <div class="flex items-center p-3 bg-[#f8f8f8] rounded-lg mb-4">
          <img :src="fileIcon" class="w-8 h-8 mr-3 shrink-0" alt="">
          <div class="flex-1 min-w-0">
            <div class="text-sm font-medium text-text-primary break-all leading-snug">{{ fileName }}</div>
            <div class="text-xs text-text-tertiary mt-1">{{ formatFileSize(transferred) }} / {{ formatFileSize(total) }}</div>
          </div>
        </div>

        <div class="mb-3">
          <div class="h-2 bg-gray-200 rounded-full overflow-hidden mb-2">
            <div class="h-full rounded-full transition-[width] duration-300" :class="progressColor" :style="{width: progress + '%'}"></div>
          </div>
          <div class="text-center text-sm text-text-secondary">{{ Math.round(progress) }}%</div>
        </div>

        <div v-if="status === 'completed'" class="text-center py-2">
          <div class="text-[15px] font-medium text-primary mb-1.5">文件接收完成</div>
          <div class="text-xs text-text-tertiary break-all">{{ savePath }}</div>
        </div>

        <div v-if="status === 'failed'" class="text-center py-2">
          <div class="text-[15px] font-medium text-danger">文件接收失败</div>
        </div>

        <div v-if="status === 'cancelled'" class="text-center py-2">
          <div class="text-[15px] font-medium text-warning">文件接收已取消</div>
        </div>
      </div>
      <div class="flex justify-end gap-3 px-5 py-3 pb-4">
        <button v-if="status === 'completed'" class="px-5 py-2 border-none rounded-md text-sm cursor-pointer bg-primary text-white transition-colors hover:bg-primary-hover" @click="openLocation">
          打开文件所在位置
        </button>
        <button v-if="status === 'receiving'" class="px-5 py-2 border-none rounded-md text-sm cursor-pointer bg-danger text-white transition-colors hover:bg-danger-hover" @click="cancelTransfer">
          取消接收
        </button>
        <button v-if="status === 'completed' || status === 'failed' || status === 'cancelled'" class="px-5 py-2 border-none rounded-md text-sm cursor-pointer bg-gray-100 text-gray-600 transition-colors hover:bg-gray-200" @click="close">
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import {ref, computed, onUnmounted} from 'vue'
import {EventsOn, EventsOff} from '../../wailsjs/runtime/runtime'
import {OpenFileLocation, CancelFile} from '../../wailsjs/go/app/App'
import {getFileIcon, formatFileSize} from '../utils/file'

const props = defineProps({
  show: Boolean,
  transferId: String,
  fileName: String,
  fileSize: Number,
  savePath: String
})

const emit = defineEmits(['close'])

const progress = ref(0)
const transferred = ref(0)
const total = ref(0)
const status = ref('receiving')

const panelTitle = computed(() => {
  switch (status.value) {
    case 'receiving': return '正在接收文件'
    case 'completed': return '接收完成'
    case 'failed': return '接收失败'
    case 'cancelled': return '已取消'
    default: return '接收文件'
  }
})

const progressColor = computed(() => {
  if (status.value === 'completed') return 'bg-primary'
  if (status.value === 'failed') return 'bg-danger'
  if (status.value === 'cancelled') return 'bg-warning'
  return 'bg-primary'
})

const fileIcon = computed(() => getFileIcon(props.fileName))

function setupListeners() {
  EventsOn('file:transfer_progress', (data) => {
    if (data.transfer_id !== props.transferId) return
    progress.value = data.progress || 0
    transferred.value = data.transferred || 0
    total.value = data.total || props.fileSize || 0
  })

  EventsOn('file:transfer_completed', (data) => {
    if (data.transfer_id !== props.transferId) return
    progress.value = 100
    transferred.value = total.value
    status.value = 'completed'
    cleanup()
  })

  EventsOn('file:transfer_failed', (data) => {
    if (data.transfer_id !== props.transferId) return
    status.value = 'failed'
    cleanup()
  })

  EventsOn('file:transfer_cancelled', (data) => {
    if (data.transfer_id !== props.transferId) return
    status.value = 'cancelled'
    cleanup()
  })
}

function cleanup() {
  EventsOff('file:transfer_progress')
  EventsOff('file:transfer_completed')
  EventsOff('file:transfer_failed')
  EventsOff('file:transfer_cancelled')
}

async function openLocation() {
  if (!props.savePath) return
  try {
    await OpenFileLocation(props.savePath)
  } catch (err) {
    console.error('Failed to open file location:', err)
  }
}

function close() {
  emit('close')
}

async function cancelTransfer() {
  if (props.transferId) {
    try {
      await CancelFile(props.transferId)
    } catch (err) {
      console.error('Failed to cancel transfer:', err)
    }
  }
  status.value = 'cancelled'
}

function reset() {
  progress.value = 0
  transferred.value = 0
  total.value = 0
  status.value = 'receiving'
}

defineExpose({ setupListeners, reset })

onUnmounted(() => {
  cleanup()
})
</script>
