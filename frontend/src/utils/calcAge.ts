// 根据身份证号计算年龄
export function calcAge(idCardNo?: string): number {
  if (!idCardNo || idCardNo.length < 14) return 0;
  const birth = idCardNo.slice(6, 14);
  const year = Number(birth.slice(0, 4));
  const month = Number(birth.slice(4, 6));
  const day = Number(birth.slice(6, 8));
  if (!year || !month || !day) return 0;
  const now = new Date();
  let age = now.getFullYear() - year;
  const m = now.getMonth() + 1;
  if (m < month || (m === month && now.getDate() < day)) age--;
  return Math.max(0, age);
}
