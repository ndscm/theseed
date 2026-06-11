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
    ...prefix("person", [
      index("./routes/person/index.tsx"), //
      ...prefix(":handle", [
        layout("./routes/person/$handle/layout.tsx", [
          index("./routes/person/$handle/index.tsx"), //
          ...prefix("chat", [
            index("./routes/person/$handle/chat/index.tsx"), //
          ]),
        ]),
      ]),
    ]),
  ]),
]

export default routeConfig
