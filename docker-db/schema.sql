-- Places
CREATE TABLE places (
	id SERIAL PRIMARY KEY,
	abbr VARCHAR (5) NOT NULL UNIQUE,
	name VARCHAR (100) NOT NULL UNIQUE,
	email_domain VARCHAR (100) NOT NULL UNIQUE
);

CREATE UNIQUE INDEX ind_places_id ON places (id);
CREATE UNIQUE INDEX ind_places_email_domain ON places (email_domain);

-- Initial data
INSERT INTO places (abbr, name, email_domain)
VALUES  ('RU', 'Rutgers University', 'rutgers.edu'),
        ('CUNY', 'City University of New York', 'cuny.edu'),
        ('GM', 'Gmail Tests', 'gmail.com');

-- User table
CREATE TABLE users (
	id serial primary key,
	username varchar(40) unique not null,
	display_name varchar(100) not null,
	email_address varchar(100) unique not null,
	password varchar(50) not null,
	salt varchar(50) not null,
	activation varchar(100) not null,
	place_id int not null references places(id) ON DELETE CASCADE,
	created timestamp with time zone default (now()),
	modified timestamp with time zone default (now())
);

CREATE UNIQUE INDEX ind_users_id ON users (id);
CREATE UNIQUE INDEX ind_users_username ON users (username);
CREATE UNIQUE INDEX ind_users_email_address ON users (email_address);

-- Sessions Table
CREATE TABLE sessions (
  user_id int not null references users(id),
  session_id varchar(64) not null unique,
  session_secret varchar(64) not null,
  csrf_token varchar(64) not null,
  browser_agent varchar(200) not null,
  created timestamp with time zone default (now()),
  modified timestamp with time zone default (now())
);

CREATE INDEX ind_sessions_user_id ON sessions (user_id);
CREATE UNIQUE INDEX ind_sessions_session_id_session_secret ON sessions (session_id, session_secret);

-- Listings Table
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
CREATE INDEX ind_listings_place_id_type ON listings (place_id, type);
CREATE INDEX ind_listings_type ON listings (type);

-- Offers Table
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

-- Messages Table
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

-- Images Table
CREATE TABLE images (
    id serial primary key,
    media varchar(20),
    media_id int,
    ordinal int,
    url varchar(255),
    user_id int references users(id) on delete cascade,
    created timestamp with time zone default(now()),
    modified timestamp with time zone default(now())
);

CREATE UNIQUE INDEX ind_images_id ON images (id);
CREATE INDEX ind_images_media_media_id ON images (media, media_id);
CREATE INDEX ind_images_user_id ON images (user_id);

-- Notifications Table
CREATE TABLE notifications (
  id serial primary key,
  user_id int references users(id) on delete cascade,
  notification_value text,
  is_read boolean,
  created timestamp with time zone default(now())
);

CREATE UNIQUE INDEX ind_notifications_id ON notifications (id);
CREATE INDEX ind_notifications_user_id ON notifications (user_id);

-- Password Recovery Table
CREATE TABLE password_recovery_requests (
  user_id INT PRIMARY KEY NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  recovery_string VARCHAR(100) NOT NULL,
  recovery_code VARCHAR(7) NOT NULL,
  is_valid BOOLEAN NOT NULL,
  created TIMESTAMP WITH TIME ZONE DEFAULT(now()),
  modified TIMESTAMP WITH TIME ZONE DEFAULT(now())
);

CREATE INDEX ind_password_recovery_requests_user_id ON password_recovery_requests (user_id);

-- Search Entries
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
CREATE INDEX ind_search_entries_word_place_id ON search_entries (word, place_id);
CREATE INDEX ind_search_entries_listing_type ON search_entries (listing_type);
