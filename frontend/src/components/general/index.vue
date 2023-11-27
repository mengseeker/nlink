<template>
  <div class="nlink-ui-content nlink-ui-general">
    <div class="nlink-ui-general-title">
      <span>Nlink&nbsp;&nbsp;</span>
      <span style="font-size: 14px">v0.1.0</span>
    </div>
    <!-- <div class="nlink-ui-general-operate nlink-ui-common-group">
      <div class="nlink-ui-general-operate-header nlink-ui-common-group-header">
        操作
      </div>
      <div class="nlink-ui-general-operate-panel nlink-ui-common-group-panel">
        <button @click="restartNlink">重启服务</button>
        <button @click="closeNlink">关闭服务</button>
      </div>
    </div> -->
    <div class="nlink-ui-general-operate">
      <div class="nlink-ui-general-operate-header">
        配置预览(修改内容后需要点击应用修改再进行重启)
      </div>
      <div class="nlink-ui-general-operate-panel">
        <n-button size="small" type="info" @click="updateSettings">应用修改</n-button>
        <div class="nlink-ui-general-settings-panel">
          使用端口:
          <n-input
            v-model:value="client.Listen" size="small" placeholder="示例: :7899"
            style="width: 200px;display: inline-block;" />
        </div>
        <div class="nlink-ui-general-settings-panel">
          系统代理:
          <n-switch v-model:value="client.System" />
        </div>
        <div class="nlink-ui-general-settings-panel">
          代理协议:
          <n-select
            v-model:value="client.Net" size="small" 
            :options="netTypes"
            style="display: inline-block;width: 200px" />
        </div>
        <div class="nlink-ui-general-settings-panel">
          cert文件: <n-button @click="selectFile('cert')">选择(pem后缀)</n-button>
          &nbsp;&nbsp;
          <span>当前路径：{{ client.Cert }}</span>
        </div>
        <div class="nlink-ui-general-settings-panel">
          key 文件: <n-button @click="selectFile('key')">选择(pem后缀)</n-button>
          &nbsp;&nbsp;
          <span>当前路径：{{ client.Key }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ipcEmit } from '@ipc'
import { startLogs } from '@utils/logs'
import { useProfilerStore } from '@store'

import { ref } from 'vue'
import { NInput } from 'naive-ui'

const profiler = useProfilerStore()

let url = ref('')

const reload = () => {
  window.location.href = url.value
}

// 重启服务
const restartNlink = async () => {
  const res = await ipcEmit('restart', profiler.currentProfile.content)

  // 重启后开始记录日志
  startLogs()
  window.$message.success('重启成功')
}

const closeNlink = () => {
  ipcEmit('close', {})
}

// 获取端口、代理类型配置
let client = ref({})
if (profiler.currentProfile) {
  const profileContent = JSON.parse(profiler.currentProfile.content)
  if (profileContent && profileContent.Listen) {
    client.value = profileContent
  }
}
const netTypes = [
  { value: 'udp', label: 'udp' },
  { value: 'tcp', label: 'tcp' }]

const updateSettings = () => {
  if (!profiler.currentProfile) {
    window.$message.warning('当前配置为空')
    return false
  }

  let profileContent = JSON.parse(profiler.currentProfile.content)
  profileContent = {
    ...profileContent,
    ...client.value
  }
  // TODO: 后续封装放utils里
  profiler.updateCurrentProfile(JSON.stringify(profileContent, null, 2))
}

const selectFile = async (type = 'cert') => {
  const res = await ipcEmit('select_file')
  if (!res || res === '') return false

  switch (type) {
    case 'cert':
      client.value.Cert = res
      break
    case 'key':
      client.value.Key = res
      break
  }
  window.$message.success('已修改文件，请点击应用修改')
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
.nlink-ui-general-settings-panel {
  margin-top: 10px;
}
</style>