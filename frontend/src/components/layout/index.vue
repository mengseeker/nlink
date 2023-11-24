<template>
  <div class="nlink-ui-layout">
    <div class="nlink-ui-left-panel">
      <div class="left-header-panel">
        <!-- <div>上传速率：{{ uploadSpeed }} k/s</div>
        <div>下载速率：{{ downloadSpeed }} k/s</div> -->
        <div style="font-size: 20px;margin-bottom: 10px;">
          Nlink
        </div>
        <div style="margin-bottom: 10px;">
          <n-button size="tiny" type="info"
            style="margin-right: 5px;"
            @click="restartNlink" >
            重启服务
          </n-button>
          <n-button size="tiny" type="info" @click="closeNlink">
            关闭服务
          </n-button>
        </div>
      </div>
      <div class="nlink-ui-menu">
        <div
          class="nlink-ui-menu-item"
          :class="{ 'active': modelValue === item.value }"
          v-for="item in menu" :key="item.value"
          @click="clickMenu(item)">
          {{ item.label }}
        </div>
      </div>
    </div>
    <div class="nlink-ui-right-panel">
      <slot></slot>
    </div>
  </div>
</template>

<script setup>
import { ipcEmit } from '@ipc'
import { startLogs } from '@utils/logs'
import { useProfilerStore } from '@store'

import { defineProps, defineEmits, onMounted } from 'vue'

import { NConfigProvider, useMessage } from 'naive-ui'

const profiler = useProfilerStore()

defineProps({
  modelValue: [String, Number],
  uploadSpeed: {
    type: [String, Number],
    default: 0
  },
  downloadSpeed: {
    type: [String, Number],
    default: 0
  },
})
const emit = defineEmits(['update:modelValue'])

const menu = [
  { value: 'general', label: '全局配置' },
  { value: 'proxies', label: '代理服务' },
  { value: 'profiles', label: '配置管理' },
  { value: 'log', label: '日志管理' },
  // { value: 'connections', label: '连接管理' },
  { value: 'settings', label: '客户端设置' },
]

const clickMenu = (item) => {
  emit('update:modelValue', item.value)
}

// 重启服务
const restartNlink = async () => {
  await ipcEmit('restart', profiler.currentProfile.content)

  // 重启后开始记录日志
  startLogs()
  window.$message.success('重启成功')
}
// 关闭服务
const closeNlink = () => {
  ipcEmit('close', {})
}

onMounted(() => {
  window.$message = useMessage()
})

</script>

<style scoped>
.nlink-ui-layout {
  width: 100%;
  height: 100%;
  padding: 10px;
  display: flex;
  background-color: var(--bg-color);
  color: var(--font-color);
}
.nlink-ui-left-panel {
  width: 150px;
  height: 100%;
}
.nlink-ui-right-panel {
  flex: 1;
  height: 100%;
  overflow: auto;
  padding: 10px;
  background-color: var(--content-bg-color);
}

.nlink-ui-left-panel .left-header-panel {
  padding: 10px;
  border-bottom: solid 2px #fff;
  text-align: center;
}

.nlink-ui-menu {}
.nlink-ui-menu-item {
  padding: 10px 20px;
  cursor: pointer;
  text-align: center;
  border-radius: 5px;
  margin-bottom: 4px;
}
.nlink-ui-menu-item.active {
  background-color: var(--select-bg-color);
}
</style>