package repo

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
	"github.com/samber/mo"

	"github.com/btschwartz12/forum/repo/db"
)

const (
	maxPostUploadMb   = 50
	maxPostUploadSize = maxPostUploadMb << 20
)

var (
	allowedExtensionsRe = regexp.MustCompile(`\.(jpe?g|png|gif)$`)

	ErrStorageFull      = fmt.Errorf("storage full")
	ErrInvalidExtension = fmt.Errorf("invalid file extension")
	ErrFileTooLarge     = fmt.Errorf("file too large (max %d MB)", maxPostUploadMb)
	ErrPostNotFound     = fmt.Errorf("post not found")
)

func (r *Repo) InsertPost(
	ctx context.Context,
	post *Post,
	file mo.Option[multipart.File],
	header mo.Option[*multipart.FileHeader],
) (int64, error) {
	if file.IsPresent() && header.IsPresent() {
		if r.storageFull() {
			return 0, ErrStorageFull
		}
		if header.MustGet().Size > maxPostUploadSize {
			return 0, ErrFileTooLarge
		}
		ext := filepath.Ext(header.MustGet().Filename)
		if !allowedExtensionsRe.MatchString(ext) {
			return 0, ErrInvalidExtension
		}
		newName := uuid.New().String() + ext
		newPath := filepath.Join(r.varDir, postUploadDir, newName)
		newFile, err := os.Create(newPath)
		if err != nil {
			return 0, fmt.Errorf("error creating file: %w", err)
		}
		defer newFile.Close()
		if _, err := io.Copy(newFile, file.MustGet()); err != nil {
			return 0, fmt.Errorf("error copying file: %w", err)
		}
		post.ImageFilename = newName
	}
	q := db.New(r.db)
	postId, err := q.InsertPost(ctx, post.toDb())
	if err != nil {
		return 0, fmt.Errorf("error inserting post: %w", err)
	}
	return postId, nil
}

func (r *Repo) GetAllPosts(ctx context.Context) ([]Post, error) {
	q := db.New(r.db)
	rows, err := q.GetAllPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting posts: %w", err)
	}

	posts := make([]Post, len(rows))
	for i, row := range rows {
		posts[i] = Post{}
		posts[i].fromDb(&row)
	}

	return posts, nil
}

func (r *Repo) GetPathForPost(filename string) string {
	return filepath.Join(r.varDir, postUploadDir, filename)
}

func (r *Repo) DeletePost(ctx context.Context, postId int64) error {
	q := db.New(r.db)
	imageFilename, err := q.GetImageFilenameForPost(ctx, postId)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("error getting image filename for post: %w", err)
		}
	} else {
		if imageFilename.Valid {
			if err := os.Remove(r.GetPathForPost(imageFilename.String)); err != nil {
				return fmt.Errorf("error deleting image file: %w", err)
			}
		}
	}
	if err := q.DeletePost(ctx, postId); err != nil {
		return fmt.Errorf("error deleting post: %w", err)
	}
	return nil
}

func (r *Repo) UpdatePostContent(ctx context.Context, postId int64, content, ip string) error {
	q := db.New(r.db)
	err := q.UpdatePostContent(ctx, db.UpdatePostContentParams{
		ID:      postId,
		Content: content,
		Ip:      ip,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPostNotFound
		}
		return fmt.Errorf("error updating post content: %w", err)
	}
	return nil
}

func (r *Repo) GetPostAuthor(ctx context.Context, postId int64) (*User, error) {
	q := db.New(r.db)
	row, err := q.GetPostAuthor(ctx, postId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("error getting post author: %w", err)
	}
	user := User{}
	user.fromDb(&row)
	return &user, nil
}
