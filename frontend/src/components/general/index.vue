<template>
  <div class="nlink-ui-content nlink-ui-general">
    <div class="nlink-ui-general-title">
      <span>Nlink&nbsp;&nbsp;</span>
      <span style="font-size: 14px">v0.1.0</span>
    </div>
    <div class="nlink-ui-general-operate">
      <div class="nlink-ui-general-operate-header">
        操作
      </div>
      <div class="nlink-ui-general-operate-panel">
        <button @click="restartNlink">重启服务</button>
        <button @click="closeNlink">关闭服务</button>
      </div>
    </div>
    <div class="nlink-ui-general-operate">
      <div class="nlink-ui-general-operate-header">
        开发
      </div>
      <div class="nlink-ui-general-operate-panel">
        <div style="margin-bottom: 10px">
          <input v-model="url">
        </div>
        <button @click="reload">加载自定义链接服务</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ipcEmit } from '../../../ipc/index'
import { useProfilerStore } from '../../store/index.js'
import { ref } from 'vue'

const profiler = useProfilerStore()

let url = ref('')

const reload = () => {
  window.location.href = url.value
}

// 重启服务
const restartNlink = () => {
  ipcEmit('restart', profiler.currentProfile.content)
}

const closeNlink = () => {
  ipcEmit('close', {})
}

</script>

<style scoped>
.nlink-ui-general {
  padding: 20px;
}
.nlink-ui-general-title {
  font-size: 24px;
  text-align: center;
}
.nlink-ui-general-operate {
  margin-top: 20px;
  padding: 20px;
  border: solid 1px var(--block-bg-color);
  border-radius: 5px;
}
.nlink-ui-general-operate-header {
  font-size: 18px;
  padding-bottom: 10px;
  border-bottom: solid 1px var(--block-bg-color);
}
.nlink-ui-general-operate-panel {
  padding: 20px 0;
}
</style>