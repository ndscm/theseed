import React from "react"
import { useTranslation } from "react-i18next"

import tw from "../../../../../devprod/ts/grouping-tailwind"
import KurisuTopBar from "../../components/KurisuTopBar"

const HomePage: React.FC<{}> = ({}) => {
  const { t } = useTranslation("home")

  return (
    <>
      <header className={tw({ layout: "sticky top-0" })}>
        <KurisuTopBar title={t("page.title", "Kurisu")} />
      </header>
      <main
        className={tw({ layout: "min-h-0 flex-1 overflow-auto px-7 py-6" })}
      >
        {t("system.welcome", "Kurisu is the portal of steins agent system!")}
      </main>
    </>
  )
}

export default HomePage
