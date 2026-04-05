import {
  type RouteConfig,
  index,
  layout,
  prefix,
} from "@react-router/dev/routes"

const routeConfig: RouteConfig = [
  layout("./layout.tsx", [
    layout("./layout/StuffAppBarLayout.tsx", [
      index("./page.tsx"), //
    ]),
    ...prefix("print", [
      index("./print/page.tsx"), //
    ]),
  ]),
]

export default routeConfig
