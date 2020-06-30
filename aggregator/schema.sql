CREATE TABLE users (
  id CHAR(36) PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL
);

CREATE TABLE devices (
  uid CHAR(36) NOT NULL,
  id INTEGER NOT NULL,
  name VARCHAR(255) NOT NULL,
  kind INTEGER NOT NULL,
  PRIMARY KEY(uid, id)
);

CREATE TABLE events (
  uid CHAR(36) NOT NULL,
  did INTEGER NOT NULL,
  id INTEGER NOT NULL,
  kind INTEGER NOT NULL,
  starttime INTEGER NOT NULL,
  endtime INTEGER NOT NULL,
  PRIMARY KEY(uid, did, id)
);

CREATE TABLE apps (
  name VARCHAR(255) PRIMARY KEY,
  label VARCHAR(255)
);

CREATE TABLE app_switch_events (
  uid CHAR(36) NOT NULL,
  did INTEGER NOT NULL,
  id INTEGER NOT NULL,
  app VARCHAR(255) NOT NULL,
  PRIMARY KEY(uid, did, id)
);

INSERT INTO users VALUES('9e9b23c8-8cf1-4891-b201-5bc0467ba535','test@productimon.com','$2a$10$18SpmyR9yo4pegsfy/a1W.SuYTmgYSMNoNmuS0T9EQE6OQPh40rLK'); -- password: test
INSERT INTO devices VALUES('9e9b23c8-8cf1-4891-b201-5bc0467ba535', 0, 'test device (Linux)', 1);
