#<up "1.00">
#<depend "users:1.00">
#<depend "offers:1.00">
CREATE TABLE messages (
  id serial primary key,
  message varchar(200),
  seen boolean default(false),
  sender_id int references users(id) on delete cascade,
  recepient_id int references users(id) on delete cascade,
  offer_id int references offers(id) on delete cascade,
  created timestamp with time zone default(now()),
  modified timestamp with time zone default(now())
);

CREATE UNIQUE INDEX ind_messages_id ON messages (id);
CREATE INDEX ind_messages_sender_id ON messages (sender_id);
CREATE INDEX ind_messages_recepient_id ON messages (recepient_id);
CREATE INDEX ind_messages_offer_id ON messages (offer_id);
#<end>

#<down "1.00">
DROP TABLE messages
#<end>
