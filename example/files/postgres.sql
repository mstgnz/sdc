-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS blogs_id_seq;

-- Table Definition
CREATE TABLE "public"."blogs" (
    "id" int8 NOT NULL DEFAULT nextval('blogs_id_seq'::regclass),
    "title" varchar NOT NULL,
    "slug" varchar,
    "user_id" int8 NOT NULL,
    "short_text" varchar NOT NULL,
    "long_text" text NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT now(),
    "updated_at" timestamp NOT NULL,
    CONSTRAINT "blogs_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);

-- Column Comment
COMMENT ON COLUMN "public"."blogs"."slug" IS 'Title Slug';

-- Trigger
CREATE OR REPLACE FUNCTION public.delete_blogs_when_user_delete()
    RETURNS trigger
    LANGUAGE plpgsql
AS $function$
BEGIN
    DELETE FROM blogs WHERE user_id = OLD.id;
    RETURN OLD;
END;
$function$