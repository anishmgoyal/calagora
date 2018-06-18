#<up "1.00">
#<depend "users:1.00">
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
CREATE INDEX ind_images_media ON images (media);
CREATE INDEX ind_images_media_id ON images (media_id);
CREATE INDEX ind_images_user_id ON images (user_id);
#<end>

#<down "1.00">
DROP TABLE images;
#<end>
