package functions

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

func GetConfig() *AppConfig {
	config := &AppConfig{
		DataBucket:        os.Getenv("DATA_BUCKET"),
		CommentsFile:      os.Getenv("COMMENTS_FILE"),
		NewCommentsPrefix: os.Getenv("NEW_PREFIX"),
		AdminKey:          os.Getenv("ADMIN_KEY"),
	}
	if len(config.DataBucket) == 0 {
		panic("config not set") // TODO
	}
	return config
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func readCloudObject(ctx context.Context, obj *storage.ObjectHandle) ([]byte, error) {
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("readCloudObject NewReader failed: %w", err)
	}
	defer r.Close()

	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("readCloudObject ReadAll failed: %w", err)
	}
	return bytes, nil
}

func getTimestampForFilename(time time.Time) string {
	return time.Format("20060102_150405")
}
