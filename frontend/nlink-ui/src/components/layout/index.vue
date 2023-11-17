<template>
  <div class="nlink-ui-layout">
    <div class="nlink-ui-left-panel">
      <div class="left-header-panel">
        <!-- 显示网速 -->
        <div>上传速率：{{ uploadSpeed }} k/s</div>
        <div>下载速率：{{ downloadSpeed }} k/s</div>
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
import { defineProps, defineEmits } from 'vue'

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
  { value: 'profiles', label: '订阅管理' },
  { value: 'log', label: '日志管理' },
  { value: 'connections', label: '连接管理' },
  { value: 'settings', label: '设置' },
]

const clickMenu = (item) => {
  emit('update:modelValue', item.value)
}

</script>

<style scoped>
.nlink-ui-layout {
  width: 100%;
  height: 100%;
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
  background-color: var(--content-bg-color);
}

.nlink-ui-left-panel .left-header-panel {
  padding: 10px;
  border-bottom: solid 2px #fff;
}

.nlink-ui-menu {}
.nlink-ui-menu-item {
  padding: 10px 20px;
  cursor: pointer;
}
.nlink-ui-menu-item.active {
  background-color: #fff;
}
</style>