package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
)

func execSQL(c *hazelcast.Client, text string, w io.Writer) (*sql.Rows, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	lt := strings.ToLower(text)
	if strings.HasPrefix(lt, "select") || strings.HasPrefix(lt, "show") {
		return query(c, text, w)
	}
	return nil, exec(c, text, w)
}

func query(c *hazelcast.Client, text string, w io.Writer) (*sql.Rows, error) {
	rows, err := c.QuerySQL(context.Background(), text)
	if err != nil {
		return rows, fmt.Errorf("querying: %w", err)
	}
	return rows, nil
	//defer rows.Close()
	//cols, err := rows.Columns()
	//if err != nil {
	//	return fmt.Errorf("retrieving columns: %w", err)
	//}
	//fmt.Fprintln(w, strings.Join(cols, "    "))
	////fmt.Println(w,"---")
	//row := make([]interface{}, len(cols))
	//for i := 0; i < len(cols); i++ {
	//	row[i] = new(interface{})
	//}
	//rowStr := make([]string, len(cols))
	//for rows.Next() {
	//	if err := rows.Scan(row...); err != nil {
	//		return fmt.Errorf("scanning row: %w", err)
	//	}
	//	for i, v := range row {
	//		val := *(v.(*interface{}))
	//		rowStr[i] = fmt.Sprintf("%v", val)
	//	}
	//	fmt.Fprintln(w, strings.Join(rowStr, "    "))
	//}
	//return nil
}

func exec(db *hazelcast.Client, text string, w io.Writer) error {
	r, err := db.ExecSQL(context.Background(), text)
	if err != nil {
		return fmt.Errorf("executing: %w", err)
	}
	ra, err := r.RowsAffected()
	if err != nil {
		return err
	}
	return fmt.Errorf("---\nAffected rows: %d\n\n", ra)
	//fmt.Fprintf(w, "---\nAffected rows: %d\n\n", ra)
	//return nil
}

func fatal(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, text)
	os.Exit(1)
}
