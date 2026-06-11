import React from "react"
import { useTranslation } from "react-i18next"

import tw from "../../../../../../../../devprod/ts/grouping-tailwind"

const PersonWorkstationPage: React.FC<{ params: { handle: string } }> = ({
  params,
}) => {
  const { handle } = params
  const { t } = useTranslation("person")

  const personHandle = handle
    .trim()
    .replace(/^@/, "")
    .replace(/@$/, "")
    .toLowerCase()
    .trim()

  return (
    <main
      className={tw({ layout: "min-h-0 flex-1 overflow-auto px-7 py-6" })}
    ></main>
  )
}

export default PersonWorkstationPage
