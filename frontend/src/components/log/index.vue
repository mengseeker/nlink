<template>
  <div>
    <div>
      <div>日志</div>
      <div>
        <input v-model="name" @blur="getLog">
      </div>
    </div>
    <div>
      <div v-for="item in logs" :key="'log' + item.id">
        <div>
          item
        </div>
        <!-- <div>
          <span>{{ item.date }}</span>
          <span>{{ item.type }}</span>
        </div>
        <div>
          {{ item.message }}
        </div> -->
      </div>
    </div>
  </div>
</template>

<script setup>
import { onUnmounted, ref } from 'vue'
import { ipcEmit } from '@ipc'

let logs = ref([])
let name = ref('')

let timer = null
const getLog = async () => {
  // 清除定时器
  if (timer) clearTimeout(timer)

  const res = await ipcEmit('logs', { name: name.value })
  // if (!res) return

  logs.value = res

  // 自动调用
  timer = setTimeout(() => {
    getLog()
  }, 1000)
}

// 销毁定时器
onUnmounted(() => {
  clearTimeout(timer)
})

</script>