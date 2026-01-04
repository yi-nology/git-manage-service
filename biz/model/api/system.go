package api

type ListDirsReq struct {
	Path   string `query:"path"`
	Search string `query:"search"`
}

type DirItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ListDirsResp struct {
	Parent  string    `json:"parent"`
	Current string    `json:"current"`
	Dirs    []DirItem `json:"dirs"`
}

type SSHKey struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ConfigReq struct {
	DebugMode   bool   `json:"debug_mode"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
}
