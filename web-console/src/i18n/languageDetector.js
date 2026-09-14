import { store } from "../store/persister"

export const reduxLanguageDetector = {
  name: "redux_language_detector",

  lookup(_options) {
    // options -> are passed in options
    
    // get selected language from redux
    return store.getState().entities.locale.language
  },

  cacheUserLanguage(_lng, _options) {
    // options -> are passed in options
    // lng -> current language, will be called after init and on changeLanguage
    // store it
    // nothing to do here, changed will be stored
  },
}
