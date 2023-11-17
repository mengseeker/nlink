const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('electronAPI', {
  emitFunc: (func, args) => ipcRenderer.send(func, args)
})