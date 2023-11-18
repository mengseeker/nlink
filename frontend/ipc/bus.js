import { Greet } from '../wailsjs/go/main/App'

export const emitFunc = async (name, args) => {
  const res = await Greet(name, args)

  return res
}