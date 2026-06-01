// See: https://ui.shadcn.com/docs/installation/manual#add-a-cn-helper
import { twMerge } from "tailwind-merge"

export type CommonClassGroups = {
  component?: string
  layout?: string
  appearance?: string
  state?: string
}

export const tw = (...inputs: (CommonClassGroups | string | boolean)[]) => {
  const classes: string[] = []
  for (const input of inputs) {
    if (typeof input === "boolean") {
      continue
    }
    if (typeof input === "string") {
      classes.push(input)
      continue
    }
    for (const v of Object.values(input)) {
      if (v) {
        classes.push(v)
      }
    }
  }
  return twMerge(...classes)
}

export default tw
