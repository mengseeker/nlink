import { Restart, Logs } from "../wailsjs/go/client/WailsApp"

export const ipcEmit = async (name, args) => {
  console.log('ipcEmit', name, args)
  let res = {}
  switch (name) {
    case 'restart':
      res = await Restart(args)
    case 'logs':
      res = await Logs(args)
  }
  return res
}