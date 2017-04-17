#<up "1.00">
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
CREATE UNIQUE INDEX ind_sessions_session_id ON sessions (session_id);
CREATE INDEX ind_sessions_session_secret ON sessions (session_secret);
#end

#<down "1.00">
drop table sessions
#end
