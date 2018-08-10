package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	gogitlab "github.com/xanzy/go-gitlab"
	"github.com/zachomedia/composerrepo/pkg/composer"
)

type GitLabConnector struct {
	ID     string
	Client *gogitlab.Client
	Group  *gogitlab.Group
}

func NewConnector(id string, client *gogitlab.Client, group interface{}) (*GitLabConnector, error) {
	connector := &GitLabConnector{
		ID:     id,
		Client: client,
	}

	if groupObj, ok := group.(*gogitlab.Group); ok {
		connector.Group = groupObj
	} else {
		groupObj, _, err := connector.Client.Groups.GetGroup(group)
		if err != nil {
			return nil, err
		}

		connector.Group = groupObj
	}

	return connector, nil
}

func (conn *GitLabConnector) GetID() string {
	return conn.ID
}

func (conn *GitLabConnector) GetName() string {
	return strings.Replace(strings.ToLower(conn.Group.FullPath), "/", "-", -1)
}

func (conn *GitLabConnector) getProjects() ([]*gogitlab.Project, error) {
	projects := make([]*gogitlab.Project, 0)

	simple := true
	opts := &gogitlab.ListGroupProjectsOptions{
		Simple: &simple,
		ListOptions: gogitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	inProjects, res, err := conn.Client.Groups.ListGroupProjects(conn.Group.ID, opts)
	projects = append(projects, inProjects...)
	if err != nil {
		return nil, err
	}
	for page := 2; page < res.TotalPages; page++ {
		opts.ListOptions.Page = page
		inProjects, _, err := conn.Client.Groups.ListGroupProjects(conn.Group.ID, opts)
		if err != nil {
			return nil, err
		}
		projects = append(projects, inProjects...)
	}

	return projects, nil
}

func (conn *GitLabConnector) getProjectRefs(project *gogitlab.Project) ([]interface{}, error) {
	refs := make([]interface{}, 0)

	// Get branches
	branchesOpts := &gogitlab.ListBranchesOptions{
		PerPage: 100,
		Page:    1,
	}
	branches, branchesResp, err := conn.Client.Branches.ListBranches(project.ID, branchesOpts)
	if err != nil {
		return nil, err
	}
	for _, branch := range branches {
		refs = append(refs, branch)
	}

	for page := 1; page < branchesResp.TotalPages; page++ {
		branchesOpts.Page = page
		branches, _, err := conn.Client.Branches.ListBranches(project.ID, branchesOpts)
		if err != nil {
			return nil, err
		}
		for _, branch := range branches {
			refs = append(refs, branch)
		}
	}

	// Get tags
	tagsOpts := &gogitlab.ListTagsOptions{
		ListOptions: gogitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}
	tags, tagsResp, err := conn.Client.Tags.ListTags(project.ID, tagsOpts)
	if err != nil {
		return nil, err
	}
	for _, branch := range tags {
		refs = append(refs, branch)
	}

	for page := 1; page < tagsResp.TotalPages; page++ {
		tagsOpts.ListOptions.Page = page
		tags, _, err := conn.Client.Tags.ListTags(project.ID, tagsOpts)
		if err != nil {
			return nil, err
		}
		for _, branch := range tags {
			refs = append(refs, branch)
		}
	}

	return refs, nil
}

func (conn *GitLabConnector) getRefPackage(project *gogitlab.Project, ref string) (*composer.Package, error) {
	var pkg composer.Package

	// Check for a composer.json file
	composerJSON, _, err := conn.Client.RepositoryFiles.GetRawFile(project.ID, "composer.json", &gogitlab.GetRawFileOptions{
		Ref: &ref,
	})

	if err == nil {
		err = json.Unmarshal(composerJSON, &pkg)
		if err != nil {
			return nil, err
		}
	}

	pkg.Name = strings.ToLower(project.PathWithNamespace)

	pkg.Source = &composer.Source{
		URL:       fmt.Sprintf("%s.git", project.WebURL),
		Type:      "git",
		Reference: ref,
	}

	return &pkg, nil
}

func (conn *GitLabConnector) getProjectVersions(project *gogitlab.Project) (map[string]*composer.Package, error) {
	versions := make(map[string]*composer.Package)

	refs, err := conn.getProjectRefs(project)
	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		if branch, ok := ref.(*gogitlab.Branch); ok {
			pkg, err := conn.getRefPackage(project, branch.Name)
			if err != nil {
				return nil, err
			}

			// Set version
			pkg.Version = fmt.Sprintf("dev-%s", branch.Name)

			// Set source by commit
			pkg.Source.Reference = branch.Commit.ID

			versions[pkg.Version] = pkg
		} else if tag, ok := ref.(*gogitlab.Tag); ok {
			pkg, err := conn.getRefPackage(project, tag.Name)
			if err != nil {
				return nil, err
			}

			// Set version
			pkg.Version = tag.Name

			// Set source by commit
			pkg.Source.Reference = tag.Commit.ID

			// Convert Drupal version number to valid version number
			r, err := regexp.Compile("\\d+\\.x-(\\d+\\.\\d+(-.*)?)")
			if err != nil {
				return nil, err
			}

			versionMatch := r.FindStringSubmatch(pkg.Version)
			if len(versionMatch) > 0 {
				log.Printf("Changing version %q to %q", pkg.Version, versionMatch[1])
				pkg.Version = versionMatch[1]
			}

			// Confirm we have a valid version
			match, err := regexp.MatchString("v?\\d+\\.\\d+(\\.\\d+)?(-(dev|p|patch|a|alpha|b|beta|RC|rc)\\d*)?", pkg.Version)
			if err != nil {
				return nil, err
			}

			if match {
				versions[pkg.Version] = pkg
			} else {
				log.Printf("Skipping tag %q as it is not a valid version number", pkg.Version)
			}
		} else {
			return nil, errors.New("Unknown ref type")
		}
	}

	return versions, nil
}

func (conn *GitLabConnector) GetPackages() (composer.Packages, error) {
	packages := make(composer.Packages)

	projects, err := conn.getProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		log.Printf("Loading %q", project.PathWithNamespace)

		versions, err := conn.getProjectVersions(project)
		if err != nil {
			return nil, err
		}

		packages[strings.ToLower(project.PathWithNamespace)] = versions
	}

	return packages, nil
}

func (conn *GitLabConnector) GetPackage(packageName string) (composer.PackageVersions, error) {
	project, _, err := conn.Client.Projects.GetProject(packageName)
	if err != nil {
		return nil, err
	}

	return conn.getProjectVersions(project)
}
