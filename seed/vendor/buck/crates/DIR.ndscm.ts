import { type DirConfig } from "//seed/devprod/ndscm/config/DIR"

export default {
  vendor: {
    vendor: {
      target: "vendor",
      watchRepo: "Cargo.toml",
      watch: "^reindeer.toml$",
      buck: {
        run: "reindeer vendor",
      },
    },
    buck: {
      target: "BUCK",
      watchRepo: "seed/vendor/buck/crates/vendor",
      watch: "^reindeer.toml$",
      buck: {
        run: "reindeer buckify",
      },
    },
  },
} satisfies DirConfig
