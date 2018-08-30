package composer

import "time"

type Dist struct {
	URL  string `json:"url" msgpack:"url"`
	Type string `json:"type" msgpack:"type"`
}

type Source struct {
	URL       string `json:"url" msgpack:"url"`
	Type      string `json:"type" msgpack:"type"`
	Reference string `json:"reference" msgpack:"reference"`
}

type Author struct {
	Name     string `json:"name,omitempty" msgpack:"name"`
	Email    string `json:"email,omitempty" msgpack:"email"`
	Homepage string `json:"homepage,omitempty" msgpack:"homepage"`
	Role     string `json:"role,omitempty" msgpack:"role"`
}

type Support struct {
	Email  string `json:"email,omitempty" msgpack:"email"`
	Issues string `json:"issues,omitempty" msgpack:"issues"`
	Forum  string `json:"forum,omitempty" msgpack:"forum"`
	Wiki   string `json:"wiki,omitempty" msgpack:"wiki"`
	IRC    string `json:"irc,omitempty" msgpack:"irc"`
	Source string `json:"source,omitempty" msgpack:"source"`
	Docs   string `json:"docs,omitempty" msgpack:"docs"`
	RSS    string `json:"rss,omitempty" msgpack:"rss"`
}

type ArchiveOptions struct {
	Exclude []string `json:"exclude,omitempty" msgpack:"exclude"`
}

type PackageLink map[string]string

type Package struct {
	UID                string                 `json:"uid,omitempty" msgpack:"uid"`
	Name               string                 `json:"name,omitempty" msgpack:"name"`
	Description        string                 `json:"description,omitempty" msgpack:"description"`
	Version            string                 `json:"version,omitempty" msgpack:"version"`
	Type               string                 `json:"type,omitempty" msgpack:"type"`
	Keywords           []string               `json:"keywords,omitempty" msgpack:"keywords"`
	Homepage           string                 `json:"homepage,omitempty" msgpack:"homepage"`
	Readme             string                 `json:"readme,omitempty" msgpack:"readme"`
	Time               *time.Time             `json:"time,omitempty" msgpack:"time"`
	License            interface{}            `json:"license,omitempty" msgpack:"license"`
	Authors            []Author               `json:"authors,omitempty" msgpack:"authors"`
	Support            *Support               `json:"support,omitempty" msgpack:"support"`
	Require            PackageLink            `json:"require,omitempty" msgpack:"require"`
	RequireDev         PackageLink            `json:"require-dev,omitempty" msgpack:"require-dev"`
	Conflict           PackageLink            `json:"conflict,omitempty" msgpack:"conflict"`
	Replace            PackageLink            `json:"replace,omitempty" msgpack:"replace"`
	Provide            PackageLink            `json:"provide,omitempty" msgpack:"provide"`
	Suggest            PackageLink            `json:"suggest,omitempty" msgpack:"suggest"`
	Autoload           map[string]interface{} `json:"autoload,omitempty" msgpack:"autoload"`
	AutoloadDev        map[string]interface{} `json:"autoload-dev,omitempty" msgpack:"autoload-dev"`
	IncludePath        []string               `json:"include-path,omitempty" msgpack:"include-path"`
	TargetDir          string                 `json:"target-dir,omitempty" msgpack:"target-dir"`
	MinimumStability   string                 `json:"minimum-stability,omitempty" msgpack:"minimum-stability"`
	PreferStable       bool                   `json:"prefer-stable,omitempty" msgpack:"prefer-stable"`
	Repositories       interface{}            `json:"repositories,omitempty" msgpack:"repositories"`
	Config             interface{}            `json:"config,omitempty" msgpack:"config"`
	Scripts            interface{}            `json:"scripts,omitempty" msgpack:"scripts"`
	Extra              interface{}            `json:"extra,omitempty" msgpack:"extra"`
	Bin                []string               `json:"bin,omitempty" msgpack:"bin"`
	Archive            *ArchiveOptions        `json:"archive,omitempty" msgpack:"archive"`
	Abandoned          bool                   `json:"abandoned,omitempty" msgpack:"abondoned"`
	NonFeatureBranches []string               `json:"non-feature-branches,omitempty" msgpack:"non-feature-branches"`
	Dist               *Dist                  `json:"dist,omitempty" msgpack:"dist"`
	Source             *Source                `json:"source,omitempty" msgpack:"source"`
}

type PackageVersions map[string]*Package
