import {
  type RouteConfig,
  index,
  layout,
  prefix,
} from "@react-router/dev/routes"

const routeConfig: RouteConfig = [
  layout("./routes/layout.tsx", [
    index("./routes/index.tsx"), //
    ...prefix(".link", [
      ...prefix(":linkKey", [
        index("./routes/link/$linkKey/index.tsx"), //
      ]),
    ]),
    ...prefix(":linkKey", [
      index("./routes/index.tsx", { id: "link-key-page" }), //
    ]),
  ]),
]

export default routeConfig
