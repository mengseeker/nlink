import { defineStore } from 'pinia'

export const useProfilerStore = defineStore('profiler', {
  persist: true,
  state: () => {
    return {
      currentProfile: {
        content: ''
      },
      profiles: [] // { id: 1, content: '', name: '', type: '', url: '' }
    }
  },
  // 也可以这样定义
  // state: () => ({ count: 0 })
  actions: {
    pushProfile (profile) {
      this.profiles.push(profile)
    },
    setCurrentProfile (profile) {
      this.currentProfile = profile
    },
    updateCurrentProfile (content) {
      this.currentProfile.content = content
    }
  },
})

export const useLogStore = defineStore('logger', {
  persist: true,
  state: () => {
    return {
      logs: []
    }
  },
  getters: {
    showLogs: (state) => state.logs.slice(-100)
  },
  actions: {
    pushLog (log) {
      // 最多存一万条
      if (this.logs.length >= 10000) {
        this.logs.shift()
      }
      this.logs.push(log)
    },
    pushLogs (logs) {
      // 最多存一万条
      while (this.logs.length >= 10000) {
        this.logs.shift()
      }
      // 避免重新遍历，一个一个加
      for (const log of logs) {
        this.logs.push(log)
      }
    },
    getLogs (count = 100) {
      // 默认返回一百条
      this.logs.slice(0, count)
    }
  },
})
