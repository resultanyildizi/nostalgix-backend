CREATE TABLE album
(
    id         VARCHAR PRIMARY KEY,
    name       VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

create table public.user 
(
    id uuid primary key not null,
    name varchar(50) not null,
    customer_id uuid not null,
    auth_method varchar(20) not null,
    auth_id text not null,
    is_new_user boolean not null,
    fcm_token text null,
    credits bigint not null,
    credits_expires_at TIMESTAMPTZ null,
    subscription_plan varchar null,
    subscription_type varchar null,
    subscription_period varchar null,
    subscription_status varchar null,
    subscription_expires_at TIMESTAMPTZ null,
    created_at TIMESTAMPTZ not null,
    updated_at TIMESTAMPTZ not null,
    deleted_at TIMESTAMPTZ null
);