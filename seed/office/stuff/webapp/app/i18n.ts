import i18next from "i18next"
import { initReactI18next } from "react-i18next"

import LocalesEnCommon from "../locales/en/common.json"

const resources = {
  en: {
    common: LocalesEnCommon,
  },
}

const fallbackLng = {
  default: ["en"],
  "en-US": ["en"],
}

i18next //
  .use(initReactI18next)
  .init({
    lng: "en",
    resources,
    fallbackLng,
    load: "currentOnly",
    interpolation: {
      escapeValue: false,
    },
    debug: !!(
      typeof window !== "undefined" && window.localStorage.getItem("debug")
    ),
  })

export default i18next
