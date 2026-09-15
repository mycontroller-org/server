import i18n from "i18next"
import LanguageDetector from "i18next-browser-languagedetector"
import backend from "i18next-http-backend"
import YAML from "js-yaml"
import { initReactI18next } from "react-i18next"
import { DEFAULT_LANGUAGE } from "../Constants/Common"
import { languages } from "./languages"
import { reduxLanguageDetector } from "./languageDetector"

const supportedLngs = languages.map((l) => l.lng)

const normalizeLng = (lng) => {
  if (!lng) {
    return DEFAULT_LANGUAGE
  }
  const normalized = String(lng).replace(/-/g, "_")
  if (supportedLngs.includes(normalized)) {
    return normalized
  }
  const prefix = normalized.split("_")[0]
  const match = supportedLngs.find((s) => s.startsWith(`${prefix}_`))
  return match || DEFAULT_LANGUAGE
}

const languageDetector = new LanguageDetector()
languageDetector.addDetector(reduxLanguageDetector)

i18n
  .use(initReactI18next) // passes i18n down to react-i18next
  .use(languageDetector)
  .use(backend)
  .init({
    ns: "translation",
    fallbackLng: DEFAULT_LANGUAGE,
    supportedLngs,
    load: "currentOnly",
    nonExplicitSupportedLngs: false,
    debug: !process.env.NODE_ENV || process.env.NODE_ENV === "development",
    detection: {
      order: ["redux_language_detector", "navigator"],
      caches: [],
      convertDetectedLanguage: normalizeLng,
    },
    backend: {
      loadPath: "/locales/{{lng}}.yaml",
      parse: function (data) {
        return YAML.load(data)
      },
    },
    interpolation: {
      escapeValue: false, // react already safes from xss
    },
    react: {
      useSuspense: true,
    },
  })

export { normalizeLng }
export default i18n
