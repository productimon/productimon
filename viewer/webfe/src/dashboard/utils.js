// Number of miliseconds per unit
export const timeUnits = {
  Seconds: 10 ** 3,
  Minutes: 10 ** 3 * 60,
  Hours: 10 ** 3 * 60 * 60,
  Days: 10 ** 3 * 60 * 60 * 24,
  Weeks: 10 ** 3 * 60 * 60 * 24 * 7,
  Months: 10 ** 3 * 60 * 60 * 24 * 30,
  Years: 10 ** 3 * 60 * 60 * 24 * 365,
};

export function calculateDate(unit, val) {
  const mult = timeUnits[unit];
  return new Date().getTime() - val * mult;
}
