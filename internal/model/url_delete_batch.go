package model

type URLToDelete struct {
	UserUUID string
	ShortID  string
}

type URLDeleteBatch []URLToDelete
