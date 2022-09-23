package main

import (
	"context"
	"fmt"
)

func Insert(str string) error {

	// fmt.Println(largeMapNode.Get("102343214"))
	// fmt.Printf("%#v", Pool.Config().MinConns)

	// tx, err := Pool.Begin(context.Background())
	// if err != nil {
	// 	return fmt.Errorf("begin: %d,\n %s ", err, str)
	// 	// return fmt.Errorf("begin: %d,\n %s,\n\n %s", err, str[:1000], str[len(str)-500:])
	// }
	_, err := Pool.Exec(context.Background(), str)
	if err != nil {
		// return fmt.Errorf("exec: %d,\n %s ", err, str)
		return fmt.Errorf("exec:, %d,\n %s,\n\n %s", err, str[:1000], str[len(str)-500:])
	}
	// err = tx.Commit(context.Background())
	// if err != nil {
	// 	return fmt.Errorf("commit: %d,\n %s ", err, str)
	// 	// return fmt.Errorf("commit:, %d,\n %s,\n\n %s", err, str[:1000], str[len(str)-500:])
	// }
	return nil
}
