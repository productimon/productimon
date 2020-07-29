CREATE TABLE users (
  id CHAR(36) PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  verified BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE devices (
  uid CHAR(36) NOT NULL,
  id INTEGER NOT NULL,
  name VARCHAR(255) NOT NULL,
  kind INTEGER NOT NULL,
  PRIMARY KEY(uid, id),
  FOREIGN KEY(uid) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE events (
  uid CHAR(36) NOT NULL,
  did INTEGER NOT NULL,
  id INTEGER NOT NULL,
  kind INTEGER NOT NULL,
  starttime INTEGER NOT NULL,
  endtime INTEGER NOT NULL,
  PRIMARY KEY(uid, did, id),
  FOREIGN KEY (uid, did) REFERENCES devices(uid, id) ON DELETE CASCADE
);

CREATE TABLE intervals (
  uid CHAR(36) NOT NULL,
  did INTEGER NOT NULL,
  starttime INTEGER NOT NULL,
  endtime INTEGER NOT NULL,
  activetime INTEGER NOT NULL,
  app VARCHAR(255) NOT NULL,
  PRIMARY KEY(uid, did, starttime),
  FOREIGN KEY (uid, did) REFERENCES devices(uid, id) ON DELETE CASCADE
);

CREATE TABLE default_apps (
  name VARCHAR(255) PRIMARY KEY,
  label VARCHAR(255)
);

CREATE TABLE user_apps (
  name VARCHAR(255),
  uid CHAR(36) NOT NULL,
  label VARCHAR(255),
  PRIMARY KEY(name, uid),
  FOREIGN KEY (uid) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE app_switch_events (
  uid CHAR(36) NOT NULL,
  did INTEGER NOT NULL,
  id INTEGER NOT NULL,
  app VARCHAR(255) NOT NULL,
  PRIMARY KEY(uid, did, id),
  FOREIGN KEY (uid, did, id) REFERENCES events(uid, did, id) ON DELETE CASCADE
);

CREATE TABLE activity_events (
  uid CHAR(36) NOT NULL,
  did INTEGER NOT NULL,
  id INTEGER NOT NULL,
  keystrokes INTEGER NOT NULL,
  mouseclicks INTEGER NOT NULL,
  PRIMARY KEY(uid, did, id),
  FOREIGN KEY (uid, did, id) REFERENCES events(uid, did, id) ON DELETE CASCADE
);

INSERT INTO users VALUES('9e9b23c8-8cf1-4891-b201-5bc0467ba535','test@productimon.com','$2a$10$18SpmyR9yo4pegsfy/a1W.SuYTmgYSMNoNmuS0T9EQE6OQPh40rLK', TRUE); -- password: test
