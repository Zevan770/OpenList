package anyshare

import "time"

// ===== Entry doc lib (root listing) =====

type EntryDocLib struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Rev        string    `json:"rev"`
	Attr       int       `json:"attr"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
	Type       string    `json:"type"`
}

// ===== Paginated folder listing =====

type ListResponse struct {
	Dirs       []DirInfo  `json:"dirs"`
	Files      []FileInfo `json:"files"`
	NextMarker string     `json:"next_marker"`
}

type DirInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Rev        string    `json:"rev"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type FileInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Rev        string    `json:"rev"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

// ===== Legacy dir listing =====

type LegacyListReq struct {
	DocID string `json:"docid"`
}

type LegacyListResp struct {
	Dirs  []LegacyDirInfo  `json:"dirs"`
	Files []LegacyFileInfo `json:"files"`
}

type LegacyDirInfo struct {
	DocID      string `json:"docid"`
	Name       string `json:"name"`
	Rev        string `json:"rev"`
	Size       int64  `json:"size"`
	CreateTime int64  `json:"create_time"`
	Modified   int64  `json:"modified"`
}

type LegacyFileInfo struct {
	DocID      string `json:"docid"`
	Name       string `json:"name"`
	Rev        string `json:"rev"`
	Size       int64  `json:"size"`
	CreateTime int64  `json:"create_time"`
	Modified   int64  `json:"modified"`
}

// ===== File download =====

type OsDownloadReq struct {
	DocID string `json:"docid"`
	Rev   string `json:"rev"`
}

type OsDownloadResp struct {
	AuthRequest []AuthRequest `json:"authrequest"`
	Name        string        `json:"name"`
	Rev         string        `json:"rev"`
	Size        int64         `json:"size"`
}

type AuthRequest struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// ===== File upload (simple) =====

type OsBeginUploadReq struct {
	DocID       string `json:"docid"`
	Length      int64  `json:"length"`
	Name        string `json:"name"`
	Ondup       int    `json:"ondup"`
	ClientMtime int64  `json:"client_mtime,omitempty"`
}

type OsBeginUploadResp struct {
	AuthRequest []AuthRequest `json:"authrequest"`
	DocID       string        `json:"docid"`
	Name        string        `json:"name"`
	Rev         string        `json:"rev"`
}

type OsEndUploadReq struct {
	DocID string `json:"docid"`
	Rev   string `json:"rev"`
}

type OsEndUploadResp struct {
	Name     string `json:"name"`
	Modified int64  `json:"modified"`
}

// ===== File upload (multipart) =====

type OsInitMultiUploadReq struct {
	DocID  string `json:"docid"`
	Length int64  `json:"length"`
	Name   string `json:"name"`
	Ondup  int    `json:"ondup"`
}

type OsInitMultiUploadResp struct {
	DocID    string `json:"docid"`
	Name     string `json:"name"`
	Rev      string `json:"rev"`
	UploadID string `json:"uploadid"`
}

type PartInfo struct {
	PartNumber int   `json:"partnumber"`
	Start      int64 `json:"start"`
	End        int64 `json:"end"`
}

type OsUploadPartReq struct {
	DocID    string     `json:"docid"`
	Rev      string     `json:"rev"`
	UploadID string     `json:"uploadid"`
	Parts    []PartInfo `json:"parts"`
}

type OsUploadPartResp struct {
	AuthRequests []AuthRequest `json:"authrequests"`
}

type PartInfoComplete struct {
	PartNumber int    `json:"partnumber"`
	Size       int64  `json:"size"`
	Etag       string `json:"etag,omitempty"`
}

type OsCompleteMultiUploadReq struct {
	DocID    string             `json:"docid"`
	Rev      string             `json:"rev"`
	UploadID string             `json:"uploadid"`
	PartInfo []PartInfoComplete `json:"partinfo"`
}

// ===== Upload options =====

type OsOptionResp struct {
	PartMinSize int64 `json:"partminsize"`
	PartMaxSize int64 `json:"partmaxsize"`
	PartMaxNum  int   `json:"partmaxnum"`
}

// ===== Directory operations =====

type DirCreateReq struct {
	DocID string `json:"docid"`
	Name  string `json:"name"`
}

type DirCreateResp struct {
	DocID      string `json:"docid"`
	Rev        string `json:"rev"`
	Name       string `json:"name"`
	Modified   int64  `json:"modified"`
	CreateTime int64  `json:"create_time"`
}

type DocIDReq struct {
	DocID string `json:"docid"`
}

type DirRenameReq struct {
	DocID string `json:"docid"`
	Name  string `json:"name"`
}

type DirMoveReq struct {
	DocID      string `json:"docid"`
	DestParent string `json:"destparent"`
}

type DirCopyReq struct {
	DocID      string `json:"docid"`
	DestParent string `json:"destparent"`
}

// ===== File operations =====

type FileRenameReq struct {
	DocID string `json:"docid"`
	Name  string `json:"name"`
}

type FileMoveReq struct {
	DocID      string `json:"docid"`
	DestParent string `json:"destparent"`
}

type FileCopyReq struct {
	DocID      string `json:"docid"`
	DestParent string `json:"destparent"`
}

// ===== Quota =====

type QuotaResp struct {
	Used      int64 `json:"used"`
	Allocated int64 `json:"allocated"`
}
