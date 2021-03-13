package database

import "database/sql"

type EventDatabase struct {
	storage *sql.DB
}





func NewAlbumRepositoryRealisation(db *sql.DB) AlbumRepositoryRealisation {
	return AlbumRepositoryRealisation{storage: db}

}
