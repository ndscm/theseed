-- Create "link" table
CREATE TABLE "link" (
  "key" character varying NOT NULL,
  "target" character varying NOT NULL,
  "public" boolean NOT NULL DEFAULT true,
  "owner" character varying NULL,
  "hit_count" bigint NOT NULL DEFAULT 0,
  "created_time" timestamptz NOT NULL,
  "updated_time" timestamptz NOT NULL,
  PRIMARY KEY ("key")
);
