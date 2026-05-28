import React from "react"

import tw from "../../../../devprod/ts/grouping-tailwind"

const KurisuAvatar: React.FC<{
  size?: "small" | "large"
  organic?: "carbon" | "silicon"
  children?: React.ReactNode
}> = ({ size, organic, children }) => {
  return (
    <div
      className={tw({
        component: "avatar avatar-placeholder",
        layout: "shrink-0",
      })}
    >
      <div
        className={tw(
          {
            layout: "flex shrink-0 items-center justify-center",
            appearance: "border-2 bg-transparent",
          },
          size === "small"
            ? { layout: "h-9 w-9", appearance: "rounded-lg" }
            : size === "large"
              ? { layout: "h-16 w-16", appearance: "rounded-2xl" }
              : { layout: "h-12 w-12", appearance: "rounded-xl" },
          organic === "carbon"
            ? { appearance: "border-carbon text-carbon" }
            : organic === "silicon"
              ? { appearance: "border-silicon text-silicon" }
              : { appearance: "border-primary text-primary" },
        )}
      >
        {children}
      </div>
    </div>
  )
}

export default KurisuAvatar
