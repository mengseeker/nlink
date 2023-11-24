import { Restart, Logs, OpenFileDialog } from "../wailsjs/go/client/WailsApp"

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