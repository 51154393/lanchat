<template>
  <div class="fixed inset-0 z-[2000] flex items-center justify-center" v-if="show">
    <div class="fixed inset-0 bg-black/40"></div>
    <div class="relative bg-white rounded-xl w-[380px] shadow-panel overflow-hidden">
      <div class="px-5 pt-4 pb-3">
        <h3 class="text-base font-semibold text-text-primary m-0">收到文件</h3>
      </div>
      <div class="px-5 py-4">
        <div class="flex items-center mb-4">
          <img
            :src="senderAvatar"
            class="w-10 h-10 rounded-full mr-3 shrink-0"
            alt=""
          >
          <div class="flex-1 min-w-0">
            <div class="text-sm font-medium text-text-primary">{{ senderName || '未知设备' }}</div>
            <div class="text-xs text-text-tertiary mt-0.5">{{ senderIP }}</div>
          </div>
        </div>
        <div class="flex items-center p-3 bg-[#f8f8f8] rounded-lg">
          <img :src="fileIcon" class="w-8 h-8 mr-3 shrink-0" alt="">
          <div class="flex-1 min-w-0">
            <div class="text-sm font-medium text-text-primary break-all leading-snug">{{ transfer?.file_name }}</div>
            <div class="text-xs text-text-tertiary mt-1">{{ formatFileSize(transfer?.file_size) }}</div>
          </div>
        </div>
      </div>
      <div class="flex gap-3 px-5 py-3 pb-4">
        <button
          class="flex-1 py-2.5 border-none rounded-lg text-sm font-medium cursor-pointer transition-all bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-60 disabled:cursor-not-allowed"
          @click="reject"
          :disabled="loading"
        >
          拒绝
        </button>
        <button
          class="flex-1 py-2.5 border-none rounded-lg text-sm font-medium cursor-pointer transition-all bg-primary text-white hover:bg-primary-hover disabled:opacity-60 disabled:cursor-not-allowed"
          @click="accept"
          :disabled="loading"
        >
          {{ loading ? '处理中...' : '接收' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import {ref, computed} from 'vue'
import {AcceptFile, RejectFile, SelectSavePath} from '../../wailsjs/go/app/App'
import {getFileIcon, formatFileSize} from '../utils/file'
import winAvatar from '@/assets/images/avatar/win.svg'
import macAvatar from '@/assets/images/avatar/mac.svg'
import linuxAvatar from '@/assets/images/avatar/linux.svg'

const props = defineProps({
  show: Boolean,
  transfer: Object,
  senderName: String,
  senderIP: String,
  senderPlatform: String
})

const emit = defineEmits(['close', 'accepted', 'rejected'])

const loading = ref(false)

const senderAvatar = computed(() => {
  switch (props.senderPlatform) {
    case 'windows': return winAvatar
    case 'darwin': return macAvatar
    case 'linux': return linuxAvatar
    default: return winAvatar
  }
})

const fileIcon = computed(() => getFileIcon(props.transfer?.file_name))

async function accept() {
  if (!props.transfer?.transfer_id) return
  loading.value = true
  try {
    const savePath = await SelectSavePath(props.transfer.file_name)
    if (!savePath) {
      loading.value = false
      return
    }
    await AcceptFile(props.transfer.transfer_id, savePath)
    emit('accepted', savePath)
    emit('close')
  } catch (err) {
    console.error('Failed to accept file:', err)
  } finally {
    loading.value = false
  }
}

function reject() {
  if (!props.transfer?.transfer_id) return
  loading.value = true
  try {
    RejectFile(props.transfer.transfer_id)
    emit('rejected')
    emit('close')
  } catch (err) {
    console.error('Failed to reject file:', err)
  } finally {
    loading.value = false
  }
}
</script>
