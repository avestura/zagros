package market

import (
	"fmt"
	"net/url"
	"strings"
)

// RefKind says how a reference is resolved.
type RefKind int

const (
	RefRegistry RefKind = iota // registry id
	RefGitHub                  // github:owner/repo[@tag]
	RefURL                     // direct https URL to a .zagros/.zip archive
)

// Ref is a parsed install reference.
type Ref struct {
	Kind  RefKind
	ID    string // RefRegistry
	Owner string // RefGitHub
	Repo  string // RefGitHub
	Tag   string // RefGitHub: release tag or git ref ("" = latest release)
	URL   string // RefURL
}

// String returns the canonical form of the reference.
func (r Ref) String() string {
	switch r.Kind {
	case RefGitHub:
		s := "github:" + r.Owner + "/" + r.Repo
		if r.Tag != "" {
			s += "@" + r.Tag
		}
		return s
	case RefURL:
		return r.URL
	}
	return r.ID
}

// ParseRef understands:
//
//	hello-world                         registry id
//	github:owner/repo                   latest release
//	github:owner/repo@v1.2.0            specific tag (or branch/commit)
//	owner/repo[@tag]                    same as github:
//	https://github.com/owner/repo       (also .git, /tree/<ref>, /releases/tag/<tag>)
//	https://host/path/ext.zagros        direct archive
func ParseRef(s string) (Ref, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Ref{}, fmt.Errorf("empty reference")
	}
	if strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://") {
		return parseURLRef(s)
	}
	if rest, ok := strings.CutPrefix(s, "github:"); ok {
		return parseGitHubRef(rest)
	}
	if strings.Contains(s, "/") {
		return parseGitHubRef(s)
	}
	if !idRe.MatchString(s) {
		return Ref{}, fmt.Errorf("invalid extension id %q", s)
	}
	return Ref{Kind: RefRegistry, ID: s}, nil
}

func parseGitHubRef(s string) (Ref, error) {
	spec, tag, _ := strings.Cut(s, "@")
	spec = strings.TrimSuffix(spec, ".git")
	if !repoRe.MatchString(spec) {
		return Ref{}, fmt.Errorf("invalid GitHub reference %q (want github:owner/repo[@tag])", s)
	}
	owner, repo, _ := strings.Cut(spec, "/")
	return Ref{Kind: RefGitHub, Owner: owner, Repo: repo, Tag: tag}, nil
}

func parseURLRef(s string) (Ref, error) {
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return Ref{}, fmt.Errorf("invalid URL %q", s)
	}
	host := strings.ToLower(u.Hostname())
	if host == "github.com" || host == "www.github.com" {
		p := strings.Split(strings.Trim(u.Path, "/"), "/")
		isDownload := len(p) >= 4 && p[2] == "releases" && p[3] == "download"
		if len(p) >= 2 && !isDownload {
			r, err := parseGitHubRef(p[0] + "/" + p[1])
			if err != nil {
				return Ref{}, err
			}
			if len(p) >= 5 && p[2] == "releases" && p[3] == "tag" {
				r.Tag = p[4]
			} else if len(p) >= 4 && p[2] == "tree" {
				r.Tag = strings.Join(p[3:], "/")
			}
			return r, nil
		}
	}
	return Ref{Kind: RefURL, URL: u.String()}, nil
}
