create table if not exists users (
  id bigserial primary key,
  username text not null unique,
  password_hash text not null,
  created_at timestamptz not null default now()
);

create table if not exists refresh_tokens (
  id bigserial primary key,
  user_id bigint not null references users(id) on delete cascade,
  token_hash text not null,
  revoked_at timestamptz null,
  expires_at timestamptz not null,
  created_at timestamptz not null default now(),
  unique (user_id, token_hash)
);

create table if not exists rooms (
  id bigserial primary key,
  name text not null unique,
  created_at timestamptz not null default now()
);

create table if not exists messages (
  id bigserial primary key,
  room_id bigint not null references rooms(id) on delete cascade,
  user_id bigint not null references users(id) on delete cascade,
  text text not null,
  created_at timestamptz not null default now()
);

create index if not exists idx_messages_room_id_id on messages(room_id, id desc);
