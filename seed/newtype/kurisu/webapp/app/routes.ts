import {
  type RouteConfig,
  index,
  layout,
  prefix,
} from "@react-router/dev/routes"

const routeConfig: RouteConfig = [
  layout("./routes/layout.tsx", [
    index("./routes/index.tsx"), //
    ...prefix("team", [
      layout("./routes/team/layout.tsx", [
        index("./routes/team/index.tsx"), //
        ...prefix("members", [
          index("./routes/team/members/index.tsx"), //
        ]),
      ]),
    ]),
  ]),
]

export default routeConfig
