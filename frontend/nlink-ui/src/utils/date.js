/**
 * 字符串相关处理模块
 */

let dateUtils = {
  /**
   * 格式化日期
   * @param date
   * @param formatStr
   * @return 格式化后的日期
   */
  formate_date (date, formatStr = 'yyyy-MM-dd') {
    if (!(date instanceof Date)) {
      return date
    }
    let str = formatStr
    const Week = ['日', '一', '二', '三', '四', '五', '六']

    // 年
    str = str.replace(/yyyy|YYYY/, date.getFullYear())
    str = str.replace(/yy|YY/, (date.getYear() % 100) > 9 ? (date.getYear() % 100).toString() : '0' + (date.getYear() % 100))

    // 月
    const mon = Number.parseInt(date.getMonth()) + 1
    str = str.replace(/MM/, mon > 9 ? mon.toString() : '0' + mon)
    str = str.replace(/M/g, mon)

    // 季度
    const currentQ = Math.floor(mon / 3) + 1
    str = str.replace(/Q/g, `Q${currentQ}`)

    // 周
    const days = this.get_days(date)
    str = str.replace(/W/g, `W${Math.floor(days/7)}`) // 获取当前是第几周
    str = str.replace(/w/g, Week[date.getDay()])

    // 日
    str = str.replace(/dd|DD/, date.getDate() > 9 ? date.getDate().toString() : '0' + date.getDate())
    str = str.replace(/d|D/g, date.getDate())

    // 时分
    str = str.replace(/hh|HH/, date.getHours() > 9 ? date.getHours().toString() : '0' + date.getHours())
    str = str.replace(/h|H/g, date.getHours())
    str = str.replace(/mm/, date.getMinutes() > 9 ? date.getMinutes().toString() : '0' + date.getMinutes())
    str = str.replace(/m/g, date.getMinutes())

    // 秒
    str = str.replace(/ss|SS/, date.getSeconds() > 9 ? date.getSeconds().toString() : '0' + date.getSeconds())
    str = str.replace(/s|S/g, date.getSeconds())

    return str
  },
  /* 获取今天起始 00:00:00 */
  start_of_day (date) {
    let res = new Date(date)
    res.setHours(0)
    res.setMinutes(0)
    res.setSeconds(0)
    return res
  },
  /* 获取今天结束 23:59:59 */
  end_of_day (date) {
    let res = new Date(date)
    res.setHours(23)
    res.setMinutes(59)
    res.setSeconds(59)
    return res
  },
  /* 获取上一周 */
  last_week_day (date) {
    return this.count_day(date, -7)
  },
  /* 计算日期（x天前或x天后） */
  count_day (date, day) {
    //Date()返回当日的日期和时间。
    //getTime()返回 1970 年 1 月 1 日至今的毫秒数。
    let gettimes = date.getTime() + 1000 * 60 * 60 * 24 * day
    let newDate = new Date(gettimes)
    return newDate
  },
  /**
   * 获取日期是一年中的第几天
   * @param {date | string} date 
   * @return {number}
   */
  get_days (date) {
    const currentYear = new Date().getFullYear().toString()
    // 今天减今年的第一天（xxxx年01月01日）
    const hasTimestamp = new Date(date) - new Date(currentYear)
    // 86400000 = 24 * 60 * 60 * 1000
    let hasDays = Math.ceil(hasTimestamp / 86400000)
    return hasDays
  },
  /**
   *  判断日期相差几天
   * @param date1 {date | string} date 
   * @param date2 {date | string} date 
   * @return {number}
   */
  get_day_difference (date1, date2) {
    // 不能加这段，为空需要判断 < 1 结果为 false
    // if (!date1 || !date2) return 0

    // 将日期字符串转换为 Date 对象
    const d1 = new Date(date1)
    const d2 = new Date(date2)

    // 计算两个日期的时间差（毫秒）
    const timeDiff = Math.abs(d2.getTime() - d1.getTime())

    // 将时间差转换为天数
    const daysDiff = Math.floor(timeDiff / (1000 * 3600 * 24))

    return daysDiff
  },
}

export default dateUtils
