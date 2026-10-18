<template>
  <div class="flex flex-col h-full bg-bg-chat">
    <div class="flex-1 overflow-y-auto px-6 py-5 bg-[#f0f0f0] scrollbar-thin" ref="messagesContainer">
      <div v-if="messages.length === 0" class="text-center py-24 px-5">
        <div class="mb-3">
          <img src="@/assets/images/nomsg.svg" class="w-[100px] h-[100px] mx-auto opacity-60">
        </div>
        <div class="text-gray-400 text-sm">暂无消息</div>
      </div>
      <div v-for="msg in messages" :key="msg.id" class="flex mb-4 items-start" :class="{'flex-row-reverse': isSelf(msg)}">
        <div class="max-w-[65%] flex flex-col">
          <div class="text-xs text-gray-400 mb-1 pl-1" v-if="!isSelf(msg)">{{ msg.sender_name }}</div>
          <div class="flex items-start" :class="{'flex-row-reverse': isSelf(msg)}">
            <div class="bubble-arrow" v-if="!isSelf(msg)"></div>
            <div class="bubble-arrow-self" v-if="isSelf(msg)"></div>
            <div
              v-if="msg.type !== 'file'"
              class="relative px-3.5 py-2.5 rounded-md text-sm leading-relaxed text-text-primary break-words shadow-bubble"
              :class="isSelf(msg) ? 'bg-bg-bubble-self text-black' : 'bg-white'"
            >
              {{ msg.content }}
            </div>
            <div
              v-else
              class="relative px-3.5 py-2.5 rounded-md text-sm shadow-bubble min-w-[200px]"
              :class="isSelf(msg) ? 'bg-bg-bubble-self' : 'bg-white'"
            >
              <div class="flex items-center mb-2">
                <img :src="getFileIcon(msg.file_name)" class="w-8 h-8 mr-2.5 shrink-0" alt="">
                <div class="flex-1 min-w-0">
                  <div class="text-[13px] font-medium text-text-primary break-all leading-tight">{{ msg.file_name }}</div>
                  <div class="text-[11px] text-text-tertiary mt-0.5">{{ formatFileSize(msg.file_size) }}</div>
                </div>
              </div>
              <div class="text-xs text-[#576b95] cursor-pointer pt-1 border-t border-black/5 hover:text-[#3457d4]" @click="openFileLocation(msg.file_path)">
                查看文件
              </div>
            </div>
          </div>
          <div class="text-[11px] text-gray-400 mt-1 pl-1" :class="{'text-right pr-1 pl-0': isSelf(msg)}">
            {{ formatTime(msg.timestamp) }}
          </div>
        </div>
      </div>
    </div>
    <div class="bg-[#f7f7f7] border-t border-border-light px-6 py-3 shrink-0">
      <div class="flex flex-col gap-2">
        <textarea
          ref="textareaRef"
          v-model="messageInput"
          @keydown.enter.exact.prevent="sendMessage"
          @input="autoResize"
          placeholder="输入消息，Enter 发送"
          class="w-full px-3.5 py-2.5 bg-white border border-[#e0e0e0] rounded-md text-sm text-gray-700 font-[inherit] outline-none resize-none leading-relaxed transition-colors overflow-y-auto focus:border-primary placeholder:text-gray-300 scrollbar-thin"
          rows="2"
        ></textarea>
        <div class="flex items-center justify-end gap-2">
          <button
            class="w-8 h-8 border-none bg-transparent rounded-md cursor-pointer flex items-center justify-center transition-colors hover:bg-gray-200"
            title="发送文件"
            @click="showFilePanel = true"
          >
            <img src="@/assets/images/attachment.svg" class="w-5 h-5" alt="">
          </button>
          <button
            @click="sendMessage"
            class="px-5 py-1.5 bg-primary border-none rounded-md text-white text-[13px] font-medium cursor-pointer transition-colors hover:bg-primary-hover disabled:bg-gray-300 disabled:cursor-not-allowed"
            :disabled="!messageInput.trim()"
          >
            发送
          </button>
        </div>
      </div>
    </div>
    <FileTransferPanel :show="showFilePanel" @close="showFilePanel = false" />
  </div>
</template>

<script setup>
import {ref, computed, nextTick, onMounted} from 'vue'
import {useChatStore} from '../stores/chat'
import {useDevicesStore} from '../stores/devices'
import {SendMessage, GetMessages, OpenFileLocation} from '../../wailsjs/go/app/App'
import FileTransferPanel from './FileTransferPanel.vue'
import {getFileIcon, formatFileSize} from '../utils/file'

const chatStore = useChatStore()
const devicesStore = useDevicesStore()
const messages = computed(() => chatStore.messages)
const onlineCount = computed(() => devicesStore.onlineCount)
const messageInput = ref('')
const messagesContainer = ref(null)
const textareaRef = ref(null)
const showFilePanel = ref(false)

function isSelf(msg) {
  return msg.sender_id === window.__DEVICE_ID__
}

function formatTime(timestamp) {
  const date = new Date(timestamp)
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${hours}:${minutes}`
}

async function openFileLocation(filePath) {
  if (!filePath) return
  try {
    await OpenFileLocation(filePath)
  } catch (err) {
    console.error('Failed to open file location:', err)
  }
}

async function sendMessage() {
  if (!messageInput.value.trim()) return
  try {
    await SendMessage(messageInput.value)
    messageInput.value = ''
    resetTextareaHeight()
    await loadMessages()
  } catch (err) {
    console.error('Failed to send message:', err)
  }
}

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  const maxHeight = 5 * 21 + 20 + 2
  el.style.height = Math.min(el.scrollHeight, maxHeight) + 'px'
}

function resetTextareaHeight() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
}

async function loadMessages() {
  try {
    const msgs = await GetMessages()
    chatStore.setMessages(msgs)
    scrollToBottom()
  } catch (err) {
    console.error('Failed to load messages:', err)
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

onMounted(() => {
  loadMessages()
})

defineExpose({ scrollToBottom })
</script>

<style scoped>
.scrollbar-thin::-webkit-scrollbar {
  width: 6px;
}
.scrollbar-thin::-webkit-scrollbar-thumb {
  background: #ccc;
  border-radius: 3px;
}
.scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background: #b0b0b0;
}
.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}

.bubble-arrow {
  width: 0;
  height: 0;
  border-top: 6px solid transparent;
  border-bottom: 6px solid transparent;
  border-right: 8px solid #fff;
  margin-top: 12px;
  margin-right: -2px;
  flex-shrink: 0;
  filter: drop-shadow(-1px 0 1px rgba(0, 0, 0, 0.04));
}

.bubble-arrow-self {
  width: 0;
  height: 0;
  border-top: 6px solid transparent;
  border-bottom: 6px solid transparent;
  border-left: 8px solid #95ec69;
  margin-top: 12px;
  margin-left: -2px;
  flex-shrink: 0;
  filter: drop-shadow(1px 0 1px rgba(0, 0, 0, 0.04));
}
</style>
