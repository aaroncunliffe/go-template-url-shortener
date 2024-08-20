-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";
-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'standard public schema';
-- Create "links" table
CREATE TABLE "public"."links" ("id" bigserial NOT NULL, "short_path" text NOT NULL, "original_url" text NOT NULL, "created_at" timestamp NOT NULL DEFAULT now(), "updated_at" timestamp NULL, PRIMARY KEY ("id"), CONSTRAINT "links_short_path_key" UNIQUE ("short_path"));
