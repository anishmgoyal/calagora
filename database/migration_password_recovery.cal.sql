#<up "1.00">
#<depend "user:1.00">
CREATE TABLE password_recovery_requests (
  user_id INT PRIMARY KEY NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  recovery_string VARCHAR(100) NOT NULL,
  recovery_code VARCHAR(7) NOT NULL,
  is_valid BOOLEAN NOT NULL,
  created TIMESTAMP WITH TIME ZONE DEFAULT(now()),
  modified TIMESTAMP WITH TIME ZONE DEFAULT(now())
);

CREATE INDEX ind_password_recovery_requests_user_id ON password_recovery_requests (user_id);
#<end>

#<down "1.00">
DROP TABLE password_recovery_requests;
#<end>
