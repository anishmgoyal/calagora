#<up "1.00">
CREATE TABLE notifications (
  id serial primary key,
  user_id int references users(id) on delete cascade,
  notification_value text,
  is_read boolean,
  created timestamp with time zone default(now())
);

CREATE UNIQUE INDEX ind_notifications_id ON notifications (id);
CREATE INDEX ind_notifications_user_id ON notifications (user_id);
#<end>

#<down "1.00">
DROP TABLE notifications;
#<end>
