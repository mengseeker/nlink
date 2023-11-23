import { Restart, Logs } from "../wailsjs/go/client/WailsApp"

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
  }
  if (!res || !res.Success) {
    alert(res && res.Mes ? res.Mes : '出错了')
    throw res
  }

  console.log(name, res, 'ipcEmitReturn')
  return res.Result
}