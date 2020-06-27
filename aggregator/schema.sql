CREATE TABLE users (
  id CHAR(36) PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL
);

INSERT INTO users VALUES('9e9b23c8-8cf1-4891-b201-5bc0467ba535','test@productimon.com','$2a$10$18SpmyR9yo4pegsfy/a1W.SuYTmgYSMNoNmuS0T9EQE6OQPh40rLK'); -- password: test
