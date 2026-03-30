package cloud_storage

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/bucket"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/object"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/object_list"
)

func testServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return NewClientWithBaseUrl(u)
}

func TestInsertBucket(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/storage/v1/b") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("project") != "my-project" {
			t.Errorf("unexpected project: %s", r.URL.Query().Get("project"))
		}

		var input bucket.Bucket
		json.NewDecoder(r.Body).Decode(&input)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&bucket.Bucket{
			Kind:         "storage#bucket",
			Name:         input.Name,
			Location:     "US",
			StorageClass: "STANDARD",
		})
	})

	b, err := client.InsertBucket(context.Background(), "my-project", &bucket.Bucket{Name: "test-bucket"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Name != "test-bucket" {
		t.Errorf("expected bucket name 'test-bucket', got %q", b.Name)
	}
	if b.Kind != "storage#bucket" {
		t.Errorf("expected kind 'storage#bucket', got %q", b.Kind)
	}
}

func TestInsertBucket_EmptyProject(t *testing.T) {
	client := NewClient()
	_, err := client.InsertBucket(context.Background(), "", &bucket.Bucket{Name: "b"})
	if err == nil {
		t.Fatal("expected error for empty project")
	}
}

func TestInsertBucket_NilConfig(t *testing.T) {
	client := NewClient()
	b, err := client.InsertBucket(context.Background(), "project", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b != nil {
		t.Error("expected nil for nil config")
	}
}

func TestInsertBucket_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.InsertBucket(ctx, "project", &bucket.Bucket{Name: "b"})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestPatchBucket(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/b/my-bucket") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&bucket.Bucket{
			Name:         "my-bucket",
			StorageClass: "NEARLINE",
		})
	})

	b, err := client.PatchBucket(context.Background(), "my-bucket", &bucket.Bucket{StorageClass: "NEARLINE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.StorageClass != "NEARLINE" {
		t.Errorf("expected storage class 'NEARLINE', got %q", b.StorageClass)
	}
}

func TestPatchBucket_EmptyName(t *testing.T) {
	client := NewClient()
	_, err := client.PatchBucket(context.Background(), "", &bucket.Bucket{})
	if err == nil {
		t.Fatal("expected error for empty bucket name")
	}
}

func TestPatchBucket_NilConfig(t *testing.T) {
	client := NewClient()
	b, err := client.PatchBucket(context.Background(), "bucket", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b != nil {
		t.Error("expected nil for nil config")
	}
}

func TestGetObject(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/b/my-bucket/o/my-object.txt") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&object.Object{
			Kind:         "storage#object",
			Name:         "my-object.txt",
			Bucket:       "my-bucket",
			Size:         "1024",
			StorageClass: "STANDARD",
		})
	})

	obj, err := client.GetObject(context.Background(), "my-bucket", "my-object.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Name != "my-object.txt" {
		t.Errorf("expected object name 'my-object.txt', got %q", obj.Name)
	}
	if obj.Bucket != "my-bucket" {
		t.Errorf("expected bucket 'my-bucket', got %q", obj.Bucket)
	}
}

func TestGetObject_EmptyBucketName(t *testing.T) {
	client := NewClient()
	_, err := client.GetObject(context.Background(), "", "obj")
	if err == nil {
		t.Fatal("expected error for empty bucket name")
	}
}

func TestGetObject_EmptyObjectName(t *testing.T) {
	client := NewClient()
	_, err := client.GetObject(context.Background(), "bucket", "")
	if err == nil {
		t.Fatal("expected error for empty object name")
	}
}

func TestDownloadObject(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("alt") != "media" {
			t.Errorf("expected alt=media, got %q", r.URL.Query().Get("alt"))
		}
		w.Write([]byte("file content here"))
	})

	data, err := client.DownloadObject(context.Background(), "my-bucket", "my-file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "file content here" {
		t.Errorf("expected 'file content here', got %q", string(data))
	}
}

func TestDownloadObject_EmptyBucketName(t *testing.T) {
	client := NewClient()
	_, err := client.DownloadObject(context.Background(), "", "obj")
	if err == nil {
		t.Fatal("expected error for empty bucket name")
	}
}

func TestDownloadObject_EmptyObjectName(t *testing.T) {
	client := NewClient()
	_, err := client.DownloadObject(context.Background(), "bucket", "")
	if err == nil {
		t.Fatal("expected error for empty object name")
	}
}

func TestListObjects(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("prefix") != "logs/" {
			t.Errorf("unexpected prefix: %s", r.URL.Query().Get("prefix"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&object_list.ObjectList{
			Kind: "storage#objects",
			Items: []*object.Object{
				{Name: "logs/2024-01.txt"},
				{Name: "logs/2024-02.txt"},
			},
		})
	})

	list, err := client.ListObjects(context.Background(), "my-bucket", url.Values{"prefix": {"logs/"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list.Items))
	}
	if list.Items[0].Name != "logs/2024-01.txt" {
		t.Errorf("unexpected first item: %q", list.Items[0].Name)
	}
}

func TestListObjects_EmptyBucketName(t *testing.T) {
	client := NewClient()
	_, err := client.ListObjects(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty bucket name")
	}
}

func TestListObjects_NilQuery(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query, got %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&object_list.ObjectList{Kind: "storage#objects"})
	})

	_, err := client.ListObjects(context.Background(), "bucket", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteObject(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/b/my-bucket/o/my-object.txt") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteObject(context.Background(), "my-bucket", "my-object.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteObject_EmptyBucketName(t *testing.T) {
	client := NewClient()
	err := client.DeleteObject(context.Background(), "", "obj")
	if err == nil {
		t.Fatal("expected error for empty bucket name")
	}
}

func TestDeleteObject_EmptyObjectName(t *testing.T) {
	client := NewClient()
	err := client.DeleteObject(context.Background(), "bucket", "")
	if err == nil {
		t.Fatal("expected error for empty object name")
	}
}

func TestDeleteObject_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := client.DeleteObject(ctx, "bucket", "obj")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestInsertObject(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/upload/storage/v1/b/my-bucket/o") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("uploadType") != "multipart" {
			t.Errorf("unexpected uploadType: %s", r.URL.Query().Get("uploadType"))
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/related") {
			t.Errorf("unexpected content-type: %s", contentType)
		}

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "hello world") {
			t.Error("request body missing object data")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&object.Object{
			Kind:   "storage#object",
			Name:   "test.txt",
			Bucket: "my-bucket",
			Size:   "11",
		})
	})

	obj, err := client.InsertObject(
		context.Background(),
		"my-bucket",
		&object.Object{Name: "test.txt"},
		[]byte("hello world"),
		"text/plain",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Name != "test.txt" {
		t.Errorf("expected object name 'test.txt', got %q", obj.Name)
	}
}

func TestInsertObject_EmptyBucketName(t *testing.T) {
	client := NewClient()
	_, err := client.InsertObject(context.Background(), "", &object.Object{Name: "o"}, []byte("x"), "text/plain")
	if err == nil {
		t.Fatal("expected error for empty bucket name")
	}
}

func TestInsertObject_EmptyContentType(t *testing.T) {
	client := NewClient()
	_, err := client.InsertObject(context.Background(), "bucket", &object.Object{Name: "o"}, []byte("x"), "")
	if err == nil {
		t.Fatal("expected error for empty content type")
	}
}

func TestInsertObject_NilMetadata(t *testing.T) {
	client := NewClient()
	obj, err := client.InsertObject(context.Background(), "bucket", nil, []byte("x"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj != nil {
		t.Error("expected nil for nil metadata")
	}
}

func TestInsertObject_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.InsertObject(ctx, "bucket", &object.Object{Name: "o"}, []byte("x"), "text/plain")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
