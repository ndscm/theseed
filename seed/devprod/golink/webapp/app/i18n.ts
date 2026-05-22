import i18next from "i18next"
import { initReactI18next } from "react-i18next"

import en from "../locales/en/en"
import es from "../locales/es/es"
import fallbackLng from "../locales/fallback"

declare const BUILD_LANGUAGE: string

const resources = {
  en,
  es,
}

i18next //
  .use(initReactI18next)
  .init({
    lng: BUILD_LANGUAGE,
    resources,
    fallbackLng,
    load: "currentOnly",
    interpolation: {
      escapeValue: false, // not needed for react
    },
    debug: !!(
      typeof window !== "undefined" && window.localStorage.getItem("debug")
    ),
  })

export default i18next
