package anyshare

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/OpenListTeam/OpenList/v4/drivers/base"
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	log "github.com/sirupsen/logrus"
)

type AnyShare struct {
	model.Storage
	Addition
}

func (d *AnyShare) Config() driver.Config {
	return config
}

func (d *AnyShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *AnyShare) Init(ctx context.Context) error {
	// Validate the connection by fetching entry doc libs
	var entries []EntryDocLib
	if err := d.get(ctx, "/efast/v1/entry-doc-lib", &entries); err != nil {
		return fmt.Errorf("failed to connect to AnyShare: %w", err)
	}
	log.Infof("AnyShare: connected, found %d entry doc libs", len(entries))
	return nil
}

func (d *AnyShare) Drop(ctx context.Context) error {
	return nil
}

func (d *AnyShare) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	id := dir.GetID()

	// If root ID is empty, list entry doc libs
	if id == "" {
		return d.listEntryDocLibs(ctx)
	}

	return d.listFolder(ctx, id)
}

func (d *AnyShare) listEntryDocLibs(ctx context.Context) ([]model.Obj, error) {
	var entries []EntryDocLib
	if err := d.get(ctx, "/efast/v1/entry-doc-lib", &entries); err != nil {
		return nil, fmt.Errorf("list entry doc libs: %w", err)
	}

	var objs []model.Obj
	for _, e := range entries {
		objs = append(objs, &model.Object{
			ID:       e.ID,
			Name:     e.Name,
			IsFolder: true,
			Modified: e.ModifiedAt,
			Ctime:    e.CreatedAt,
		})
	}
	return objs, nil
}

func (d *AnyShare) listFolder(ctx context.Context, folderID string) ([]model.Obj, error) {
	var objs []model.Obj
	marker := ""
	limit := 200

	for {
		apiPath := fmt.Sprintf("/efast/v1/folders/%s/sub_objects?limit=%d", url.PathEscape(folderID), limit)
		if marker != "" {
			apiPath += "&marker=" + url.QueryEscape(marker)
		}

		var resp ListResponse
		if err := d.get(ctx, apiPath, &resp); err != nil {
			return nil, fmt.Errorf("list folder %s: %w", folderID, err)
		}

		for _, dir := range resp.Dirs {
			objs = append(objs, &model.Object{
				ID:       dir.ID,
				Name:     dir.Name,
				Size:     dir.Size,
				IsFolder: true,
				Modified: dir.ModifiedAt,
				Ctime:    dir.CreatedAt,
			})
		}

		for _, file := range resp.Files {
			objs = append(objs, &model.Object{
				ID:       file.ID,
				Name:     file.Name,
				Size:     file.Size,
				IsFolder: false,
				Modified: file.ModifiedAt,
				Ctime:    file.CreatedAt,
			})
		}

		if resp.NextMarker == "" {
			break
		}
		marker = resp.NextMarker
	}

	return objs, nil
}

func (d *AnyShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp OsDownloadResp
	reqBody := OsDownloadReq{
		DocID: file.GetID(),
		Rev:   "", // empty to get latest revision
	}
	if err := d.post(ctx, "/efast/v1/file/osdownload", reqBody, &resp); err != nil {
		return nil, fmt.Errorf("get download link: %w", err)
	}

	if len(resp.AuthRequest) == 0 {
		return nil, fmt.Errorf("no download URL returned for file %s", file.GetName())
	}

	auth := resp.AuthRequest[0]
	header := http.Header{}
	for k, v := range auth.Headers {
		header.Set(k, v)
	}

	return &model.Link{
		URL:    auth.URL,
		Header: header,
	}, nil
}

func (d *AnyShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	req := DirCreateReq{
		DocID: parentDir.GetID(),
		Name:  dirName,
	}
	var resp DirCreateResp
	if err := d.post(ctx, "/efast/v1/dir/create", req, &resp); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	return &model.Object{
		ID:       resp.DocID,
		Name:     resp.Name,
		IsFolder: true,
		Modified: time.Unix(resp.Modified, 0),
		Ctime:    time.Unix(resp.CreateTime, 0),
	}, nil
}

func (d *AnyShare) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	var apiPath string
	if srcObj.IsDir() {
		apiPath = "/efast/v1/dir/move"
	} else {
		apiPath = "/efast/v1/file/move"
	}

	req := DirMoveReq{
		DocID:      srcObj.GetID(),
		DestParent: dstDir.GetID(),
	}
	if err := d.post(ctx, apiPath, req, nil); err != nil {
		return nil, fmt.Errorf("move %s: %w", srcObj.GetName(), err)
	}

	return nil, nil
}

func (d *AnyShare) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	var apiPath string
	if srcObj.IsDir() {
		apiPath = "/efast/v1/dir/rename"
	} else {
		apiPath = "/efast/v1/file/rename"
	}

	req := DirRenameReq{
		DocID: srcObj.GetID(),
		Name:  newName,
	}
	if err := d.post(ctx, apiPath, req, nil); err != nil {
		return nil, fmt.Errorf("rename %s: %w", srcObj.GetName(), err)
	}

	return nil, nil
}

func (d *AnyShare) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	var apiPath string
	if srcObj.IsDir() {
		apiPath = "/efast/v1/dir/copy"
	} else {
		apiPath = "/efast/v1/file/copy"
	}

	req := DirCopyReq{
		DocID:      srcObj.GetID(),
		DestParent: dstDir.GetID(),
	}
	if err := d.post(ctx, apiPath, req, nil); err != nil {
		return nil, fmt.Errorf("copy %s: %w", srcObj.GetName(), err)
	}

	return nil, nil
}

func (d *AnyShare) Remove(ctx context.Context, obj model.Obj) error {
	var apiPath string
	if obj.IsDir() {
		apiPath = "/efast/v1/dir/delete"
	} else {
		apiPath = "/efast/v1/file/delete"
	}

	req := DocIDReq{
		DocID: obj.GetID(),
	}
	if err := d.post(ctx, apiPath, req, nil); err != nil {
		return fmt.Errorf("remove %s: %w", obj.GetName(), err)
	}

	return nil
}

func (d *AnyShare) Put(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	fileSize := file.GetSize()

	// Use multipart upload for files > 50MB, simple upload otherwise
	const multipartThreshold = 50 * 1024 * 1024

	if fileSize > multipartThreshold {
		return d.multipartUpload(ctx, dstDir, file, up)
	}
	return d.simpleUpload(ctx, dstDir, file, up)
}

func (d *AnyShare) simpleUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	// Step 1: Begin upload
	beginReq := OsBeginUploadReq{
		DocID:  dstDir.GetID(),
		Length: file.GetSize(),
		Name:   file.GetName(),
		Ondup:  2, // auto-rename on conflict
	}
	var beginResp OsBeginUploadResp
	if err := d.post(ctx, "/efast/v1/file/osbeginupload", beginReq, &beginResp); err != nil {
		return nil, fmt.Errorf("begin upload: %w", err)
	}

	if len(beginResp.AuthRequest) == 0 {
		return nil, fmt.Errorf("no upload URL returned")
	}

	// Step 2: Upload data to object storage
	auth := beginResp.AuthRequest[0]
	if err := d.uploadToObjectStorage(ctx, auth, file, file.GetSize(), up); err != nil {
		return nil, fmt.Errorf("upload data: %w", err)
	}

	// Step 3: End upload
	endReq := OsEndUploadReq{
		DocID: beginResp.DocID,
		Rev:   beginResp.Rev,
	}
	var endResp OsEndUploadResp
	if err := d.post(ctx, "/efast/v1/file/osendupload", endReq, &endResp); err != nil {
		return nil, fmt.Errorf("end upload: %w", err)
	}

	return &model.Object{
		ID:       beginResp.DocID,
		Name:     endResp.Name,
		Size:     file.GetSize(),
		IsFolder: false,
		Modified: time.Unix(endResp.Modified, 0),
	}, nil
}

func (d *AnyShare) multipartUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	fileSize := file.GetSize()

	// Step 1: Get upload options for part size
	var optResp OsOptionResp
	if err := d.post(ctx, "/efast/v1/file/osoption", nil, &optResp); err != nil {
		log.Warnf("AnyShare: failed to get upload options, using defaults: %v", err)
		optResp.PartMinSize = 5 * 1024 * 1024  // 5MB
		optResp.PartMaxSize = 50 * 1024 * 1024 // 50MB
		optResp.PartMaxNum = 10000
	}

	// Calculate part size
	partSize := optResp.PartMinSize
	if partSize == 0 {
		partSize = 5 * 1024 * 1024 // 5MB default
	}
	numParts := (fileSize + partSize - 1) / partSize
	if numParts > int64(optResp.PartMaxNum) && optResp.PartMaxNum > 0 {
		partSize = (fileSize + int64(optResp.PartMaxNum) - 1) / int64(optResp.PartMaxNum)
		numParts = (fileSize + partSize - 1) / partSize
	}

	// Step 2: Init multipart upload
	initReq := OsInitMultiUploadReq{
		DocID:  dstDir.GetID(),
		Length: fileSize,
		Name:   file.GetName(),
		Ondup:  2,
	}
	var initResp OsInitMultiUploadResp
	if err := d.post(ctx, "/efast/v1/file/osinitmultiupload", initReq, &initResp); err != nil {
		return nil, fmt.Errorf("init multipart upload: %w", err)
	}

	// Step 3: Upload parts
	var completeParts []PartInfoComplete
	var uploaded int64

	for partNum := int64(1); partNum <= numParts; partNum++ {
		start := (partNum - 1) * partSize
		end := start + partSize
		if end > fileSize {
			end = fileSize
		}
		currentPartSize := end - start

		// Get auth for this part
		parts := []PartInfo{
			{
				PartNumber: int(partNum),
				Start:      start,
				End:        end - 1, // end is inclusive
			},
		}
		partReq := OsUploadPartReq{
			DocID:    initResp.DocID,
			Rev:      initResp.Rev,
			UploadID: initResp.UploadID,
			Parts:    parts,
		}
		var partResp OsUploadPartResp
		if err := d.post(ctx, "/efast/v1/file/osuploadpart", partReq, &partResp); err != nil {
			return nil, fmt.Errorf("get upload part auth (part %d): %w", partNum, err)
		}

		if len(partResp.AuthRequests) == 0 {
			return nil, fmt.Errorf("no upload auth returned for part %d", partNum)
		}

		// Read part data
		partData := make([]byte, currentPartSize)
		if _, err := io.ReadFull(file, partData); err != nil {
			return nil, fmt.Errorf("read part %d data: %w", partNum, err)
		}

		// Upload part
		partAuth := partResp.AuthRequests[0]
		if err := d.uploadToObjectStorage(ctx, partAuth, bytes.NewReader(partData), currentPartSize, nil); err != nil {
			return nil, fmt.Errorf("upload part %d: %w", partNum, err)
		}

		completeParts = append(completeParts, PartInfoComplete{
			PartNumber: int(partNum),
			Size:       currentPartSize,
		})

		uploaded += currentPartSize
		if up != nil {
			up(float64(uploaded) / float64(fileSize) * 100)
		}
	}

	// Step 4: Complete multipart upload
	completeReq := OsCompleteMultiUploadReq{
		DocID:    initResp.DocID,
		Rev:      initResp.Rev,
		UploadID: initResp.UploadID,
		PartInfo: completeParts,
	}
	if err := d.post(ctx, "/efast/v1/file/oscompleteupload", completeReq, nil); err != nil {
		return nil, fmt.Errorf("complete multipart upload: %w", err)
	}

	return &model.Object{
		ID:       initResp.DocID,
		Name:     file.GetName(),
		Size:     fileSize,
		IsFolder: false,
		Modified: time.Now(),
	}, nil
}

// uploadToObjectStorage uploads data to the object storage URL provided by AnyShare.
func (d *AnyShare) uploadToObjectStorage(ctx context.Context, auth AuthRequest, reader io.Reader, size int64, up driver.UpdateProgress) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, auth.URL, reader)
	if err != nil {
		return fmt.Errorf("create upload request: %w", err)
	}

	for k, v := range auth.Headers {
		req.Header.Set(k, v)
	}
	req.ContentLength = size
	req.Header.Set("Content-Length", strconv.FormatInt(size, 10))

	resp, err := base.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload to object storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("object storage upload returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (d *AnyShare) GetDetails(ctx context.Context) (*model.StorageDetails, error) {
	var quota QuotaResp
	if err := d.get(ctx, "/efast/v1/quota/user", &quota); err != nil {
		return nil, fmt.Errorf("get quota: %w", err)
	}

	return &model.StorageDetails{
		DiskUsage: model.DiskUsage{
			TotalSpace: quota.Allocated,
			UsedSpace:  quota.Used,
		},
	}, nil
}

var _ driver.Driver = (*AnyShare)(nil)
