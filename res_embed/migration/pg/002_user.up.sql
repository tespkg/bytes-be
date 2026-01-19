create table if not exists users
(
    "id"                            bigserial                   primary key not null,
    "uuid"                          text                        not null,
    "email"                         text                        default null,
    "phone"                         text                        default null,
    "password"                      bytea                       default null,
    "is_enabled"                    bool                        not null default false, -- 是否启用
    "created_at"                    timestamp with time zone    not null default now() ,
    "updated_at"                    timestamp with time zone    not null default now() ,
    "deleted_at"                    timestamp with time zone    default null,

    constraint email_or_phone_not_null check (email is not null or phone is not null)
);

create unique index if not exists uidx_users_uuid on users (uuid);
create unique index if not exists uidx_users_email_when_enabled on users(email) WHERE is_enabled = true;
create unique index if not exists uidx_users_phone_when_enabled on users(phone) WHERE is_enabled = true;
create index if not exists idx_users_uuid on users(uuid);
create index if not exists idx_users_email on users(email);
create index if not exists idx_users_phone on users(phone);


create table if not exists roles
(
    "id"                            bigserial                   primary key not null,
    "role"                          varchar(20)                 not null
);

create unique index if not exists idx_roles_role on roles(role);
INSERT INTO roles(role) VALUES('admin'),('merchant'),('customer'),('driver');


create table if not exists user_roles
(
    "id"                            bigserial                   primary key not null,
    "user_id"                       bigint                      not null references users(id),
    "role_id"                       bigint                      not null references roles(id),
    "created_at"                    timestamp with time zone    not null default now(),
    "updated_at"                    timestamp with time zone    not null default now(),
    "deleted_at"                    timestamp with time zone    default null
);

create unique index if not exists uidx_user_roles_user_id_role_id on user_roles(user_id, role_id) WHERE deleted_at IS NULL;


create table if not exists customers
(
    "id"                            bigserial                   primary key not null,
    "user_id"                       bigint                      not null references users(id),
    "first_name"                    text                        default null,
    "last_name"                     text                        default null,
    "gender"                        int                         default null, -- null:unknown, 0:male, 1:female
    "birthday"                      text                        default null,
    "nationality"                   text                        default null,
    "created_at"                    timestamp with time zone    not null default now() ,
    "updated_at"                    timestamp with time zone    not null default now() ,
    "deleted_at"                    timestamp with time zone    default null
);

create unique index if not exists uidx_customer_user_id on customers(user_id);


create table if not exists customer_addresses
(
    "id"                            bigserial                   primary key not null,
    "customer_id"                   bigint                      not null references customers(id),
    "country"                       text                        default null,
    "state"                         text                        default null,
    "city"                          text                        default null,
    "street"                        text                        default null,
    "zip_code"                      text                        default null,
    "address"                       text                        not null,
    "longitude"                     numeric(20,10)              not null ,
    "latitude"                      numeric(20,10)              not null ,
    "created_at"                    timestamp with time zone    not null default now() ,
    "updated_at"                    timestamp with time zone    not null default now() ,
    "deleted_at"                    timestamp with time zone    default null
);

create index if not exists idx_customer_addresses_customer_id on customer_addresses(customer_id);
