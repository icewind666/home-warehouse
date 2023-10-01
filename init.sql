create table if not exists items
(
    id          serial
        constraint id
            primary key,
    title       varchar not null,
    description text,
    quantity    integer,
    expires     date
);