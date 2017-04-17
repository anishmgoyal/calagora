#<up "1.00">
CREATE TABLE places (
	id SERIAL PRIMARY KEY,
	abbr VARCHAR (5) NOT NULL UNIQUE,
	name VARCHAR (100) NOT NULL UNIQUE,
	email_domain VARCHAR (100) NOT NULL UNIQUE
);

CREATE UNIQUE INDEX ind_places_id ON places (id);
CREATE UNIQUE INDEX ind_places_email_domain ON places (email_domain);
#end

#<up "1.01">
INSERT INTO places (abbr, name, email_domain)
VALUES ('RU', 'Rutgers University', 'rutgers.edu'),
('CUNY', 'City University of New York', 'cuny.edu')
#<end>

#<down "1.01">
DELETE FROM places WHERE abbr IN ('RU', 'CUNY')
#<end>

#<down "1.00">
DROP TABLE places;
#<end>
