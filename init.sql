create extension if not exists CITEXT;

drop table if exists users cascade ;

create table users (
user_id integer not null,
nickname CITEXT NOT NULL,
fullname CITEXT
);

CREATE TABLE events (
event_id serial,
event_name CITEXT,
user_id integer not null
);