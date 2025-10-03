package warlot

// ---- Project Models ----

type InitProjectRequest struct {
	HolderID      string `json:"holder_id"`
	ProjectName   string `json:"project_name"`
	OwnerAddress  string `json:"owner_address"`
	EpochSet      int    `json:"epoch_set"`
	CycleEnd      int    `json:"cycle_end"`
	WritersLen    int    `json:"writers_len"`
	TrackBackLen  int    `json:"track_back_len"`
	DraftEpochDur int    `json:"draft_epoch_dur"`
	IncludePass   bool   `json:"include_pass"`
	Deletable     bool   `json:"deletable"`
}

type InitProjectResponse struct {
	ProjectID    string `json:"ProjectID"`
	DBID         string `json:"DBID"`
	WriterPassID string `json:"WriterPassID"`
	BlobID       string `json:"BlobID"`
	TxDigest     string `json:"TxDigest"`
	CSVHashHex   string `json:"CSVHashHex"`
	DigestHex    string `json:"DigestHex"`
	SignatureHex string `json:"SignatureHex"`
}

type IssueKeyRequest struct {
	ProjectID     string `json:"projectId"`
	ProjectHolder string `json:"projectHolder"`
	ProjectName   string `json:"projectName"`
	User          string `json:"user"`
}

type IssueKeyResponse struct {
	APIKey string `json:"apiKey"`
	URL    string `json:"url"`
}

type ResolveProjectRequest struct {
	HolderID    string `json:"holder_id"`
	ProjectName string `json:"project_name"`
}

// ResolveProjectResponse accepts both modern snake_case and legacy PascalCase.
type ResolveProjectResponse struct {
	// Current/observed fields.
	ExistsMeta  bool   `json:"exists_meta"`
	ExistsChain bool   `json:"exists_chain"`
	ProjectID   string `json:"project_id"`
	DBID        string `json:"db_id"`
	Action      string `json:"action"`
	// Legacy fields.
	LegacyProjectID string `json:"ProjectID,omitempty"`
	LegacyDBID      string `json:"DBID,omitempty"`
}

type TableCountResponse struct {
	ProjectID  string `json:"project_id"`
	TableCount int    `json:"table_count"`
}

// ---- SQL Models ----

type SQLRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}

// SQLResponse supports both DDL/DML and SELECT shapes.
type SQLResponse struct {
	OK       bool                     `json:"ok"`
	RowCount *int                     `json:"row_count,omitempty"`
	Rows     []map[string]interface{} `json:"rows,omitempty"`
	Error    string                   `json:"error,omitempty"`
}

// ---- Tables and Status Models ----

type ListTablesResponse struct {
	Tables []string `json:"tables"`
}

type BrowseRowsResponse struct {
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
	Table  string                   `json:"table"`
	Rows   []map[string]interface{} `json:"rows"`
}

// TableSchema is intentionally open to allow backend evolution.
type TableSchema = map[string]interface{}

// ProjectStatus and CommitResponse are intentionally open.
type ProjectStatus = map[string]interface{}
type CommitResponse = map[string]interface{}
