import  { request } from './http'
import { v4 as uuidv4 } from 'uuid'
import dateUtils from './date'

const defaultContent = {
  "Listen": ":7890",
  "System": false,
  "Net": "tcp",
  "Cert": ".dev/tls/client/xingbiao_cert.pem",
  "Key": ".dev/tls/client/xingbiao_key.pem",
  "Resolver": [
      {
          "DNS": "114.114.114.114"
      },
      {
          "DoH": "https://223.6.6.6/dns-query"
      },
      {
          "DoT": "223.6.6.6"
      },
      {
          "DoT": "dns.pub"
      },
      {
          "DoH": "https://doh.pub/dns-query"
      },
      {
          "DoT": "185.222.222.222"
      }
  ],
  "Servers": [
      {
          "Name": "tokyo",
          "Addr": "localhost:8899"
      }
  ],
  "Rules": [
      "host-match: ad\\.com, reject",
      "host-match: \\.cn, direct",
      "ip-cidr: 127.0.0.1/8, direct",
      "ip-cidr: 172.16.0.0/12, direct",
      "ip-cidr: 192.168.1.201/16, direct",
      "geoip: CN, direct",
      "match, forward: tokyo"
  ]
}
export const getDefaultProfile = () => {
  return {
    id: uuidv4(),
    content: JSON.stringify(defaultContent, null, 2),
    name: '本地配置',
    type: 'local', // ['link', 'local']
    lastUpdatedAt: dateUtils.formate_date(new Date()),
    createdAt: dateUtils.formate_date(new Date())
  }
}

export const requestRemoteProfile = async (url) => {

  const res = await request(url)
  if (!res) return false

  // 生成一个 UUID
  const uuid = uuidv4();
  const urlArr = url.split('/')

  return {
    id: uuid,
    content: res.data,
    name: urlArr[urlArr.length - 1],
    type: 'link',
    lastUpdatedAt: dateUtils.formate_date(new Date()),
    createdAt: dateUtils.formate_date(new Date())
  }
}

export const setCurrentProfile = (id) => {

}