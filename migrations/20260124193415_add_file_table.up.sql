CREATE TABLE file (
    id uuid primary key not null,
    user_id uuid not null references public.user(id),
    size bigint not null, -- file size in bytes
    content_type varchar(20) not null,
    subject varchar(20) not null,
    created_at TIMESTAMPTZ not null,
    updated_at TIMESTAMPTZ not null,
    deleted_at TIMESTAMPTZ null
);