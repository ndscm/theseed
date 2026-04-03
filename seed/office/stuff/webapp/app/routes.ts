import { type RouteConfig, index, layout } from "@react-router/dev/routes"

const routeConfig: RouteConfig = [
  layout("./layout.tsx", [
    layout("./layout/StuffAppBarLayout.tsx", [
      index("./page.tsx"), //
    ]),
  ]),
]

export default routeConfig
