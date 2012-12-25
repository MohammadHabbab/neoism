// Copyright (c) 2012 Jason McVetta.  This is Free Software, released under the 
// terms of the GPL v3.  See http://www.gnu.org/copyleft/gpl.html for details.

package neo4j

import (
	"github.com/jmcvetta/restclient"
	"sort"
	"strconv"
	"strings"
)

type RelationshipManager struct {
	db *Database
}

// do is a convenience wrapper around the embedded restclient's Do() method.
func (m *RelationshipManager) do(rr *restclient.RestRequest) (status int, err error) {
	return m.db.rc.Do(rr)
}

// GetRelationship fetches a Relationship from the DB by id.
func (m *RelationshipManager) Get(id int) (*Relationship, error) {
	rel := Relationship{}
	rel.db = m.db
	res := new(relationshipResponse)
	uri := join(m.db.url.String(), "relationship", strconv.Itoa(id))
	ne := new(neoError)
	rr := restclient.RestRequest{
		Url:    uri,
		Method: restclient.GET,
		Result: &res,
		Error:  &ne,
	}
	status, err := m.do(&rr)
	if err != nil {
		logPretty(ne)
		return &rel, err
	}
	switch status {
	default:
		logPretty(ne)
		err = BadResponse
	case 200:
		err = nil // Success!
	case 404:
		err = NotFound
	}
	rel.populate(res)
	return &rel, err
}

// Types lists all existing relationship types
func (m *RelationshipManager) Types() ([]string, error) {
	reltypes := []string{}
	if m.db.info.RelTypes == "" {
		return reltypes, FeatureUnavailable
	}
	ne := new(neoError)
	c := restclient.RestRequest{
		Url:    m.db.info.RelTypes,
		Method: restclient.GET,
		Result: &reltypes,
		Error:  &ne,
	}
	status, err := m.db.rc.Do(&c)
	if err != nil {
		logPretty(ne)
		return reltypes, err
	}
	if status == 200 {
		sort.Sort(sort.StringSlice(reltypes))
		return reltypes, nil // Success!
	}
	logPretty(ne)
	return reltypes, BadResponse
}

// A Relationship is a directional connection between two Nodes, with an
// optional set of arbitrary properties.
type Relationship struct {
	entity
	HrefStart string
	HrefType  string
	HrefEnd   string
}

// populate uses the values from a relationshipResponse object to populate the
// fields on this Relationship.
func (r *Relationship) populate(res *relationshipResponse) {
	r.HrefProperty = res.HrefProperty
	r.HrefProperties = res.HrefProperties
	r.HrefSelf = res.HrefSelf
	r.HrefStart = res.HrefStart
	r.HrefType = res.HrefType
	r.HrefEnd = res.HrefEnd
}

// A relationshipResponse represents data returned by the Neo4j server on a
// Relationship operation.
type relationshipResponse struct {
	HrefProperty   string `json:"property"`
	HrefProperties string `json:"properties"`
	HrefSelf       string `json:"self"`
	// HrefData       interface{} `json:"data"`
	// HrefExtensions interface{} `json:"extensions"`
	//
	HrefStart string `json:"start"`
	HrefType  string `json:"type"`
	HrefEnd   string `json:"end"`
}

// Id gets the ID number of this Relationship
func (r *Relationship) Id() int {
	parts := strings.Split(r.HrefSelf, "/")
	s := parts[len(parts)-1]
	id, err := strconv.Atoi(s)
	if err != nil {
		// Are both r.Info and r.Node valid?
		panic(err)
	}
	return id
}

// Start gets the starting Node of this Relationship.
func (r *Relationship) Start() (*Node, error) {
	// log.Println("INFO", r.Info)
	return r.db.Nodes.getNodeByUri(r.HrefStart)
}

// End gets the ending Node of this Relationship.
func (r *Relationship) End() (*Node, error) {
	return r.db.Nodes.getNodeByUri(r.HrefEnd)
}

// Type gets the type of this relationship
func (r *Relationship) Type() string {
	return r.HrefType
}
