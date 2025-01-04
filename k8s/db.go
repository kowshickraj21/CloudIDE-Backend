package k8s

import "database/sql"

type Stash struct {
	Name string `json:"name"`
	Image string `json:"image"`
	Owner string `json:"owner"`
	Port int32 `json:"port"`
}

func CreateStash(db *sql.DB, stash Stash) (sql.Result,error) {
	query := `INSERT INTO Stashes (name, image, owner, port) VALUES ($1, $2, $3, $4)`
	res,err := db.Exec(query, stash.Name, stash.Image, stash.Owner, stash.Port)
	if err != nil {
		return nil,err
	}
	return res,nil
}

func GetStashes(db *sql.DB, owner string) ([]Stash, error) {
	query := `SELECT * FROM Stashes WHERE owner = $1`
	rows,err := db.Query(query,owner)
	if err != nil {
		return nil,err;
	}
	defer rows.Close()

	var stashes []Stash
	for rows.Next() {
		var stash Stash
		err := rows.Scan(&stash.Name, &stash.Image, &stash.Owner, &stash.Port)
		if err != nil {
			return nil, err
		}

		stashes = append(stashes, stash)
	}
	return stashes, nil
}

func FindStash(db *sql.DB, name string) Stash{
	query := `SELECT * FROM Stashes WHERE name = $1`
	row := db.QueryRow(query,name)
	var stash Stash
	_ = row.Scan(&stash.Name, &stash.Image, &stash.Owner, &stash.Port)
	return stash
}
