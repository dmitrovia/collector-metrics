CREATE TABLE gauges (
   id serial primary key,
   name varchar not null,
   value double precision,
   created_at TIMESTAMP default now()
);

CREATE TABLE counters (
   id serial primary key,
   name varchar not null,
   value bigint,
   created_at TIMESTAMP default now()
);