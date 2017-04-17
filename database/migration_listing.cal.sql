#<up "1.00">
#<depend "user:1.00">
#<depend "place:1.00">

CREATE TYPE listing_type AS ENUM
  ('misc', 'textbook', 'homegoods', 'housing', 'automotive', 'electronics',
  'clothing', 'athletics');

CREATE TYPE listing_condition AS ENUM
  ('na', 'new', 'excellent', 'good', 'fair', 'poor', 'forparts');

CREATE TYPE listing_status AS ENUM
  ('listed', 'transaction', 'sold');

CREATE TABLE listings (
  id serial primary key,
  name varchar(255) not null,
  price int not null,
  type listing_type not null,
  condition listing_condition not null,
  status listing_status not null,
  description varchar(65535),
  published boolean not null default (FALSE),
  place_id int not null references places(id) ON DELETE CASCADE,
  user_id int not null references users(id) ON DELETE CASCADE,
  created timestamp with time zone default (now()),
  modified timestamp with time zone default (now())
);

CREATE UNIQUE INDEX ind_listings_id ON listings (id);
CREATE INDEX ind_listings_user_id ON listings (user_id);
CREATE INDEX ind_listings_place_id ON listings (place_id);
CREATE INDEX ind_listings_type ON listings (type);
#<end>

#<down "1.00">
DROP TABLE listings;
DROP TYPE listing_status;
DROP TYPE listing_condition;
DROP TYPE listing_type;
#<end>
