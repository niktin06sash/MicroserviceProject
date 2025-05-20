CREATE TABLE users (
    userid UUID PRIMARY KEY,
    username TEXT,
    useremail TEXT UNIQUE,
    userpassword TEXT
);