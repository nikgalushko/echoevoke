create table images (
    id integer primary key autoincrement,
    etag text,
    data blob,
    unique (etag)
);