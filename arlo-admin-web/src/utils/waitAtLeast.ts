/** 保证从 startedAt 起至少经过 minMs，避免过快结束导致 loading 闪一下看不见 */
export async function waitAtLeast(startedAt: number, minMs = 360) {
  const left = minMs - (Date.now() - startedAt)
  if (left > 0) {
    await new Promise((r) => setTimeout(r, left))
  }
}
