import React from "react"

import tw from "../../../../devprod/ts/grouping-tailwind"

const KurisuPanel: React.FC<{
  title: React.ReactNode
  subtitle?: React.ReactNode
  action?: React.ReactNode
  children: React.ReactNode
}> = ({ title, subtitle, action, children }) => {
  return (
    <div
      className={tw({
        appearance:
          "bg-base-100 border-base-300 rounded-xl border shadow-sm transition-shadow",
        state: "hover:shadow-md",
      })}
    >
      <div
        className={tw({
          layout: "flex items-center justify-between gap-3 px-5 py-3",
          appearance: "border-base-200 border-b",
        })}
      >
        <div className={tw({ layout: "min-w-0 flex-1" })}>
          <div
            className={tw({
              appearance:
                "text-base-content truncate text-sm font-semibold tracking-tight",
            })}
          >
            {title}
          </div>
          {subtitle && (
            <div
              className={tw({
                layout: "mt-0.5",
                appearance: "text-neutral truncate text-xs",
              })}
            >
              {subtitle}
            </div>
          )}
        </div>
        {action && <div className={tw({ layout: "shrink-0" })}>{action}</div>}
      </div>
      <div className={tw({ layout: "px-5 py-4" })}>{children}</div>
    </div>
  )
}

export default KurisuPanel
