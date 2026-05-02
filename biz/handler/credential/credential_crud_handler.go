package credential

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func List(ctx context.Context, c *app.RequestContext) {
	dao := db.NewCredentialDAO()
	creds, err := dao.FindAll()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch credentials: "+err.Error())
		return
	}

	sshKeyMap := buildSSHKeyMap(creds)

	result := make([]api.CredentialDTO, 0, len(creds))
	for _, cred := range creds {
		result = append(result, toCredentialDTO(&cred, sshKeyMap))
	}

	response.Success(c, result)
}

func Create(ctx context.Context, c *app.RequestContext) {
	var req api.CreateCredentialReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}
	if req.Type == "" {
		response.BadRequest(c, "type is required")
		return
	}
	if req.Type != "ssh_key" && req.Type != "http_basic" && req.Type != "http_token" && req.Type != "platform_token" {
		response.BadRequest(c, "type must be ssh_key, http_basic, http_token or platform_token")
		return
	}

	if req.Type == "ssh_key" && req.SSHKeyID == 0 && req.SSHKeyPath == "" {
		response.BadRequest(c, "ssh_key type requires ssh_key_id or ssh_key_path")
		return
	}
	if (req.Type == "http_basic" || req.Type == "http_token") && req.Username == "" && req.Secret == "" {
		response.BadRequest(c, "http type requires username or secret")
		return
	}

	if req.SSHKeyID > 0 {
		sshKeyDAO := db.NewSSHKeyDAO()
		if _, err := sshKeyDAO.FindByID(req.SSHKeyID); err != nil {
			response.BadRequest(c, "SSH key not found with the given id")
			return
		}
	}

	dao := db.NewCredentialDAO()
	exists, err := dao.ExistsByName(req.Name)
	if err != nil {
		response.InternalServerError(c, "Failed to check credential name: "+err.Error())
		return
	}
	if exists {
		response.BadRequest(c, "Credential with this name already exists")
		return
	}

	cred := &po.Credential{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		SSHKeyID:    req.SSHKeyID,
		SSHKeyPath:  req.SSHKeyPath,
		Username:    req.Username,
		Secret:      req.Secret,
		URLPattern:  req.URLPattern,
	}

	if err := dao.Create(cred); err != nil {
		response.InternalServerError(c, "Failed to create credential: "+err.Error())
		return
	}

	response.Success(c, toCredentialDTO(cred, nil))
}

func Get(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	dao := db.NewCredentialDAO()
	cred, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "Credential not found")
		return
	}

	response.Success(c, toCredentialDTO(cred, nil))
}

func Update(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req api.UpdateCredentialReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dao := db.NewCredentialDAO()
	cred, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "Credential not found")
		return
	}

	if req.Name != "" && req.Name != cred.Name {
		exists, err := dao.ExistsByNameExcludeID(req.Name, id)
		if err != nil {
			response.InternalServerError(c, "Failed to check credential name: "+err.Error())
			return
		}
		if exists {
			response.BadRequest(c, "Credential with this name already exists")
			return
		}
		cred.Name = req.Name
	}

	if req.Description != "" {
		cred.Description = req.Description
	}
	if req.SSHKeyID > 0 {
		sshKeyDAO := db.NewSSHKeyDAO()
		if _, err := sshKeyDAO.FindByID(req.SSHKeyID); err != nil {
			response.BadRequest(c, "SSH key not found with the given id")
			return
		}
		cred.SSHKeyID = req.SSHKeyID
	}
	if req.SSHKeyPath != "" {
		cred.SSHKeyPath = req.SSHKeyPath
	}
	if req.Username != "" {
		cred.Username = req.Username
	}
	if req.Secret != "" {
		cred.Secret = req.Secret
	}
	if req.URLPattern != "" {
		cred.URLPattern = req.URLPattern
	}

	if err := dao.Save(cred); err != nil {
		response.InternalServerError(c, "Failed to update credential: "+err.Error())
		return
	}

	response.Success(c, toCredentialDTO(cred, nil))
}

func Delete(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	dao := db.NewCredentialDAO()
	if _, err := dao.FindByID(id); err != nil {
		response.NotFound(c, "Credential not found")
		return
	}

	repoDAO := db.NewRepoDAO()
	repos, _ := repoDAO.FindAll()
	for _, repo := range repos {
		if repo.DefaultCredentialID == id {
			response.BadRequest(c, "Credential is referenced by repo: "+repo.Name)
			return
		}
		if repo.RemoteCredentials != nil {
			for remoteName, credID := range repo.RemoteCredentials {
				if credID == id {
					response.BadRequest(c, "Credential is referenced by repo "+repo.Name+" remote "+remoteName)
					return
				}
			}
		}
	}

	pcDAO := db.NewProviderConfigDAO()
	configs, _ := pcDAO.FindAll()
	for _, cfg := range configs {
		if cfg.CredentialID == id {
			response.BadRequest(c, "Credential is referenced by provider config: "+cfg.Name)
			return
		}
	}

	if err := dao.Delete(id); err != nil {
		response.InternalServerError(c, "Failed to delete credential: "+err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "Credential deleted successfully"})
}
