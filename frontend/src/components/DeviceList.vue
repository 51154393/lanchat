<template>
  <div class="flex flex-col h-full bg-[#fcfcfc]">
    <div class="px-4 pt-5 pb-4 flex items-center justify-between">
      <div class="text-md font-semibold">局域网</div>
      <div class="text-xs text-gray-500">在线设备 ({{ onlineCount }})</div>
    </div>
    <div class="flex-1 overflow-y-auto px-2">
      <div v-if="onlineDevices.length === 0" class="text-center py-10 px-4">
        <div class="mb-3">
          <img src="@/assets/images/chat.svg" class="w-[100px] h-[100px] mx-auto opacity-60">
        </div>
        <div class="text-gray-500 text-sm">暂无在线设备</div>
      </div>
      <div
        v-for="device in onlineDevices"
        :key="device.device_id"
        class="flex items-center px-3 py-3 mb-1 rounded-lg transition-colors cursor-pointer hover:bg-gray-100"
      >
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
        <div class="flex-1 ml-3 min-w-0">
          <div class="text-sm font-medium mb-1 truncate">{{ device.username }}@{{ device.device_name }}</div>
          <div class="flex items-center text-gray-500 text-xs">
            <span class="w-1.5 h-1.5 rounded-full bg-[#4caf50] mr-1.5 shrink-0"></span>
            {{ device.ip }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {computed} from 'vue'
import {useDevicesStore} from '../stores/devices'

const devicesStore = useDevicesStore()
const onlineDevices = computed(() => devicesStore.onlineDevices)
const onlineCount = computed(() => devicesStore.onlineCount)
</script>
