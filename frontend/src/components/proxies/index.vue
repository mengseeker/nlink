<template>
  <div class="nlink-ui-proxies">
    <div class="nlink-ui-proxies-group nlink-ui-common-group">
      <div class="nlink-ui-general-title nlink-ui-common-group-header">服务器列表</div>
      <div class="nlink-ui-proxies-services nlink-ui-common-group-panel">
        <div
          class="nlink-ui-proxies-service-item"
          v-for="item in servers" :key="item.Name">
          <div >
            {{ item.Name }}
          </div>
          <div>
            {{ item.Addr }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useProfilerStore } from '../../store/index'

const profiler = useProfilerStore()

// 获取服务器配置
let servers = ref([])
if (profiler.currentProfile) {
  const profileContent = JSON.parse(profiler.currentProfile.content)
  if (profileContent.client && profileContent.client.Servers) {
    servers.value = profileContent.client.Servers
  }
}

</script>

<style scoped>
.nlink-ui-proxies-group {

}

.nlink-ui-proxies-services {
  display: flex;
}
.nlink-ui-proxies-service-item {
  width: 40%;
  border-left: solid 2px transparent;
  border-radius: 4px;
  background-color: var(--block-bg-color);
  padding: 10px;
  margin-left: 10px;
  position: relative;
  cursor: pointer;
}
</style>