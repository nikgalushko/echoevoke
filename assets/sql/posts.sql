create table posts (
    id integer primary key,
    channel_id text not null,
    date integer not null,
    message text not null
);


create table post_images (
    post_id integer not null,
    image_id integer not null,
    foreign key (post_id) references posts(id),
    foreign key (image_id) references images(id),
);