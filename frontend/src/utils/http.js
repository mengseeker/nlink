import axios from 'axios'

// 发起 GET 请求获取远程文件的内容
export const request = async (url) => {
  const res = await axios.get(url)
  return res
}