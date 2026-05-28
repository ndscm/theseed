import React from "react"

import tw from "../../../../devprod/ts/grouping-tailwind"

const KurisuPanelInfoRow: React.FC<{
  label: string
  value: React.ReactNode
  mono?: boolean
  last?: boolean
}> = ({ label, value, mono, last }) => {
  return (
    <div
      className={tw(
        {
          layout: "flex gap-4 py-3",
        },
        last ? {} : { appearance: "border-base-200 border-b" },
      )}
    >
      <div
        className={tw({
          layout: "w-32 shrink-0",
          appearance: "text-neutral text-sm font-medium",
        })}
      >
        {label}
      </div>
      <div
        className={tw(
          {
            layout: "min-w-0 flex-1",
            appearance: "text-base-content text-sm font-medium",
          },
          mono ? { appearance: "font-mono" } : {},
        )}
      >
        {value}
      </div>
    </div>
  )
}

export default KurisuPanelInfoRow
