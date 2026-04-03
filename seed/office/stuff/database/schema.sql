-- Create "categories" table
CREATE TABLE "categories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "primary" character varying NULL,
  "secondary" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "stuffs" table
CREATE TABLE "stuffs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order" character varying NOT NULL,
  "data" jsonb NULL,
  "owner" character varying NULL,
  "created_time" timestamptz NOT NULL,
  "updated_time" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
