// 参考值范围格式化与异常比对（与后端 service/exam_result_service.go IsAbnormal 保持一致）
export function formatReferenceRange(rng?: string): string {
  return rng ? rng.replace('~', ' ~ ') : '-';
}

export function isAbnormal(refRange: string, value: string): boolean {
  const ref = (refRange || '').trim();
  const val = (value || '').trim();
  if (!ref || !val) return false;
  const v = Number(val);
  if (Number.isNaN(v)) return false;
  if (ref.includes('-') && !ref.includes('>') && !ref.includes('<')) {
    const [lo, hi] = ref.split('-').map((s) => Number(s.trim()));
    if (!Number.isNaN(lo) && !Number.isNaN(hi)) return v < lo || v > hi;
  }
  if (ref.startsWith('>')) {
    const lim = Number(ref.slice(1).trim());
    if (!Number.isNaN(lim)) return v <= lim;
  }
  if (ref.startsWith('<')) {
    const lim = Number(ref.slice(1).trim());
    if (!Number.isNaN(lim)) return v >= lim;
  }
  return false;
}
