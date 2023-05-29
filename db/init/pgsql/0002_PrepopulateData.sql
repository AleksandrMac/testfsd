insert  into room ("name") values ('room1'),('room2'),('room3'),('room4');

insert  into "user" ("name") values ('user1'),('user2'),('user3'),('user4');

-- делаем 3ю комнату приватной для user3, user4
insert  into room_user (room_id, user_id) values(3,3),(3,4);