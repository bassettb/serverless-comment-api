package functions

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

func BaseHandler(w http.ResponseWriter, r *http.Request) {
	/*
		var sb strings.Builder
		sb.WriteString("URI: " + string(r.RequestURI) + "\n")
		sb.WriteString("Host: " + string(r.Host) + "\n")
		sb.WriteString("URL.Path: " + string(r.URL.Path) + "\n")
		for name, values := range r.Header {
			for _, value := range values {
				sb.WriteString(name + ": " + value + "\n")
			}
		}
		fmt.Println(sb.String())
	*/
	paths, ok := r.Header["X-Envoy-Original-Path"]
	if !ok || len(paths) == 0 || len(paths[0]) == 0 {
		http.Error(w, "no envoy header", 404)
		fmt.Fprintln(os.Stderr, "no envoy header")
	}
	path := strings.Split(paths[0], "?")[0]

	if path == "/comment" {
		CommentHandler(w, r)
	} else if path == "/admin/new" {
		GetNewCommentsHandler(w, r)
	} else if path == "/admin/approve" {
		ApproveNewCommentHandler(w, r)
	} else {
		http.Error(w, "bad request", 404)
		fmt.Fprintln(os.Stderr, "bad url "+path)
	}
}

func CommentHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "GET" {
		GetComments(w, r)
	} else if r.Method == "POST" {
		PostComment(w, r)
	} else {
		http.Error(w, "method not allowed", 405)
	}
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	comments, err := loadComments(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		fmt.Fprintf(os.Stderr, "GetComments error: %v\n", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(comments)
	fmt.Printf("GetComments returned %d comments\n", len(comments))
}

func PostComment(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", 400)
		fmt.Fprintf(os.Stderr, "PostComment error: %v\n", err.Error())
		return
	}
	var newComment NewComment
	err = json.Unmarshal(bytes, &newComment)
	if err != nil {
		http.Error(w, "invalid payload", 400)
		fmt.Fprintf(os.Stderr, "PostComment invalid payload: %v\n", err.Error())
		return
	}

	err = addNewComment(r.Context(), newComment)
	if err != nil {
		http.Error(w, "command failed", 500)
		fmt.Fprintf(os.Stderr, "PostComment failed: %v\n", err.Error())
		return
	}
	fmt.Println("PostComment succeeded")
}

func loadComments(ctx context.Context) ([]Comment, error) {
	config := GetConfig()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("loadComments NewClient failed: %w", err)
	}
	defer client.Close()
	bkt := client.Bucket(config.DataBucket)

	obj := bkt.Object(config.CommentsFile)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("loadComments NewReader failed: %w", err)
	}
	defer r.Close()

	return parseComments(r)
}

func parseComments(reader io.Reader) ([]Comment, error) {
	var comments []Comment
	//scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner := bufio.NewScanner(reader)
	var id int64 = 1
	for scanner.Scan() {
		text := scanner.Text()
		//fmt.Println(text)
		var comment Comment
		err := json.Unmarshal([]byte(text), &comment)
		if err != nil {
			return nil, fmt.Errorf("parseComments Unmarshal failed: %w", err)
		}
		comment.Id = id
		comments = append(comments, comment)
		id++
	}
	return comments, nil
}

func addNewComment(ctx context.Context, newComment NewComment) error {
	config := GetConfig()
	now := time.Now().UTC()
	filename := getTimestampForFilename(now)

	comment := Comment{
		Name:      newComment.Name,
		Email:     newComment.Email,
		Msg:       newComment.Msg,
		Timestamp: now,
	}
	bytes, err := json.Marshal(&comment)
	if err != nil {
		return fmt.Errorf("addNewComment Marshal failed: %w", err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("addNewComment storage.NewClient failed: %w", err)
	}
	defer client.Close()
	bkt := client.Bucket(config.DataBucket)

	obj := bkt.Object(config.NewCommentsPrefix + filename + ".json")
	w := obj.NewWriter(ctx)
	defer w.Close()

	_, err = w.Write(bytes)
	if err != nil {
		return fmt.Errorf("addNewComment Write failed: %w", err)
	}
	return nil
}
