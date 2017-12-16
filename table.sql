create table users (
  id int auto_increment,
  username varchar(255) not null,
  password varchar(255) not null,
  primary key (id)
);

create table posts (
  id int auto_increment,
  user_id int not null,
  msg varchar(255) not null,
  created_at timestamp not null default now(),
  primary key (id),
  foreign key (user_id) references users (id)
);
