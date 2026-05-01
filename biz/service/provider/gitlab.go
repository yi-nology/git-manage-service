package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type gitlabProvider struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewGitLabProvider(baseURL, token string, skipTLS bool) *gitlabProvider {
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	transport := &http.Transport{}
	if skipTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &gitlabProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second, Transport: transport},
	}
}

func (g *gitlabProvider) Platform() Platform { return PlatformGitLab }

func (g *gitlabProvider) TestConnection(ctx context.Context) (*TestConnectionResult, error) {
	var user struct {
		Username string `json:"username"`
	}
	if err := g.doRequest(ctx, "GET", "/user", nil, &user); err != nil {
		return &TestConnectionResult{Connected: false, Message: err.Error()}, nil
	}
	return &TestConnectionResult{Connected: true, Platform: string(g.Platform()), UserName: user.Username}, nil
}

func (g *gitlabProvider) ListRepos(ctx context.Context, opts ListRepoOptions) ([]*PlatformRepo, error) {
	path := "/projects"
	if opts.Owner != "" {
		path = fmt.Sprintf("/groups/%s/projects", opts.Owner)
	}
	if opts.Page == 0 {
		opts.Page = 1
	}
	if opts.PerPage == 0 {
		opts.PerPage = 20
	}
	path = fmt.Sprintf("%s?page=%d&per_page=%d", path, opts.Page, opts.PerPage)
	var projects []struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		PathWithNS    string `json:"path_with_namespace"`
		Description   string `json:"description"`
		HTTPURL       string `json:"http_url_to_repo"`
		SSHURL        string `json:"ssh_url_to_repo"`
		DefaultBranch string `json:"default_branch"`
		Visibility    string `json:"visibility"`
	}
	if err := g.doRequest(ctx, "GET", path, nil, &projects); err != nil {
		return nil, err
	}
	repos := make([]*PlatformRepo, 0, len(projects))
	for _, p := range projects {
		parts := strings.SplitN(p.PathWithNS, "/", 2)
		owner := ""
		if len(parts) == 2 {
			owner = parts[0]
		}
		repos = append(repos, &PlatformRepo{
			ID: int64(p.ID), FullName: p.PathWithNS, Name: p.Name, Owner: owner,
			Description: p.Description, CloneURL: p.HTTPURL, SSHURL: p.SSHURL,
			DefaultBranch: p.DefaultBranch, Private: p.Visibility != "public", Platform: g.Platform(),
		})
	}
	return repos, nil
}

func (g *gitlabProvider) GetRepo(ctx context.Context, owner, repo string) (*PlatformRepo, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	var p struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		PathWithNS    string `json:"path_with_namespace"`
		Description   string `json:"description"`
		HTTPURL       string `json:"http_url_to_repo"`
		SSHURL        string `json:"ssh_url_to_repo"`
		DefaultBranch string `json:"default_branch"`
		Visibility    string `json:"visibility"`
	}
	if err := g.doRequest(ctx, "GET", "/projects/"+encoded, nil, &p); err != nil {
		return nil, err
	}
	parts := strings.SplitN(p.PathWithNS, "/", 2)
	ownerR := ""
	if len(parts) == 2 {
		ownerR = parts[0]
	}
	return &PlatformRepo{
		ID: int64(p.ID), FullName: p.PathWithNS, Name: p.Name, Owner: ownerR,
		Description: p.Description, CloneURL: p.HTTPURL, SSHURL: p.SSHURL,
		DefaultBranch: p.DefaultBranch, Private: p.Visibility != "public", Platform: g.Platform(),
	}, nil
}

func (g *gitlabProvider) CreateCR(ctx context.Context, opts CreateCROptions) (*ChangeRequest, error) {
	encoded := fmt.Sprintf("%s%%2F%s", opts.Owner, opts.Repo)
	body := map[string]interface{}{
		"source_branch": opts.SourceBranch, "target_branch": opts.TargetBranch,
		"title": opts.Title, "description": opts.Description,
		"remove_source_branch": opts.RemoveSourceBranch,
	}
	if len(opts.Labels) > 0 {
		body["labels"] = strings.Join(opts.Labels, ",")
	}
	var mr gitlabMR
	if err := g.doRequest(ctx, "POST", "/projects/"+encoded+"/merge_requests", body, &mr); err != nil {
		return nil, err
	}
	return mr.toCR(), nil
}

func (g *gitlabProvider) GetCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	var mr gitlabMR
	if err := g.doRequest(ctx, "GET", fmt.Sprintf("/projects/%s/merge_requests/%d", encoded, number), nil, &mr); err != nil {
		return nil, err
	}
	return mr.toCR(), nil
}

func (g *gitlabProvider) ListCRs(ctx context.Context, opts ListCROptions) ([]*ChangeRequest, int, error) {
	encoded := fmt.Sprintf("%s%%2F%s", opts.Owner, opts.Repo)
	if opts.Page == 0 {
		opts.Page = 1
	}
	if opts.PerPage == 0 {
		opts.PerPage = 20
	}
	path := fmt.Sprintf("/projects/%s/merge_requests?page=%d&per_page=%d", encoded, opts.Page, opts.PerPage)
	if opts.State != "" {
		path += "&state=" + string(opts.State)
	}
	if opts.SourceBranch != "" {
		path += "&source_branch=" + opts.SourceBranch
	}
	if opts.TargetBranch != "" {
		path += "&target_branch=" + opts.TargetBranch
	}
	var mrs []gitlabMR
	if err := g.doRequest(ctx, "GET", path, nil, &mrs); err != nil {
		return nil, 0, err
	}
	crs := make([]*ChangeRequest, 0, len(mrs))
	for i := range mrs {
		crs = append(crs, mrs[i].toCR())
	}
	return crs, len(crs), nil
}

func (g *gitlabProvider) MergeCR(ctx context.Context, owner, repo string, number int, opts MergeCROptions) (*ChangeRequest, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	var existingMR gitlabMR
	if err := g.doRequest(ctx, "GET", fmt.Sprintf("/projects/%s/merge_requests/%d", encoded, number), nil, &existingMR); err == nil {
		if existingMR.MergeStatus != "" && existingMR.MergeStatus != "can_be_merged" && existingMR.MergeStatus != "checking" {
			return nil, fmt.Errorf("MR cannot be merged (status: %s). It may have conflicts or an active pipeline", existingMR.MergeStatus)
		}
		if existingMR.State != "opened" {
			return nil, fmt.Errorf("MR is not in 'opened' state (current: %s)", existingMR.State)
		}
	}
	body := map[string]interface{}{}
	if opts.MergeCommitMessage != "" {
		body["merge_commit_message"] = opts.MergeCommitMessage
	}
	if opts.Squash {
		body["squash"] = true
	}
	if opts.RemoveSourceBranch {
		body["should_remove_source_branch"] = true
	}
	var mr gitlabMR
	if err := g.doRequest(ctx, "PUT", fmt.Sprintf("/projects/%s/merge_requests/%d/merge", encoded, number), body, &mr); err != nil {
		return nil, fmt.Errorf("merge failed: %w. The MR may have conflicts, an active pipeline, or branch protection rules preventing merge", err)
	}
	return mr.toCR(), nil
}

func (g *gitlabProvider) CloseCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	body := map[string]interface{}{"state_event": "close"}
	var mr gitlabMR
	if err := g.doRequest(ctx, "PUT", fmt.Sprintf("/projects/%s/merge_requests/%d", encoded, number), body, &mr); err != nil {
		return nil, err
	}
	return mr.toCR(), nil
}

func (g *gitlabProvider) CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (*PlatformWebhook, error) {
	encoded := fmt.Sprintf("%s%%2F%s", opts.Owner, opts.Repo)
	body := map[string]interface{}{"url": opts.URL, "token": opts.Secret}
	if len(opts.Events) > 0 {
		em := map[string]bool{}
		for _, e := range opts.Events {
			em[e] = true
		}
		body["push_events"] = em["push"]
		body["merge_requests_events"] = em["cr"]
		body["tag_push_events"] = em["tag"]
	}
	var wh struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := g.doRequest(ctx, "POST", "/projects/"+encoded+"/hooks", body, &wh); err != nil {
		return nil, err
	}
	return &PlatformWebhook{ID: int64(wh.ID), URL: wh.URL}, nil
}

func (g *gitlabProvider) DeleteWebhook(ctx context.Context, owner, repo string, webhookID int64) error {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	return g.doRequest(ctx, "DELETE", fmt.Sprintf("/projects/%s/hooks/%d", encoded, webhookID), nil, nil)
}

func (g *gitlabProvider) ListWebhooks(ctx context.Context, owner, repo string) ([]*PlatformWebhook, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	var whs []struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := g.doRequest(ctx, "GET", "/projects/"+encoded+"/hooks", nil, &whs); err != nil {
		return nil, err
	}
	result := make([]*PlatformWebhook, 0, len(whs))
	for _, wh := range whs {
		result = append(result, &PlatformWebhook{ID: int64(wh.ID), URL: wh.URL})
	}
	return result, nil
}

func (g *gitlabProvider) ParseWebhookEvent(r *http.Request, secret string) (*NormalizedEvent, error) {
	if err := g.ValidateWebhookSignature(r, secret); err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	var pl struct {
		ObjectKind string `json:"object_kind"`
		User       struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Name     string `json:"name"`
		} `json:"user"`
		Project struct {
			PathWithNS string `json:"path_with_namespace"`
		} `json:"project"`
		ObjectAttributes struct {
			IID          int       `json:"iid"`
			Title        string    `json:"title"`
			Description  string    `json:"description"`
			State        string    `json:"state"`
			SourceBranch string    `json:"source_branch"`
			TargetBranch string    `json:"target_branch"`
			Action       string    `json:"action"`
			MergeStatus  string    `json:"merge_status"`
			URL          string    `json:"url"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
		} `json:"object_attributes"`
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(body, &pl); err != nil {
		return nil, err
	}

	parts := strings.SplitN(pl.Project.PathWithNS, "/", 2)
	er := &EventRepo{FullName: pl.Project.PathWithNS}
	if len(parts) == 2 {
		er.Owner = parts[0]
		er.Name = parts[1]
	}
	actor := &CRUser{ID: int64(pl.User.ID), Username: pl.User.Username, Name: pl.User.Name}

	event := &NormalizedEvent{
		ID:     fmt.Sprintf("gl-%d-%d", time.Now().UnixNano(), pl.ObjectAttributes.IID),
		Source: g.Platform(), Timestamp: time.Now(), Actor: actor, Repo: er,
	}

	switch pl.ObjectKind {
	case "merge_request":
		state := mapGLState(pl.ObjectAttributes.State)
		action := pl.ObjectAttributes.Action
		if action == "merge" {
			action = "merged"
		}
		event.Type = "cr." + action
		event.CR = &ChangeRequest{
			ID: int64(pl.ObjectAttributes.IID), Number: pl.ObjectAttributes.IID,
			Title: pl.ObjectAttributes.Title, Description: pl.ObjectAttributes.Description,
			State: state, SourceBranch: pl.ObjectAttributes.SourceBranch,
			TargetBranch: pl.ObjectAttributes.TargetBranch, MergeStatus: pl.ObjectAttributes.MergeStatus,
			WebURL: pl.ObjectAttributes.URL, Author: actor,
			CreatedAt: pl.ObjectAttributes.CreatedAt, UpdatedAt: pl.ObjectAttributes.UpdatedAt,
		}
	case "push":
		event.Type = "push"
		event.Branch = strings.TrimPrefix(pl.Ref, "refs/heads/")
	case "tag_push":
		event.Type = "tag.created"
		event.Tag = strings.TrimPrefix(pl.Ref, "refs/tags/")
	}
	return event, nil
}

func (g *gitlabProvider) ValidateWebhookSignature(r *http.Request, secret string) error {
	token := r.Header.Get("X-Gitlab-Token")
	if token == "" || token != secret {
		return fmt.Errorf("invalid GitLab webhook token")
	}
	return nil
}

func (g *gitlabProvider) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", g.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("GitLab API %s %s returned %d: %s", method, path, resp.StatusCode, string(respBody))
	}
	if result != nil && resp.StatusCode != http.StatusNoContent {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

type gitlabMR struct {
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	State        string `json:"state"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Author       struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"author"`
	Labels      []string  `json:"labels"`
	MergeStatus string    `json:"merge_status"`
	WebURL      string    `json:"web_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (mr *gitlabMR) toCR() *ChangeRequest {
	return &ChangeRequest{
		ID: int64(mr.IID), Number: mr.IID, Title: mr.Title, Description: mr.Description,
		State: mapGLState(mr.State), SourceBranch: mr.SourceBranch, TargetBranch: mr.TargetBranch,
		Author: &CRUser{ID: int64(mr.Author.ID), Username: mr.Author.Username, Name: mr.Author.Name},
		Labels: mr.Labels, MergeStatus: mr.MergeStatus, WebURL: mr.WebURL,
		CreatedAt: mr.CreatedAt, UpdatedAt: mr.UpdatedAt,
	}
}

func mapGLState(state string) CRState {
	switch state {
	case "merged":
		return CRStateMerged
	case "closed":
		return CRStateClosed
	default:
		return CRStateOpened
	}
}

func (g *gitlabProvider) ListBranches(ctx context.Context, owner, repo string) ([]*PlatformBranch, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	path := fmt.Sprintf("/projects/%s/repository/branches?per_page=100", encoded)
	var branches []struct {
		Name string `json:"name"`
	}
	if err := g.doRequest(ctx, "GET", path, nil, &branches); err != nil {
		return nil, err
	}
	result := make([]*PlatformBranch, 0, len(branches))
	for _, b := range branches {
		result = append(result, &PlatformBranch{Name: b.Name})
	}
	return result, nil
}

func (g *gitlabProvider) CreateBranch(ctx context.Context, owner, repo, branch, ref string) (*PlatformBranch, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	body := map[string]string{"branch": branch, "ref": ref}
	var res struct {
		Name string `json:"name"`
	}
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/projects/%s/repository/branches", encoded), body, &res); err != nil {
		return nil, err
	}
	return &PlatformBranch{Name: res.Name}, nil
}

func (g *gitlabProvider) DeleteBranch(ctx context.Context, owner, repo, branch string) error {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	return g.doRequest(ctx, "DELETE", fmt.Sprintf("/projects/%s/repository/branches/%s", encoded, branch), nil, nil)
}

func (g *gitlabProvider) GetCRDiff(ctx context.Context, owner, repo string, number int) (*MergeDiff, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	var resp struct {
		Changes []struct {
			OldPath     string `json:"old_path"`
			NewPath     string `json:"new_path"`
			Diff        string `json:"diff"`
			NewFile     bool   `json:"new_file"`
			DeletedFile bool   `json:"deleted_file"`
			RenamedFile bool   `json:"renamed_file"`
			Binary      bool   `json:"binary"`
		} `json:"changes"`
	}
	if err := g.doRequest(ctx, "GET", fmt.Sprintf("/projects/%s/merge_requests/%d/changes", encoded, number), nil, &resp); err != nil {
		return nil, err
	}
	diff := &MergeDiff{}
	for _, c := range resp.Changes {
		additions, deletions := countDiffLines(c.Diff)
		cf := &ChangedFile{
			OldPath: c.OldPath, NewPath: c.NewPath, Diff: c.Diff,
			Additions: additions, Deletions: deletions,
			IsNew: c.NewFile, IsDeleted: c.DeletedFile, IsRenamed: c.RenamedFile, IsBinary: c.Binary,
		}
		diff.Files = append(diff.Files, cf)
		diff.TotalAdd += additions
		diff.TotalDel += deletions
		diff.RawDiff += c.Diff + "\n"
	}
	return diff, nil
}

func (g *gitlabProvider) GetCRFiles(ctx context.Context, owner, repo string, number int) ([]*ChangedFile, error) {
	diff, err := g.GetCRDiff(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return diff.Files, nil
}

func (g *gitlabProvider) CreateNote(ctx context.Context, owner, repo string, number int, body string) (string, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	payload := map[string]string{"body": body}
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/projects/%s/merge_requests/%d/notes", encoded, number), payload, &resp); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", resp.ID), nil
}

func (g *gitlabProvider) CreateDiscussion(ctx context.Context, owner, repo string, number int, opts DiscussionOptions) (string, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	payload := map[string]interface{}{"body": opts.Body}
	if opts.FilePath != "" {
		position := map[string]interface{}{
			"base_sha":      "head",
			"start_sha":     "head",
			"head_sha":      "head",
			"position_type": "text",
			"new_path":      opts.FilePath,
			"new_line":      opts.NewLine,
		}
		if opts.OldLine > 0 {
			position["old_path"] = opts.FilePath
			position["old_line"] = opts.OldLine
		}
		payload["position"] = position
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/projects/%s/merge_requests/%d/discussions", encoded, number), payload, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (g *gitlabProvider) CreateCommitStatus(ctx context.Context, owner, repo, sha string, opts CommitStatusOptions) error {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	payload := map[string]string{
		"state":       opts.State,
		"context":     opts.Context,
		"description": opts.Description,
		"target_url":  opts.TargetURL,
	}
	return g.doRequest(ctx, "POST", fmt.Sprintf("/projects/%s/statuses/%s", encoded, sha), payload, nil)
}

func (g *gitlabProvider) GetFileContent(ctx context.Context, owner, repo, filePath, ref string) (string, error) {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	apiPath := fmt.Sprintf("/projects/%s/repository/files/%s?ref=%s", encoded, filePath, ref)
	var resp struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := g.doRequest(ctx, "GET", apiPath, nil, &resp); err != nil {
		return "", err
	}
	if resp.Encoding == "base64" {
		decoded, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(resp.Content)))
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}
	return resp.Content, nil
}

func (g *gitlabProvider) UpdateCRLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	encoded := fmt.Sprintf("%s%%2F%s", owner, repo)
	payload := map[string]interface{}{"labels": strings.Join(labels, ",")}
	return g.doRequest(ctx, "PUT", fmt.Sprintf("/projects/%s/merge_requests/%d", encoded, number), payload, nil)
}

func countDiffLines(diff string) (additions, deletions int) {
	for _, line := range strings.Split(diff, "\n") {
		if len(line) == 0 {
			continue
		}
		if line[0] == '+' && !strings.HasPrefix(line, "+++") {
			additions++
		} else if line[0] == '-' && !strings.HasPrefix(line, "---") {
			deletions++
		}
	}
	return
}
