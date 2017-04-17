#<up "1.00">
#<depend "listing:1.00">

CREATE TABLE search_entries (
  id SERIAL PRIMARY KEY,
  word VARCHAR(255) NOT NULL,
  count INT NOT NULL,
  listing_id INT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
  listing_name VARCHAR(255) NOT NULL,
  listing_price INT NOT NULL,
  listing_image VARCHAR(255) NOT NULL,
  listing_type listing_type NOT NULL,
  place_id INT NOT NULL REFERENCES places(id) ON DELETE CASCADE,
  created TIMESTAMP WITH TIME ZONE DEFAULT(now()),
  modified TIMESTAMP WITH TIME ZONE DEFAULT(now())
);

CREATE UNIQUE INDEX ind_search_entries_id ON search_entries (id);
CREATE INDEX ind_search_entries_word ON search_entries (word);
CREATE INDEX ind_search_entries_count ON search_entries (count);
CREATE INDEX ind_search_entries_place_id ON search_entries (place_id);
CREATE INDEX ind_search_entries_listing_type ON search_entries (listing_type);
#<end>

#<down "1.00">
DROP TABLE search_entry;
#<end>
