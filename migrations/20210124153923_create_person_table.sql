-- +swan Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS persons
(
  id bigserial primary key,
  name text not null,
  age bigint not null,
  height text not null,
  weight text not null,
  created_at timestamp with time zone
);

-- +swan Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS persons;
