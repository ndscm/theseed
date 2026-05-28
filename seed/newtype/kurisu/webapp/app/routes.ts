import {
  type RouteConfig,
  index,
  layout,
  prefix,
} from "@react-router/dev/routes"

const routeConfig: RouteConfig = [
  layout("./routes/layout.tsx", [
    index("./routes/index.tsx"), //
  ]),
]

export default routeConfig
