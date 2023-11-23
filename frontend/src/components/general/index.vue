<template>
  <div class="nlink-ui-content nlink-ui-general">
    <div class="nlink-ui-general-title">
      <span>Nlink&nbsp;&nbsp;</span>
      <span style="font-size: 14px">v0.1.0</span>
    </div>
    <div class="nlink-ui-general-operate nlink-ui-common-group">
      <div class="nlink-ui-general-operate-header nlink-ui-common-group-header">
        操作
      </div>
      <div class="nlink-ui-general-operate-panel nlink-ui-common-group-panel">
        <button @click="restartNlink">重启服务</button>
        <button @click="closeNlink">关闭服务</button>
      </div>
    </div>
    <div class="nlink-ui-general-operate">
      <div class="nlink-ui-general-operate-header">
        配置预览(修改内容后需要点击应用修改再进行重启)
      </div>
      <div class="nlink-ui-general-operate-panel">
        <button @click="updateSettings">应用修改</button>
        <div class="nlink-ui-general-settings-panel">
          使用端口:
          <input v-model="client.Listen">
        </div>
        <div class="nlink-ui-general-settings-panel">
          系统代理:
          <select v-model="client.System">
            <option v-for="item in systemProxies"
              :key="item"
              :value="item">
              {{ item }}
            </option>
          </select>
        </div>
        <div class="nlink-ui-general-settings-panel">
          代理协议:
          <select v-model="client.Net">
            <option v-for="item in netTypes"
              :key="item"
              :value="item">
              {{ item }}
            </option>
          </select>
        </div>
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
const restartNlink = async () => {
  const res = await ipcEmit('restart', profiler.currentProfile.content)

  alert('重启成功')
}

const closeNlink = () => {
  ipcEmit('close', {})
}

// 获取端口、代理类型配置
let client = ref({})
if (profiler.currentProfile) {
  const profileContent = JSON.parse(profiler.currentProfile.content)
  if (profileContent.client && profileContent.client.Listen) {
    client.value = profileContent.client
  }
}
const netTypes = ['udp', 'http', 'sock5']
const systemProxies = [true, false]

const updateSettings = () => {
  if (!profiler.currentProfile) {
    alert('当前配置为空')
    return false
  }

  const profileContent = JSON.parse(profiler.currentProfile.content)
  profileContent.client = client.value
  // TODO: 后续封装放utils里
  profiler.updateCurrentProfile(JSON.stringify(profileContent, null, 2))
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