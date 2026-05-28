import React from "react"

import { PanelLeftOpenIcon } from "lucide-react"

import tw from "../../../../devprod/ts/grouping-tailwind/dist"

const KurisuTopBar: React.FC<{
  title: React.ReactNode
  children?: React.ReactNode
}> = ({ title, children }) => {
  return (
    <nav
      className={tw({
        layout: "flex h-16 shrink-0 items-center gap-4 px-6",
        appearance: "border-base-300 bg-base-100 border-b",
        state: "max-sm:gap-3 max-sm:px-4",
      })}
    >
      <label
        htmlFor="kurisu-sidebar-toggle"
        className={tw({
          layout: "flex h-9 w-9 shrink-0 items-center justify-center",
          appearance:
            "border-base-300 bg-base-100 text-neutral-content cursor-pointer transition-colors",
          state: "hover:bg-base-200 hover:text-base-content lg:hidden",
        })}
      >
        <PanelLeftOpenIcon className={tw({ layout: "h-5 w-5 shrink-0" })} />
      </label>
      <div
        className={tw({
          appearance: "text-base-content text-lg font-semibold tracking-tight",
        })}
      >
        {title}
      </div>
      <div className={tw({ layout: "ml-auto flex items-center gap-3" })}>
        {children}
      </div>
    </nav>
  )
}

export default KurisuTopBar
