const { ipcMain } = require('electron')

// 重启服务（启动）
ipcMain.on('restart_nlink', () => {
  console.log('restart_nlink')
})

// 关闭服务
ipcMain.on('close_nlink', () => {
  console.log('close_nlink')
})

// 重新配置订阅（传入文件路径？）
ipcMain.on('reset_subscription', () => {
  console.log('reset_subscription')
})

// 获取日志列表（包括搜索、筛选）
ipcMain.on('get_logs', () => {
  console.log('get_logs')
})

// 获取当前端口
ipcMain.on('get_port', () => {
  console.log('get_port')
})

// 设置当前端口
ipcMain.on('set_port', () => {
  console.log('set_port')
})
