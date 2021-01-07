package main

import (
	"flag"
	"fmt"
	"net/http"

	"example.com/functions"

	"github.com/joho/godotenv"
)

func main() {

	//ctx := context.Background()

	xmlFile := flag.String("xml", "", "")
	flag.Parse()

	fmt.Println(*xmlFile)
	if len(*xmlFile) > 0 {
		convert(*xmlFile)
		return
	}

	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}

	// Validate config
	functions.GetConfig()

	mux := http.NewServeMux()

	//mux.Handle("/", http.HandlerFunc(functions.BaseHandler))
	mux.Handle("/comment", http.HandlerFunc(functions.CommentHandler))
	mux.Handle("/admin/new", http.HandlerFunc(functions.GetNewCommentsHandler))
	mux.Handle("/admin/approve", http.HandlerFunc(functions.ApproveNewCommentHandler))

	err := http.ListenAndServe(":5000", mux)
	if err != nil {
		panic("http Listen failed")
	}
}
