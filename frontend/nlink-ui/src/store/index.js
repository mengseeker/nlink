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
