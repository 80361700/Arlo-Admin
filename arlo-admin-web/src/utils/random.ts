/** 生成随机 len 位数字（流程节点 key） */
export function randomLenNum(len = 4, date = false): string {
  let random = Math.ceil(Math.random() * 100000000000000)
    .toString()
    .substr(0, len || 4)
  if (date) random = random + Date.now()
  return random
}
