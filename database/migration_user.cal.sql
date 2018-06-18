#<up "1.00">
#<depend "place:1.00">
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
#<end>

#<down "1.00">
DROP TABLE users
#<end>
