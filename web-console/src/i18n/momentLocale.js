import moment from "moment"
import "moment/locale/de"
import "moment/locale/en-gb"
import "moment/locale/es"
import "moment/locale/fr"
import "moment/locale/he"
import "moment/locale/hi"
import "moment/locale/it"
import "moment/locale/kn"
import "moment/locale/ml"
import "moment/locale/nl"
import "moment/locale/pl"
import "moment/locale/pt"
import "moment/locale/ro"
import "moment/locale/ru"
import "moment/locale/ta"
import "moment/locale/te"
import "moment/locale/zh-cn"
import "moment/locale/zh-tw"

const SPECIAL = {
  en_GB: "en-gb",
  en_US: "en",
  zh_CN: "zh-cn",
  zh_TW: "zh-tw",
}

// Keep ASCII digits in the console. Moment's ta/hi/kn locales otherwise
// rewrite 5 to ௫ / ५ / ೫, which does not match other table values.
;["ta", "hi", "kn"].forEach((locale) => {
  moment.updateLocale(locale, {
    preparse: (string) => string,
    postformat: (string) => string,
  })
})

export const toMomentLocale = (lng) => {
  if (SPECIAL[lng]) {
    return SPECIAL[lng]
  }
  const prefix = String(lng || "")
    .split("_")[0]
    .toLowerCase()
  return prefix || "en-gb"
}

export const applyMomentLocale = (lng) => {
  moment.locale(toMomentLocale(lng))
}
