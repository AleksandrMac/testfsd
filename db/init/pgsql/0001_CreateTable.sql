CREATE TABLE room (
	room_id SERIAL PRIMARY KEY,
	name varchar(100),
	last_message_duration_in_minutes int not null default 10,
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NULL DEFAULT NULL,
	deleted_at timestamp null default null);

CREATE TABLE "user" (
	user_id SERIAL PRIMARY KEY,
	name varchar(100),
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NULL DEFAULT NULL,
	deleted_at timestamp null default null);

CREATE TABLE room_user (
	room_id integer not null,
	user_id integer not null,
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NULL DEFAULT NULL,
	deleted_at timestamp null default null,
	UNIQUE(room_id, user_id),
	CONSTRAINT room_user_to_room_room_id_fkey FOREIGN KEY (room_id)
		REFERENCES room (room_id) MATCH SIMPLE
		ON UPDATE RESTRICT
		ON DELETE restrict,
	CONSTRAINT room_user_to_user_user_id_fkey FOREIGN KEY (user_id)
		REFERENCES "user" (user_id) MATCH SIMPLE
		ON UPDATE RESTRICT
		ON DELETE restrict);

CREATE INDEX room_user_room_indx ON room_user (room_id);
CREATE INDEX room_user_user_indx ON room_user (user_id);

CREATE TABLE message (
	message_id SERIAL primary key,
	text varchar(1000) not null,
	user_id integer not null,
	room_id integer not null,	
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NULL DEFAULT NULL,
	deleted_at timestamp null default null,
	CONSTRAINT message_to_room_room_id_fkey FOREIGN KEY (room_id)
		REFERENCES room (room_id) MATCH SIMPLE
		ON UPDATE RESTRICT
		ON DELETE restrict,
	CONSTRAINT message_to_user_user_id_fkey FOREIGN KEY (user_id)
		REFERENCES "user" (user_id) MATCH SIMPLE
		ON UPDATE RESTRICT
		ON DELETE restrict)


CREATE INDEX message_room_indx ON message (room_id);
CREATE INDEX message_user_indx ON message (user_id);