CREATE TABLE ticks (
    timestamp BIGINT  PRIMARY KEY,
    symbol varchar(7),
    bid FLOAT,
    ask FLOAT
);

