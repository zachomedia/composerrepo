package composer

import "time"

type Dist struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type Source struct {
	URL       string `json:"url"`
	Type      string `json:"type"`
	Reference string `json:"reference"`
}

type Author struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Homepage string `json:"homepage,omitempty"`
	Role     string `json:"role,omitempty"`
}

type Support struct {
	Email  string `json:"email,omitempty"`
	Issues string `json:"issues,omitempty"`
	Forum  string `json:"forum,omitempty"`
	Wiki   string `json:"wiki,omitempty"`
	IRC    string `json:"irc,omitempty"`
	Source string `json:"source,omitempty"`
	Docs   string `json:"docs,omitempty"`
	RSS    string `json:"rss,omitempty"`
}

type ArchiveOptions struct {
	Exclude []string `json:"exclude,omitempty"`
}

type PackageLink map[string]string

type Package struct {
	UID                string                 `json:"uid,omitempty"`
	Name               string                 `json:"name,omitempty"`
	Description        string                 `json:"description,omitempty"`
	Version            string                 `json:"version,omitempty"`
	Type               string                 `json:"type,omitempty"`
	Keywords           []string               `json:"keywords,omitempty"`
	Homepage           string                 `json:"homepage,omitempty"`
	Readme             string                 `json:"readme,omitempty"`
	Time               *time.Time             `json:"time,omitempty"`
	License            interface{}            `json:"license,omitempty"`
	Authors            []Author               `json:"authors,omitempty"`
	Support            *Support               `json:"support,omitempty"`
	Require            PackageLink            `json:"require,omitempty"`
	RequireDev         PackageLink            `json:"require-dev,omitempty"`
	Conflict           PackageLink            `json:"conflict,omitempty"`
	Replace            PackageLink            `json:"replace,omitempty"`
	Provide            PackageLink            `json:"provide,omitempty"`
	Suggest            PackageLink            `json:"suggest,omitempty"`
	Autoload           map[string]interface{} `json:"autoload,omitempty"`
	AutoloadDev        map[string]interface{} `json:"autoload-dev,omitempty"`
	IncludePath        []string               `json:"include-path,omitempty"`
	TargetDir          string                 `json:"target-dir,omitempty"`
	MinimumStability   string                 `json:"minimum-stability,omitempty"`
	PreferStable       bool                   `json:"prefer-stable,omitempty"`
	Repositories       interface{}            `json:"repositories,omitempty"`
	Config             interface{}            `json:"config,omitempty"`
	Scripts            interface{}            `json:"scripts,omitempty"`
	Extra              interface{}            `json:"extra,omitempty"`
	Bin                []string               `json:"bin,omitempty"`
	Archive            *ArchiveOptions        `json:"archive,omitempty"`
	Abandoned          bool                   `json:"abandoned,omitempty"`
	NonFeatureBranches []string               `json:"non-feature-branches,omitempty"`
	Dist               *Dist                  `json:"dist,omitempty"`
	Source             *Source                `json:"source,omitempty"`
}

type PackageVersions map[string]*Package
