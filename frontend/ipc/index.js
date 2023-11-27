import { Restart, Logs, OpenFileDialog, Stop } from "../wailsjs/go/wails/WailsApp"

export const ipcEmit = async (name, args) => {
  console.log('ipcEmit', name, args)
  let res = {}
  switch (name) {
    case 'restart':
      res = await Restart(args)
      break
    case 'logs':
      res = await Logs(args)
      break
    case 'select_file':
      res = await OpenFileDialog(args)
      break
    case 'stop':
      res = await Stop(args)
      break
  }
  console.log(name, res, 'ipcEmitReturn')
  if (!res || !res.Success) {
    window.$message.error(
      res && res.Msg ? res.Msg : '出错了'
    )
    throw {
      name: 'IpcError',
      message: res.Msg
    }
  }

  return res.Result
}