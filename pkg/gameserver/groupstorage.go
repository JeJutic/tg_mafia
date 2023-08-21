package gameserver

import (
	"database/sql"
	"errors"
	"fmt"
)

type groupStorage interface {
	createGroup(creator int64, name string) error                  // error if name is invalid
	joinGroup(int64, string) error                                 // error if group doesn't exist
	getGroupMembers(userAsked int64, name string) ([]int64, error) // error if userAsked isn't group member
}

type groupsDb struct {
	driverName string
	dbUrl      string
}

func (g groupsDb) open() (*sql.DB, error) {
	return sql.Open(g.driverName, g.dbUrl)
}

func wrapUnableToConnect(err error) error {
	return fmt.Errorf("unable to connect to database: %w", err)
}

const createGroupQuery = `
INSERT INTO groups (name) VALUES ($1)
`

func (g *groupsDb) createGroup(creator int64, group string) error {
	db, err := g.open()
	if err != nil {
		return wrapUnableToConnect(err)
	}
	defer db.Close()

	rows, err := db.Query(createGroupQuery, group)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}

const alreadyInGroupQuery = `
SELECT * FROM
groups INNER JOIN group_user USING(group_id)
WHERE groups.name=$1 AND user_id=$2
`

func isUserInGroup(db *sql.DB, user int64, group string) (bool, error) {
	rows, err := db.Query(alreadyInGroupQuery, group, user)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

const groupExistsQuery = `
SELECT * FROM
groups
WHERE groups.name=$1
`

func groupExists(db *sql.DB, group string) (bool, error) {
	rows, err := db.Query(groupExistsQuery, group)
	if err != nil {
		return false, err
	}

	return rows.Next(), nil
}

const joinGroupQuery = `
INSERT INTO group_user (group_id, user_id)
SELECT group_id, $2
FROM groups
WHERE groups.name=$1
`

func (g *groupsDb) joinGroup(user int64, group string) error {
	db, err := g.open()
	if err != nil {
		return wrapUnableToConnect(err)
	}
	defer db.Close()

	rows, err := db.Query(alreadyInGroupQuery, group, user)
	if err != nil {
		return err
	}
	defer rows.Close()

	inGroup, err := isUserInGroup(db, user, group)
	if err != nil {
		return err
	} else if inGroup {
		return errors.New("user is already in this group")
	}

	exists, err := groupExists(db, group)
	if err != nil {
		return err
	} else if !exists {
		return errors.New("the group doesn't exist")
	}

	rows, err = db.Query(joinGroupQuery, group, user)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}

const groupMembersQuery = `
SELECT user_id 
FROM
groups INNER JOIN group_user USING(group_id)
WHERE groups.name=$1
`

func (g *groupsDb) getGroupMembers(user int64, group string) ([]int64, error) {
	db, err := g.open()
	if err != nil {
		return nil, wrapUnableToConnect(err)
	}
	defer db.Close()

	inGroup, err := isUserInGroup(db, user, group)
	if err != nil {
		return nil, err
	} else if !inGroup {
		return nil, errors.New("user not in the group, permission restricted")
	}

	rows, err := db.Query(groupMembersQuery, group)
	if err != nil {
		return nil, err
	}

	users := make([]int64, 0)

	for i := 0; rows.Next(); i++ {
		var user int64
		err = rows.Scan(&user)

		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
