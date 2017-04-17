#<up "1.00">
CREATE TABLE offers (
  id serial primary key,
  price int not null,
  counter int not null,
  is_countered boolean not null default(false),
  buyer_comment varchar(140),
  seller_comment varchar(140),
  status varchar(20),
  listing_id int not null references listings(id) on delete cascade,
  seller_id int not null references users(id) on delete cascade,
  buyer_id int not null references users(id) on delete cascade,
  created timestamp with time zone default (now()),
  modified timestamp with time zone default (now()),
  unique (listing_id, buyer_id)
);

CREATE UNIQUE INDEX ind_offers_id ON offers (id);
CREATE INDEX ind_listing_id ON offers (listing_id);
CREATE INDEX ind_seller_id ON offers (seller_id);
CREATE INDEX ind_buyer_id ON offers (buyer_id);
#<end>

#<down "1.00">
DROP TABLE offers;
#<end>
