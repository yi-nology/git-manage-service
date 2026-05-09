package spec

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	specModel "github.com/yi-nology/git-manage-service/biz/model/spec"
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

func buildSpecTree(repoPath string) ([]*specModel.SpecFile, error) {
	nodes := make(map[string]*specModel.SpecFile)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		relPath, _ := filepath.Rel(repoPath, path)

		if strings.HasSuffix(info.Name(), ".spec") || info.IsDir() {
			name := info.Name()
			p := relPath
			isDir := info.IsDir()
			size := info.Size()
			modTime := info.ModTime().Format("2006-01-02T15:04:05Z")
			nodes[relPath] = &specModel.SpecFile{
				Name:    &name,
				Path:    &p,
				IsDir:   &isDir,
				Size:    &size,
				ModTime: &modTime,
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	childrenMap := make(map[string][]*specModel.SpecFile)

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

	var buildTree func(path string) *specModel.SpecFile
	buildTree = func(path string) *specModel.SpecFile {
		node := nodes[path]
		if node == nil {
			return nil
		}

		children := childrenMap[path]
		if len(children) > 0 {
			node.Children = children
			for _, child := range children {
				buildTree(*child.Path)
			}
		}

		return node
	}

	root := buildTree(".")
	if root == nil {
		return []*specModel.SpecFile{}, nil
	}

	filterTree(root)

	if len(root.Children) > 0 {
		return root.Children, nil
	}

	return []*specModel.SpecFile{}, nil
}

func createDirChain(pathMap map[string]*specModel.SpecFile, path string, repoPath string) *specModel.SpecFile {
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

	name := filepath.Base(path)
	p := path
	isDir := true
	modTime := info.ModTime().Format("2006-01-02T15:04:05Z")
	dir := &specModel.SpecFile{
		Name:    &name,
		Path:    &p,
		IsDir:   &isDir,
		ModTime: &modTime,
	}
	pathMap[path] = dir

	parentPath := filepath.Dir(path)
	if parentPath == "" {
		parentPath = "."
	}

	parent := createDirChain(pathMap, parentPath, repoPath)
	if parent != nil {
		parent.Children = append(parent.Children, dir)
	}

	return dir
}

func filterTree(node *specModel.SpecFile) bool {
	if !*node.IsDir {
		return strings.HasSuffix(*node.Name, ".spec")
	}

	var hasSpecFile bool
	var filteredChildren []*specModel.SpecFile

	for _, child := range node.Children {
		if filterTree(child) {
			filteredChildren = append(filteredChildren, child)
			hasSpecFile = true
		}
	}

	node.Children = filteredChildren
	return hasSpecFile
}
