package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func GetNewCommentsHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		fmt.Fprintln(os.Stderr, "GetNewCommentsHandler error: method not allowed")
		return
	}
	if !ValidateKey(w, r) {
		return
	}

	newCommentFiles, err := getNewComments(r.Context())
	if err != nil {
		http.Error(w, "failed retrieving comments", 500)
		fmt.Fprintln(os.Stderr, "GetNewCommentsHandler error: failed retrieving comments: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8") // TODO
	json.NewEncoder(w).Encode(newCommentFiles)
	fmt.Println("GetNewCommentsHandler returned %d comment files", len(newCommentFiles))
}

func ApproveNewCommentHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		fmt.Fprintln(os.Stderr, "ApproveNewCommentHandler error: method not allowed")
		return
	}
	if !ValidateKey(w, r) {
		return
	}

	filenames, ok := r.URL.Query()["file"]
	if !ok || len(filenames[0]) < 1 {
		http.Error(w, "must specify file parameter", 404)
		fmt.Fprintln(os.Stderr, "ApproveNewCommentHandler error: must specify file parameter")
		return
	}
	if err := doApproval(r.Context(), filenames[0]); err != nil {
		http.Error(w, err.Error(), 500)
		fmt.Fprintln(os.Stderr, "ApproveNewCommentHandler error: %v", err.Error())
		return
	}
	fmt.Println("ApproveNewCommentHandler succeeded")
}

func ValidateKey(w http.ResponseWriter, r *http.Request) bool {
	keys, ok := r.URL.Query()["key"]

	if !ok || keys[0] != GetConfig().AdminKey {
		http.Error(w, "invalid key", 401)
		fmt.Fprintln(os.Stderr, "ValidateKey: invalid key")
		return false
	}
	return true
}

func getNewComments(ctx context.Context) ([]NewCommentFile, error) {
	newCommentFiles := make([]NewCommentFile, 0)

	config := GetConfig()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getNewComments storage.NewClient failed: %w", err)
	}
	defer client.Close()
	bkt := client.Bucket(config.DataBucket)

	query := storage.Query{
		Prefix: config.NewCommentsPrefix,
	}
	it := bkt.Objects(ctx, &query)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("getNewComments it.Next failed: %w", err)
		}

		if attrs.Size == 0 {
			continue
		}
		obj := bkt.Object(attrs.Name)
		bytes, err := readCloudObject(ctx, obj)
		if err != nil {
			return nil, err
		}
		var comment Comment
		err = json.Unmarshal(bytes, &comment)
		newCommentFile := NewCommentFile{
			Filename: strings.TrimPrefix(attrs.Name, config.NewCommentsPrefix),
			Comment:  comment,
		}
		newCommentFiles = append(newCommentFiles, newCommentFile)
	}
	return newCommentFiles, nil
}

func doApproval(ctx context.Context, filename string) error {
	config := GetConfig()
	delim := []byte{0x0A}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("doApproval storage.NewClient failed: %w", err)
	}
	defer client.Close()
	bkt := client.Bucket(config.DataBucket)

	obj1 := bkt.Object(config.CommentsFile)

	commentsBytes, err := readCloudObject(ctx, obj1)
	err = backupComments(ctx, bkt, obj1)
	if err != nil {
		return err
	}

	obj2 := bkt.Object(config.NewCommentsPrefix + filename + ".json")
	newBytes, err := readCloudObject(ctx, obj2)
	if err != nil {
		return err
	}
	err = archiveNewComment(ctx, bkt, obj2)

	w := obj1.NewWriter(ctx)
	defer w.Close()

	_, err = w.Write(commentsBytes)
	_, err = w.Write(newBytes)
	_, err = w.Write(delim)

	return nil
}

func backupComments(ctx context.Context, bkt *storage.BucketHandle, obj *storage.ObjectHandle) error {
	dstName := "archive/comments_" + getTimestampForFilename(time.Now().UTC()) + ".dat"
	dstObj := bkt.Object(dstName)

	if _, err := dstObj.CopierFrom(obj).Run(ctx); err != nil {
		return fmt.Errorf("backupComments CopierFrom failed: %w", err)
	}
	return nil
}

func archiveNewComment(ctx context.Context, bkt *storage.BucketHandle, srcObj *storage.ObjectHandle) error {
	dstName := strings.Replace(srcObj.ObjectName(), "new/", "archive/", 1)
	dstObj := bkt.Object(dstName)

	if _, err := dstObj.CopierFrom(srcObj).Run(ctx); err != nil {
		return fmt.Errorf("CopierFrom failed: %v", err)
	}

	if err := srcObj.Delete(ctx); err != nil {
		return fmt.Errorf("Delete failed: %v", err)
	}

	return nil
}
