package main

type APIDoc struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	APIGroups   []*APIGroup `json:"api_groups"`
}
