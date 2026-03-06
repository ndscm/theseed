import {
  type RouteConfig,
  index,
  layout,
  prefix,
} from "@react-router/dev/routes"

const routeConfig: RouteConfig = [
  layout("./layout.tsx", [
    index("./page.tsx"), //
    ...prefix(".link", [
      ...prefix(":linkKey", [
        index("./link/$linkKey/page.tsx"), //
      ]),
    ]),
    ...prefix(":linkKey", [
      index("./page.tsx", { id: "link-key-page" }), //
    ]),
  ]),
]

export default routeConfig
