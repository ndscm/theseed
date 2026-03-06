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

// nagi: The automatic language detector will trigger react hydration error,
// because the server is rendering with default language while the client is
// rendering with navigator language. We have to to detect language on our own
// in the login context and change to our expected language dynamically during
// the second round rendering.

i18next //
  .use(initReactI18next)
  .init({
    lng: "en",
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
