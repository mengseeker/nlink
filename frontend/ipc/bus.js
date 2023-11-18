

export const emitFunc =  (name, args) => {
  if (window.electronAPI) {
    window.electronAPI.emitFunc(name, args)
  } else {
    console.log('emitFunc', name, args)
  }
}