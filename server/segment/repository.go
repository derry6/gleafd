package segment

import (
	"context"
	"database/sql"
	"errors"
)

type Repository interface {
	List(ctx context.Context) ([]*Segment, error)
	Get(ctx context.Context, biztag string) (*Segment, error)
	UpdateMaxID(ctx context.Context, biztag string) (*Segment, error)
	UpdateMaxIDWithStep(ctx context.Context, biztag string, step int32) (*Segment, error)
	ListBizTags(ctx context.Context) ([]string, error)
}

type defaultRepository struct {
	db *sql.DB
}

func (r *defaultRepository) List(ctx context.Context) (segs []*Segment, err error) {
	q := "SELECT `biz_tag`,`max_id`,`step`,`updated` FROM segments"
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var seg Segment
		if err = rows.Scan(&seg.BizTag, &seg.MaxID, &seg.Step, &seg.Updated); err != nil {
			return nil, err
		}
		segs = append(segs, &seg)
	}
	return segs, nil
}

func (r *defaultRepository) getSegment(ctx context.Context, tx *sql.Tx, biztag string) (*Segment, error) {
	var seg Segment
	q := "SELECT `biz_tag`,`max_id`,`step`,`updated` FROM `segments` WHERE `biz_tag`=?"
	row := tx.QueryRowContext(ctx, q, biztag)
	if err := row.Scan(&seg.BizTag, &seg.MaxID, &seg.Step, &seg.Updated); err != nil {
		return nil, err
	}
	return &seg, nil
}

func (r *defaultRepository) Get(ctx context.Context, biztag string) (*Segment, error) {
	var seg Segment
	q := "SELECT `biz_tag`,`max_id`,`step`,`updated` FROM `segments` WHERE `biz_tag`=?"
	row := r.db.QueryRowContext(ctx, q, biztag)
	if err := row.Scan(&seg.BizTag, &seg.MaxID, &seg.Step, &seg.Updated); err != nil {
		return nil, err
	}
	return &seg, nil
}

func (r *defaultRepository) UpdateMaxID(ctx context.Context, biztag string) (*Segment, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	q := "UPDATE `segments` SET `max_id`=`max_id`+`step` WHERE `biz_tag`=?"
	rs, err := tx.ExecContext(ctx, q, biztag)
	if err != nil {
		return nil, err
	}
	n, err := rs.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("segment not found")
	}
	seg, err := r.getSegment(ctx, tx, biztag)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return seg, nil
}

func (r *defaultRepository) UpdateMaxIDWithStep(ctx context.Context, biztag string, step int32) (*Segment, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	q := `UPDATE segments SET max_id = max_id + ? WHERE biz_tag=?`
	rs, err := tx.ExecContext(ctx, q, step, biztag)
	if err != nil {
		return nil, err
	}
	n, err := rs.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("segment not found")
	}
	seg, err := r.getSegment(ctx, tx, biztag)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return seg, nil
}

func (r *defaultRepository) ListBizTags(ctx context.Context) (biztags []string, err error) {
	q := "SELECT `biz_tag` FROM segments"
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var biztag string
		if err = rows.Scan(&biztag); err != nil {
			return biztags, err
		}
		biztags = append(biztags, biztag)
	}
	return biztags, nil
}

func (r *defaultRepository) createTables() (err error) {
	_, err = r.db.Exec(
		"CREATE TABLE IF NOT EXISTS `segments`(" +
			"	`biz_tag` VARCHAR(128) NOT NULL DEFAULT ''," +
			"	`max_id` 	BIGINT(20) NOT NULL DEFAULT '1'," +
			"	`step` 	INT(11) NOT NULL," +
			"	`desc` 	VARCHAR(256)  DEFAULT NULL," +
			"	`updated` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
			"	PRIMARY KEY (`biz_tag`)" +
			");")
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		"INSERT INTO `segments`(`biz_tag`,`step`,`desc`) " +
			"VALUES('example', 1000, 'gleafd example')")
	// Ignore error
	return nil
}

func NewDefaultRepository(db *sql.DB) (Repository, error) {
	r := &defaultRepository{db: db}
	if err := r.createTables(); err != nil {
		return nil, err
	}
	return r, nil
}
