import {
  type RouteConfig,
  index,
  layout,
  prefix,
} from "@react-router/dev/routes"

const routeConfig: RouteConfig = [
  layout("./routes/layout.tsx", [
    layout("./routes/StuffAppBarLayout.tsx", [
      index("./routes/index.tsx"), //
    ]),
    ...prefix("print", [
      index("./routes/print/index.tsx"), //
    ]),
  ]),
]

export default routeConfig
