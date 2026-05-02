package spec

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func GetSpecTree(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	tree, err := buildSpecTree(repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, tree)
}

func buildSpecTree(repoPath string) ([]api.SpecFile, error) {
	nodes := make(map[string]*api.SpecFile)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		relPath, _ := filepath.Rel(repoPath, path)

		if strings.HasSuffix(info.Name(), ".spec") || info.IsDir() {
			nodes[relPath] = &api.SpecFile{
				Name:    info.Name(),
				Path:    relPath,
				IsDir:   info.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	childrenMap := make(map[string][]*api.SpecFile)

	for path, node := range nodes {
		if path == "." {
			continue
		}

		parentPath := filepath.Dir(path)
		if parentPath == "" {
			parentPath = "."
		}

		childrenMap[parentPath] = append(childrenMap[parentPath], node)
	}

	var buildTree func(path string) *api.SpecFile
	buildTree = func(path string) *api.SpecFile {
		node := nodes[path]
		if node == nil {
			return nil
		}

		children := childrenMap[path]
		if len(children) > 0 {
			node.Children = make([]api.SpecFile, 0, len(children))
			for _, child := range children {
				buildTree(child.Path)
				node.Children = append(node.Children, *child)
			}
		}

		return node
	}

	root := buildTree(".")
	if root == nil {
		return []api.SpecFile{}, nil
	}

	filterTree(root)

	if len(root.Children) > 0 {
		return root.Children, nil
	}

	return []api.SpecFile{}, nil
}

func createDirChain(pathMap map[string]*api.SpecFile, path string, repoPath string) *api.SpecFile {
	if path == "." || path == "" {
		return pathMap["."]
	}

	if dir, exists := pathMap[path]; exists {
		return dir
	}

	info, err := os.Stat(filepath.Join(repoPath, path))
	if err != nil {
		return nil
	}

	dir := &api.SpecFile{
		Name:    filepath.Base(path),
		Path:    path,
		IsDir:   true,
		ModTime: info.ModTime(),
	}
	pathMap[path] = dir

	parentPath := filepath.Dir(path)
	if parentPath == "" {
		parentPath = "."
	}

	parent := createDirChain(pathMap, parentPath, repoPath)
	if parent != nil {
		parent.Children = append(parent.Children, *dir)
	}

	return dir
}

func filterTree(node *api.SpecFile) bool {
	if !node.IsDir {
		return strings.HasSuffix(node.Name, ".spec")
	}

	var hasSpecFile bool
	var filteredChildren []api.SpecFile

	for i := range node.Children {
		if filterTree(&node.Children[i]) {
			filteredChildren = append(filteredChildren, node.Children[i])
			hasSpecFile = true
		}
	}

	node.Children = filteredChildren
	return hasSpecFile
}
