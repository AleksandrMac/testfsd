package datastores

import (
	"context"
	"errors"
	"fmt"

	"github.com/AleksandrMac/testfsd/internal/entities"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type commonDatastorePGSql struct {
	db *pgxpool.Pool
}

// NewcommonDatastorePGSql get operation external api
func NewCommonDatastorePGSql(db *pgxpool.Pool) CommonDatastore {
	return &commonDatastorePGSql{db}
}

func (x *commonDatastorePGSql) GetMessages(ctx context.Context, roomId int64, afterMessageId uint64) ([]entities.Message, error) {
	f, fa := func() (s string, args []interface{}) {
		if afterMessageId == 0 {
			return `from (select * from message m where room_id = $1 ) m`, []interface{}{roomId}
		}
		return `from (select * from message m where room_id = $1
	AND created_at < (select created_at from message where message_id = $2)) m`, []interface{}{roomId, afterMessageId}
	}()
	w, wa := func() (s string, args []interface{}) {
		if afterMessageId == 0 {
			return `WHERE m.created_at > NOW() - concat(r.last_message_duration_in_minutes, ' minutes')::interval`, nil
		}
		return fmt.Sprintf(`WHERE m.created_at > (select created_at from message where message_id = $3)) - concat(r.last_message_duration_in_minutes, ' minutes')::interval`), []interface{}{afterMessageId}
	}()
	fa = append(fa, wa...)

	sql := fmt.Sprintf(`
select m.message_id, m.text, u.user_id, u."name", m.created_at, m.updated_at, m.deleted_at
%s
left join room r on m.room_id = r.room_id
left join "user" u on m.user_id = u.user_id
%s
ORDER BY m.created_at desc`, f, w)
	rows, err := x.db.Query(ctx, sql, fa...)
	if err != nil {
		return nil, fmt.Errorf("failed get messages: %w", err)
	}

	var msgs []entities.Message
	var msg *entities.Message
	for rows.Next() {
		msgs = append(msgs, entities.Message{})
		msg = &msgs[len(msgs)-1]
		if err = rows.Scan(&msg.ID, &msg.Text, &msg.Author.Id, &msg.Author.Name, &msg.CreatedAt, &msg.UpdatedAt, &msg.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed scan message: %w", err)
		}
	}
	return msgs, nil
}

func (x *commonDatastorePGSql) GetRooms(ctx context.Context, userId int64) ([]entities.Room, error) {
	sql := `
SELECT r.room_id, r."name", r.created_at, r.updated_at, r.deleted_at
FROM room r 
LEFT JOIN room_user ru ON r.room_id = ru.room_id 
LEFT JOIN "user" u ON ru.user_id = u.user_id
WHERE u.user_id IS NULL OR u.user_id = $1`

	rows, err := x.db.Query(ctx, sql, userId)
	if err != nil {
		return nil, fmt.Errorf("failed get rooms: %w", err)
	}

	var rooms []entities.Room
	var room *entities.Room
	for rows.Next() {
		rooms = append(rooms, entities.Room{})
		room = &rooms[len(rooms)-1]
		if err = rows.Scan(
			&room.ID,
			&room.Name,
			&room.CreatedAt,
			&room.UpdatedAt,
			&room.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed scan result into room: %w", err)
		}
	}

	return rooms, nil
}

func (x *commonDatastorePGSql) WriteMessage(ctx context.Context, msg *entities.Message) (insertedId int64, err error) {
	// `INSERT INTO message (text, user_id, room_id) VALUES($1,$2,$3)  RETURNING message_id;`,,
	sql := `
INSERT INTO message (text, user_id, room_id)
	SELECT $1, $2, $3
	FROM room_user ru
	WHERE ru.user_id = $2 AND ru.room_id = $3
		OR (SELECT count(room_id) = 0 FROM room_user WHERE room_id = $3)
	LIMIT 1
RETURNING message_id;`
	if msg == nil {
		return 0, fmt.Errorf("message is nil")
	}
	if err = x.db.QueryRow(ctx, sql, msg.Text, msg.Author.Id, msg.RoomId).Scan(&insertedId); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			err = fmt.Errorf("room not available")
		default:
			err = fmt.Errorf("failed write message: %w", err)
		}
		return
	}
	return
}
