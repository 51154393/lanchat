import {defineStore} from 'pinia'
import {ref, computed} from 'vue'

export const useDevicesStore = defineStore('devices', () => {
  const devices = ref([])

  const onlineDevices = computed(() => {
    return devices.value.filter(d => d.status === 1)
  })

  const onlineCount = computed(() => onlineDevices.value.length)

  function setDevices(newDevices) {
    devices.value = newDevices
  }

  function addDevice(device) {
    const existing = devices.value.find(d => d.device_id === device.device_id)
    if (existing) {
      Object.assign(existing, device)
    } else {
      devices.value.push(device)
    }
  }

  function removeDevice(deviceId) {
    devices.value = devices.value.filter(d => d.device_id !== deviceId)
  }

  function updateDeviceStatus(deviceId, status) {
    const device = devices.value.find(d => d.device_id === deviceId)
    if (device) {
      device.status = status
    }
  }

  return {
    devices,
    onlineDevices,
    onlineCount,
    setDevices,
    addDevice,
    removeDevice,
    updateDeviceStatus
  }
})
