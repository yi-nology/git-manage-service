package credential

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func Test(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req api.TestCredentialReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.URL == "" {
		response.BadRequest(c, "url is required")
		return
	}

	dao := db.NewCredentialDAO()
	cred, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "Credential not found")
		return
	}

	gitSvc := git.NewGitService()
	var testErr error

	switch cred.Type {
	case "ssh_key":
		if cred.SSHKeyID > 0 {
			sshKeyDAO := db.NewSSHKeyDAO()
			sshKey, err := sshKeyDAO.FindByID(cred.SSHKeyID)
			if err != nil {
				response.InternalServerError(c, "Failed to load SSH key: "+err.Error())
				return
			}
			testErr = gitSvc.TestRemoteConnectionWithDBKey(req.URL, sshKey.PrivateKey, sshKey.Passphrase)
		} else if cred.SSHKeyPath != "" {
			testErr = gitSvc.TestRemoteConnectionWithLocalKey(req.URL, cred.SSHKeyPath, cred.Secret)
		} else {
			response.BadRequest(c, "SSH key credential has no key configured")
			return
		}
	case "http_basic", "http_token":
		testErr = gitSvc.TestRemoteConnectionWithHTTP(req.URL, cred.Username, cred.Secret)
	default:
		testErr = gitSvc.TestRemoteConnection(req.URL)
	}

	if testErr != nil {
		response.Success(c, map[string]interface{}{
			"success": false,
			"message": testErr.Error(),
		})
		return
	}

	if err := dao.UpdateLastUsed(id); err != nil {
		_ = err
	}

	response.Success(c, map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	})
}

func Match(ctx context.Context, c *app.RequestContext) {
	var req api.MatchCredentialReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.URL == "" {
		response.BadRequest(c, "url is required")
		return
	}

	dao := db.NewCredentialDAO()
	recommended, others, err := dao.FindMatchingURL(req.URL)
	if err != nil {
		response.InternalServerError(c, "Failed to match credentials: "+err.Error())
		return
	}

	resp := api.MatchCredentialResp{
		Recommended: make([]api.CredentialDTO, 0, len(recommended)),
		Others:      make([]api.CredentialDTO, 0, len(others)),
	}
	for _, cred := range recommended {
		resp.Recommended = append(resp.Recommended, toCredentialDTO(&cred, nil))
	}
	for _, cred := range others {
		resp.Others = append(resp.Others, toCredentialDTO(&cred, nil))
	}

	response.Success(c, resp)
}
