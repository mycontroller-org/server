export const toIntlLocale = (lng) => {
  if (!lng) {
    return "en-GB"
  }
  return String(lng).replace(/_/g, "-")
}

// Same buckets as moment.fromNow(): under ~45s is "a few seconds", not "12 seconds".
export const formatFromNow = (date, lng, t) => {
  const then = date instanceof Date ? date : new Date(date)
  if (Number.isNaN(then.getTime())) {
    return "-"
  }
  const diffMs = then.getTime() - Date.now()
  const sign = diffMs >= 0 ? 1 : -1
  const absMs = Math.abs(diffMs)
  const seconds = Math.round(absMs / 1000)
  const minutes = Math.round(absMs / 60000)
  const hours = Math.round(absMs / 3600000)
  const days = Math.round(absMs / 86400000)
  const months = Math.round(absMs / 2629800000)
  const years = Math.round(absMs / 31557600000)
  const rtf = new Intl.RelativeTimeFormat(toIntlLocale(lng), { numeric: "auto", style: "long" })

  if (seconds <= 44) {
    if (typeof t === "function") {
      return t(sign >= 0 ? "in_a_few_seconds" : "a_few_seconds_ago")
    }
    return sign >= 0 ? "in a few seconds" : "a few seconds ago"
  }
  if (minutes <= 1) {
    return rtf.format(sign, "minute")
  }
  if (minutes < 45) {
    return rtf.format(sign * minutes, "minute")
  }
  if (hours <= 1) {
    return rtf.format(sign, "hour")
  }
  if (hours < 22) {
    return rtf.format(sign * hours, "hour")
  }
  if (days <= 1) {
    return rtf.format(sign, "day")
  }
  if (days < 26) {
    return rtf.format(sign * days, "day")
  }
  if (months <= 1) {
    return rtf.format(sign, "month")
  }
  if (months < 11) {
    return rtf.format(sign * months, "month")
  }
  if (years <= 1) {
    return rtf.format(sign, "year")
  }
  return rtf.format(sign * years, "year")
}

export const formatAbsolute = (date, lng) => {
  const then = date instanceof Date ? date : new Date(date)
  if (Number.isNaN(then.getTime())) {
    return "-"
  }
  return new Intl.DateTimeFormat(toIntlLocale(lng), {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    second: "2-digit",
    hour12: true,
  }).format(then)
}
