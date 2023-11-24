import { useLogStore } from '@store'

import { ipcEmit } from '@ipc'

let logging = false

export const startLogs = async () => {
  // 避免重复获取
  if (logging) return true

  const logger = useLogStore()

  logging = true
  const res = await ipcEmit('logs')
  console.log('res', res)
  if (res && res.length > 0) logger.pushLogs(res)
  
  // 循环调用
  // 基于promise与go那边通知原理挂起，go那边通知后会await继续往下走
  logging = false
  startLogs()
}

export const getLogs = (count = 100) => {
  const logger = useLogStore()

  return logger.getLogs(count)
}