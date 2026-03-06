-- Create "golink" table
CREATE TABLE "golink" (
  "key" text NOT NULL,
  "target" text NOT NULL,
  "public" boolean NOT NULL DEFAULT true,
  "owner" text NULL,
  "hit_count" bigint NOT NULL DEFAULT 0,
  "created_time" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_time" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("key")
);
