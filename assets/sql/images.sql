create table if not exists images (
    id integer primary key autoincrement,
    etag text,
    data blob,
    unique (etag)
);