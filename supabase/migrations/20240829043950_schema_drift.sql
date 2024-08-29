create table "public"."account" (
    "id" bigint generated by default as identity not null,
    "created_at" timestamp with time zone not null default now(),
    "name" text not null,
    "info" text not null,
    "location" text not null,
    "email" text not null,
    "experience_level" smallint not null default '0'::smallint
);


alter table "public"."account" enable row level security;

create table "public"."passwords" (
    "id" bigint generated by default as identity not null,
    "created_at" timestamp with time zone not null default now(),
    "account_email" text not null,
    "hash" text not null,
    "salt" text not null
);


alter table "public"."passwords" enable row level security;

CREATE UNIQUE INDEX account_email_key ON public.account USING btree (email);

CREATE UNIQUE INDEX account_pkey ON public.account USING btree (id);

CREATE UNIQUE INDEX pass_account_id_key ON public.passwords USING btree (account_email);

CREATE UNIQUE INDEX pass_pkey ON public.passwords USING btree (id);

CREATE UNIQUE INDEX passwords_hash_key ON public.passwords USING btree (hash);

CREATE UNIQUE INDEX passwords_salt_key ON public.passwords USING btree (salt);

alter table "public"."account" add constraint "account_pkey" PRIMARY KEY using index "account_pkey";

alter table "public"."passwords" add constraint "pass_pkey" PRIMARY KEY using index "pass_pkey";

alter table "public"."account" add constraint "account_email_key" UNIQUE using index "account_email_key";

alter table "public"."passwords" add constraint "pass_account_id_key" UNIQUE using index "pass_account_id_key";

alter table "public"."passwords" add constraint "passwords_account_email_fkey" FOREIGN KEY (account_email) REFERENCES account(email) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."passwords" validate constraint "passwords_account_email_fkey";

alter table "public"."passwords" add constraint "passwords_hash_key" UNIQUE using index "passwords_hash_key";

alter table "public"."passwords" add constraint "passwords_salt_key" UNIQUE using index "passwords_salt_key";

grant delete on table "public"."account" to "anon";

grant insert on table "public"."account" to "anon";

grant references on table "public"."account" to "anon";

grant select on table "public"."account" to "anon";

grant trigger on table "public"."account" to "anon";

grant truncate on table "public"."account" to "anon";

grant update on table "public"."account" to "anon";

grant delete on table "public"."account" to "authenticated";

grant insert on table "public"."account" to "authenticated";

grant references on table "public"."account" to "authenticated";

grant select on table "public"."account" to "authenticated";

grant trigger on table "public"."account" to "authenticated";

grant truncate on table "public"."account" to "authenticated";

grant update on table "public"."account" to "authenticated";

grant delete on table "public"."account" to "service_role";

grant insert on table "public"."account" to "service_role";

grant references on table "public"."account" to "service_role";

grant select on table "public"."account" to "service_role";

grant trigger on table "public"."account" to "service_role";

grant truncate on table "public"."account" to "service_role";

grant update on table "public"."account" to "service_role";

grant delete on table "public"."passwords" to "anon";

grant insert on table "public"."passwords" to "anon";

grant references on table "public"."passwords" to "anon";

grant select on table "public"."passwords" to "anon";

grant trigger on table "public"."passwords" to "anon";

grant truncate on table "public"."passwords" to "anon";

grant update on table "public"."passwords" to "anon";

grant delete on table "public"."passwords" to "authenticated";

grant insert on table "public"."passwords" to "authenticated";

grant references on table "public"."passwords" to "authenticated";

grant select on table "public"."passwords" to "authenticated";

grant trigger on table "public"."passwords" to "authenticated";

grant truncate on table "public"."passwords" to "authenticated";

grant update on table "public"."passwords" to "authenticated";

grant delete on table "public"."passwords" to "service_role";

grant insert on table "public"."passwords" to "service_role";

grant references on table "public"."passwords" to "service_role";

grant select on table "public"."passwords" to "service_role";

grant trigger on table "public"."passwords" to "service_role";

grant truncate on table "public"."passwords" to "service_role";

grant update on table "public"."passwords" to "service_role";


