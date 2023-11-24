<template>
  <div>
    <div>
      <div>
        <!-- 搜索待做 -->
        <n-input v-model:value="name" round placeholder="搜索">
          <template #suffix>
            <n-icon :component="FlashOutline" />
          </template>
        </n-input>
      </div>
    </div>
    <div style="margin-top: 10px;padding: 10px;background-color: black;color: #367c71">
      <n-log
        ref="logInstRef"
        :log="showLogs"
        :loading="loading"
        language="self-log"
        trim
        :line-height="1.5"
        :font-size="16"
        :rows="size"
        :hljs="hljs"
        @require-top="handlerequireTop"
      />
      <!-- <n-list v-else hoverable clickable>
        <n-list-item v-for="item in logger.showLogs" :key="'log-' + item">
          {{ item }}
        </n-list-item>
      </n-list> -->
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, watchEffect, onMounted, nextTick } from 'vue'
// import { ipcEmit } from '@ipc'
import { useLogStore } from '@store'

import { FlashOutline } from '@vicons/ionicons5'
import hljs from 'highlight.js/lib/core'

// import { LogInst } from 'naive-ui'

const logger = useLogStore()

const loading = ref(false)
const name = ref('')
const index = ref(1)
const size = ref(10)
const logInstRef = ref(null)

const showLogs = computed(() => {
  return logger.logs.slice(index.value * (-size)).join('\n')
})

const handlerequireTop = () => {
  if (loading.value) return

  loading.value = true
  console.log('scroll1', index)
  setTimeout(() => {
  console.log('scroll2', index)
    index.value = index.value + 1
    loading.value = false
  }, 1000)
}
hljs.registerLanguage('self-log', () => ({
  Keywords: [name.value]
}))

watch(name, () => {
})

onMounted(() => {
  watchEffect(() => {
    if (showLogs.value) {
      nextTick(() => {
        logInstRef.value.scrollTo({ position: 'bottom', silent: true })
      })
    }
  })
})

</script>