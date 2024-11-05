package k8s

import "database/sql"

type Stash struct {
	StashId string
	Image string
	Owner string  
}

func CreateStash(db *sql.DB, stash Stash) (sql.Result,error) {
	query := `INSERT INTO Users (name, email, picture, provider) VALUES ($1, $2, $3, $4)`
	res,err := db.Exec(query, stash.StashId, stash.Image, stash.Owner)
	if err != nil {
		return nil,err
	}
	return res,nil
}

func GetStashes(db *sql.DB, owner string) ([]Stash, error) {

	query := `SELECT * FROM Users WHERE owner = $1`
	rows,err := db.Query(query,owner)
	if err != nil {
		return nil,err;
	}
	defer rows.Close()

	var stashes []Stash
	for rows.Next() {
		var stash Stash
		err := rows.Scan(&stash.StashId, &stash.Image, &stash.Owner)
		if err != nil {
			return nil, err
		}

		stashes = append(stashes, stash)
	}
	return stashes, nil
}

