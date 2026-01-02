create table refresh_token (
    id uuid primary key not null,
    user_id uuid not null,
    hashed_value text not null,
    device_key text not null,
    created_at TIMESTAMPTZ not null, 
    expires_at TIMESTAMPTZ not null,
    revoked_at TIMESTAMPTZ
);

create index refresh_token_device_key_idx on refresh_token (device_key);
create index refresh_token_hashed_value_idx on refresh_token (hashed_value);