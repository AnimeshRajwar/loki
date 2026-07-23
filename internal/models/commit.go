package models

import (
	"fmt"
	"time"
)

type Commit struct {
	Tree    string
	Message string
	Author  string
	Email   string
}

func (c *Commit) Serialize() []byte {
	body := fmt.Sprintf(
		"tree %s\n"+
			"author %s <%s> %d\n\n"+
			"%s\n",
		c.Tree,
		c.Author,
		c.Email,
		time.Now().Unix(),
		c.Message,
	)

	header := fmt.Sprintf("commit %d\x00", len(body))
	return append([]byte(header), []byte(body)...)
}
